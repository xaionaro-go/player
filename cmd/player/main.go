package main

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"strconv"
	"strings"
	"time"

	fyneapp "fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	child_process_manager "github.com/AgustinSRG/go-child-process-manager"
	"github.com/facebookincubator/go-belt"
	"github.com/facebookincubator/go-belt/tool/logger"
	"github.com/facebookincubator/go-belt/tool/logger/implementation/logrus"
	"github.com/spf13/pflag"

	"github.com/xaionaro-go/observability"
	"github.com/xaionaro-go/player/pkg/player"
	"github.com/xaionaro-go/player/pkg/player/types"
	xfyne "github.com/xaionaro-go/xfyne/widget"
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
	app := fyneapp.New()

	opts := types.Options{types.OptionPathToMPV(*mpvPath)}
	if *lowLatency {
		opts = append(opts, types.OptionPreset(types.PresetLowLatency))
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

	observability.Go(ctx, func(ctx context.Context) {
		ch, err := p.EndChan(ctx)
		if err != nil {
			panic(err)
		}
		<-ch
		w := app.NewWindow("file ended")
		b := widget.NewButton("Close", func() {
			w.Close()
		})
		w.SetContent(container.NewStack(b))
		w.Show()
	})

	errorMessage := widget.NewLabel("")

	setSpeed := xfyne.NewNumericalEntry()
	setSpeed.SetText("1.0")
	setSpeed.OnSubmitted = func(s string) {
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			errorMessage.SetText(fmt.Sprintf("unable to parse speed '%s': %s", s, err))
			return
		}

		err = p.SetSpeed(ctx, f)
		if err != nil {
			errorMessage.SetText(fmt.Sprintf("unable to set speed to '%f': %s", f, err))
			return
		}
		errorMessage.SetText("")
	}

	videoTrack := xfyne.NewNumericalEntry()
	videoTrack.SetText("1")
	videoTrack.OnSubmitted = func(s string) {
		id, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			errorMessage.SetText(fmt.Sprintf("unable to parse video track ID '%s': %s", s, err))
			return
		}

		if tracks, err := p.GetVideoTracks(ctx); err == nil {
			logger.Debugf(ctx, "video tracks: %v", tracks)
			found := false
			for _, track := range tracks {
				if track.ID == id {
					found = true
					break
				}
			}
			if !found {
				errorMessage.SetText(fmt.Sprintf("there is no video track ID %d", id))
				return
			}
		} else {
			logger.Errorf(ctx, "unable to get the list of video tracks: %v", err)
		}

		p.SetVideoTrack(ctx, id)
	}

	audioTrack := xfyne.NewNumericalEntry()
	audioTrack.SetText("1")
	audioTrack.OnSubmitted = func(s string) {
		id, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			errorMessage.SetText(fmt.Sprintf("unable to parse audio track ID '%s': %s", s, err))
			return
		}

		if tracks, err := p.GetAudioTracks(ctx); err == nil {
			logger.Debugf(ctx, "audio tracks: %v", tracks)
			found := false
			for _, track := range tracks {
				if track.ID == id {
					found = true
					break
				}
			}
			if !found {
				errorMessage.SetText(fmt.Sprintf("there is no audio track ID %d", id))
				return
			}
		} else {
			logger.Errorf(ctx, "unable to get the list of audio tracks: %v", err)
		}

		p.SetAudioTrack(ctx, id)
	}

	isPaused := false
	p.SetPause(ctx, isPaused)
	var pauseUnpause *widget.Button
	pauseUnpause = widget.NewButtonWithIcon("Pause", theme.MediaPauseIcon(), func() {
		isPaused = !isPaused
		switch isPaused {
		case true:
			pauseUnpause.SetText("Unpause")
			pauseUnpause.SetIcon(theme.MediaPlayIcon())
		case false:
			pauseUnpause.SetText("Pause")
			pauseUnpause.SetIcon(theme.MediaPauseIcon())
		}
		err := p.SetPause(ctx, isPaused)
		if err != nil {
			errorMessage.SetText(fmt.Sprintf("unable to set pause to '%v': %s", isPaused, err))
			return
		}
		errorMessage.SetText("")
	})

	stopButton := widget.NewButtonWithIcon("Stop", theme.MediaStopIcon(), func() {
		p.Stop(ctx)
	})

	closeButton := widget.NewButtonWithIcon("Close", theme.WindowCloseIcon(), func() {
		p.Stop(ctx)
	})

	forwardButton := widget.NewButtonWithIcon("", theme.MediaFastForwardIcon(), func() {
		p.Seek(ctx, time.Second, true, false)
	})

	backwardButton := widget.NewButtonWithIcon("", theme.MediaFastRewindIcon(), func() {
		p.Seek(ctx, -time.Second, true, false)
	})

	forwardQuickButton := widget.NewButtonWithIcon("Q", theme.MediaFastForwardIcon(), func() {
		p.Seek(ctx, time.Second, true, true)
	})

	backwardQuickButton := widget.NewButtonWithIcon("Q", theme.MediaFastRewindIcon(), func() {
		p.Seek(ctx, -time.Second, true, true)
	})

	posLabel := widget.NewLabel("")
	observability.Go(ctx, func(ctx context.Context) {
		t := time.NewTicker(time.Millisecond * 100)
		for {
			<-t.C
			l, err := p.GetLength(ctx)
			if err != nil {
				l = -1
			}

			pos, err := p.GetPosition(ctx)
			if err != nil {
				posLabel.SetText(fmt.Sprintf("unable to get the position: %v", err))
			}

			posLabel.SetText(pos.String() + " / " + l.String())
		}
	})

	w := app.NewWindow("player controls")
	w.SetContent(container.NewBorder(
		posLabel,
		errorMessage,
		nil,
		nil,
		container.NewVBox(
			setSpeed,
			container.NewHBox(
				videoTrack,
				audioTrack,
			),
			container.NewHBox(
				backwardButton,
				forwardButton,
				backwardQuickButton,
				forwardQuickButton,
			),
			container.NewHBox(
				pauseUnpause,
				stopButton,
				closeButton,
			),
		),
	))
	w.Show()
	app.Run()
}
