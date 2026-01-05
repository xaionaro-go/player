package libav

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
	"github.com/xaionaro-go/avpipeline/node"
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

type Decoder struct {
	ImageRenderer
	AudioRenderer
	lastSeekAt            time.Time
	audioWriter           io.WriteCloser
	audioStream           audio.Stream
	locker                xsync.Gorex
	currentURL            string
	currentImage          image.Image
	previousVideoPosition time.Duration
	currentAudioPosition  atomic.Uint64
	videoStreamIndex      atomic.Uint32
	audioStreamIndex      atomic.Uint32
	closedChan            chan struct{}
	endChan               chan struct{}
	cancelFunc            context.CancelFunc
	videoFramesQueue      chan frame.Input
}

var _ types.Player = (*Decoder)(nil)

func New(
	ctx context.Context,
	imageRenderer ImageRenderer,
	audioRenderer AudioRenderer,
) *Decoder {
	p := &Decoder{
		ImageRenderer:    imageRenderer,
		AudioRenderer:    audioRenderer,
		closedChan:       make(chan struct{}),
		endChan:          make(chan struct{}),
		videoFramesQueue: make(chan frame.Input, 100),
	}
	p.init(ctx)
	p.onEnd()
	return p
}

func (*Decoder) SetupForStreaming(
	ctx context.Context,
) error {
	return nil
}

func (p *Decoder) init(ctx context.Context) {
	observability.Go(ctx, func(ctx context.Context) {
		p.videoRenderLoop(ctx)
	})
}

func (p *Decoder) videoRenderLoop(
	ctx context.Context,
) {
	logger.Debugf(ctx, "videoRenderLoop")
	defer logger.Debugf(ctx, "/videoRenderLoop")
	for {
		var f frame.Input
		select {
		case <-ctx.Done():
			logger.Debugf(ctx, "videoRenderLoop: context done")
			return
		case f = <-p.videoFramesQueue:
		}

		currentExpectedPosition := p.getCurrentAudioPosition() - BufferSizeAudio
		curPosition := f.GetPTSAsDuration()
		waitIntervalForNextFrame := curPosition - currentExpectedPosition
		p.previousVideoPosition = curPosition

		logger.Tracef(ctx, "sleeping for %v (%v - %v)", waitIntervalForNextFrame, curPosition, currentExpectedPosition)
		if waitIntervalForNextFrame > 0 {
			time.Sleep(waitIntervalForNextFrame)
		}

		switch r := p.ImageRenderer.(type) {
		case AVFrameRenderer:
			if err := r.SetAVFrame(ctx, ImageUnparsed{
				Decoder: p,
				Input:   f,
			}); err != nil {
				logger.Errorf(ctx, "unable to set the AV frame: %v", err)
				continue
			}
		case ImageRenderer:
			err := f.Data().ToImage(p.currentImage)
			if err != nil {
				logger.Errorf(ctx, "unable to convert the frame into an image: %v", err)
				continue
			}
			if _, ok := p.ImageRenderer.(RenderImageNower); !ok {
				err := r.SetImage(ctx, ImageGeneric{
					Decoder: p,
					Input:   f,
					Image:   p.currentImage,
				})
				if err != nil {
					logger.Errorf(ctx, "unable to set the image: %v", err)
					continue
				}
			}
		default:
			logger.Errorf(ctx, "an image renderer of an unexpected type %T", r)
			continue
		}

		if err := p.renderCurrentPicture(ctx, f); err != nil {
			logger.Errorf(ctx, "unable to render the picture: %v", err)
			continue
		}
	}
}

func (p *Decoder) OpenURL(
	ctx context.Context,
	link string,
) (_err error) {
	logger.Debugf(ctx, "OpenURL(ctx, '%s')", link)
	defer func() { logger.Debugf(ctx, "/OpenURL(ctx, '%s'): %v", link, _err) }()
	return xsync.DoA2R1(ctx, &p.locker, p.openURL, ctx, link)
}

func (p *Decoder) openURL(
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

	inputNode := node.NewFromKernel(
		ctx,
		input,
		processor.DefaultOptionsInput()...,
	)
	decoderNode := node.NewFromKernel(
		ctx,
		kernel.NewDecoder(ctx, codec.NewNaiveDecoderFactory(ctx, nil)),
		processor.DefaultOptionsTranscoder()...,
	)
	playerNode := node.NewFromKernel(
		ctx,
		p,
		processor.DefaultOptionsTranscoder()...,
	)
	inputNode.AddPushTo(ctx, decoderNode)
	decoderNode.AddPushTo(ctx, playerNode)

	p.onSeek(ctx)

	errCh := make(chan node.Error, 1)
	observability.Go(ctx, func(ctx context.Context) {
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
	observability.Go(ctx, func(ctx context.Context) {
		defer close(errCh)
		defer p.onEnd()
		avpipeline.Serve(ctx, avpipeline.ServeConfig{}, errCh, inputNode)
	})

	p.currentURL = link
	if p.ImageRenderer != nil {
		if v, ok := p.ImageRenderer.(SetVisibler); ok {
			if err := v.SetVisible(true); err != nil {
				return fmt.Errorf("unable to make the image renderer visible: %w", err)
			}
		}
	}

	select {
	case <-p.closedChan:
		p.closedChan = make(chan struct{})
	default:
		panic("is supposed to be impossible")
	}
	return nil
}

func (p *Decoder) getCurrentAudioPosition() time.Duration {
	return time.Duration(p.currentAudioPosition.Load())
}

func (p *Decoder) processFrame(
	ctx context.Context,
	frame frame.Input,
) error {
	logger.Tracef(ctx, "processFrame: pos: %v; pts: %v; time_base: %v", frame.GetPTSAsDuration(), frame.Pts(), frame.GetTimeBase())
	defer func() {
		logger.Tracef(ctx, "/processFrame; av-desync: %v", p.getCurrentAudioPosition()-BufferSizeAudio-p.previousVideoPosition)
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

func (p *Decoder) onSeek(
	ctx context.Context,
) {
	logger.Tracef(ctx, "onSeek")
	defer logger.Tracef(ctx, "/onSeek")

	p.lastSeekAt = time.Now()
}

func (p *Decoder) processVideoFrame(
	ctx context.Context,
	f frame.Input,
) error {
	logger.Tracef(ctx, "processVideoFrame")
	defer logger.Tracef(ctx, "/processVideoFrame")
	if p.ImageRenderer == nil {
		return nil
	}

	streamIdx := f.GetStreamIndex()

	if p.videoStreamIndex.CompareAndSwap(math.MaxUint32, uint32(streamIdx)) { // atomics are not really needed because all of this happens while holding p.locker
		if err := p.initImageFor(ctx, f); err != nil {
			return fmt.Errorf("unable to initialize an image variable for the frame: %w", err)
		}
	} else {
		oldStreamIdx := int(p.videoStreamIndex.Load())
		if oldStreamIdx != streamIdx {
			return fmt.Errorf("the index of the video stream have changed from %d to %d; the support of dynamic/multiple video tracks is not implemented, yet", oldStreamIdx, streamIdx)
		}
	}

	p.videoFramesQueue <- f
	return nil
}

func (p *Decoder) renderCurrentPicture(
	ctx context.Context,
	_ frame.Input,
) (_err error) {
	logger.Tracef(ctx, "renderCurrentPicture")
	defer func() { logger.Tracef(ctx, "/renderCurrentPicture: %v", _err) }()
	if r, ok := p.ImageRenderer.(RenderImageNower); ok {
		return r.RenderImageNow(ctx)
	}
	return nil
}

func (p *Decoder) processAudioFrame(
	ctx context.Context,
	frame frame.Input,
) (_err error) {
	logger.Tracef(ctx, "processAudioFrame")
	defer func() { logger.Tracef(ctx, "/processAudioFrame: %v", _err) }()
	if p.AudioRenderer == nil {
		return nil
	}

	p.currentAudioPosition.Store(uint64(frame.GetPTSAsDuration()))
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

		codecParams := frame.CodecParameters
		sampleRate := codecParams.SampleRate()
		channels := codecParams.ChannelLayout().Channels()
		pcmFormatAV := codecParams.SampleFormat()
		codecID := codecParams.CodecID()
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

func (p *Decoder) onEnd() {
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
		select {
		case <-p.closedChan:
		default:
			close(p.closedChan)
		}
		if p.ImageRenderer != nil {
			if v, ok := p.ImageRenderer.(SetVisibler); ok {
				if err := v.SetVisible(false); err != nil {
					logger.Errorf(ctx, "unable to hide the image renderer: %v", err)
				}
			}
		}
	})
}

func (p *Decoder) EndChan(
	ctx context.Context,
) (<-chan struct{}, error) {
	return p.endChan, nil
}

func (p *Decoder) IsEnded(
	ctx context.Context,
) (bool, error) {
	return xsync.DoR1(ctx, &p.locker, p.isEnded), nil
}

func (p *Decoder) isEnded() bool {
	return p.currentURL == ""
}

func (p *Decoder) GetPosition(
	ctx context.Context,
) (_ret time.Duration, _err error) {
	logger.Tracef(ctx, "GetPosition")
	defer func() { logger.Tracef(ctx, "/GetPosition: %v %v", _ret, _err) }()
	return xsync.DoR2(ctx, &p.locker, func() (time.Duration, error) {
		if p.isEnded() {
			return 0, fmt.Errorf("the player is not started or already ended")
		}
		currentAudioPosition := p.getCurrentAudioPosition()
		switch {
		case p.previousVideoPosition != 0 && currentAudioPosition != 0:
			return (p.previousVideoPosition + currentAudioPosition) / 2, nil
		case currentAudioPosition != 0:
			return currentAudioPosition, nil
		case p.previousVideoPosition != 0:
			return p.previousVideoPosition, nil
		}
		return 0, nil
	})
}

func (p *Decoder) GetAudioPosition(
	ctx context.Context,
) (_ret time.Duration, _err error) {
	logger.Tracef(ctx, "GetAudioPosition")
	defer func() { logger.Tracef(ctx, "/GetAudioPosition: %v %v", _ret, _err) }()
	return p.GetPosition(ctx)
}

func (p *Decoder) GetLength(
	ctx context.Context,
) (_ret time.Duration, _err error) {
	logger.Tracef(ctx, "GetLength")
	defer func() { logger.Tracef(ctx, "/GetLength: %v %v", _ret, _err) }()
	return xsync.DoR2(ctx, &p.locker, func() (time.Duration, error) {
		if p.isEnded() {
			return 0, fmt.Errorf("the player is not started or already ended")
		}

		return 0, fmt.Errorf("not implemented, yet")
	})
}

func (p *Decoder) ProcessTitle(
	ctx context.Context,
) (string, error) {
	if titler, ok := p.ImageRenderer.(interface{ Title() string }); ok {
		return titler.Title(), nil
	}
	return "", nil
}

func (p *Decoder) GetLink(
	ctx context.Context,
) (string, error) {
	return xsync.DoR2(ctx, &p.locker, func() (string, error) {
		if p.isEnded() {
			return "", fmt.Errorf("the player is not started or already ended")
		}

		return p.currentURL, nil
	})
}

func (*Decoder) GetSpeed(
	ctx context.Context,
) (float64, error) {
	return 1, nil
}

func (*Decoder) SetSpeed(
	ctx context.Context,
	speed float64,
) error {
	logger.Errorf(ctx, "SetSpeed is not implemented, yet")
	return nil
}

func (*Decoder) GetPause(
	ctx context.Context,
) (bool, error) {
	return false, nil
}

func (*Decoder) SetPause(
	ctx context.Context,
	pause bool,
) error {
	logger.Errorf(ctx, "SetPause is not implemented, yet")
	return nil
}

func (*Decoder) Seek(
	ctx context.Context,
	pos time.Duration,
	isRelative bool,
	quick bool,
) error {
	return fmt.Errorf("not implemented, yet")
}

func (*Decoder) GetVideoTracks(
	ctx context.Context,
) (types.VideoTracks, error) {
	return nil, fmt.Errorf("not implemented, yet")
}

func (*Decoder) GetAudioTracks(
	ctx context.Context,
) (types.AudioTracks, error) {
	return nil, fmt.Errorf("not implemented, yet")
}

func (*Decoder) GetSubtitlesTracks(
	ctx context.Context,
) (types.SubtitlesTracks, error) {
	return nil, fmt.Errorf("not implemented, yet")
}

func (*Decoder) SetVideoTrack(
	ctx context.Context,
	vid int64,
) error {
	return fmt.Errorf("not implemented, yet")
}

func (*Decoder) SetAudioTrack(
	ctx context.Context,
	aid int64,
) error {
	return fmt.Errorf("not implemented, yet")
}

func (*Decoder) SetSubtitlesTrack(
	ctx context.Context,
	sid int64,
) error {
	return fmt.Errorf("not implemented, yet")
}

func (*Decoder) Stop(
	ctx context.Context,
) error {
	panic("not implemented, yet")
}
