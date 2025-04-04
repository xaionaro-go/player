package player

import (
	"context"
	"fmt"

	"github.com/facebookincubator/go-belt/tool/logger"
	"github.com/xaionaro-go/player/pkg/player/types"
	"github.com/xaionaro-go/xsync"
)

type Manager struct {
	Config types.Config

	PlayersLocker xsync.Mutex
	Players       []Player
}

func NewManager(opts ...types.Option) *Manager {
	return &Manager{
		Config: types.Options(opts).Config(),
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
) (Player, error) {
	logger.Debugf(ctx, "NewPlayer: '%s' '%s'", title, backend)
	switch backend {
	case BackendBuiltinLibAV:
		return m.NewBuiltinLibAV(ctx, title)
	case BackendLibVLC:
		return m.NewLibVLC(ctx, title)
	case BackendMPV:
		return m.NewMPV(ctx, title)
	default:
		return nil, fmt.Errorf("unexpected backend type: '%s'", backend)
	}
}
