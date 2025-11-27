//go:build with_libvlc
// +build with_libvlc

package client

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/facebookincubator/go-belt/tool/logger"
	"github.com/xaionaro-go/observability"
	"github.com/xaionaro-go/player/pkg/player/protobuf/go/player_grpc"
	"github.com/xaionaro-go/player/pkg/player/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	Title  string
	Target string
}

var _ types.Player = (*Client)(nil)

func New(title, target string) *Client {
	return &Client{Title: title, Target: target}
}

func (c *Client) grpcClient() (player_grpc.PlayerClient, *grpc.ClientConn, error) {
	conn, err := grpc.NewClient(
		c.Target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to initialize a gRPC client: %w", err)
	}

	client := player_grpc.NewPlayerClient(conn)
	return client, conn, nil
}

func (c *Client) ProcessTitle(
	ctx context.Context,
) (string, error) {
	client, conn, err := c.grpcClient()
	if err != nil {
		return "", err
	}
	defer conn.Close()

	resp, err := client.ProcessTitle(ctx, &player_grpc.ProcessTitleRequest{})
	if err != nil {
		return "", fmt.Errorf("query error: %w", err)
	}
	return resp.GetTitle(), nil
}

func logLevelGo2Protobuf(logLevel logger.Level) player_grpc.LoggingLevel {
	switch logLevel {
	case logger.LevelFatal:
		return player_grpc.LoggingLevel_LoggingLevelFatal
	case logger.LevelPanic:
		return player_grpc.LoggingLevel_LoggingLevelPanic
	case logger.LevelError:
		return player_grpc.LoggingLevel_LoggingLevelError
	case logger.LevelWarning:
		return player_grpc.LoggingLevel_LoggingLevelWarn
	case logger.LevelInfo:
		return player_grpc.LoggingLevel_LoggingLevelInfo
	case logger.LevelDebug:
		return player_grpc.LoggingLevel_LoggingLevelDebug
	case logger.LevelTrace:
		return player_grpc.LoggingLevel_LoggingLevelTrace
	default:
		return player_grpc.LoggingLevel_LoggingLevelWarn
	}
}

func (c *Client) SetupForStreaming(
	ctx context.Context,
) error {
	client, conn, err := c.grpcClient()
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = client.SetupForStreaming(ctx, &player_grpc.SetupForStreamingRequest{})
	if err != nil {
		return fmt.Errorf("query error: %w", err)
	}
	return nil
}

func (c *Client) OpenURL(
	ctx context.Context,
	link string,
) error {
	client, conn, err := c.grpcClient()
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = client.Open(ctx, &player_grpc.OpenRequest{
		Link:         link,
		Title:        c.Title,
		LoggingLevel: logLevelGo2Protobuf(logger.FromCtx(ctx).Level()),
	})
	if err != nil {
		return fmt.Errorf("query error: %w", err)
	}
	return nil
}

func (c *Client) GetLink(
	ctx context.Context,
) (string, error) {
	client, conn, err := c.grpcClient()
	if err != nil {
		return "", err
	}
	defer conn.Close()

	resp, err := client.GetLink(ctx, &player_grpc.GetLinkRequest{})
	if err != nil {
		return "", fmt.Errorf("query error: %w", err)
	}
	return resp.GetLink(), nil
}

func (c *Client) EndChan(ctx context.Context) (<-chan struct{}, error) {
	client, conn, err := c.grpcClient()
	if err != nil {
		return nil, err
	}

	waiter, err := client.EndChan(ctx, &player_grpc.EndChanRequest{})
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}

	result := make(chan struct{})
	waiter.CloseSend()
	observability.Go(ctx, func(ctx context.Context) {
		defer conn.Close()
		defer func() {
			close(result)
		}()

		_, err := waiter.Recv()
		if err == io.EOF {
			logger.Debugf(ctx, "the receiver is closed: %v", err)
			return
		}
		if err != nil {
			logger.Errorf(ctx, "unable to read data: %v", err)
			return
		}
	})

	return result, nil
}

func (c *Client) IsEnded(
	ctx context.Context,
) (bool, error) {
	client, conn, err := c.grpcClient()
	if err != nil {
		return false, err
	}
	defer conn.Close()

	resp, err := client.IsEnded(ctx, &player_grpc.IsEndedRequest{})
	if err != nil {
		return false, fmt.Errorf("query error: %w", err)
	}
	return resp.GetIsEnded(), nil
}

func (c *Client) GetPosition(
	ctx context.Context,
) (time.Duration, error) {
	client, conn, err := c.grpcClient()
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	resp, err := client.GetPosition(ctx, &player_grpc.GetPositionRequest{})
	if err != nil {
		return 0, fmt.Errorf("query error: %w", err)
	}
	return time.Duration(resp.GetPositionSecs() * float64(time.Second)), nil
}

func (c *Client) GetAudioPosition(
	ctx context.Context,
) (time.Duration, error) {
	client, conn, err := c.grpcClient()
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	resp, err := client.GetAudioPosition(ctx, &player_grpc.GetAudioPositionRequest{})
	if err != nil {
		return 0, fmt.Errorf("query error: %w", err)
	}
	return time.Duration(resp.GetPositionSecs() * float64(time.Second)), nil
}

func (c *Client) GetLength(
	ctx context.Context,
) (time.Duration, error) {
	client, conn, err := c.grpcClient()
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	resp, err := client.GetLength(ctx, &player_grpc.GetLengthRequest{})
	if err != nil {
		return 0, fmt.Errorf("query error: %w", err)
	}
	return time.Duration(resp.GetLengthSecs() * float64(time.Second)), nil
}

func (c *Client) GetSpeed(
	ctx context.Context,
) (float64, error) {
	client, conn, err := c.grpcClient()
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	resp, err := client.GetSpeed(ctx, &player_grpc.GetSpeedRequest{})
	if err != nil {
		return 0, fmt.Errorf("query error: %w", err)
	}
	return resp.GetSpeed(), nil
}

func (c *Client) SetSpeed(
	ctx context.Context,
	speed float64,
) error {
	client, conn, err := c.grpcClient()
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = client.SetSpeed(ctx, &player_grpc.SetSpeedRequest{Speed: speed})
	if err != nil {
		return fmt.Errorf("query error: %w", err)
	}
	return nil
}

func (c *Client) GetPause(
	ctx context.Context,
) (bool, error) {
	client, conn, err := c.grpcClient()
	if err != nil {
		return false, err
	}
	defer conn.Close()

	resp, err := client.GetPause(ctx, &player_grpc.GetPauseRequest{})
	if err != nil {
		return false, fmt.Errorf("query error: %w", err)
	}
	return resp.IsPaused, nil
}

func (c *Client) SetPause(
	ctx context.Context,
	pause bool,
) error {
	client, conn, err := c.grpcClient()
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = client.SetPause(ctx, &player_grpc.SetPauseRequest{
		IsPaused: pause,
	})
	if err != nil {
		return fmt.Errorf("query error: %w", err)
	}
	return nil
}

func (c *Client) Seek(
	ctx context.Context,
	pos time.Duration,
	isRelative bool,
	quick bool,
) error {
	client, conn, err := c.grpcClient()
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = client.Seek(ctx, &player_grpc.SeekRequest{
		Pos:        pos.Nanoseconds(),
		IsRelative: isRelative,
		IsQuick:    quick,
	})
	if err != nil {
		return fmt.Errorf("query error: %w", err)
	}
	return nil
}

func (c *Client) GetVideoTracks(
	ctx context.Context,
) (types.VideoTracks, error) {
	client, conn, err := c.grpcClient()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	resp, err := client.GetVideoTracks(ctx, &player_grpc.GetVideoTracksRequest{})
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}

	var result types.VideoTracks
	for _, track := range resp.GetVideoTrack() {
		result = append(result, types.VideoTrack{ID: track.GetId()})
	}
	return result, nil
}

func (c *Client) GetAudioTracks(
	ctx context.Context,
) (types.AudioTracks, error) {
	client, conn, err := c.grpcClient()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	resp, err := client.GetAudioTracks(ctx, &player_grpc.GetAudioTracksRequest{})
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}

	var result types.AudioTracks
	for _, track := range resp.GetAudioTrack() {
		result = append(result, types.AudioTrack{ID: track.GetId()})
	}
	return result, nil
}

func (c *Client) GetSubtitlesTracks(
	ctx context.Context,
) (types.SubtitlesTracks, error) {
	client, conn, err := c.grpcClient()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	resp, err := client.GetSubtitlesTracks(ctx, &player_grpc.GetSubtitlesTracksRequest{})
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}

	var result types.SubtitlesTracks
	for _, track := range resp.GetSubtitlesTrack() {
		result = append(result, types.SubtitlesTrack{ID: track.GetId()})
	}
	return result, nil
}

func (c *Client) SetVideoTrack(
	ctx context.Context,
	vid int64,
) error {
	client, conn, err := c.grpcClient()
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = client.SetVideoTrack(ctx, &player_grpc.SetVideoTrackRequest{
		VideoTrackID: vid,
	})
	if err != nil {
		return fmt.Errorf("query error: %w", err)
	}
	return nil
}

func (c *Client) SetAudioTrack(
	ctx context.Context,
	aid int64,
) error {
	client, conn, err := c.grpcClient()
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = client.SetAudioTrack(ctx, &player_grpc.SetAudioTrackRequest{
		AudioTrackID: aid,
	})
	if err != nil {
		return fmt.Errorf("query error: %w", err)
	}
	return nil
}

func (c *Client) SetSubtitlesTrack(
	ctx context.Context,
	sid int64,
) error {
	client, conn, err := c.grpcClient()
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = client.SetSubtitlesTrack(ctx, &player_grpc.SetSubtitlesTrackRequest{
		SubtitlesTrackID: sid,
	})
	if err != nil {
		return fmt.Errorf("query error: %w", err)
	}
	return nil
}

func (c *Client) Stop(
	ctx context.Context,
) error {
	client, conn, err := c.grpcClient()
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = client.Stop(ctx, &player_grpc.StopRequest{})
	if err != nil {
		return fmt.Errorf("query error: %w", err)
	}
	return nil
}

func (c *Client) Close(ctx context.Context) error {
	client, conn, err := c.grpcClient()
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = client.Close(ctx, &player_grpc.CloseRequest{})
	if err != nil {
		return fmt.Errorf("query error: %w", err)
	}
	return nil
}
