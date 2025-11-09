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
	if SupportedGStreamerEbiten {
		result = append(result, BackendGStreamerEbiten)
	}
	if SupportedLibVLC {
		result = append(result, BackendLibVLC)
	}
	if SupportedMPV {
		result = append(result, BackendMPV)
	}
	if SupportedLibAVEbiten {
		result = append(result, BackendLibAVEbiten)
	}
	if SupportedLibAVFyne {
		result = append(result, BackendLibAVFyne)
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
	opts = m.opts(opts)
	logger.Debugf(ctx, "NewPlayer: '%s' '%s' (%v)", title, backend, opts)
	switch backend {
	case BackendLibAVFyne:
		return m.NewLibAVFyne(ctx, title, opts...)
	case BackendLibAVEbiten:
		return m.NewLibAVEbiten(ctx, title, opts...)
	case BackendLibVLC:
		return m.NewLibVLC(ctx, title, opts...)
	case BackendGStreamerFyne:
		return nil, fmt.Errorf("not implemented")
	case BackendGStreamerEbiten:
		return m.NewGStreamerEbiten(ctx, title, opts...)
	case BackendMPV:
		return m.NewMPV(ctx, title, opts...)
	default:
		return nil, fmt.Errorf("unexpected backend type: '%s'", backend)
	}
}

func (m *Manager) opts(opts []types.Option) types.Options {
	result := make([]types.Option, 0, len(m.CommonOptions)+len(opts))
	result = append(result, m.CommonOptions...)
	result = append(result, opts...)
	return result
}
