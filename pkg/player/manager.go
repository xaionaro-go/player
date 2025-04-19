package player

import (
	"context"
	"fmt"

	"github.com/facebookincubator/go-belt/tool/logger"
	"github.com/xaionaro-go/player/pkg/player/types"
	"github.com/xaionaro-go/xsync"
)

type Manager struct {
	CommonOptions []types.Option

	PlayersLocker xsync.Mutex
	Players       []Player
}

func NewManager(opts ...types.Option) *Manager {
	return &Manager{
		CommonOptions: opts,
	}
}

func SupportedBackends() []Backend {
	var result []Backend
	if SupportedLibVLC {
		result = append(result, BackendLibVLC)
	}
	if SupportedMPV {
		result = append(result, BackendMPV)
	}
	if SupportedBuiltinLibAV {
		result = append(result, BackendBuiltinLibAV)
	}
	return result
}

func (m *Manager) SupportedBackends() []Backend {
	return SupportedBackends()
}

func (m *Manager) NewPlayer(
	ctx context.Context,
	title string,
	backend Backend,
	opts ...types.Option,
) (Player, error) {
	logger.Debugf(ctx, "NewPlayer: '%s' '%s'", title, backend)
	switch backend {
	case BackendBuiltinLibAV:
		return m.NewBuiltinLibAV(ctx, title, m.opts(opts)...)
	case BackendLibVLC:
		return m.NewLibVLC(ctx, title, m.opts(opts)...)
	case BackendMPV:
		return m.NewMPV(ctx, title, m.opts(opts)...)
	default:
		return nil, fmt.Errorf("unexpected backend type: '%s'", backend)
	}
}

func (m *Manager) opts(opts []types.Option) []types.Option {
	result := make([]types.Option, 0, len(m.CommonOptions)+len(opts))
	result = append(result, m.CommonOptions...)
	result = append(result, opts...)
	return result
}
