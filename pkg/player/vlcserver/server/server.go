//go:build with_libvlc
// +build with_libvlc

package server

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/facebookincubator/go-belt"
	"github.com/facebookincubator/go-belt/tool/logger"
	"github.com/facebookincubator/go-belt/tool/logger/implementation/logrus"
	"github.com/xaionaro-go/observability"
	"github.com/xaionaro-go/player/pkg/player/protobuf/go/player_grpc"
	"github.com/xaionaro-go/player/pkg/player/vlcserver/player"
	"github.com/xaionaro-go/xsync"
	"google.golang.org/grpc"
)

const (
	timeoutCloseVLC = 10 * time.Second
)

type GRPCServer struct {
	player_grpc.UnimplementedPlayerServer
	GRPCServer *grpc.Server

	VLCLocker xsync.Mutex
	VLC       *player.VLC
	Belt      *belt.Belt
}

func NewServer() *GRPCServer {
	srv := &GRPCServer{
		GRPCServer: grpc.NewServer(),
	}
	player_grpc.RegisterPlayerServer(srv.GRPCServer, srv)
	return srv
}

func (srv *GRPCServer) Serve(
	listener net.Listener,
) error {
	return srv.GRPCServer.Serve(listener)
}

func logLevelProtobuf2Go(logLevel player_grpc.LoggingLevel) logger.Level {
	switch logLevel {
	case player_grpc.LoggingLevel_LoggingLevelNone:
		return logger.LevelFatal
	case player_grpc.LoggingLevel_LoggingLevelFatal:
		return logger.LevelFatal
	case player_grpc.LoggingLevel_LoggingLevelPanic:
		return logger.LevelPanic
	case player_grpc.LoggingLevel_LoggingLevelError:
		return logger.LevelError
	case player_grpc.LoggingLevel_LoggingLevelWarn:
		return logger.LevelWarning
	case player_grpc.LoggingLevel_LoggingLevelInfo:
		return logger.LevelInfo
	case player_grpc.LoggingLevel_LoggingLevelDebug:
		return logger.LevelDebug
	case player_grpc.LoggingLevel_LoggingLevelTrace:
		return logger.LevelTrace
	default:
		return logger.LevelUndefined
	}
}

func (srv *GRPCServer) Open(
	ctx context.Context,
	req *player_grpc.OpenRequest,
) (*player_grpc.OpenReply, error) {
	return xsync.DoR2(ctx, &srv.VLCLocker, func() (*player_grpc.OpenReply, error) {
		if err := srv.closePlayer(ctx); err != nil {
			logger.Errorf(ctx, "unable to close the player: %w", err)
		}

		var err error
		srv.VLC, err = player.NewVLC(req.GetTitle())
		if err != nil {
			return nil, fmt.Errorf("unable to initialize the VLC player: %w", err)
		}

		if err := srv.VLC.OpenURL(ctx, req.Link); err != nil {
			return nil, fmt.Errorf("unable to open link '%s': %w", req.Link, err)
		}

		l := logrus.Default().WithLevel(logLevelProtobuf2Go(req.LoggingLevel))
		srv.Belt = logger.BeltWithLogger(belt.New(), l)

		return &player_grpc.OpenReply{}, nil
	})
}

func (srv *GRPCServer) SetupForStreaming(
	ctx context.Context,
	req *player_grpc.SetupForStreamingRequest,
) (*player_grpc.SetupForStreamingReply, error) {
	return &player_grpc.SetupForStreamingReply{}, nil
}

func (srv *GRPCServer) ctx(ctx context.Context) context.Context {
	return belt.CtxWithBelt(ctx, srv.Belt)
}

func (srv *GRPCServer) isInited() error {
	ctx := context.TODO()
	return xsync.DoR1(ctx, &srv.VLCLocker, func() error {
		if srv.VLC == nil {
			return fmt.Errorf("call Open first")
		}
		return nil
	})
}

func (srv *GRPCServer) ProcessTitle(
	ctx context.Context,
	req *player_grpc.ProcessTitleRequest,
) (*player_grpc.ProcessTitleReply, error) {
	if err := srv.isInited(); err != nil {
		return nil, err
	}
	title, err := srv.VLC.ProcessTitle(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get the title: %w", err)
	}
	return &player_grpc.ProcessTitleReply{
		Title: title,
	}, nil
}

func (srv *GRPCServer) GetLink(
	ctx context.Context,
	req *player_grpc.GetLinkRequest,
) (*player_grpc.GetLinkReply, error) {
	if err := srv.isInited(); err != nil {
		return nil, err
	}
	link, err := srv.VLC.GetLink(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get the title: %w", err)
	}
	return &player_grpc.GetLinkReply{
		Link: link,
	}, nil
}

func (srv *GRPCServer) EndChan(
	req *player_grpc.EndChanRequest,
	server player_grpc.Player_EndChanServer,
) (_ret error) {
	ctx := srv.ctx(server.Context())
	logger.Tracef(ctx, "EndChan()")
	defer func() {
		logger.Tracef(ctx, "/EndChan(): %v", _ret)
	}()

	if err := srv.isInited(); err != nil {
		return err
	}

	for {
		ch, err := srv.VLC.EndChan(ctx)
		if err != nil {
			return fmt.Errorf("unable to get the EndChan: %w", err)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ch:
		}

		return server.Send(&player_grpc.EndChanReply{})
	}
}

func (srv *GRPCServer) IsEnded(
	ctx context.Context,
	req *player_grpc.IsEndedRequest,
) (*player_grpc.IsEndedReply, error) {
	if err := srv.isInited(); err != nil {
		return nil, err
	}
	isEnded, err := srv.VLC.IsEnded(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get if it is already ended: %w", err)
	}
	return &player_grpc.IsEndedReply{
		IsEnded: isEnded,
	}, nil
}

func (srv *GRPCServer) GetPosition(
	ctx context.Context,
	req *player_grpc.GetPositionRequest,
) (*player_grpc.GetPositionReply, error) {
	if err := srv.isInited(); err != nil {
		return nil, err
	}
	pos, err := srv.VLC.GetPosition(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get the position: %w", err)
	}
	return &player_grpc.GetPositionReply{
		PositionSecs: pos.Seconds(),
	}, nil
}

func (srv *GRPCServer) GetLength(
	ctx context.Context,
	req *player_grpc.GetLengthRequest,
) (*player_grpc.GetLengthReply, error) {
	if err := srv.isInited(); err != nil {
		return nil, err
	}
	length, err := srv.VLC.GetLength(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get the length: %w", err)
	}
	return &player_grpc.GetLengthReply{
		LengthSecs: length.Seconds(),
	}, nil
}

func (srv *GRPCServer) GetSpeed(
	ctx context.Context,
	req *player_grpc.GetSpeedRequest,
) (*player_grpc.GetSpeedReply, error) {
	if err := srv.isInited(); err != nil {
		return nil, err
	}
	speed, err := srv.VLC.GetSpeed(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to set speed to '%v': %w", speed, err)
	}
	return &player_grpc.GetSpeedReply{
		Speed: speed,
	}, nil
}

func (srv *GRPCServer) SetSpeed(
	ctx context.Context,
	req *player_grpc.SetSpeedRequest,
) (*player_grpc.SetSpeedReply, error) {
	if err := srv.isInited(); err != nil {
		return nil, err
	}
	if err := srv.VLC.SetSpeed(ctx, req.GetSpeed()); err != nil {
		return nil, fmt.Errorf("unable to set speed to '%v': %w", req.GetSpeed(), err)
	}
	return &player_grpc.SetSpeedReply{}, nil
}

func (srv *GRPCServer) GetPause(
	ctx context.Context,
	req *player_grpc.GetPauseRequest,
) (*player_grpc.GetPauseReply, error) {
	if err := srv.isInited(); err != nil {
		return nil, err
	}
	isPaused, err := srv.VLC.GetPause(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get the info if it is paused: %w", err)
	}
	return &player_grpc.GetPauseReply{
		IsPaused: isPaused,
	}, nil
}

func (srv *GRPCServer) SetPause(
	ctx context.Context,
	req *player_grpc.SetPauseRequest,
) (*player_grpc.SetPauseReply, error) {
	if err := srv.isInited(); err != nil {
		return nil, err
	}
	if err := srv.VLC.SetPause(ctx, req.GetIsPaused()); err != nil {
		return nil, fmt.Errorf("unable to set paused state to '%v': %w", req.GetIsPaused(), err)
	}
	return &player_grpc.SetPauseReply{}, nil
}

func (srv *GRPCServer) Seek(
	ctx context.Context,
	req *player_grpc.SeekRequest,
) (*player_grpc.SeekReply, error) {
	if err := srv.isInited(); err != nil {
		return nil, err
	}
	pos := time.Nanosecond * time.Duration(req.GetPos())
	isRel := req.GetIsRelative()
	isQuick := req.GetIsQuick()
	if err := srv.VLC.Seek(ctx, pos, isRel, isQuick); err != nil {
		return nil, fmt.Errorf("unable to seek to %v (rel:%t, quick:%t): %w", pos, isRel, isQuick, err)
	}
	return &player_grpc.SeekReply{}, nil
}

func (srv *GRPCServer) GetVideoTracks(
	ctx context.Context,
	req *player_grpc.GetVideoTracksRequest,
) (*player_grpc.GetVideoTracksReply, error) {
	if err := srv.isInited(); err != nil {
		return nil, err
	}
	result, err := srv.VLC.GetVideoTracks(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get video tracks: %w", err)
	}
	resp := &player_grpc.GetVideoTracksReply{}
	for _, track := range result {
		resp.VideoTrack = append(resp.VideoTrack, &player_grpc.VideoTrack{
			Id:       track.ID,
			IsActive: track.IsActive,
		})
	}
	return resp, nil
}

func (srv *GRPCServer) GetAudioTracks(
	ctx context.Context,
	req *player_grpc.GetAudioTracksRequest,
) (*player_grpc.GetAudioTracksReply, error) {
	if err := srv.isInited(); err != nil {
		return nil, err
	}
	result, err := srv.VLC.GetAudioTracks(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get audio tracks: %w", err)
	}
	resp := &player_grpc.GetAudioTracksReply{}
	for _, track := range result {
		resp.AudioTrack = append(resp.AudioTrack, &player_grpc.AudioTrack{
			Id:       track.ID,
			IsActive: track.IsActive,
		})
	}
	return resp, nil
}

func (srv *GRPCServer) GetSubtitlesTracks(
	ctx context.Context,
	req *player_grpc.GetSubtitlesTracksRequest,
) (*player_grpc.GetSubtitlesTracksReply, error) {
	if err := srv.isInited(); err != nil {
		return nil, err
	}
	result, err := srv.VLC.GetSubtitlesTracks(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get subtitles tracks: %w", err)
	}
	resp := &player_grpc.GetSubtitlesTracksReply{}
	for _, track := range result {
		resp.SubtitlesTrack = append(resp.SubtitlesTrack, &player_grpc.SubtitlesTrack{
			Id:       track.ID,
			IsActive: track.IsActive,
		})
	}
	return resp, nil
}

func (srv *GRPCServer) SetVideoTrack(
	ctx context.Context,
	req *player_grpc.SetVideoTrackRequest,
) (*player_grpc.SetVideoTrackReply, error) {
	if err := srv.isInited(); err != nil {
		return nil, err
	}
	if err := srv.VLC.SetVideoTrack(ctx, req.GetVideoTrackID()); err != nil {
		return nil, fmt.Errorf("unable to set video track ID to '%v': %w", req.GetVideoTrackID(), err)
	}
	return &player_grpc.SetVideoTrackReply{}, nil
}

func (srv *GRPCServer) SetAudioTrack(
	ctx context.Context,
	req *player_grpc.SetAudioTrackRequest,
) (*player_grpc.SetAudioTrackReply, error) {
	if err := srv.isInited(); err != nil {
		return nil, err
	}
	if err := srv.VLC.SetAudioTrack(ctx, req.GetAudioTrackID()); err != nil {
		return nil, fmt.Errorf("unable to set audio track ID to '%v': %w", req.GetAudioTrackID(), err)
	}
	return &player_grpc.SetAudioTrackReply{}, nil
}

func (srv *GRPCServer) SetSubtitlesTrack(
	ctx context.Context,
	req *player_grpc.SetSubtitlesTrackRequest,
) (*player_grpc.SetSubtitlesTrackReply, error) {
	if err := srv.isInited(); err != nil {
		return nil, err
	}
	if err := srv.VLC.SetSubtitlesTrack(ctx, req.GetSubtitlesTrackID()); err != nil {
		return nil, fmt.Errorf("unable to set subtitles track ID to '%v': %w", req.GetSubtitlesTrackID(), err)
	}
	return &player_grpc.SetSubtitlesTrackReply{}, nil
}

func (srv *GRPCServer) Stop(
	ctx context.Context,
	req *player_grpc.StopRequest,
) (*player_grpc.StopReply, error) {
	if err := srv.isInited(); err != nil {
		return nil, err
	}
	if err := srv.VLC.Stop(ctx); err != nil {
		return nil, fmt.Errorf("unable to stop the playback: %w", err)
	}
	return &player_grpc.StopReply{}, nil
}

func (srv *GRPCServer) Close(
	ctx context.Context,
	req *player_grpc.CloseRequest,
) (*player_grpc.CloseReply, error) {
	if err := srv.isInited(); err != nil {
		return nil, nil
	}
	return xsync.DoR2(ctx, &srv.VLCLocker, func() (*player_grpc.CloseReply, error) {
		if err := srv.close(ctx); err != nil {
			return nil, err
		}
		return &player_grpc.CloseReply{}, nil
	})
}

func (srv *GRPCServer) closePlayer(
	ctx context.Context,
) (_err error) {
	logger.Debugf(ctx, "closePlayer")
	defer func() { logger.Debugf(ctx, "/closePlayer: %v", _err) }()
	defer func() {
		srv.VLC = nil
	}()

	if srv.VLC == nil {
		return nil
	}

	ctx, cancelFn := context.WithTimeout(ctx, timeoutCloseVLC)
	defer cancelFn()
	errCh := make(chan error, 1)
	observability.Go(ctx, func() {
		defer close(errCh)
		if err := srv.VLC.Close(ctx); err != nil {
			errCh <- fmt.Errorf("unable to stop the playback: %w", err)
			return
		}
		errCh <- nil
	})
	select {
	case <-ctx.Done():
		return fmt.Errorf("closing takes too long: %w", ctx.Err())
	case err := <-errCh:
		return err
	}
}

func (srv *GRPCServer) close(
	ctx context.Context,
) (_err error) {
	logger.Debugf(ctx, "close")
	defer func() { logger.Debugf(ctx, "/close: %v", _err) }()

	defer func() {
		if srv.GRPCServer != nil {
			logger.Debugf(ctx, "closing the GRPCServer")
			srv.GRPCServer.Stop()
		}
		srv.Belt = nil
	}()

	err := srv.closePlayer(ctx)
	if err != nil {
		return fmt.Errorf("unable to close the player: %w", err)
	}

	return nil
}
