package ebiten

import (
	"context"
	"fmt"
	"image"
	"sync/atomic"

	"github.com/facebookincubator/go-belt"
	"github.com/facebookincubator/go-belt/tool/logger"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/xaionaro-go/observability"
	"github.com/xaionaro-go/player/pkg/player/imagerenderer"
	"github.com/xaionaro-go/player/pkg/player/types"
)

// this implementation is heavily inspired by https://github.com/realskyquest/ebiten-gstreamer

type Window struct {
	inputImage      *image.RGBA
	renderImage     *ebiten.Image
	shouldTerminate atomic.Bool
	observability   *belt.Belt
	prevFPS         float64
}

var _ imagerenderer.ImageRenderer = (*Window)(nil)
var _ imagerenderer.RenderImageNower = (*Window)(nil)

func NewWindow(
	ctx context.Context,
	title string,
	opts ...types.Option,
) (_ret *Window, _err error) {
	logger.Debugf(ctx, "NewWindow(ctx, '%s', %#+v)", title, opts)
	defer func() { logger.Debugf(ctx, "/NewWindow: %v %v", _ret, _err) }()

	cfg := types.Options(opts).Config()
	_ = cfg

	r := &Window{
		observability: belt.CtxBelt(ctx),
	}

	ebiten.SetWindowSize(1920, 1080)
	ebiten.SetWindowTitle(title)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetVsyncEnabled(true)
	observability.Go(ctx, func(ctx context.Context) {
		logger.Debugf(ctx, "ebiten.RunGame")
		defer logger.Debugf(ctx, "/ebiten.RunGame")
		if err := ebiten.RunGame(r); err != nil {
			logger.Errorf(ctx, "unable to run ebiten window: %w", err)
		}
	})
	return r, nil
}

func (r *Window) logger() logger.Logger {
	return logger.FromBelt(r.observability)
}

func (r *Window) SetImage(
	ctx context.Context,
	img imagerenderer.ImageGetter,
) error {
	var ok bool
	r.inputImage, ok = img.GetImage().(*image.RGBA)
	if !ok {
		return fmt.Errorf("only RGBA images are supported")
	}
	return nil
}

func (r *Window) RenderImageNow(
	ctx context.Context,
) (_err error) {
	logger.Tracef(ctx, "RenderImageNow")
	defer func() { logger.Tracef(ctx, "/RenderImageNow: %v", _err) }()

	assert(ctx, r.inputImage != nil, "inputImage must be set before RenderImageNow is called")
	if r.renderImage == nil || r.renderImage.Bounds() != r.inputImage.Bounds() {
		r.renderImage = ebiten.NewImageFromImageWithOptions(
			r.inputImage,
			&ebiten.NewImageFromImageOptions{
				Unmanaged: true,
			},
		)
	}
	r.renderImage.WritePixels(r.inputImage.Pix)
	return nil
}

func (r *Window) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}

func (r *Window) Draw(screen *ebiten.Image) {
	if newFPS := ebiten.ActualFPS(); newFPS != r.prevFPS {
		r.logger().Debugf("FPS: %0.2f", ebiten.ActualFPS())
		r.prevFPS = newFPS
	}

	inputSize := r.inputImage.Bounds().Size()
	screenSize := screen.Bounds().Size()
	r.logger().Tracef("screenSize=%v inputSize=%v", screenSize, inputSize)
	if screenSize == inputSize {
		screen.DrawImage(r.renderImage, nil)
		return
	}

	sw, sh := screenSize.X, screenSize.Y
	iw, ih := inputSize.X, inputSize.Y
	opts := &ebiten.DrawImageOptions{
		Filter: ebiten.FilterLinear,
	}
	scale := min(float64(sw)/float64(iw), float64(sh)/float64(ih))
	opts.GeoM.Scale(scale, scale) // scaling
	opts.GeoM.Translate(
		float64(sw-int(float64(iw)*scale))/2,
		float64(sh-int(float64(ih)*scale))/2,
	) // recentering
	screen.DrawImage(r.renderImage, opts)
}

func (r *Window) Update() error {
	if r.shouldTerminate.Load() {
		return ebiten.Termination
	}
	return nil
}

func (r *Window) Close() error {
	r.shouldTerminate.Store(true)
	return nil
}
