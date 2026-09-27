package heresphere

import (
	"context"
	"stash-vr/internal/library"
	"stash-vr/internal/multipart"
	"time"

	"github.com/rs/zerolog/log"
)

func newPlayback(vd *library.VideoData, fileId string) *playbackState {
	targetItem := multipart.TargetItem(vd.SceneId(), vd.SceneParts.Files, fileId)

	return &playbackState{
		sceneId:       vd.SceneId(),
		fileId:        fileId,
		videoDuration: targetItem.File.Duration,
		lastPlayTime:  time.Now(),
		isPlaying:     true,
	}
}

func (ps *playbackState) handleStop(ctx context.Context, libraryService *library.Service, minPlayFraction *float64) {
	if ps.isPlaying {
		currentPlayDuration := time.Since(ps.lastPlayTime)
		ps.accumulatedPlayTime += currentPlayDuration
		if !ps.thresholdReached && minPlayFraction != nil && ps.accumulatedPlayTime.Seconds() >= ps.videoDuration*(*minPlayFraction) {
			ps.thresholdReached = true
			log.Ctx(ctx).Debug().Str("total play time", ps.accumulatedPlayTime.Round(time.Second).String()).Msg("Incrementing play count")
			err := libraryService.IncrementPlayCount(ctx, ps.sceneId)
			if err != nil {
				log.Ctx(ctx).Warn().Err(err).Msg("Failed to increment play count")
			}
		}
		log.Ctx(ctx).Debug().Str("duration", currentPlayDuration.Round(time.Second).String()).Msg("Adding play duration")
		err := libraryService.AddPlayDuration(ctx, ps.sceneId, currentPlayDuration)
		if err != nil {
			log.Ctx(ctx).Warn().Err(err).Msg("Failed to add play duration")
		}
	}
	ps.isPlaying = false
}

func (ps *playbackState) handleResume() {
	if !ps.isPlaying {
		ps.lastPlayTime = time.Now()
	}
	ps.isPlaying = true
}

type playbackState struct {
	sceneId       string
	fileId        string
	videoDuration float64

	accumulatedPlayTime time.Duration
	thresholdReached    bool
	lastPlayTime        time.Time
	isPlaying           bool
}
