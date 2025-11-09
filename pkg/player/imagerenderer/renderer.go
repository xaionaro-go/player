package imagerenderer

import (
	"context"
	"image"
	"io"
)

type ImageGetter interface {
	GetImage() image.Image
}

type ImageRenderer interface {
	io.Closer // TODO: remove this from here
	SetImage(ctx context.Context, img ImageGetter) error
}

type RenderImageNower interface {
	RenderImageNow(context.Context) error
}

type SetVisibler interface {
	SetVisible(bool) error
}
