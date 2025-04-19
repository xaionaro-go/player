//go:build with_libvlc
// +build with_libvlc

package player

import (
	"context"

	"github.com/xaionaro-go/player/pkg/player/types"
	"github.com/xaionaro-go/player/pkg/player/vlcserver"
)

const SupportedLibVLC = true

func (m *Manager) NewLibVLC(
	ctx context.Context,
	title string,
	opts ...types.Option,
) (*LibVLC, error) {
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
