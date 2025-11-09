package main

import (
	"context"
	"net/http"
	_ "net/http/pprof"

	child_process_manager "github.com/AgustinSRG/go-child-process-manager"
	"github.com/facebookincubator/go-belt"
	"github.com/facebookincubator/go-belt/tool/logger"
	"github.com/facebookincubator/go-belt/tool/logger/implementation/logrus"
	"github.com/spf13/pflag"

	"github.com/xaionaro-go/observability"
	"github.com/xaionaro-go/player/pkg/player/decoder/libav"
	"github.com/xaionaro-go/player/pkg/player/imagerenderer/libavv4l2"
	"github.com/xaionaro-go/xsync"

	"github.com/xaionaro-go/audio/pkg/audio"
	_ "github.com/xaionaro-go/audio/pkg/audio/backends/oto"
	//_ "github.com/xaionaro-go/audio/pkg/audio/backends/pulseaudio"
)

func assertNoError(ctx context.Context, err error) {
	if err == nil {
		return
	}
	logger.Fatal(ctx, err)
}

func main() {
	loggerLevel := logger.LevelInfo
	pflag.Var(&loggerLevel, "log-level", "Log level")
	netPprofAddr := pflag.String("net-pprof-listen-addr", "", "an address to listen for incoming net/pprof connections")
	pflag.Parse()

	l := logrus.Default().WithLevel(loggerLevel)
	ctx := xsync.WithNoLogging(logger.CtxWithLogger(context.Background(), l), true)
	logger.Default = func() logger.Logger {
		return l
	}
	defer belt.Flush(ctx)

	if pflag.NArg() != 2 {
		l.Fatal("exactly two arguments expected: <path-to-V4L2-device> <media-file>")
	}
	v4l2Device := pflag.Arg(0)
	mediaPath := pflag.Arg(1)

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

	imageRenderer, err := libavv4l2.NewImageRendererV4L2Output(ctx, v4l2Device)
	assertNoError(ctx, err)

	p := libav.New(ctx, imageRenderer, audio.NewPlayerAuto(ctx))

	err = p.OpenURL(ctx, mediaPath)
	if err != nil {
		logger.Fatalf(ctx, "unable to open the url '%s': %v", mediaPath, err)
	}

	ch, err := p.EndChan(ctx)
	assertNoError(ctx, err)
	<-ch
}
