//go:build with_gstreamer && with_ebiten
// +build with_gstreamer,with_ebiten

package player

import (
	"context"
	"errors"
	"fmt"

	"github.com/xaionaro-go/audio/pkg/audio"
	"github.com/xaionaro-go/player/pkg/player/audiorenderer"
	"github.com/xaionaro-go/player/pkg/player/decoder/gstreamer"
	"github.com/xaionaro-go/player/pkg/player/imagerenderer/ebiten"
	"github.com/xaionaro-go/player/pkg/player/types"
)

const SupportedGStreamerEbiten = true

type GStreamerEbiten struct {
	*gstreamer.Decoder
	*ebiten.Window
	audiorenderer.AudioRenderer
}

func NewGStreamerEbiten(
	ctx context.Context,
	title string,
	opts ...types.Option,
) (*GStreamerEbiten, error) {
	videoRenderer, err := ebiten.NewWindow(ctx, title, opts...)
	if err != nil {
		return nil, fmt.Errorf("unable to create an ebiten window: %w", err)
	}
	audioRenderer := audio.NewPlayerAuto(ctx)
	decoder, err := gstreamer.New(ctx, videoRenderer, audioRenderer)
	if err != nil {
		return nil, fmt.Errorf("unable to create a gstreamer decoder: %w", err)
	}
	return &GStreamerEbiten{
		Decoder:       decoder,
		Window:        videoRenderer,
		AudioRenderer: audioRenderer,
	}, nil
}

func (*Manager) NewGStreamerEbiten(
	ctx context.Context,
	title string,
	opts ...types.Option,
) (*GStreamerEbiten, error) {
	return NewGStreamerEbiten(ctx, title)
}

func (p *GStreamerEbiten) Close(
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
