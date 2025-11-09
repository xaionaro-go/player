package libav

import (
	"context"
	"image"

	"github.com/xaionaro-go/avpipeline/frame"
	"github.com/xaionaro-go/player/pkg/player/audiorenderer"
	"github.com/xaionaro-go/player/pkg/player/imagerenderer"
)

type ImageGeneric struct {
	*Decoder
	frame.Input
	image.Image
}

func (img ImageGeneric) GetImage() image.Image {
	return img.Image
}

type ImageUnparsed struct {
	*Decoder
	frame.Input
}

type ImageRenderer = imagerenderer.ImageRenderer
type AudioRenderer = audiorenderer.AudioRenderer
type SetVisibler = imagerenderer.SetVisibler
type RenderImageNower = imagerenderer.RenderImageNower

type AVFrameRenderer interface {
	SetAVFrame(ctx context.Context, img ImageUnparsed) error
}
