package fyne

import (
	"context"
	"image"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"github.com/facebookincubator/go-belt/tool/logger"
	"github.com/xaionaro-go/player/pkg/player/imagerenderer"
	"github.com/xaionaro-go/player/pkg/player/types"
)

type Window struct {
	window       fyne.Window
	imageRaster  *canvas.Raster
	currentImage image.Image
	resizeOnce   sync.Once
}

var _ imagerenderer.ImageRenderer = (*Window)(nil)
var _ imagerenderer.RenderImageNower = (*Window)(nil)
var _ imagerenderer.SetVisibler = (*Window)(nil)

func (r *Window) SetImage(
	ctx context.Context,
	img imagerenderer.ImageGetter,
) (_err error) {
	logger.Tracef(ctx, "SetImage")
	defer func() { logger.Tracef(ctx, "/SetImage: %v", _err) }()
	r.currentImage = img.GetImage()
	r.resizeOnce.Do(func() {
		bounds := r.currentImage.Bounds()
		size := fyne.NewSize(float32(bounds.Dx()), float32(bounds.Dy()))
		if size.Width == 0 {
			size.Width = 1920
		}
		if size.Height == 0 {
			size.Height = 1080
		}
		r.window.Resize(size)
		r.imageRaster = canvas.NewRaster(r.GetImage)
		r.imageRaster.ScaleMode = canvas.ImageScaleFastest
		r.imageRaster.Resize(size)
		r.imageRaster.SetMinSize(size)
		r.imageRaster.Show()
		r.window.SetContent(container.NewStack(r.imageRaster))
	})
	return nil
}

func (r *Window) RenderImageNow(
	ctx context.Context,
) (_err error) {
	logger.Tracef(ctx, "RenderImageNow")
	defer func() { logger.Tracef(ctx, "/RenderImageNow: %v", _err) }()
	r.imageRaster.Refresh()
	return nil
}

func (r *Window) SetVisible(visible bool) error {
	if visible {
		r.window.Show()
	} else {
		r.window.Hide()
	}
	return nil
}

func (r *Window) GetImage(w, h int) image.Image {
	logger.Tracef(context.TODO(), "GetImage(%d, %d)", w, h)
	return r.currentImage
}

func (r *Window) Close() error {
	r.window.Hide()
	r.window.Close()
	return nil
}

func NewWindow(
	ctx context.Context,
	title string,
	opts ...types.Option,
) *Window {
	logger.Debugf(ctx, "NewWindow(ctx, '%s', %#+v)", title, opts)
	defer func() { logger.Debugf(ctx, "/NewWindow") }()
	cfg := types.Options(opts).Config()
	r := &Window{
		window: fyne.CurrentApp().NewWindow(title),
	}
	if !cfg.HideWindow {
		r.window.Show()
	}
	return r
}
