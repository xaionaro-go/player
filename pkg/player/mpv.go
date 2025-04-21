package player

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	child_process_manager "github.com/AgustinSRG/go-child-process-manager"
	"github.com/blang/mpv"
	"github.com/davecgh/go-spew/spew"
	"github.com/dexterlb/mpvipc"
	"github.com/facebookincubator/go-belt/tool/experimental/errmon"
	"github.com/facebookincubator/go-belt/tool/logger"
	"github.com/xaionaro-go/logwriter"
	"github.com/xaionaro-go/observability"
	"github.com/xaionaro-go/player/pkg/player/types"
	"github.com/xaionaro-go/xpath"
	"github.com/xaionaro-go/xsync"
)

const SupportedMPV = true

const (
	restartMPV = true
)

var mpvCount uint64

type MPV struct {
	PlayerCommon
	PathToMPV  string
	SocketPath string
	Cmd        *exec.Cmd
	IPCClient  *mpv.IPCClient
	MPVClient  *mpv.Client
	MPVConn    *mpvipc.Connection
	CancelFunc context.CancelFunc
	isClosed   bool

	EndChInitialized bool
	EndChMutex       xsync.Mutex
	EndCh            chan struct{}

	OpenLinkOnRerun string
}

var _ Player = (*MPV)(nil)

func (m *Manager) NewMPV(
	ctx context.Context,
	title string,
	opts ...types.Option,
) (*MPV, error) {
	cfg := types.Options(opts).Config()
	r, err := NewMPV(ctx, title, cfg.PathToMPV, cfg.LowLatency, cfg.CacheLength, cfg.CacheMaxSize)
	if err != nil {
		return nil, err
	}

	m.PlayersLocker.Do(ctx, func() {
		m.Players = append(m.Players, r)
	})
	return r, nil
}

func NewMPV(
	ctx context.Context,
	title string,
	pathToMPV string,
	lowLatency bool,
	cacheDuration *time.Duration,
	cacheMaxSize uint64,
) (_ret *MPV, _err error) {
	logger.Debugf(ctx, "NewMPV()")
	defer func() { logger.Debugf(ctx, "/NewMPV(): %#+v %v", spew.Sdump(_ret), _err) }()

	if pathToMPV == "" {
		pathToMPV = "mpv"
		switch runtime.GOOS {
		case "windows":
			pathToMPV += ".exe"
		}
	}

	execPathToMPV, err := xpath.GetExecPath(pathToMPV, "mpv")
	if err != nil {
		return nil, fmt.Errorf("unable to locate the executable of MPV: '%s': %w", pathToMPV, err)
	}

	ctx, cancelFn := context.WithCancel(ctx)
	p := &MPV{
		PlayerCommon: PlayerCommon{
			Title:         title,
			LowLatency:    lowLatency,
			CacheDuration: cacheDuration,
			CacheMaxSize:  cacheMaxSize,
		},
		PathToMPV:  execPathToMPV,
		EndCh:      make(chan struct{}),
		CancelFunc: cancelFn,
	}
	err = p.execMPV(ctx)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (p *MPV) execMPV(ctx context.Context) (_err error) {
	logger.Debugf(ctx, "execMPV()")
	defer func() { logger.Debugf(ctx, "/execMPV(): %v", _err) }()

	myPid := os.Getpid()
	mpvID := atomic.AddUint64(&mpvCount, 1)
	var socketPath string
	switch runtime.GOOS {
	case "windows":
		socketPath = `\\.\pipe\` + fmt.Sprintf("mpv-ipc-%d-%d", myPid, mpvID)
	default:
		tempDir := os.TempDir()
		socketPath = path.Join(tempDir, fmt.Sprintf("mpv-ipc-%d-%d.sock", myPid, mpvID))
	}
	logger.Debugf(ctx, "socket path: '%s'", socketPath)

	err := os.Remove(socketPath)
	logger.Debugf(ctx, "socket deletion result: '%s': %v", socketPath, err)

	args := []string{
		p.PathToMPV,
		"--idle",
		"--keep-open=always",
		"--keep-open-pause=no",
		"--no-hidpi-window-scale",
		"--no-osc",
		"--no-osd-bar",
		"--window-scale=1",
		"--input-ipc-server=" + socketPath,
		fmt.Sprintf("--title=%s", p.Title),
	}
	if p.LowLatency {
		args = append(args,
			"--profile=low-latency",
		)
	}
	if p.CacheDuration != nil {
		if *p.CacheDuration == 0 {
			args = append(args, "--cache=no")
		} else {
			args = append(args, fmt.Sprintf("--cache-secs=%v", p.CacheDuration.Seconds()))
		}
	}
	if p.CacheMaxSize > 0 {
		args = append(args,
			fmt.Sprintf("--demuxer-max-bytes=%d", p.CacheMaxSize),
		)
	}
	switch observability.LogLevelFilter.GetLevel() {
	case logger.LevelPanic, logger.LevelFatal:
		args = append(args, "--msg-level=all=no")
	case logger.LevelError:
		args = append(args, "--msg-level=all=error")
	case logger.LevelWarning:
		args = append(args, "--msg-level=all=warn")
	case logger.LevelInfo:
		args = append(args, "--msg-level=all=info")
	case logger.LevelDebug, logger.LevelTrace:
		args = append(args, "--msg-level=all=debug")
	}
	logger.Debugf(ctx, "running command '%s %s'", args[0], strings.Join(args[1:], " "))
	cmd := exec.Command(args[0], args[1:]...)

	cmd.Stdout = logwriter.NewLogWriter(
		ctx,
		logger.FromCtx(ctx).
			WithField("log_writer_target", "mpv").
			WithField("output_type", "stdout"),
		logger.LevelTrace,
	)
	cmd.Stderr = logwriter.NewLogWriter(
		ctx,
		logger.FromCtx(ctx).
			WithField("log_writer_target", "mpv").
			WithField("output_type", "stderr"),
		logger.LevelTrace,
	)
	err = child_process_manager.ConfigureCommand(cmd)
	errmon.ObserveErrorCtx(ctx, err)
	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("unable to start mpv: %w", err)
	}
	err = child_process_manager.AddChildProcess(cmd.Process)
	if err != nil {
		if runtime.GOOS == "windows" {
			// this is actually an error, but I have no idea how to fix it, so demoting to a debug message
			logger.Debugf(ctx, "unable to register the command %v to be auto-killed: %v", args, err)
		} else {
			logger.Errorf(ctx, "unable to register the command %v to be auto-killed: %v", args, err)
		}
	}
	logger.Debugf(ctx, "started command '%s %s'", args[0], strings.Join(args[1:], " "))

	logger.Debugf(ctx, "waiting for the socket '%s' to get ready", socketPath)

	mpvConn := mpvipc.NewConnection(socketPath)
	t := time.NewTicker(100 * time.Millisecond)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.C:
		}
		if cmd.ProcessState != nil {
			logger.Errorf(ctx, "mpv unexpectedly exited: exitcode: %d", cmd.ProcessState.ExitCode())
		}
		err := mpvConn.Open()
		if err == nil {
			break
		}
		logger.Tracef(ctx, "mpvConn.Open() err: %v", err)
	}
	logger.Debugf(ctx, "socket '%s' is ready", socketPath)
	p.SocketPath = socketPath
	p.Cmd = cmd
	p.MPVConn = mpvConn

	if restartMPV {
		observability.Go(ctx, func() {
			err := p.Cmd.Wait()
			logger.Debugf(ctx, "player was closed: %v", err)
			link := p.OpenLinkOnRerun
			if link == "" {
				logger.Debugf(ctx, "not going to open any links")
			} else {
				logger.Debugf(ctx, "going to open link '%s'", link)
			}
			err = p.cleanup(ctx)
			logger.Debugf(ctx, "cleanup result: %v", err)
			select {
			case <-ctx.Done():
				logger.Debugf(ctx, "context is closed, not rerunning the player")
				return
			default:
			}
			logger.Debugf(ctx, "rerunning the player")
			err = p.execMPV(ctx)
			if err != nil {
				logger.Error(ctx, "unable to rerun the player: %v", err)
			}
			logger.Debugf(ctx, "successfully reran the player")
			if link != "" {
				logger.Debugf(ctx, "reopen link '%s'", link)
				err := p.OpenURL(ctx, link)
				if err != nil {
					logger.Errorf(ctx, "unable to reopen link '%v'", err)
				}
			}
		})
	}
	return nil
}

func (p *MPV) SetupForStreaming(
	ctx context.Context,
) (_err error) {
	logger.Debugf(ctx, "SetupForStreaming()")
	defer func() { logger.Debugf(ctx, "/SetupForStreaming(): %v", _err) }()

	return p.SetDisplayScale(ctx, 1)
}

func (p *MPV) OpenURL(
	ctx context.Context,
	link string,
) error {
	logger.Debugf(ctx, "OpenURL(ctx, '%s')", link)
	p.OpenLinkOnRerun = link
	_, err := p.mpvCall(ctx, "loadfile", link, "replace")
	return err
}

func (p *MPV) getString(
	ctx context.Context,
	key string,
) (string, error) {
	r, err := p.mpvGet(ctx, key)
	if err != nil {
		return "", fmt.Errorf("unable to get '%s' from the MPV: %w", key, err)
	}
	s, ok := r.(string)
	if !ok {
		s = fmt.Sprint(r)
	}
	return s, nil
}

const requestTimeout = time.Second

func (p *MPV) timeboxedCall(
	ctx context.Context,
	fn func() error,
) error {
	ctx, cancelFn := context.WithTimeout(ctx, requestTimeout)
	defer cancelFn()

	endedCh := make(chan struct{})

	var err error
	observability.Go(ctx, func() {
		err = fn()
		close(endedCh)
	})

	select {
	case <-ctx.Done():
		logger.Errorf(ctx, "timed out on a request")
		return ctx.Err()
	case <-endedCh:
	}

	return err
}

func (p *MPV) mpvSet(
	ctx context.Context,
	key string,
	value any,
) (_err error) {
	logger.Debugf(ctx, "mpvSet(ctx, '%s', %v)", key, value)
	defer func() { logger.Debugf(ctx, "/mpvSet(ctx, '%s', %v): %v", key, value, _err) }()
	return p.timeboxedCall(ctx, func() error {
		return p.MPVConn.Set(key, value)
	})
}

func (p *MPV) mpvGet(
	ctx context.Context,
	key string,
) (_ret any, _err error) {
	logger.Tracef(ctx, "mpvGet(ctx, '%s')", key)
	defer func() { logger.Tracef(ctx, "/mpvGet(ctx, '%s'): %v %v", key, _ret, _err) }()
	var result any
	err := p.timeboxedCall(ctx, func() error {
		var err error
		result, err = p.MPVConn.Get(key)
		return err
	})
	return result, err
}

func (p *MPV) mpvCall(
	ctx context.Context,
	args ...any,
) (_ret any, _err error) {
	logger.Debugf(ctx, "mpvCall(ctx, %v)", args)
	defer func() { logger.Debugf(ctx, "/mpvCall(ctx, %v): %v %v", args, _ret, _err) }()
	var result any
	err := p.timeboxedCall(ctx, func() error {
		var err error
		result, err = p.MPVConn.Call(args...)
		return err
	})
	return result, err
}

func (p *MPV) getFloat64(
	ctx context.Context,
	key string,
) (float64, error) {
	r, err := p.mpvGet(ctx, key)
	if err != nil {
		return 0, fmt.Errorf("unable to get '%s' from the MPV: %w", key, err)
	}
	switch r := r.(type) {
	case float64:
		return r, nil
	case string:
		return strconv.ParseFloat(r, 64)
	default:
		return 0, fmt.Errorf("unexpected type %T", r)
	}
}

func (p *MPV) getBool(
	ctx context.Context,
	key string,
) (bool, error) {
	r, err := p.mpvGet(ctx, key)
	if err != nil {
		return false, fmt.Errorf("unable to get '%s' from the MPV: %w", key, err)
	}
	switch r := r.(type) {
	case bool:
		return r, nil
	case string:
		return strconv.ParseBool(r)
	default:
		return false, fmt.Errorf("unexpected type %T", r)
	}
}

func (p *MPV) GetLink(
	ctx context.Context,
) (string, error) {
	return p.getString(ctx, "filename")
}

func (p *MPV) EndChan(
	ctx context.Context,
) (<-chan struct{}, error) {
	return xsync.DoR2(ctx, &p.EndChMutex, func() (<-chan struct{}, error) {
		p.initEndCh(ctx)
		return p.EndCh, nil
	})
}

func (p *MPV) initEndCh(
	ctx context.Context,
) {
	if p.EndChInitialized {
		return
	}
	observability.Go(ctx, func() {
		func() {
			t := time.NewTimer(time.Millisecond * 100)
			defer t.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-t.C:
				}
				isEnded, _ := p.IsEnded(ctx)
				if !isEnded {
					return
				}
			}
		}()
		p.EndChMutex.Do(ctx, func() {
			var oldCh chan struct{}
			oldCh, p.EndCh = p.EndCh, make(chan struct{})
			close(oldCh)
			p.EndChInitialized = false
		})
	})
}

func (p *MPV) IsEnded(
	ctx context.Context,
) (bool, error) {
	link, err := p.GetLink(ctx)
	if err != nil {
		return false, nil
	}
	return link != "", nil
}

func (p *MPV) GetPosition(
	ctx context.Context,
) (time.Duration, error) {
	ts, err := p.getFloat64(ctx, "time-pos")
	if err != nil {
		return 0, err
	}

	return time.Duration(ts * float64(time.Second)), nil
}

func (p *MPV) GetLength(
	ctx context.Context,
) (time.Duration, error) {
	ts, err := p.getFloat64(ctx, "duration")
	if err != nil {
		return 0, err
	}

	return time.Duration(ts * float64(time.Second)), nil
}

func (p *MPV) GetCachedDuration(
	ctx context.Context,
) (time.Duration, error) {
	dur, err := p.getFloat64(ctx, "demuxer-cache-duration")
	if err != nil {
		return 0, err
	}

	return time.Duration(dur * float64(time.Second)), nil
}

func (p *MPV) Seek(
	ctx context.Context,
	pos time.Duration,
	isRelative bool,
	quick bool,
) (_err error) {
	logger.Tracef(ctx, "Seek(ctx, %v, %t, %t)", pos, isRelative, quick)
	defer func() { logger.Tracef(ctx, "/Seek(ctx, %v, %t, %t): %v", pos, isRelative, quick, _err) }()
	var flags []string
	if isRelative {
		flags = append(flags, "relative")
	} else {
		flags = append(flags, "absolute")
	}
	if quick {
		flags = append(flags, "keyframes")
	} else {
		flags = append(flags, "exact")
	}
	args := []any{"seek", pos.Seconds()}
	if len(flags) > 0 {
		args = append(args, strings.Join(flags, "+"))
	}
	_, err := p.mpvCall(ctx, args...)
	if err != nil {
		return fmt.Errorf("unable to request 'seek'-ing: %w", err)
	}
	return nil
}

func (p *MPV) SetSpeed(
	ctx context.Context,
	speed float64,
) error {
	return p.mpvSet(ctx, "speed", speed)
}

func (p *MPV) GetSpeed(
	ctx context.Context,
) (float64, error) {
	return p.getFloat64(ctx, "speed")
}

func (p *MPV) GetPause(
	ctx context.Context,
) (bool, error) {
	return p.getBool(ctx, "pause")
}

func (p *MPV) SetPause(
	ctx context.Context,
	pause bool,
) error {
	return p.mpvSet(ctx, "pause", pause)
}

func (p *MPV) Stop(
	ctx context.Context,
) error {
	_, err := p.mpvCall(ctx, "stop")
	if err != nil {
		return fmt.Errorf("unable to request 'stop'-ing: %w", err)
	}
	return nil
}

func (p *MPV) Quit(ctx context.Context, exitCode uint8) error {
	_, err := p.mpvCall(ctx, "quit", exitCode)
	if err != nil {
		return fmt.Errorf("unable to request 'quit'-ing: %w", err)
	}
	return nil
}

func (p *MPV) GetDisplayScale(ctx context.Context) (float64, error) {
	scale, err := p.getFloat64(ctx, "window-scale")
	if err != nil {
		return 0, err
	}

	return scale, nil
}

func (p *MPV) SetDisplayScale(ctx context.Context, scale float64) error {
	return p.mpvSet(ctx, "window-scale", scale)
}

func getTracks[E any, T []E](
	ctx context.Context,
	p *MPV,
	trackType string,
	fn func(trackID int64, isActive bool) E,
) (T, error) {
	resp, err := p.mpvGet(ctx, "track-list")
	if err != nil {
		return nil, fmt.Errorf("unable to get the %s track list: %w", trackType, exec.ErrDot)
	}
	list, ok := resp.([]any)
	if !ok {
		return nil, fmt.Errorf("expected a slice of values, but received %T", resp)
	}

	result := make(T, 0, len(list))
	for idx, item := range list {
		logger.Tracef(ctx, "%#+v", item)
		m, ok := item.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("item #%d is expected to be a map[string]any, but received %T", idx, item)
		}

		typeI, ok := m["type"]
		if !ok {
			return nil, fmt.Errorf("item #%d does not has field 'type'", idx)
		}
		typ, ok := typeI.(string)
		if !ok {
			return nil, fmt.Errorf("item #%d has field 'type' of an unexpected type: %T", idx, typeI)
		}

		if typ != trackType {
			continue
		}

		trackIDI, ok := m["id"]
		if !ok {
			return nil, fmt.Errorf("item #%d does not has field 'id'", idx)
		}
		trackID, ok := trackIDI.(float64)
		if !ok {
			return nil, fmt.Errorf("item #%d has field 'id' of an unexpected type: %T", idx, trackIDI)
		}

		selectedI, ok := m["selected"]
		if !ok {
			return nil, fmt.Errorf("item #%d does not has field 'selected'", idx)
		}
		selected, ok := selectedI.(bool)
		if !ok {
			return nil, fmt.Errorf("item #%d has field 'selected' of an unexpected type: %T", idx, selectedI)
		}

		result = append(result, fn(int64(trackID), selected))
	}

	return result, nil
}

func (p *MPV) GetVideoTracks(
	ctx context.Context,
) (types.VideoTracks, error) {
	return getTracks(ctx, p, "video", func(trackID int64, isActive bool) types.VideoTrack {
		return types.VideoTrack{
			ID:       trackID,
			IsActive: isActive,
		}
	})
}

func (p *MPV) GetAudioTracks(
	ctx context.Context,
) (types.AudioTracks, error) {
	return getTracks(ctx, p, "audio", func(trackID int64, isActive bool) types.AudioTrack {
		return types.AudioTrack{
			ID:       trackID,
			IsActive: isActive,
		}
	})
}

func (p *MPV) GetSubtitlesTracks(
	ctx context.Context,
) (types.SubtitlesTracks, error) {
	return getTracks[types.SubtitlesTrack](ctx, p, "sub", func(trackID int64, isActive bool) types.SubtitlesTrack {
		return types.SubtitlesTrack{
			ID:       trackID,
			IsActive: isActive,
		}
	})
}

func (p *MPV) SetVideoTrack(ctx context.Context, vid int64) error {
	return p.mpvSet(ctx, "vid", vid)
}

func (p *MPV) SetAudioTrack(ctx context.Context, aid int64) error {
	return p.mpvSet(ctx, "aid", aid)
}

func (p *MPV) SetSubtitlesTrack(ctx context.Context, sid int64) error {
	return p.mpvSet(ctx, "sid", sid)
}

const mpvQuitTimeout = time.Second

func (p *MPV) Close(ctx context.Context) (_err error) {
	logger.Debugf(ctx, "Close()")
	defer func() { logger.Debugf(ctx, "/Close(): %v", _err) }()

	if p.isClosed {
		return nil
	}
	p.isClosed = true
	p.CancelFunc()
	p.OpenLinkOnRerun = ""
	return p.cleanup(ctx)
}

func (p *MPV) cleanup(ctx context.Context) (_err error) {
	if p.MPVConn.IsClosed() {
		if err := p.Cmd.Process.Kill(); err != nil {
			logger.Debugf(ctx, "unable to kill the process: %v", err)
		}
		if err := os.Remove(p.SocketPath); err != nil {
			logger.Tracef(ctx, "unable to remove the socket file: %v", err)
		}
		return
	}

	if err := p.Quit(ctx, 0); err != nil {
		logger.Errorf(ctx, "unable to request the player to quit: %v", err)
	}
	quitCtx, quittedFn := context.WithCancel(ctx)
	go func() {
		p.Cmd.Process.Wait()
		quittedFn()
	}()
	select {
	case <-time.After(mpvQuitTimeout):
		logger.Warnf(ctx, "timed out on waiting until MPV would die, so killing it forcefully")
		if err := p.Cmd.Process.Kill(); err != nil {
			logger.Errorf(ctx, "unable to kill the process: %v", err)
		}
	case <-quitCtx.Done():
		logger.Debugf(ctx, "the process successfully quitted")
	}
	if err := p.MPVConn.Close(); err != nil {
		logger.Errorf(ctx, "unable to close old socket: %v", err)
	}
	if err := os.Remove(p.SocketPath); err != nil {
		logger.Tracef(ctx, "unable to remove the socket file: %v", err)
	}
	return nil
}
