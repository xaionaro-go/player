//go:build !with_libvlc
// +build !with_libvlc

package player

import (
	"context"
	"fmt"
	"time"

	"github.com/xaionaro-go/player/pkg/player/types"
)

const SupportedLibVLC = false

type LibVLC struct{}

func NewLibVLC(
	ctx context.Context,
	title string,
	opts ...types.Option,
) (*LibVLC, error) {
	return nil, fmt.Errorf("compiled without LibVLC")
}

func (*Manager) NewLibVLC(
	ctx context.Context,
	title string,
	opts ...types.Option,
) (*LibVLC, error) {
	return NewLibVLC(ctx, title)
}

func (*LibVLC) SetupForStreaming(
	ctx context.Context,
) error {
	panic("compiled without LibVLC support")
}

func (*LibVLC) OpenURL(
	ctx context.Context,
	link string,
) error {
	panic("compiled without LibVLC support")
}

func (*LibVLC) EndChan(
	ctx context.Context,
) (<-chan struct{}, error) {
	panic("compiled without LibVLC support")
}

func (*LibVLC) IsEnded(
	ctx context.Context,
) (bool, error) {
	panic("compiled without LibVLC support")
}

func (p *LibVLC) GetPosition(
	ctx context.Context,
) (time.Duration, error) {
	panic("compiled without LibVLC support")
}

func (p *LibVLC) GetLength(
	ctx context.Context,
) (time.Duration, error) {
	panic("compiled without LibVLC support")
}

func (p *LibVLC) ProcessTitle(
	ctx context.Context,
) (string, error) {
	panic("compiled without LibVLC support")
}

func (p *LibVLC) GetLink(
	ctx context.Context,
) (string, error) {
	panic("compiled without LibVLC support")
}

func (*LibVLC) GetSpeed(
	ctx context.Context,
) (float64, error) {
	panic("compiled without LibVLC support")
}

func (*LibVLC) SetSpeed(
	ctx context.Context,
	speed float64,
) error {
	panic("compiled without LibVLC support")
}

func (*LibVLC) GetPause(
	ctx context.Context,
) (bool, error) {
	panic("compiled without LibVLC support")
}

func (*LibVLC) SetPause(
	ctx context.Context,
	pause bool,
) error {
	panic("compiled without LibVLC support")
}

func (*LibVLC) Seek(
	ctx context.Context,
	pos time.Duration,
	isRelative bool,
	quick bool,
) error {
	return fmt.Errorf("not implemented, yet")
}

func (*LibVLC) GetVideoTracks(
	ctx context.Context,
) (types.VideoTracks, error) {
	panic("compiled without LibVLC support")
}

func (*LibVLC) GetAudioTracks(
	ctx context.Context,
) (types.AudioTracks, error) {
	panic("compiled without LibVLC support")
}

func (*LibVLC) GetSubtitlesTracks(
	ctx context.Context,
) (types.SubtitlesTracks, error) {
	panic("compiled without LibVLC support")
}

func (*LibVLC) SetVideoTrack(
	ctx context.Context,
	vid int64,
) error {
	panic("compiled without LibVLC support")
}

func (*LibVLC) SetAudioTrack(
	ctx context.Context,
	aid int64,
) error {
	panic("compiled without LibVLC support")
}

func (*LibVLC) SetSubtitlesTrack(
	ctx context.Context,
	sid int64,
) error {
	panic("compiled without LibVLC support")
}

func (*LibVLC) Stop(
	ctx context.Context,
) error {
	panic("compiled without LibVLC support")
}

func (*LibVLC) Close(ctx context.Context) error {
	panic("compiled without LibVLC support")
}
