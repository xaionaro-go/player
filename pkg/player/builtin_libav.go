//go:build with_libav
// +build with_libav

package player

import (
	"context"

	"github.com/xaionaro-go/player/pkg/player/builtin"
)

const SupportedBuiltinLibAV = true

type BuiltinLibAV = builtin.Player

func NewBuiltinLibAV(
	ctx context.Context,
	title string,
) (*BuiltinLibAV, error) {
	return builtin.NewWindow(ctx, title), nil
}

func (m *Manager) NewBuiltinLibAV(
	ctx context.Context,
	title string,
) (*BuiltinLibAV, error) {
	return NewBuiltinLibAV(ctx, title)
}
