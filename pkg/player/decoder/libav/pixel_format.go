package libav

import (
	"context"
	"fmt"

	"github.com/facebookincubator/go-belt/tool/logger"
	"github.com/xaionaro-go/avpipeline/frame"
)

func (p *Decoder) initImageFor(
	ctx context.Context,
	frame frame.Input,
) (_err error) {
	logger.Debugf(ctx, "initImageFor")
	defer func() { logger.Debugf(ctx, "/initImageFor: %v", _err) }()

	if _, ok := p.ImageRenderer.(AVFrameRenderer); ok {
		// no need to decode frames
		return nil
	}

	var err error
	p.currentImage, err = frame.Data().GuessImageFormat()
	if err != nil {
		return fmt.Errorf("unable to guess the image format: %w", err)
	}

	err = p.ImageRenderer.SetImage(ctx, ImageGeneric{
		Image: p.currentImage,
		Input: frame,
	})
	if err != nil {
		return fmt.Errorf("unable to render the image: %w", err)
	}
	return nil
}
