package heresphere

import (
	"context"
	"github.com/rs/zerolog/log"
	"stash-vr/internal/library"
	"time"
)

func newPlayback(vd *library.VideoData, minPlayFraction float64) *playbackState {
	return &playbackState{
		minPlayFraction: minPlayFraction,
		videoId:         vd.Id(),
		videoDuration:   vd.SceneParts.Files[0].Duration,
		lastPlayTime:    time.Now(),
		isPlaying:       true,
	}
}

func (ps *playbackState) handleStop(ctx context.Context, libraryService *library.Service) {
	if ps.isPlaying {
		if ps.accumulatePlayTime() {
			err := libraryService.IncrementPlayCount(ctx, ps.videoId)
			if err != nil {
				log.Ctx(ctx).Warn().Err(err).Msg("Failed to increment play count")
			}
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

func (ps *playbackState) accumulatePlayTime() bool {
	ps.accumulatedPlayTime += time.Now().Sub(ps.lastPlayTime)
	shouldIncrement := !ps.thresholdReached && ps.accumulatedPlayTime.Seconds() >= ps.videoDuration*ps.minPlayFraction
	if shouldIncrement {
		ps.thresholdReached = true
	}
	return shouldIncrement
}

type playbackState struct {
	minPlayFraction float64

	videoId       string
	videoDuration float64

	accumulatedPlayTime time.Duration
	thresholdReached    bool
	lastPlayTime        time.Time
	isPlaying           bool
}
