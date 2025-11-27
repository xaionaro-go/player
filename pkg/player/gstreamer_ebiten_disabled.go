//go:build !with_gstreamer || !with_ebiten
// +build !with_gstreamer !with_ebiten

package player

import (
	"context"
	"fmt"
	"time"

	"github.com/xaionaro-go/player/pkg/player/types"
)

const SupportedGStreamerEbiten = false

type GStreamerEbiten struct{}

func NewGStreamerEbiten(
	ctx context.Context,
	title string,
	opts ...types.Option,
) (*GStreamerEbiten, error) {
	return nil, fmt.Errorf("compiled without GStreamerEbiten")
}

func (*Manager) NewGStreamerEbiten(
	ctx context.Context,
	title string,
	opts ...types.Option,
) (*GStreamerEbiten, error) {
	return NewGStreamerEbiten(ctx, title)
}

func (*GStreamerEbiten) SetupForStreaming(
	ctx context.Context,
) error {
	panic("compiled without GStreamerEbiten support")
}

func (*GStreamerEbiten) OpenURL(
	ctx context.Context,
	link string,
) error {
	panic("compiled without GStreamerEbiten support")
}

func (*GStreamerEbiten) EndChan(
	ctx context.Context,
) (<-chan struct{}, error) {
	panic("compiled without GStreamerEbiten support")
}

func (*GStreamerEbiten) IsEnded(
	ctx context.Context,
) (bool, error) {
	panic("compiled without GStreamerEbiten support")
}

func (p *GStreamerEbiten) GetPosition(
	ctx context.Context,
) (time.Duration, error) {
	panic("compiled without GStreamerEbiten support")
}

func (p *GStreamerEbiten) GetAudioPosition(
	ctx context.Context,
) (time.Duration, error) {
	panic("compiled without GStreamerEbiten support")
}

func (p *GStreamerEbiten) GetLength(
	ctx context.Context,
) (time.Duration, error) {
	panic("compiled without GStreamerEbiten support")
}

func (p *GStreamerEbiten) ProcessTitle(
	ctx context.Context,
) (string, error) {
	panic("compiled without GStreamerEbiten support")
}

func (p *GStreamerEbiten) GetLink(
	ctx context.Context,
) (string, error) {
	panic("compiled without GStreamerEbiten support")
}

func (*GStreamerEbiten) GetSpeed(
	ctx context.Context,
) (float64, error) {
	panic("compiled without GStreamerEbiten support")
}

func (*GStreamerEbiten) SetSpeed(
	ctx context.Context,
	speed float64,
) error {
	panic("compiled without GStreamerEbiten support")
}

func (*GStreamerEbiten) GetPause(
	ctx context.Context,
) (bool, error) {
	panic("compiled without GStreamerEbiten support")
}

func (*GStreamerEbiten) SetPause(
	ctx context.Context,
	pause bool,
) error {
	panic("compiled without GStreamerEbiten support")
}

func (*GStreamerEbiten) Seek(
	ctx context.Context,
	pos time.Duration,
	isRelative bool,
	quick bool,
) error {
	return fmt.Errorf("not implemented, yet")
}

func (*GStreamerEbiten) GetVideoTracks(
	ctx context.Context,
) (types.VideoTracks, error) {
	panic("compiled without GStreamerEbiten support")
}

func (*GStreamerEbiten) GetAudioTracks(
	ctx context.Context,
) (types.AudioTracks, error) {
	panic("compiled without GStreamerEbiten support")
}

func (*GStreamerEbiten) GetSubtitlesTracks(
	ctx context.Context,
) (types.SubtitlesTracks, error) {
	panic("compiled without GStreamerEbiten support")
}

func (*GStreamerEbiten) SetVideoTrack(
	ctx context.Context,
	vid int64,
) error {
	panic("compiled without GStreamerEbiten support")
}

func (*GStreamerEbiten) SetAudioTrack(
	ctx context.Context,
	aid int64,
) error {
	panic("compiled without GStreamerEbiten support")
}

func (*GStreamerEbiten) SetSubtitlesTrack(
	ctx context.Context,
	sid int64,
) error {
	panic("compiled without GStreamerEbiten support")
}

func (*GStreamerEbiten) Stop(
	ctx context.Context,
) error {
	panic("compiled without GStreamerEbiten support")
}

func (*GStreamerEbiten) Close(ctx context.Context) error {
	panic("compiled without GStreamerEbiten support")
}
