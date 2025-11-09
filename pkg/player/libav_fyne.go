//go:build with_libav && with_fyne
// +build with_libav,with_fyne

package player

import (
	"context"
	"errors"
	"fmt"

	"github.com/xaionaro-go/audio/pkg/audio"
	"github.com/xaionaro-go/player/pkg/player/audiorenderer"
	"github.com/xaionaro-go/player/pkg/player/decoder/libav"
	"github.com/xaionaro-go/player/pkg/player/imagerenderer/fyne"
	"github.com/xaionaro-go/player/pkg/player/types"
)

const SupportedLibAVFyne = true

type LibAVFyne struct {
	*libav.Decoder
	*fyne.Window
	audiorenderer.AudioRenderer
}

func NewLibAVFyne(
	ctx context.Context,
	title string,
	opts ...types.Option,
) (*LibAVFyne, error) {
	videoRenderer := fyne.NewWindow(ctx, title, opts...)
	audioRenderer := audio.NewPlayerAuto(ctx)
	decoder := libav.New(ctx, videoRenderer, audioRenderer)
	return &LibAVFyne{
		Decoder:       decoder,
		Window:        videoRenderer,
		AudioRenderer: audioRenderer,
	}, nil
}

func (m *Manager) NewLibAVFyne(
	ctx context.Context,
	title string,
	opts ...types.Option,
) (*LibAVFyne, error) {
	return NewLibAVFyne(ctx, title, opts...)
}

func (p *LibAVFyne) Close(
	ctx context.Context,
) error {
	var errs []error
	if err := p.Decoder.Close(ctx); err != nil {
		errs = append(errs, fmt.Errorf("unable to close decoder: %w", err))
	}
	if err := p.AudioRenderer.Close(); err != nil {
		errs = append(errs, fmt.Errorf("unable to close audio renderer: %w", err))
	}
	if err := p.Window.Close(); err != nil {
		errs = append(errs, fmt.Errorf("unable to close window: %w", err))
	}
	return errors.Join(errs...)
}
