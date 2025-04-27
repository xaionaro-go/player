//go:build with_libvlc
// +build with_libvlc

package player

import (
	"context"

	"github.com/facebookincubator/go-belt/tool/logger"
	"github.com/xaionaro-go/player/pkg/player/types"
	"github.com/xaionaro-go/player/pkg/player/vlcserver"
)

const SupportedLibVLC = true

func (m *Manager) NewLibVLC(
	ctx context.Context,
	title string,
	opts ...types.Option,
) (*LibVLC, error) {
	cfg := types.Options(opts).Config()

	if cfg.HideWindow {
		logger.Errorf(ctx, "hiding the VLC window is not supported, et")
	}

	r, err := NewLibVLC(ctx, title)
	if err != nil {
		return nil, err
	}

	m.PlayersLocker.Do(ctx, func() {
		m.Players = append(m.Players, r)
	})
	return r, nil
}

type LibVLC = vlcserver.VLC

var _ Player = (*LibVLC)(nil)

func NewLibVLC(
	ctx context.Context,
	title string,
) (*LibVLC, error) {
	return vlcserver.Run(ctx, title)
}
