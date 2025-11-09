//go:build with_libav && with_ebiten
// +build with_libav,with_ebiten

package player

import (
	"context"
	"errors"
	"fmt"

	"github.com/xaionaro-go/audio/pkg/audio"
	"github.com/xaionaro-go/player/pkg/player/audiorenderer"
	"github.com/xaionaro-go/player/pkg/player/decoder/libav"
	"github.com/xaionaro-go/player/pkg/player/imagerenderer/ebiten"
	"github.com/xaionaro-go/player/pkg/player/types"
)

const SupportedLibAVEbiten = true

type LibAVEbiten struct {
	*libav.Decoder
	*ebiten.Window
	audiorenderer.AudioRenderer
}

func NewLibAVEbiten(
	ctx context.Context,
	title string,
	opts ...types.Option,
) (*LibAVEbiten, error) {
	videoRenderer, err := ebiten.NewWindow(ctx, title, opts...)
	if err != nil {
		return nil, fmt.Errorf("unable to create an ebiten window: %w", err)
	}
	audioRenderer := audio.NewPlayerAuto(ctx)
	return &LibAVEbiten{
		Decoder:       libav.New(ctx, videoRenderer, audioRenderer),
		Window:        videoRenderer,
		AudioRenderer: audioRenderer,
	}, nil
}

func (m *Manager) NewLibAVEbiten(
	ctx context.Context,
	title string,
	opts ...types.Option,
) (*LibAVEbiten, error) {
	return NewLibAVEbiten(ctx, title, opts...)
}

func (p *LibAVEbiten) Close(
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
