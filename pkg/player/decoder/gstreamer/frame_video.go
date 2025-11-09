package gstreamer

import (
	"image"

	"github.com/xaionaro-go/player/pkg/player/imagerenderer"
)

type FrameVideo struct {
	*image.RGBA
}

var _ imagerenderer.ImageGetter = (*FrameVideo)(nil)

func (fv FrameVideo) GetImage() image.Image {
	return fv.RGBA
}
