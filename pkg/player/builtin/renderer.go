package builtin

import (
	"context"
	"image"
	"io"
	"time"

	"github.com/xaionaro-go/audio/pkg/audio"
	"github.com/xaionaro-go/avpipeline/frame"
)

type ImageGeneric struct {
	image.Image
}

type ImageAny interface {
	ImageGeneric | frame.Input
}

type ImageRenderer[I ImageAny] interface {
	io.Closer
	SetImage(ctx context.Context, img I) error
}

type RenderNower interface {
	RenderNow() error
}

type SetVisibler interface {
	SetVisible(bool) error
}

type AudioRenderer interface {
	io.Closer
	PlayPCM(
		ctx context.Context,
		sampleRate audio.SampleRate,
		channels audio.Channel,
		format audio.PCMFormat,
		bufferSize time.Duration,
		reader io.Reader,
	) (audio.PlayStream, error)
}
