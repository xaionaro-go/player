package libav

import (
	"context"
	"errors"
	"fmt"

	"github.com/facebookincubator/go-belt/tool/logger"
	"github.com/xaionaro-go/avpipeline/frame"
	"github.com/xaionaro-go/avpipeline/kernel"
	"github.com/xaionaro-go/avpipeline/packet"
	avptypes "github.com/xaionaro-go/avpipeline/types"
	"github.com/xaionaro-go/xsync"
)

var _ kernel.Abstract = (*Decoder)(nil)

func (p *Decoder) GetObjectID() avptypes.ObjectID {
	return avptypes.GetObjectID(p)
}

func (p *Decoder) SendInputPacket(
	ctx context.Context,
	input packet.Input,
	outputPacketsCh chan<- packet.Output,
	outputFramesCh chan<- frame.Output,
) error {
	return fmt.Errorf("player expects to receive only decoded frames")
}

func (p *Decoder) SendInputFrame(
	ctx context.Context,
	input frame.Input,
	outputPacketsCh chan<- packet.Output,
	outputFramesCh chan<- frame.Output,
) error {
	return p.processFrame(ctx, input)
}

func (p *Decoder) String() string {
	return "Player"
}

func (p *Decoder) Close(ctx context.Context) (_err error) {
	logger.Debugf(ctx, "Close()")
	defer func() { logger.Debugf(ctx, "/Close(): %v", _err) }()

	var errs []error
	var ch <-chan struct{}
	wasRunning := xsync.DoR1(ctx, &p.locker, func() bool {
		ch = p.closedChan
		if p.cancelFunc == nil {
			return false
		}
		p.cancelFunc()
		if p.ImageRenderer != nil {
			if err := p.ImageRenderer.Close(); err != nil {
				errs = append(errs, fmt.Errorf("unable to close the image renderer: %w", err))
			}
		}
		if p.AudioRenderer != nil {
			if err := p.AudioRenderer.Close(); err != nil {
				errs = append(errs, fmt.Errorf("unable to close the audio renderer: %w", err))
			}
		}
		return true
	})
	if !wasRunning {
		// already stopped
		return errors.Join(errs...)
	}
	<-ch
	return errors.Join(errs...)
}

func (p *Decoder) CloseChan() <-chan struct{} {
	ctx := context.TODO()
	ch, err := p.EndChan(ctx)
	if err != nil {
		logger.Errorf(ctx, "unable to get the EndChan: %v", err)
	}
	return ch
}

func (p *Decoder) Generate(
	ctx context.Context,
	outputPacketsCh chan<- packet.Output,
	outputFramesCh chan<- frame.Output,
) error {
	return nil
}
