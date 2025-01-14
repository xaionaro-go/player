package builtin

import (
	"context"
	"image"
	"io"
	"time"

	"github.com/xaionaro-go/audio/pkg/audio"
)

type ImageRenderer interface {
	SetImage(img image.Image) error
	Render() error
	SetVisible(bool) error
}

type AudioRenderer interface {
	PlayPCM(
		ctx context.Context,
		sampleRate audio.SampleRate,
		channels audio.Channel,
		format audio.PCMFormat,
		bufferSize time.Duration,
		reader io.Reader,
	) (audio.PlayStream, error)
}
