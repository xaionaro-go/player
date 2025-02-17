package builtin

import (
	"context"

	"github.com/xaionaro-go/avpipeline"
)

func (p *Player) NewDecoder(
	ctx context.Context,
	pkt avpipeline.InputPacket,
) (*avpipeline.Decoder, error) {
	return avpipeline.NewDecoder(ctx, "", pkt.Stream.CodecParameters(), 0, "", nil, 0)
}

func (p *Player) SendFrame(
	ctx context.Context,
	frame *avpipeline.Frame,
) error {
	return p.processFrame(ctx, frame)
}
