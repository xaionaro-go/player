package builtin

import (
	"context"
	"fmt"
	"image"
	"io"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/facebookincubator/go-belt/tool/logger"
	"github.com/xaionaro-go/audio/pkg/audio"
	"github.com/xaionaro-go/audio/pkg/audio/planar"
	"github.com/xaionaro-go/avpipeline"
	"github.com/xaionaro-go/avpipeline/codec"
	"github.com/xaionaro-go/avpipeline/frame"
	"github.com/xaionaro-go/avpipeline/kernel"
	"github.com/xaionaro-go/avpipeline/processor"
	"github.com/xaionaro-go/observability"
	"github.com/xaionaro-go/player/pkg/player/types"
	"github.com/xaionaro-go/secret"
	"github.com/xaionaro-go/xcontext"
	"github.com/xaionaro-go/xsync"
)

const (
	BufferSizeAudio = 100 * time.Millisecond
)

type Player struct {
	ImageRenderer
	AudioRenderer
	lastSeekAt            time.Time
	audioWriter           io.WriteCloser
	audioStream           audio.Stream
	locker                xsync.Gorex
	currentURL            string
	currentImage          image.Image
	currentDuration       time.Duration
	previousVideoPosition time.Duration
	currentAudioPosition  time.Duration
	startOffset           *time.Duration
	videoStreamIndex      atomic.Uint32
	audioStreamIndex      atomic.Uint32
	endChan               chan struct{}
	cancelFunc            context.CancelFunc
}

var _ types.Player = (*Player)(nil)

func New(
	ctx context.Context,
	imageRenderer ImageRenderer,
	audioRenderer AudioRenderer,
) *Player {
	p := &Player{
		ImageRenderer: imageRenderer,
		AudioRenderer: audioRenderer,
		endChan:       make(chan struct{}),
	}
	p.onEnd()
	return p
}

func (*Player) SetupForStreaming(
	ctx context.Context,
) error {
	return nil
}

func (p *Player) OpenURL(
	ctx context.Context,
	link string,
) (_err error) {
	logger.Debugf(ctx, "OpenURL(ctx, '%s')", link)
	defer func() { logger.Debugf(ctx, "/OpenURL(ctx, '%s'): %v", link, _err) }()
	return xsync.DoA2R1(ctx, &p.locker, p.openURL, ctx, link)
}

func (p *Player) openURL(
	ctx context.Context,
	link string,
) error {
	if p.cancelFunc != nil {
		return fmt.Errorf("player is already running; changing URLs is not implemented, yet")
	}
	ctx = xcontext.DetachDone(ctx)
	ctx, cancelFn := context.WithCancel(ctx)
	p.cancelFunc = cancelFn
	var once sync.Once
	stopFn := func() {
		once.Do(func() {
			p.locker.Do(ctx, func() {
				p.cancelFunc()
				p.cancelFunc = nil
			})
		})
	}

	inputCfg := kernel.InputConfig{}
	input, err := kernel.NewInputFromURL(ctx, link, secret.New(""), inputCfg)
	logger.Tracef(ctx, "NewInputFromURL(ctx, '%s', '', %#+v): %v", link, inputCfg, err)
	if err != nil {
		return fmt.Errorf("unable to open '%s': %w", link, err)
	}

	inputNode := avpipeline.NewNodeFromKernel(
		ctx,
		input,
		processor.DefaultOptionsInput()...,
	)
	decoderNode := avpipeline.NewNodeFromKernel(
		ctx,
		kernel.NewDecoder(ctx, codec.NewNaiveDecoderFactory(ctx, 0, "", nil, nil)),
		processor.DefaultOptionsRecoder()...,
	)
	playerNode := avpipeline.NewNodeFromKernel(
		ctx,
		p,
		processor.DefaultOptionsRecoder()...,
	)
	inputNode.PushPacketsTo.Add(decoderNode)
	decoderNode.PushFramesTo.Add(playerNode)

	p.onSeek(ctx)

	errCh := make(chan avpipeline.ErrNode, 1)
	observability.Go(ctx, func() {
		defer stopFn()
		select {
		case <-ctx.Done():
			return
		case err := <-errCh:
			if err.Err != nil {
				logger.Errorf(ctx, "received error: %v", err)
			} else {
				logger.Debugf(ctx, "received error: %v", err)
			}
		}
	})
	observability.Go(ctx, func() {
		defer close(errCh)
		defer p.onEnd()
		avpipeline.ServeRecursively(ctx, avpipeline.ServeConfig{}, errCh, inputNode)
	})

	p.currentURL = link
	if p.ImageRenderer != nil {
		if err := p.ImageRenderer.SetVisible(true); err != nil {
			return fmt.Errorf("unable to make the image renderer visible: %w", err)
		}
	}
	return nil
}

func (p *Player) processFrame(
	ctx context.Context,
	frame frame.Input,
) error {
	logger.Tracef(ctx, "processFrame: pos: %v; dur: %v; pts: %v; time_base: %v", frame.GetPTSAsDuration(), frame.GetStreamDurationAsDuration(), frame.Pts(), frame.GetTimeBase())
	defer func() {
		logger.Tracef(ctx, "/processFrame; av-desync: %v", p.currentAudioPosition-p.previousVideoPosition)
	}()
	return xsync.DoR1(ctx, &p.locker, func() error {
		switch frame.GetMediaType() {
		case MediaTypeVideo:
			return p.processVideoFrame(ctx, frame)
		case MediaTypeAudio:
			return p.processAudioFrame(ctx, frame)
		default:
			// we don't care about everything else
			return nil
		}
	})
}

func (p *Player) onSeek(
	ctx context.Context,
) {
	logger.Tracef(ctx, "onSeek")
	defer logger.Tracef(ctx, "/onSeek")

	p.lastSeekAt = time.Now()
}

func (p *Player) processVideoFrame(
	ctx context.Context,
	frame frame.Input,
) error {
	logger.Tracef(ctx, "processVideoFrame")
	defer logger.Tracef(ctx, "/processVideoFrame")
	if p.ImageRenderer == nil {
		return nil
	}

	p.currentDuration = frame.GetStreamDurationAsDuration()
	streamIdx := frame.GetStreamIndex()

	if p.videoStreamIndex.CompareAndSwap(math.MaxUint32, uint32(streamIdx)) { // atomics are not really needed because all of this happens while holding p.locker
		if err := p.initImageFor(ctx, frame); err != nil {
			return fmt.Errorf("unable to initialize an image variable for the frame: %w", err)
		}
	} else {
		oldStreamIdx := int(p.videoStreamIndex.Load())
		if oldStreamIdx != streamIdx {
			return fmt.Errorf("the index of the video stream have changed from %d to %d; the support of dynamic/multiple video tracks is not implemented, yet", oldStreamIdx, streamIdx)
		}
	}

	frame.Data().ToImage(p.currentImage)

	sinceStart := time.Since(p.lastSeekAt)
	currentExpectedPosition := p.previousVideoPosition + frame.GetDurationAsDuration()
	if p.startOffset == nil {
		p.startOffset = ptr(currentExpectedPosition - sinceStart)
		logger.Tracef(ctx, "set startOffset to: %v, which is %v - %v", *p.startOffset, currentExpectedPosition, sinceStart)
	}
	curPosition := sinceStart + *p.startOffset
	waitIntervalForNextFrame := currentExpectedPosition - curPosition
	if abs(waitIntervalForNextFrame) > time.Minute {
		p.startOffset = ptr(currentExpectedPosition - sinceStart)
		logger.Tracef(ctx, "update startOffset to: %v, which is %v - %v", *p.startOffset, currentExpectedPosition, sinceStart)
		waitIntervalForNextFrame = 0
	}
	p.previousVideoPosition = frame.GetPTSAsDuration()

	logger.Tracef(ctx, "sleeping for %v (%v - (%v + %v))", waitIntervalForNextFrame, currentExpectedPosition, sinceStart, *p.startOffset)
	time.Sleep(waitIntervalForNextFrame)

	if err := p.renderCurrentPicture(); err != nil {
		return fmt.Errorf("unable to render the picture: %w", err)
	}

	return nil
}

func (p *Player) renderCurrentPicture() error {
	return p.ImageRenderer.Render()
}

func (p *Player) processAudioFrame(
	ctx context.Context,
	frame frame.Input,
) error {
	logger.Tracef(ctx, "processAudioFrame")
	defer logger.Tracef(ctx, "/processAudioFrame")
	if p.AudioRenderer == nil {
		return nil
	}

	p.currentAudioPosition = frame.GetPTSAsDuration()
	streamIdx := frame.GetStreamIndex()

	if p.audioStreamIndex.CompareAndSwap(math.MaxUint32, uint32(streamIdx)) { // atomics are not really needed because all of this happens while holding p.locker
		var r io.Reader
		{
			pr, pw := io.Pipe()
			p.audioWriter = pw
			r = pr
		}
		bufSize, err := frame.SamplesBufferSize(1)
		if err != nil {
			return fmt.Errorf("unable to get the buffer size: %w", err)
		}

		codecCtx := frame.CodecContext
		sampleRate := codecCtx.SampleRate()
		channels := codecCtx.ChannelLayout().Channels()
		pcmFormatAV := codecCtx.SampleFormat()
		codecID := codecCtx.CodecID()
		logger.Debugf(ctx, "codecID == %v, sampleRate == %v, channels == %v, pcmFormat == %v", codecID, sampleRate, channels, pcmFormatAV)
		bufferSize := BufferSizeAudio
		pcmFormat := pcmFormatToAudio(pcmFormatAV)
		if isPlanar(pcmFormatAV) {
			r = planar.NewUnplanarizeReader(r, audio.Channel(channels), uint(pcmFormat.Size()), uint(bufSize))
		}
		audioStream, err := p.AudioRenderer.PlayPCM(
			ctx,
			audio.SampleRate(sampleRate),
			audio.Channel(channels),
			pcmFormat,
			bufferSize,
			r,
		)
		if err != nil {
			return fmt.Errorf("unable to initialize an audio playback: %w", err)
		}
		p.audioStream = audioStream
	} else {
		oldStreamIdx := int(p.audioStreamIndex.Load())
		if oldStreamIdx != streamIdx {
			logger.Tracef(ctx, "we do not support multiple audio streams, yet; so we ignore this new stream, index: %d (which is not %d)", streamIdx, oldStreamIdx)
			return nil
		}
	}

	align := 1
	frameBytes, err := frame.Data().Bytes(int(align))
	if err != nil {
		return fmt.Errorf("unable to get the audio frame data: %w", err)
	}

	n, err := p.audioWriter.Write(frameBytes)
	if err != nil {
		return fmt.Errorf("unable to write the audio frame into the playback subsystem: %w", err)
	}
	if n != len(frameBytes) {
		return fmt.Errorf("unable to write the full audio frame: %d != %d", n, len(frameBytes))
	}

	return nil
}

func (p *Player) onEnd() {
	ctx := context.TODO()
	logger.Debugf(ctx, "onEnd")
	defer logger.Debugf(ctx, "/onEnd")
	p.locker.Do(ctx, func() {
		p.videoStreamIndex.Store(math.MaxUint32)
		p.audioStreamIndex.Store(math.MaxUint32)
		p.currentURL = ""
		if p.audioWriter != nil {
			p.audioWriter.Close()
			p.audioWriter = nil
		}
		if p.audioStream != nil {
			p.audioStream.Close()
			p.audioStream = nil
		}

		var oldEndChan chan struct{}
		p.endChan, oldEndChan = make(chan struct{}), p.endChan
		close(oldEndChan)
		if p.ImageRenderer != nil {
			if err := p.ImageRenderer.SetVisible(false); err != nil {
				logger.Errorf(ctx, "unable to close ImageRenderer: %v", err)
			}
		}
	})
}

func (p *Player) EndChan(
	ctx context.Context,
) (<-chan struct{}, error) {
	return p.endChan, nil
}

func (p *Player) IsEnded(
	ctx context.Context,
) (bool, error) {
	return xsync.DoR1(ctx, &p.locker, p.isEnded), nil
}

func (p *Player) isEnded() bool {
	return p.currentURL == ""
}

func (p *Player) GetPosition(
	ctx context.Context,
) (_ret time.Duration, _err error) {
	logger.Tracef(ctx, "GetPosition")
	defer func() { logger.Tracef(ctx, "/GetPosition: %v %v", _ret, _err) }()
	return xsync.DoR2(ctx, &p.locker, func() (time.Duration, error) {
		if p.isEnded() {
			return 0, fmt.Errorf("the player is not started or already ended")
		}
		switch {
		case p.previousVideoPosition != 0 && p.currentAudioPosition != 0:
			return (p.previousVideoPosition + p.currentAudioPosition) / 2, nil
		case p.currentAudioPosition != 0:
			return p.currentAudioPosition, nil
		case p.previousVideoPosition != 0:
			return p.previousVideoPosition, nil
		}
		return 0, nil
	})
}

func (p *Player) GetLength(
	ctx context.Context,
) (_ret time.Duration, _err error) {
	logger.Tracef(ctx, "GetLength")
	defer func() { logger.Tracef(ctx, "/GetLength: %v %v", _ret, _err) }()
	return xsync.DoR2(ctx, &p.locker, func() (time.Duration, error) {
		if p.isEnded() {
			return 0, fmt.Errorf("the player is not started or already ended")
		}

		return p.currentDuration, nil
	})
}

func (p *Player) ProcessTitle(
	ctx context.Context,
) (string, error) {
	if titler, ok := p.ImageRenderer.(interface{ Title() string }); ok {
		return titler.Title(), nil
	}
	return "", nil
}

func (p *Player) GetLink(
	ctx context.Context,
) (string, error) {
	return xsync.DoR2(ctx, &p.locker, func() (string, error) {
		if p.isEnded() {
			return "", fmt.Errorf("the player is not started or already ended")
		}

		return p.currentURL, nil
	})
}

func (*Player) GetSpeed(
	ctx context.Context,
) (float64, error) {
	logger.Errorf(ctx, "GetSpeed is not implemented, yet")
	return 1, nil
}

func (*Player) SetSpeed(
	ctx context.Context,
	speed float64,
) error {
	logger.Errorf(ctx, "SetSpeed is not implemented, yet")
	return nil
}

func (*Player) GetPause(
	ctx context.Context,
) (bool, error) {
	return false, nil
}

func (*Player) SetPause(
	ctx context.Context,
	pause bool,
) error {
	logger.Errorf(ctx, "SetPause is not implemented, yet")
	return nil
}

func (*Player) Seek(
	ctx context.Context,
	pos time.Duration,
	isRelative bool,
	quick bool,
) error {
	return fmt.Errorf("not implemented, yet")
}

func (*Player) GetVideoTracks(
	ctx context.Context,
) (types.VideoTracks, error) {
	return nil, fmt.Errorf("not implemented, yet")
}

func (*Player) GetAudioTracks(
	ctx context.Context,
) (types.AudioTracks, error) {
	return nil, fmt.Errorf("not implemented, yet")
}

func (*Player) GetSubtitlesTracks(
	ctx context.Context,
) (types.SubtitlesTracks, error) {
	return nil, fmt.Errorf("not implemented, yet")
}

func (*Player) SetVideoTrack(
	ctx context.Context,
	vid int64,
) error {
	return fmt.Errorf("not implemented, yet")
}

func (*Player) SetAudioTrack(
	ctx context.Context,
	aid int64,
) error {
	return fmt.Errorf("not implemented, yet")
}

func (*Player) SetSubtitlesTrack(
	ctx context.Context,
	sid int64,
) error {
	return fmt.Errorf("not implemented, yet")
}

func (*Player) Stop(
	ctx context.Context,
) error {
	panic("not implemented, yet")
}
