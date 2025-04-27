package builtin

import (
	"context"
	"image"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"github.com/facebookincubator/go-belt/tool/logger"
	"github.com/xaionaro-go/audio/pkg/audio"
	"github.com/xaionaro-go/player/pkg/player/types"
)

type WindowRenderer struct {
	window       fyne.Window
	imageRaster  *canvas.Raster
	currentImage image.Image
	resizeOnce   sync.Once
}

func (r *WindowRenderer) SetImage(img image.Image) error {
	r.currentImage = img
	r.resizeOnce.Do(func() {
		bounds := img.Bounds()
		size := fyne.NewSize(float32(bounds.Dx()), float32(bounds.Dy()))
		if size.Width == 0 {
			size.Width = 1920
		}
		if size.Height == 0 {
			size.Height = 1080
		}
		r.imageRaster.Resize(size)
		r.imageRaster.SetMinSize(size)
		r.window.Resize(size)
	})
	return nil
}

func (r *WindowRenderer) Render() error {
	r.imageRaster.Refresh()
	return nil
}

func (r *WindowRenderer) SetVisible(visible bool) error {
	if visible {
		r.window.Show()
	} else {
		r.window.Hide()
	}
	return nil
}

func (r *WindowRenderer) GetImage(w, h int) image.Image {
	return r.currentImage
}

func (r *WindowRenderer) Close() error {
	r.window.Hide()
	r.window.Close()
	return nil
}

func NewWindow(
	ctx context.Context,
	title string,
	opts ...types.Option,
) *Player {
	logger.Debugf(ctx, "NewWindow(ctx, '%s', %#+v)", title, opts)
	cfg := types.Options(opts).Config()
	r := &WindowRenderer{
		window:       fyne.CurrentApp().NewWindow(title),
		currentImage: image.NewRGBA(image.Rect(0, 0, 1, 1)),
	}
	r.imageRaster = canvas.NewRaster(r.GetImage)
	r.imageRaster.ScaleMode = canvas.ImageScaleFastest
	r.imageRaster.Show()
	r.window.SetContent(container.NewStack(r.imageRaster))
	if !cfg.HideWindow {
		r.window.Show()
	}
	return New(
		ctx,
		r,
		audio.NewPlayerAuto(ctx),
	)
}
