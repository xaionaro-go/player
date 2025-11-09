//go:build with_fyne
// +build with_fyne

package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	fyneapp "fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/facebookincubator/go-belt/tool/logger"
	"github.com/xaionaro-go/observability"
	"github.com/xaionaro-go/player/pkg/player/types"
	xfyne "github.com/xaionaro-go/xfyne/widget"
)

func init() {
	fyneapp.New()
}

func runPlayerControls(
	ctx context.Context,
	p types.Player,
) {
	defer logger.Infof(ctx, "player controls ended")
	app := fyne.CurrentApp()

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

	logger.Infof(ctx, "player controls started")
	app.Run()
}
