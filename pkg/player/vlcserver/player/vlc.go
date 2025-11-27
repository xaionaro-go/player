//go:build with_libvlc
// +build with_libvlc

package player

import (
	"context"
	"fmt"
	"net/url"
	"sync/atomic"
	"time"

	vlc "github.com/adrg/libvlc-go/v3"
	"github.com/hashicorp/go-multierror"
	"github.com/xaionaro-go/player/pkg/player/types"
	"github.com/xaionaro-go/xsync"
)

type VLC struct {
	Title            string
	StatusMutex      xsync.Mutex
	Player           *vlc.Player
	Media            *vlc.Media
	EventManager     *vlc.EventManager
	DetachEventsFunc context.CancelFunc
	LastURL          string

	IsStopped bool

	EndCh chan struct{}
}

var vlcPlayerCounter int64 = 0

func NewVLC(title string) (*VLC, error) {
	if atomic.AddInt64(&vlcPlayerCounter, 1) != 1 {
		return nil, fmt.Errorf("currently we do not support more than one VLC player at once")
	}
	args := []string{fmt.Sprintf("--video-title=%s", title)}
	if err := vlc.Init(args...); err != nil {
		return nil, fmt.Errorf("unable to initialize VLC with arguments: %v", args)
	}

	p := &VLC{
		Title: title,
		EndCh: make(chan struct{}),
	}

	var err error
	p.Player, err = vlc.NewPlayer()
	if err != nil {
		return nil, fmt.Errorf("unable to initialize a VLC player: %w", err)
	}

	manager, err := p.Player.EventManager()
	if err != nil {
		return nil, fmt.Errorf("unable to initialize a VLC event manager: %w", err)
	}

	eventID, err := manager.Attach(vlc.MediaPlayerEndReached, func(e vlc.Event, i interface{}) {
		p.StatusMutex.Do(context.TODO(), func() {
			p.IsStopped = true
			var oldCh chan struct{}
			oldCh, p.EndCh = p.EndCh, make(chan struct{})
			close(oldCh)
		})
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to attach the 'EndReached' event handler: %w", err)
	}
	p.DetachEventsFunc = func() {
		manager.Detach(eventID)
	}

	return p, nil
}

func (p *VLC) ProcessTitle(
	ctx context.Context,
) (string, error) {
	return p.Title, nil
}

func (p *VLC) OpenURL(
	ctx context.Context,
	link string,
) error {
	return xsync.DoA1R1(ctx, &p.StatusMutex, p.openURL, link)
}

func (p *VLC) openURL(link string) error {
	if p.Media != nil {
		return fmt.Errorf("some media is already opened in this player")
	}

	var (
		media *vlc.Media
		err   error
	)
	if urlParsed, _err := url.Parse(link); _err == nil && urlParsed.Scheme != "" {
		media, err = p.Player.LoadMediaFromURL(link)
	} else {
		media, err = p.Player.LoadMediaFromPath(link)
	}
	if err != nil {
		return fmt.Errorf("unable to open '%s': %w", link, err)
	}
	p.Media = media
	p.LastURL = link

	if err := p.play(); err != nil {
		return fmt.Errorf("opened, but unable to start playing '%s': %w", link, err)
	}

	return nil
}

func (p *VLC) GetLink(
	ctx context.Context,
) (string, error) {
	return xsync.DoR1(ctx, &p.StatusMutex, func() string {
		return p.LastURL
	}), nil
}

func (p *VLC) EndChan(
	ctx context.Context,
) (<-chan struct{}, error) {
	return xsync.DoR1(ctx, &p.StatusMutex, func() <-chan struct{} {
		return p.EndCh
	}), nil
}

func (p *VLC) IsEnded(
	ctx context.Context,
) (bool, error) {
	return xsync.DoR1(ctx, &p.StatusMutex, func() bool {
		return p.IsStopped
	}), nil
}

func (p *VLC) GetPosition(
	ctx context.Context,
) (time.Duration, error) {
	ts, err := p.Player.MediaTime()
	if err != nil {
		return 0, fmt.Errorf("unable to get current position: %w", err)
	}
	return time.Duration(ts) * time.Millisecond, nil
}

func (p *VLC) GetAudioPosition(
	ctx context.Context,
) (time.Duration, error) {
	return p.GetPosition(ctx)
}

func (p *VLC) GetLength(
	ctx context.Context,
) (time.Duration, error) {
	ts, err := p.Player.MediaLength()
	if err != nil {
		return 0, fmt.Errorf("unable to get the total length: %w", err)
	}
	return time.Duration(ts) * time.Millisecond, nil
}

func (p *VLC) GetSpeed(
	ctx context.Context,
) (float64, error) {
	return float64(p.Player.PlaybackRate()), nil
}

func (p *VLC) SetSpeed(
	ctx context.Context,
	speed float64,
) error {
	return p.Player.SetPlaybackRate(float32(speed))
}

func (p *VLC) Play(
	ctx context.Context,
) error {
	return xsync.DoR1(ctx, &p.StatusMutex, p.play)
}

func (p *VLC) play() error {
	err := p.Player.Play()
	if err != nil {
		return err
	}
	p.IsStopped = false
	return nil
}

func (p *VLC) GetPause(
	ctx context.Context,
) (bool, error) {
	return !p.Player.IsPlaying(), nil
}

func (p *VLC) SetPause(
	ctx context.Context,
	pause bool,
) error {
	return p.Player.SetPause(pause)
}

func (p *VLC) Seek(
	ctx context.Context,
	pos time.Duration,
	isRelative bool,
	quick bool,
) error {
	return fmt.Errorf("not implemented, yet")
}

func (p *VLC) GetVideoTracks(
	ctx context.Context,
) (types.VideoTracks, error) {
	activeTrackID, err := p.Player.VideoTrackID()
	if err != nil {
		return nil, fmt.Errorf("unable to get the active track ID: %w", err)
	}

	trackDescrs, err := p.Player.VideoTrackDescriptors()
	if err != nil {
		return nil, fmt.Errorf("unable to get video track descriptors: %w", err)
	}

	result := make(types.VideoTracks, 0, len(trackDescrs))
	for _, track := range trackDescrs {
		result = append(result, types.VideoTrack{
			ID:       int64(track.ID),
			IsActive: track.ID == activeTrackID,
		})
	}
	return result, nil
}

func (p *VLC) GetAudioTracks(
	ctx context.Context,
) (types.AudioTracks, error) {
	activeTrackID, err := p.Player.AudioTrackID()
	if err != nil {
		return nil, fmt.Errorf("unable to get the active track ID: %w", err)
	}

	trackDescrs, err := p.Player.AudioTrackDescriptors()
	if err != nil {
		return nil, fmt.Errorf("unable to get video track descriptors: %w", err)
	}

	result := make(types.AudioTracks, 0, len(trackDescrs))
	for _, track := range trackDescrs {
		result = append(result, types.AudioTrack{
			ID:       int64(track.ID),
			IsActive: track.ID == activeTrackID,
		})
	}
	return result, nil
}

func (p *VLC) GetSubtitlesTracks(
	ctx context.Context,
) (types.SubtitlesTracks, error) {
	activeTrackID, err := p.Player.SubtitleTrackID()
	if err != nil {
		return nil, fmt.Errorf("unable to get the active track ID: %w", err)
	}

	trackDescrs, err := p.Player.SubtitleTrackDescriptors()
	if err != nil {
		return nil, fmt.Errorf("unable to get video track descriptors: %w", err)
	}

	result := make(types.SubtitlesTracks, 0, len(trackDescrs))
	for _, track := range trackDescrs {
		result = append(result, types.SubtitlesTrack{
			ID:       int64(track.ID),
			IsActive: track.ID == activeTrackID,
		})
	}
	return result, nil
}

func (p *VLC) SetVideoTrack(
	ctx context.Context,
	vid int64,
) error {
	return p.Player.SetVideoTrack(int(vid))
}

func (p *VLC) SetAudioTrack(
	ctx context.Context,
	aid int64,
) error {
	return p.Player.SetAudioTrack(int(aid))
}

func (p *VLC) SetSubtitlesTrack(
	ctx context.Context,
	sid int64,
) error {
	return p.Player.SetSubtitleTrack(int(sid))
}

func (p *VLC) Stop(
	ctx context.Context,
) error {
	return xsync.DoR1(ctx, &p.StatusMutex, p.stop)
}

func (p *VLC) stop() error {
	err := p.Player.Stop()
	if err != nil {
		return err
	}
	p.IsStopped = false
	return nil
}

func (p *VLC) Close(ctx context.Context) error {
	return xsync.DoA1R1(ctx, &p.StatusMutex, p.close, ctx)
}

func (p *VLC) close(
	ctx context.Context,
) error {
	if p.DetachEventsFunc != nil {
		p.DetachEventsFunc()
		p.DetachEventsFunc = nil
	}

	err := multierror.Append(
		p.Player.Stop(),
		p.Media.Release(),
		p.Player.Release(),
		vlc.Release(),
	).ErrorOrNil()
	atomic.AddInt64(&vlcPlayerCounter, -1)
	return err
}

func (p *VLC) SetupForStreaming(
	ctx context.Context,
) error {
	return nil
}
