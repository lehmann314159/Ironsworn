package model

import "fmt"

// ProgressRank determines how many ticks are marked per progress.
type ProgressRank string

const (
	RankTroublesome ProgressRank = "troublesome"
	RankDangerous   ProgressRank = "dangerous"
	RankFormidable  ProgressRank = "formidable"
	RankExtreme     ProgressRank = "extreme"
	RankEpic        ProgressRank = "epic"
)

// TicksPerMark returns the number of ticks marked per progress for this rank.
func (r ProgressRank) TicksPerMark() (int, error) {
	switch r {
	case RankTroublesome:
		return 12, nil
	case RankDangerous:
		return 8, nil
	case RankFormidable:
		return 4, nil
	case RankExtreme:
		return 2, nil
	case RankEpic:
		return 1, nil
	default:
		return 0, fmt.Errorf("unknown progress rank: %s", string(r))
	}
}

// Validate checks if the rank is valid.
func (r ProgressRank) Validate() error {
	_, err := r.TicksPerMark()
	return err
}

// ProgressTrackType categorizes progress tracks.
type ProgressTrackType string

const (
	TrackVow     ProgressTrackType = "vow"
	TrackJourney ProgressTrackType = "journey"
	TrackCombat  ProgressTrackType = "combat"
	TrackBonds   ProgressTrackType = "bonds"
)

// ProgressTrack represents a progress track (vow, journey, combat, bonds).
type ProgressTrack struct {
	ID        string            `json:"id"`
	GameID    string            `json:"game_id"`
	Name      string            `json:"name"`
	TrackType ProgressTrackType `json:"track_type"`
	Rank      ProgressRank      `json:"rank"`
	Ticks     int               `json:"ticks"` // 0-40
	Completed bool              `json:"completed"`
}

// Score returns the progress score (ticks / 4, max 10).
func (t *ProgressTrack) Score() int {
	score := t.Ticks / 4
	if score > 10 {
		return 10
	}
	return score
}

// MarkProgress adds ticks based on rank. Returns the number of ticks added.
func (t *ProgressTrack) MarkProgress() (int, error) {
	ticks, err := t.Rank.TicksPerMark()
	if err != nil {
		return 0, err
	}
	t.Ticks += ticks
	if t.Ticks > 40 {
		t.Ticks = 40
	}
	return ticks, nil
}
