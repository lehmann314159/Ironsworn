package engine

import (
	"fmt"

	"github.com/ironsworn/ironsworn/internal/model"
)

// MarkProgress marks progress on a track and returns the ticks added.
func MarkProgress(track *model.ProgressTrack) (int, error) {
	if track.Completed {
		return 0, fmt.Errorf("track '%s' is already completed", track.Name)
	}
	return track.MarkProgress()
}

// CreateProgressTrack creates a new progress track with validation.
func CreateProgressTrack(id, gameID, name string, trackType model.ProgressTrackType, rank model.ProgressRank) (*model.ProgressTrack, error) {
	if name == "" {
		return nil, fmt.Errorf("track name cannot be empty")
	}
	if err := rank.Validate(); err != nil {
		return nil, err
	}

	return &model.ProgressTrack{
		ID:        id,
		GameID:    gameID,
		Name:      name,
		TrackType: trackType,
		Rank:      rank,
		Ticks:     0,
		Completed: false,
	}, nil
}
