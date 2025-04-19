//go:build with_libav
// +build with_libav

package player

import (
	"context"

	"github.com/xaionaro-go/player/pkg/player/builtin"
	"github.com/xaionaro-go/player/pkg/player/types"
)

const SupportedBuiltinLibAV = true

type BuiltinLibAV = builtin.Player

func NewBuiltinLibAV(
	ctx context.Context,
	title string,
	opts ...types.Option,
) (*BuiltinLibAV, error) {
	return builtin.NewWindow(ctx, title, opts...), nil
}

func (m *Manager) NewBuiltinLibAV(
	ctx context.Context,
	title string,
	opts ...types.Option,
) (*BuiltinLibAV, error) {
	return NewBuiltinLibAV(ctx, title, opts...)
}
