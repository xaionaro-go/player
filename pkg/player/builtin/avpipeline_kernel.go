package builtin

import (
	"context"
	"fmt"

	"github.com/xaionaro-go/avpipeline/frame"
	"github.com/xaionaro-go/avpipeline/kernel"
	"github.com/xaionaro-go/avpipeline/packet"
)

var _ kernel.Abstract = (*Player)(nil)

func (p *Player) SendInputPacket(
	ctx context.Context,
	input packet.Input,
	outputPacketsCh chan<- packet.Output,
	outputFramesCh chan<- frame.Output,
) error {
	return fmt.Errorf("player expects to receive only decoded frames")
}

func (p *Player) SendInputFrame(
	ctx context.Context,
	input frame.Input,
	outputPacketsCh chan<- packet.Output,
	outputFramesCh chan<- frame.Output,
) error {
	return p.processFrame(ctx, input)
}

func (p *Player) String() string {
	return "Player"
}

func (p *Player) Close(ctx context.Context) error {
	return fmt.Errorf("not implemented, yet")
}

func (p *Player) CloseChan() <-chan struct{} {
	return p.endChan
}

func (p *Player) Generate(
	ctx context.Context,
	outputPacketsCh chan<- packet.Output,
	outputFramesCh chan<- frame.Output,
) error {
	return nil
}
