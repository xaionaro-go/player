package builtin

import (
	"context"
	"fmt"

	"github.com/facebookincubator/go-belt/tool/logger"
	"github.com/xaionaro-go/avpipeline/frame"
)

func (p *Player[I]) initImageFor(
	ctx context.Context,
	frame frame.Input,
) (_err error) {
	logger.Debugf(ctx, "initImageFor")
	defer func() { logger.Debugf(ctx, "/initImageFor: %v", _err) }()

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
		Input: frame,
	})
	if err != nil {
		return fmt.Errorf("unable to render the image: %w", err)
	}
	return nil
}
