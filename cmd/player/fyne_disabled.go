//go:build !with_fyne
// +build !with_fyne

package main

import (
	"context"
	"time"

	"github.com/xaionaro-go/avpipeline/logger"
	"github.com/xaionaro-go/player/pkg/player/types"
)

func runPlayerControls(
	ctx context.Context,
	p types.Player,
) {
	defer logger.Infof(ctx, "player controls ended")
	endCh, _ := p.EndChan(ctx)
	t := time.NewTicker(1 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			logger.Debugf(ctx, "context done: %v", ctx.Err())
			return
		case <-endCh:
			logger.Debugf(ctx, "player ended")
			return
		case <-t.C:
		}
		length := try(p.GetLength(ctx))
		position := try(p.GetPosition(ctx))
		speed := try(p.GetSpeed(ctx))
		paused := try(p.GetPause(ctx))
		logger.Infof(
			ctx,
			"position: %s / %s (%.2fx) paused=%v",
			position.String(),
			length.String(),
			speed,
			paused,
		)
	}
}
