package audiorenderer

import (
	"context"
	"io"
	"time"

	"github.com/xaionaro-go/audio/pkg/audio"
)

type AudioRenderer interface {
	io.Closer // TODO: remove this from here
	PlayPCM(
		ctx context.Context,
		sampleRate audio.SampleRate,
		channels audio.Channel,
		format audio.PCMFormat,
		bufferSize time.Duration,
		reader io.Reader,
	) (audio.PlayStream, error)
}
