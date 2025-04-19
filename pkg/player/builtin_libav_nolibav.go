//go:build !with_libav
// +build !with_libav

package player

import (
	"context"
	"fmt"

	"github.com/xaionaro-go/player/pkg/player/types"
)

const SupportedBuiltinLibAV = false

func NewBuiltinLibAV(
	ctx context.Context,
	title string,
	opts ...types.Option,
) (Player, error) {
	return nil, fmt.Errorf("not supported, yet")
}

func (m *Manager) NewBuiltinLibAV(
	ctx context.Context,
	title string,
	opts ...types.Option,
) (Player, error) {
	return NewBuiltinLibAV(ctx, title, opts...)
}
