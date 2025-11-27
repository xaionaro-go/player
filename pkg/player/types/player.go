package types

import (
	"context"
	"time"
)

type Player interface {
	ProcessTitle(ctx context.Context) (string, error)
	OpenURL(ctx context.Context, link string) error
	GetLink(ctx context.Context) (string, error)
	EndChan(ctx context.Context) (<-chan struct{}, error)
	IsEnded(ctx context.Context) (bool, error)
	GetPosition(ctx context.Context) (time.Duration, error)
	GetAudioPosition(ctx context.Context) (time.Duration, error)
	GetLength(ctx context.Context) (time.Duration, error)
	GetSpeed(ctx context.Context) (float64, error)
	SetSpeed(ctx context.Context, speed float64) error
	GetPause(ctx context.Context) (bool, error)
	SetPause(ctx context.Context, pause bool) error
	Seek(ctx context.Context, pos time.Duration, isRelative bool, quick bool) error
	GetVideoTracks(ctx context.Context) (VideoTracks, error)
	GetAudioTracks(ctx context.Context) (AudioTracks, error)
	GetSubtitlesTracks(ctx context.Context) (SubtitlesTracks, error)
	SetVideoTrack(ctx context.Context, vid int64) error
	SetAudioTrack(ctx context.Context, aid int64) error
	SetSubtitlesTrack(ctx context.Context, sid int64) error
	Stop(ctx context.Context) error
	Close(ctx context.Context) error
	SetupForStreaming(ctx context.Context) error
}

type PlayerCommon struct {
	Title         string
	Preset        *Preset
	AudioBuffer   *time.Duration
	CacheDuration *time.Duration
	CacheMaxSize  *uint64
}

func (p PlayerCommon) ProcessTitle(
	ctx context.Context,
) (string, error) {
	return p.Title, nil
}

type VideoTrack struct {
	ID       int64
	IsActive bool
}

type VideoTracks []VideoTrack

type AudioTrack struct {
	ID       int64
	IsActive bool
}

type AudioTracks []AudioTrack

type SubtitlesTrack struct {
	ID       int64
	IsActive bool
}

type SubtitlesTracks []SubtitlesTrack
