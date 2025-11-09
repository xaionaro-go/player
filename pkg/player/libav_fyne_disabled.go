//go:build !with_libav || !with_fyne
// +build !with_libav !with_fyne

package player

import (
	"context"
	"fmt"

	"github.com/xaionaro-go/player/pkg/player/types"
)

const SupportedLibAVFyne = false

func NewLibAVFyne(
	ctx context.Context,
	title string,
	opts ...types.Option,
) (Player, error) {
	return nil, fmt.Errorf("not supported, yet")
}

func (m *Manager) NewLibAVFyne(
	ctx context.Context,
	title string,
	opts ...types.Option,
) (Player, error) {
	return NewLibAVFyne(ctx, title, opts...)
}
