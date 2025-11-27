package gstreamer

import (
	"context"
	"errors"
	"fmt"
	"image"
	"net/url"
	"path/filepath"
	"time"

	"github.com/facebookincubator/go-belt"
	"github.com/facebookincubator/go-belt/tool/logger"
	"github.com/go-gst/go-gst/gst"
	"github.com/go-gst/go-gst/gst/app"
	"github.com/xaionaro-go/player/pkg/player/imagerenderer"
	"github.com/xaionaro-go/player/pkg/player/types"
)

func init() {
	gst.Init(nil)
}

type Decoder struct {
	Pipeline      *gst.Pipeline
	Playbin       *gst.Element
	AppSink       *app.Sink
	CurrentFrame  *image.RGBA
	AudioRenderer AudioRenderer
	ImageRenderer ImageRenderer
	observability *belt.Belt
}

var _ types.Player = (*Decoder)(nil)

// this implementation is heavily inspired by https://github.com/realskyquest/ebiten-gstreamer

func New(
	ctx context.Context,
	imageRenderer ImageRenderer,
	audioRenderer AudioRenderer,
) (_ret *Decoder, _err error) {
	d := &Decoder{
		AudioRenderer: audioRenderer,
		ImageRenderer: imageRenderer,
		CurrentFrame:  image.NewRGBA(image.Rectangle{}),
		observability: belt.CtxBelt(ctx),
	}
	logger.Errorf(ctx, "the support for audio renderer in gstreamer decoder is not implemented yet")

	if err := d.ImageRenderer.SetImage(ctx, FrameVideo{d.CurrentFrame}); err != nil {
		return nil, fmt.Errorf("unable to set initial image to the image renderer: %w", err)
	}

	defer func() {
		if _err != nil {
			d.Close(ctx)
		}
	}()

	// see https://gstreamer.freedesktop.org/documentation/app/appsink.html
	appSinkElement, err := gst.NewElement("appsink")
	if err != nil {
		return nil, fmt.Errorf("unable to create appsink element: %w", err)
	}
	appSinkElement.Set("emit-signals", true)
	appSinkElement.Set("max-buffers", 1)
	appSinkElement.Set("drop", true)
	appSink := app.SinkFromElement(appSinkElement)
	appSink.SetCaps(gst.NewCapsFromString("video/x-raw,format=RGBA"))
	appSink.SetCallbacks(&app.SinkCallbacks{
		NewSampleFunc: d.onNewSampleFunc,
	})
	d.AppSink = appSink

	// see https://gstreamer.freedesktop.org/documentation/playback/playbin.html
	playbin, err := gst.NewElement("playbin")
	if err != nil {
		return nil, fmt.Errorf("unable to create playbin element: %w", err)
	}
	playbin.Set("video-sink", appSinkElement)
	d.Playbin = playbin

	pipeline, err := gst.NewPipeline("")
	if err != nil {
		return nil, fmt.Errorf("unable to create pipeline: %w", err)
	}
	pipeline.Add(playbin)
	d.Pipeline = pipeline

	return d, nil
}

func (d *Decoder) logger() logger.Logger {
	return logger.FromBelt(d.observability)
}

func (d *Decoder) onNewSampleFunc(sink *app.Sink) (_ret gst.FlowReturn) {
	ctx := belt.CtxWithBelt(context.Background(), d.observability)
	logger.Tracef(ctx, "onNewSampleFunc called")
	defer func() { logger.Tracef(ctx, "/onNewSampleFunc: %s", _ret) }()

	sample := sink.PullSample()
	if sample == nil {
		return gst.FlowEOS
	}

	buffer := sample.GetBuffer()
	if buffer == nil {
		d.logger().Errorf("no buffer in sample")
		return gst.FlowError
	}

	caps := sample.GetCaps()
	structure := caps.GetStructureAt(0)

	width, err := structure.GetValue("width")
	if err != nil {
		d.logger().Errorf("unable to get width from structure: %v", err)
		return gst.FlowError
	}

	height, err := structure.GetValue("height")
	if err != nil {
		d.logger().Errorf("unable to get height from structure: %v", err)
		return gst.FlowError
	}

	w, ok := width.(int)
	if !ok {
		d.logger().Errorf("width is not an int: %T", width)
		return gst.FlowError
	}

	h, ok := height.(int)
	if !ok {
		d.logger().Errorf("height is not an int: %T", height)
		return gst.FlowError
	}

	bufmap := buffer.Map(gst.MapRead)
	defer buffer.Unmap()

	data := bufmap.Bytes()
	if d.CurrentFrame.Rect.Dx() != w || d.CurrentFrame.Rect.Dy() != h {
		*d.CurrentFrame = *image.NewRGBA(image.Rect(0, 0, w, h))
	}
	d.CurrentFrame.Pix = data

	if renderImageNower, ok := d.ImageRenderer.(imagerenderer.RenderImageNower); ok {
		if err := renderImageNower.RenderImageNow(ctx); err != nil {
			d.logger().Errorf("unable to render image now: %v", err)
			return gst.FlowError
		}
	} else {
		if err := d.ImageRenderer.SetImage(ctx, FrameVideo{d.CurrentFrame}); err != nil {
			d.logger().Errorf("unable to set image to the image renderer: %v", err)
			return gst.FlowError
		}
	}

	return gst.FlowOK
}

func (d *Decoder) ProcessTitle(ctx context.Context) (string, error) {
	return "", fmt.Errorf("not implemented")
}

func toURI(link string) (string, error) {
	u, err := url.Parse(link)
	if err == nil && u.Scheme != "" {
		return link, nil // already a URI (http/https/rtsp/rtmp/etc.)
	}
	if filepath.IsAbs(link) {
		// turn /abs/path into file:///abs/path (with proper escaping)
		return (&url.URL{Scheme: "file", Path: filepath.ToSlash(link)}).String(), nil
	}
	return "", fmt.Errorf("not an absolute URI or absolute path: %q", link)
}

func (d *Decoder) OpenURL(ctx context.Context, link string) (_err error) {
	logger.Debugf(ctx, "OpenURL(%q)", link)
	defer logger.Debugf(ctx, "/OpenURL(%q)", link)

	uri, err := toURI(link)
	if err != nil {
		return fmt.Errorf("unable to convert link to URI: %w", err)
	}

	logger.Debugf(ctx, "current playbin state: %s", d.Playbin.GetCurrentState())

	if err := d.Playbin.SetState(gst.StateNull); err != nil {
		return fmt.Errorf("unable to set playbin to NULL state: %w", err)
	}

	if err := d.Playbin.Set("uri", uri); err != nil {
		return fmt.Errorf("unable to set URI to playbin: %w", err)
	}

	if err := d.Playbin.SetState(gst.StatePlaying); err != nil {
		return fmt.Errorf("unable to set playbin to PLAYING state: %w", err)
	}

	return nil
}

func (d *Decoder) GetLink(ctx context.Context) (_ret string, _err error) {
	uri, err := d.Playbin.GetProperty("uri")
	if err != nil {
		return "", fmt.Errorf("unable to get URI from playbin: %w", err)
	}
	uriString, ok := uri.(string)
	if !ok {
		return "", fmt.Errorf("URI from playbin is not a string: %T", uri)
	}
	return uriString, nil
}

func (d *Decoder) EndChan(ctx context.Context) (<-chan struct{}, error) {
	return nil, fmt.Errorf("not implemented")
}

func (d *Decoder) IsEnded(ctx context.Context) (bool, error) {
	switch d.Pipeline.GetCurrentState() {
	case gst.StateReady, gst.StatePaused, gst.StatePlaying:
		return false, nil
	case gst.StateNull:
		return true, nil
	default:
		return false, fmt.Errorf("unknown pipeline state: %s", d.Pipeline.GetCurrentState().String())
	}
}

func (d *Decoder) GetPosition(ctx context.Context) (time.Duration, error) {
	ok, pos := d.Pipeline.QueryPosition(gst.FormatTime)
	if !ok {
		return 0, fmt.Errorf("unable to query position")
	}
	return time.Duration(pos) * time.Nanosecond, nil
}

func (d *Decoder) GetAudioPosition(ctx context.Context) (time.Duration, error) {
	return d.GetPosition(ctx)
}

func (d *Decoder) GetLength(ctx context.Context) (time.Duration, error) {
	ok, dur := d.Pipeline.QueryDuration(gst.FormatTime)
	if !ok {
		return 0, fmt.Errorf("unable to query duration")
	}
	return time.Duration(dur) * time.Nanosecond, nil
}

func (d *Decoder) GetSpeed(ctx context.Context) (float64, error) {
	return 1.0, nil
}

func (d *Decoder) SetSpeed(ctx context.Context, speed float64) error {
	return fmt.Errorf("not implemented")
}

func (d *Decoder) GetPause(ctx context.Context) (bool, error) {
	s := d.Pipeline.GetCurrentState()
	logger.Debugf(ctx, "GetPause: current state is %s", s.String())
	switch s {
	case gst.StatePlaying:
		return false, nil
	case gst.StatePaused:
		return true, nil
	default:
		return false, fmt.Errorf("invalid pipeline state: %s", d.Pipeline.GetCurrentState().String())
	}
}

func (d *Decoder) SetPause(ctx context.Context, pause bool) error {
	if pause {
		err := d.Pipeline.SetState(gst.StatePaused)
		if err != nil {
			return fmt.Errorf("unable to set pipeline to PAUSED state: %w", err)
		}
	} else {
		err := d.Pipeline.SetState(gst.StatePlaying)
		if err != nil {
			return fmt.Errorf("unable to set pipeline to PLAYING state: %w", err)
		}
	}
	return nil
}

func (d *Decoder) Seek(ctx context.Context, pos time.Duration, isRelative bool, quick bool) error {
	var seekFlags gst.SeekFlags = gst.SeekFlagFlush
	if !quick {
		seekFlags |= gst.SeekFlagAccurate
	} else {
		seekFlags |= gst.SeekFlagKeyUnit
	}

	var seekPos time.Duration
	if isRelative {
		currentPos, err := d.GetPosition(ctx)
		if err != nil {
			return fmt.Errorf("unable to get current position for relative seek: %w", err)
		}
		seekPos = currentPos + pos
	} else {
		seekPos = pos
	}

	ok := d.Pipeline.SeekSimple(
		seekPos.Nanoseconds(),
		gst.FormatTime,
		seekFlags,
	)
	if !ok {
		return fmt.Errorf("unable to seek to position %v (isRelative=%v, quick=%v)", pos, isRelative, quick)
	}
	return nil
}

func (d *Decoder) GetVideoTracks(ctx context.Context) (types.VideoTracks, error) {
	return nil, fmt.Errorf("not implemented")
}

func (d *Decoder) GetAudioTracks(ctx context.Context) (types.AudioTracks, error) {
	return nil, fmt.Errorf("not implemented")
}

func (d *Decoder) GetSubtitlesTracks(ctx context.Context) (types.SubtitlesTracks, error) {
	return nil, fmt.Errorf("not implemented")
}

func (d *Decoder) SetVideoTrack(ctx context.Context, vid int64) error {
	return fmt.Errorf("not implemented")
}

func (d *Decoder) SetAudioTrack(ctx context.Context, aid int64) error {
	return fmt.Errorf("not implemented")
}

func (d *Decoder) SetSubtitlesTrack(ctx context.Context, sid int64) error {
	return fmt.Errorf("not implemented")
}

func (d *Decoder) Stop(ctx context.Context) error {
	d.Pipeline.SetState(gst.StateNull)
	return nil
}

func (d *Decoder) Close(ctx context.Context) error {
	var errs []error
	if d.Playbin != nil {
		d.Playbin.SetState(gst.StateNull)
		d.Playbin.SetProperty("video-sink", nil)
		d.Playbin.SetProperty("audio-sink", nil)
		d.Playbin.Clear()
	}
	if d.Pipeline != nil {
		d.Pipeline.SetState(gst.StateNull)
		if d.AppSink != nil {
			d.Pipeline.Remove(d.AppSink.Element)
		}
		d.Pipeline.Clear()
	}
	if d.AppSink != nil {
		d.AppSink.Clear()
	}
	d.Playbin = nil
	d.AppSink = nil
	d.Pipeline = nil

	if d.ImageRenderer != nil {
		if err := d.ImageRenderer.Close(); err != nil {
			errs = append(errs, fmt.Errorf("unable to close image renderer: %w", err))
		}
		d.ImageRenderer = nil
	}

	if d.AudioRenderer != nil {
		if err := d.AudioRenderer.Close(); err != nil {
			errs = append(errs, fmt.Errorf("unable to close audio renderer: %w", err))
		}
		d.AudioRenderer = nil
	}

	return errors.Join(errs...)
}

func (d *Decoder) SetupForStreaming(ctx context.Context) error {
	return nil
}
