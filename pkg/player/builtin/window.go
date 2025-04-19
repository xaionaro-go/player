package builtin

import (
	"context"
	"image"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
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

func NewWindow(
	ctx context.Context,
	title string,
	opts ...types.Option,
) *Player {
	r := &WindowRenderer{
		window: fyne.CurrentApp().NewWindow(title),
	}
	r.imageRaster = canvas.NewRaster(r.GetImage)
	r.imageRaster.ScaleMode = canvas.ImageScaleFastest
	r.imageRaster.Show()
	r.window.SetContent(container.NewStack(r.imageRaster))
	r.window.Show()
	return New(
		ctx,
		r,
		audio.NewPlayerAuto(ctx),
	)
}
