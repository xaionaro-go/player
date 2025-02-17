package builtin

import (
	"context"
	"fmt"

	"github.com/xaionaro-go/avpipeline"
)

func (p *Player) initImageFor(
	_ context.Context,
	frame *avpipeline.Frame,
) error {
	var err error
	p.currentImage, err = frame.Data().GuessImageFormat()
	if err != nil {
		return fmt.Errorf("unable to guess the image format: %w", err)
	}
	err = p.ImageRenderer.SetImage(p.currentImage)
	if err != nil {
		return fmt.Errorf("unable to render the image: %w", err)
	}
	return nil
}
