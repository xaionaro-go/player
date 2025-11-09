package main

import (
	"context"
	"net/http"
	_ "net/http/pprof"
	"strings"

	child_process_manager "github.com/AgustinSRG/go-child-process-manager"
	"github.com/facebookincubator/go-belt"
	"github.com/facebookincubator/go-belt/tool/logger"
	"github.com/facebookincubator/go-belt/tool/logger/implementation/logrus"
	"github.com/spf13/pflag"

	"github.com/xaionaro-go/observability"
	"github.com/xaionaro-go/player/pkg/player"
	"github.com/xaionaro-go/player/pkg/player/types"
	"github.com/xaionaro-go/xsync"

	_ "github.com/xaionaro-go/audio/pkg/audio/backends/oto"
	//_ "github.com/xaionaro-go/audio/pkg/audio/backends/pulseaudio"
)

func backendsToStrings(backends []player.Backend) []string {
	result := make([]string, 0, len(backends))
	for _, s := range backends {
		result = append(result, string(s))
	}
	return result
}

func assertNoError(ctx context.Context, err error) {
	if err == nil {
		return
	}
	logger.Fatal(ctx, err)
}

func main() {
	backends := backendsToStrings(player.SupportedBackends())
	loggerLevel := logger.LevelInfo
	pflag.Var(&loggerLevel, "log-level", "Log level")
	mpvPath := pflag.String("mpv", "mpv", "path to mpv")
	backend := pflag.String("backend", backends[0], "player backend, supported values: "+strings.Join(backends, ", "))
	netPprofAddr := pflag.String("net-pprof-listen-addr", "", "an address to listen for incoming net/pprof connections")
	lowLatency := pflag.Bool("low-latency", false, "")
	cacheLength := pflag.Duration("cache-duration", 0, "")
	cacheMaxSize := pflag.Uint("cache-max-size", 0, "")
	pflag.Parse()

	l := logrus.Default().WithLevel(loggerLevel)
	ctx := xsync.WithNoLogging(logger.CtxWithLogger(context.Background(), l), true)
	logger.Default = func() logger.Logger {
		return l
	}
	defer belt.Flush(ctx)

	if pflag.NArg() != 1 {
		l.Fatal("exactly one argument expected")
	}
	mediaPath := pflag.Arg(0)

	if *netPprofAddr != "" {
		observability.Go(ctx, func(
			context.Context) {
			l.Error(http.ListenAndServe(*netPprofAddr, nil))
		})
	}

	err := child_process_manager.InitializeChildProcessManager()
	if err != nil {
		logger.Fatal(ctx, err)
	}
	defer child_process_manager.DisposeChildProcessManager()

	opts := types.Options{types.OptionPathToMPV(*mpvPath)}
	if *lowLatency {
		opts = append(opts, types.OptionPreset(types.PresetLowestLatency))
	}
	if *cacheLength > 0 {
		opts = append(opts, types.OptionCacheDuration(*cacheLength))
	}
	if *cacheMaxSize > 0 {
		opts = append(opts, types.OptionCacheMaxSize(*cacheMaxSize))
	}

	m := player.NewManager(opts...)
	p, err := m.NewPlayer(ctx, "player demonstration", player.Backend(*backend))
	assertNoError(ctx, err)

	err = p.OpenURL(ctx, mediaPath)
	if err != nil {
		logger.Fatalf(ctx, "unable to open the url '%s': %v", mediaPath, err)
	}

	err = p.SetPause(ctx, false)
	if err != nil {
		logger.Errorf(ctx, "unable to start playback: %v", err)
	}

	runPlayerControls(ctx, p)
}
