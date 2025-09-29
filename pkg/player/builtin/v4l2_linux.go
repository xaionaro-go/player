package builtin

import (
	"context"
	"fmt"
	"sync"

	"github.com/facebookincubator/go-belt/tool/logger"
	"github.com/xaionaro-go/avpipeline/codec"
	"github.com/xaionaro-go/avpipeline/frame"
	"github.com/xaionaro-go/avpipeline/kernel"
	"github.com/xaionaro-go/avpipeline/packet"
	"github.com/xaionaro-go/avpipeline/types"
	"github.com/xaionaro-go/observability"
	"github.com/xaionaro-go/secret"
	"github.com/xaionaro-go/xsync"
)

type ImageRendererV4L2Output struct {
	Output *kernel.Output

	EncoderMutex xsync.Mutex
	Encoder      *kernel.Encoder[*codec.NaiveEncoderFactory]
}

var _ ImageRenderer[frame.Input] = (*ImageRendererV4L2Output)(nil)

func NewImageRendererV4L2Output(
	ctx context.Context,
	devicePath string,
) (*ImageRendererV4L2Output, error) {
	var opts types.DictionaryItems
	opts = append(opts, types.DictionaryItem{
		Key:   "f",
		Value: "v4l2",
	})
	output, err := kernel.NewOutputFromURL(ctx, devicePath, secret.String{}, kernel.OutputConfig{
		CustomOptions: opts,
		WaitForOutputStreams: &kernel.OutputConfigWaitForOutputStreams{
			MinStreams: 0,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("unable to initialize the output: %w", err)
	}
	return &ImageRendererV4L2Output{
		Output: output,
	}, nil
}

func (r *ImageRendererV4L2Output) Close() error {
	ctx := context.TODO()
	return r.Output.Close(ctx)
}

var (
	closedPacketsCh = make(chan packet.Output)
	closedFramesCh  = make(chan frame.Output)
)

func init() {
	close(closedPacketsCh)
	close(closedFramesCh)
}

func (r *ImageRendererV4L2Output) SetImage(
	ctx context.Context,
	f frame.Input,
) (_err error) {
	logger.Tracef(ctx, "SetImage")
	defer func() { logger.Tracef(ctx, "/SetImage: %v", _err) }()
	return xsync.DoA2R1(ctx, &r.EncoderMutex, r.setImage, ctx, f)
}

func (r *ImageRendererV4L2Output) setImage(
	ctx context.Context,
	f frame.Input,
) (_err error) {
	if r.Encoder == nil {
		err := r.initEncoder(ctx, f)
		if err != nil {
			return fmt.Errorf("unable to initialize the encoder: %w", err)
		}
	}

	var err error
	packetsCh := make(chan packet.Output)
	var wg sync.WaitGroup
	wg.Add(1)
	observability.Go(ctx, func(ctx context.Context) {
		defer wg.Done()
		for p := range packetsCh {
			err = r.Output.SendInputPacket(
				ctx,
				packet.BuildInput(
					p.Packet,
					p.StreamInfo,
				),
				closedPacketsCh, closedFramesCh)

			if err != nil {
				break
			}
		}
	})

	err = r.Encoder.SendInputFrame(ctx, f, packetsCh, closedFramesCh)
	close(packetsCh)
	if err != nil {
		return fmt.Errorf("unable to send the frame to the encoder: %w", err)
	}

	wg.Wait()
	return err
}

func (r *ImageRendererV4L2Output) initEncoder(
	ctx context.Context,
	_ frame.Input,
) (_err error) {
	logger.Debugf(ctx, "initEncoder")
	defer func() { logger.Debugf(ctx, "/initEncoder: %v", _err) }()

	r.Encoder = kernel.NewEncoder(
		ctx,
		codec.NewNaiveEncoderFactory(ctx, &codec.NaiveEncoderFactoryParams{
			VideoCodec: "rawvideo",
			AudioCodec: "INTENTIONALLY-INVALID-VALUE",
		}),
		nil,
	)
	return nil
}
