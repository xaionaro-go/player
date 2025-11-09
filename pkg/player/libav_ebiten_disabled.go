//go:build !with_libav || !with_ebiten
// +build !with_libav !with_ebiten

package player

import (
	"context"
	"fmt"

	"github.com/xaionaro-go/player/pkg/player/types"
)

const SupportedLibAVEbiten = false

func NewLibAVEbiten(
	ctx context.Context,
	title string,
	opts ...types.Option,
) (Player, error) {
	return nil, fmt.Errorf("not supported, yet")
}

func (m *Manager) NewLibAVEbiten(
	ctx context.Context,
	title string,
	opts ...types.Option,
) (Player, error) {
	return NewLibAVEbiten(ctx, title, opts...)
}
