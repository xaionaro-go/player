//go:build !with_libav
// +build !with_libav

package player

import (
	"context"
	"fmt"
)

const SupportedBuiltinLibAV = false

func NewBuiltinLibAV(
	ctx context.Context,
	title string,
) (Player, error) {
	return nil, fmt.Errorf("not supported, yet")
}

func (m *Manager) NewBuiltinLibAV(
	ctx context.Context,
	title string,
) (Player, error) {
	return NewBuiltinLibAV(ctx, title)
}
