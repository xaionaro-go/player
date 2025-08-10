package builtin

import (
	"context"
	"fmt"

	"github.com/xaionaro-go/avpipeline/frame"
)

func (p *Player[I]) initImageFor(
	ctx context.Context,
	frame frame.Input,
) error {
	r, ok := p.ImageRenderer.(ImageRenderer[ImageGeneric])
	if !ok {
		return nil
	}
	var err error
	p.currentImage, err = frame.Data().GuessImageFormat()
	if err != nil {
		return fmt.Errorf("unable to guess the image format: %w", err)
	}
	err = r.SetImage(ctx, ImageGeneric{
		Image: p.currentImage,
	})
	if err != nil {
		return fmt.Errorf("unable to render the image: %w", err)
	}
	return nil
}
