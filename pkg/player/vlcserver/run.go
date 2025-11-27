//go:build with_libvlc
// +build with_libvlc

package vlcserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"time"

	child_process_manager "github.com/AgustinSRG/go-child-process-manager"
	"github.com/facebookincubator/go-belt/tool/experimental/errmon"
	"github.com/facebookincubator/go-belt/tool/logger"
	"github.com/xaionaro-go/observability"
	"github.com/xaionaro-go/player/pkg/player/types"
	"github.com/xaionaro-go/player/pkg/player/vlcserver/client"
	"github.com/xaionaro-go/xpath"
)

const (
	debugRunVLCInTheSameProcess = false
)

type VLC struct {
	Client *client.Client
	Cmd    *exec.Cmd
}

func Run(
	ctx context.Context,
	title string,
) (*VLC, error) {
	if debugRunVLCInTheSameProcess {
		return runInTheSameProcess(ctx, title)
	}
	return run(ctx, title)
}

func run(
	ctx context.Context,
	title string,
) (*VLC, error) {
	execPath, err := xpath.GetExecPath(os.Args[0])
	if err != nil {
		return nil, fmt.Errorf("unable to get self-path: %w", err)
	}
	cmd := exec.Command(execPath)
	cmd.Stderr = os.Stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("unable to initialize an stdout pipe: %w", err)
	}
	cmd.Env = append(os.Environ(), EnvKeyIsVLCServer+"=1")
	cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", EnvKeyLoggingLevel, logger.FromCtx(ctx).Level().String()))
	err = child_process_manager.ConfigureCommand(cmd)
	errmon.ObserveErrorCtx(ctx, err)
	err = cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("unable to start a subprocess to isolate VLC: %w", err)
	}
	err = child_process_manager.AddChildProcess(cmd.Process)
	if err != nil {
		if runtime.GOOS == "windows" {
			// this is actually an error, but I have no idea how to fix it, so demoting to a debug message
			logger.Debugf(ctx, "unable to register the command to be auto-killed: %v", err)
		} else {
			logger.Errorf(ctx, "unable to register the command to be auto-killed: %v", err)
		}
	}

	decoder := json.NewDecoder(stdout)
	var d ReturnedData
	err = decoder.Decode(&d)
	logger.Debugf(ctx, "got data: %#+v", d)
	if err != nil {
		return nil, fmt.Errorf("unable to un-JSON-ize the process output: %w", err)
	}

	return &VLC{
		Client: client.New(title, d.ListenAddr),
		Cmd:    cmd,
	}, nil
}

func runInTheSameProcess(
	ctx context.Context,
	title string,
) (*VLC, error) {
	addrCh := make(chan net.Addr, 1)
	errCh := make(chan error, 1)
	observability.Go(ctx, func(ctx context.Context) {
		errCh <- runVLCServer(ctx, func(reportedAddr net.Addr) error {
			addrCh <- reportedAddr
			return nil
		})
	})
	select {
	case addr := <-addrCh:
		return &VLC{
			Client: client.New(title, addr.String()),
		}, nil
	case err := <-errCh:
		return nil, err
	}
}

func (vlc *VLC) SetupForStreaming(
	ctx context.Context,
) error {
	return nil
}

func (vlc *VLC) ProcessTitle(
	ctx context.Context,
) (string, error) {
	if vlc == nil {
		return "", fmt.Errorf("vlc == nil")
	}
	if vlc.Client == nil {
		return "", fmt.Errorf("vlc.Client == nil")
	}
	return vlc.Client.ProcessTitle(ctx)
}

func (vlc *VLC) OpenURL(
	ctx context.Context,
	link string,
) error {
	if vlc == nil {
		return fmt.Errorf("vlc == nil")
	}
	if vlc.Client == nil {
		return fmt.Errorf("vlc.Client == nil")
	}
	return vlc.Client.OpenURL(ctx, link)
}

func (vlc *VLC) GetLink(
	ctx context.Context,
) (string, error) {
	if vlc == nil {
		return "", fmt.Errorf("vlc == nil")
	}
	if vlc.Client == nil {
		return "", fmt.Errorf("vlc.Client == nil")
	}
	return vlc.Client.GetLink(ctx)
}

func (vlc *VLC) EndChan(
	ctx context.Context,
) (<-chan struct{}, error) {
	if vlc == nil {
		return nil, fmt.Errorf("vlc == nil")
	}
	if vlc.Client == nil {
		return nil, fmt.Errorf("vlc.Client == nil")
	}
	return vlc.Client.EndChan(ctx)
}

func (vlc *VLC) IsEnded(
	ctx context.Context,
) (bool, error) {
	if vlc == nil {
		return false, fmt.Errorf("vlc == nil")
	}
	if vlc.Client == nil {
		return false, fmt.Errorf("vlc.Client == nil")
	}
	return vlc.Client.IsEnded(ctx)
}

func (vlc *VLC) GetPosition(
	ctx context.Context,
) (time.Duration, error) {
	if vlc == nil {
		return 0, fmt.Errorf("vlc == nil")
	}
	if vlc.Client == nil {
		return 0, fmt.Errorf("vlc.Client == nil")
	}
	return vlc.Client.GetPosition(ctx)
}

func (vlc *VLC) GetAudioPosition(
	ctx context.Context,
) (time.Duration, error) {
	if vlc == nil {
		return 0, fmt.Errorf("vlc == nil")
	}
	if vlc.Client == nil {
		return 0, fmt.Errorf("vlc.Client == nil")
	}
	return vlc.Client.GetAudioPosition(ctx)
}

func (vlc *VLC) GetLength(
	ctx context.Context,
) (time.Duration, error) {
	if vlc == nil {
		return 0, fmt.Errorf("vlc == nil")
	}
	if vlc.Client == nil {
		return 0, fmt.Errorf("vlc.Client == nil")
	}
	return vlc.Client.GetLength(ctx)
}

func (vlc *VLC) GetSpeed(
	ctx context.Context,
) (float64, error) {
	if vlc == nil {
		return 0, fmt.Errorf("vlc == nil")
	}
	if vlc.Client == nil {
		return 0, fmt.Errorf("vlc.Client == nil")
	}
	return vlc.Client.GetSpeed(ctx)
}

func (vlc *VLC) SetSpeed(
	ctx context.Context,
	speed float64,
) error {
	if vlc == nil {
		return fmt.Errorf("vlc == nil")
	}
	if vlc.Client == nil {
		return fmt.Errorf("vlc.Client == nil")
	}
	return vlc.Client.SetSpeed(ctx, speed)
}

func (vlc *VLC) GetPause(
	ctx context.Context,
) (bool, error) {
	if vlc == nil {
		return false, fmt.Errorf("vlc == nil")
	}
	if vlc.Client == nil {
		return false, fmt.Errorf("vlc.Client == nil")
	}
	return vlc.Client.GetPause(ctx)
}

func (vlc *VLC) SetPause(
	ctx context.Context,
	pause bool,
) error {
	if vlc == nil {
		return fmt.Errorf("vlc == nil")
	}
	if vlc.Client == nil {
		return fmt.Errorf("vlc.Client == nil")
	}
	return vlc.Client.SetPause(ctx, pause)
}

func (vlc *VLC) Seek(
	ctx context.Context,
	pos time.Duration,
	isRelative bool,
	quick bool,
) error {
	if vlc == nil {
		return fmt.Errorf("vlc == nil")
	}
	if vlc.Client == nil {
		return fmt.Errorf("vlc.Client == nil")
	}
	return vlc.Client.Seek(ctx, pos, isRelative, quick)
}

func (vlc *VLC) GetVideoTracks(
	ctx context.Context,
) (types.VideoTracks, error) {
	if vlc == nil {
		return nil, fmt.Errorf("vlc == nil")
	}
	if vlc.Client == nil {
		return nil, fmt.Errorf("vlc.Client == nil")
	}
	return vlc.Client.GetVideoTracks(ctx)
}

func (vlc *VLC) GetAudioTracks(
	ctx context.Context,
) (types.AudioTracks, error) {
	if vlc == nil {
		return nil, fmt.Errorf("vlc == nil")
	}
	if vlc.Client == nil {
		return nil, fmt.Errorf("vlc.Client == nil")
	}
	return vlc.Client.GetAudioTracks(ctx)
}

func (vlc *VLC) GetSubtitlesTracks(
	ctx context.Context,
) (types.SubtitlesTracks, error) {
	if vlc == nil {
		return nil, fmt.Errorf("vlc == nil")
	}
	if vlc.Client == nil {
		return nil, fmt.Errorf("vlc.Client == nil")
	}
	return vlc.Client.GetSubtitlesTracks(ctx)
}

func (vlc *VLC) SetVideoTrack(
	ctx context.Context,
	vid int64,
) error {
	if vlc == nil {
		return fmt.Errorf("vlc == nil")
	}
	if vlc.Client == nil {
		return fmt.Errorf("vlc.Client == nil")
	}
	return vlc.Client.SetVideoTrack(ctx, vid)
}

func (vlc *VLC) SetAudioTrack(
	ctx context.Context,
	aid int64,
) error {
	if vlc == nil {
		return fmt.Errorf("vlc == nil")
	}
	if vlc.Client == nil {
		return fmt.Errorf("vlc.Client == nil")
	}
	return vlc.Client.SetAudioTrack(ctx, aid)
}

func (vlc *VLC) SetSubtitlesTrack(
	ctx context.Context,
	sid int64,
) error {
	if vlc == nil {
		return fmt.Errorf("vlc == nil")
	}
	if vlc.Client == nil {
		return fmt.Errorf("vlc.Client == nil")
	}
	return vlc.Client.SetSubtitlesTrack(ctx, sid)
}

func (vlc *VLC) Stop(
	ctx context.Context,
) error {
	if vlc == nil {
		return fmt.Errorf("vlc == nil")
	}
	if vlc.Client == nil {
		return fmt.Errorf("vlc.Client == nil")
	}
	return vlc.Client.Stop(ctx)
}

func (vlc *VLC) Close(
	ctx context.Context,
) error {
	if vlc == nil {
		return fmt.Errorf("vlc == nil")
	}
	if vlc.Cmd != nil {
		defer vlc.Cmd.Process.Kill()
	}
	if vlc.Client == nil {
		return fmt.Errorf("vlc.Client == nil")
	}
	return vlc.Client.Close(ctx)
}
