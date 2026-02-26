package model

import "testing"

func TestProgressRankTicksPerMark(t *testing.T) {
	tests := []struct {
		rank  ProgressRank
		ticks int
	}{
		{RankTroublesome, 12},
		{RankDangerous, 8},
		{RankFormidable, 4},
		{RankExtreme, 2},
		{RankEpic, 1},
	}

	for _, tt := range tests {
		ticks, err := tt.rank.TicksPerMark()
		if err != nil {
			t.Errorf("%s: unexpected error: %v", tt.rank, err)
		}
		if ticks != tt.ticks {
			t.Errorf("%s: expected %d, got %d", tt.rank, tt.ticks, ticks)
		}
	}
}

func TestProgressTrackScore(t *testing.T) {
	track := &ProgressTrack{Ticks: 0}
	if track.Score() != 0 {
		t.Errorf("expected 0, got %d", track.Score())
	}

	track.Ticks = 16
	if track.Score() != 4 { // 16/4 = 4
		t.Errorf("expected 4, got %d", track.Score())
	}

	track.Ticks = 40
	if track.Score() != 10 {
		t.Errorf("expected 10, got %d", track.Score())
	}

	// Over 40 should still be capped at 10
	track.Ticks = 40
	if track.Score() != 10 {
		t.Errorf("expected 10 (capped), got %d", track.Score())
	}
}

func TestProgressTrackMarkProgress(t *testing.T) {
	track := &ProgressTrack{Rank: RankDangerous}
	ticks, err := track.MarkProgress()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ticks != 8 {
		t.Errorf("expected 8 ticks, got %d", ticks)
	}
	if track.Ticks != 8 {
		t.Errorf("expected track at 8 ticks, got %d", track.Ticks)
	}

	// Mark to cap
	track.Ticks = 38
	track.MarkProgress()
	if track.Ticks != 40 { // Capped at 40
		t.Errorf("expected capped at 40, got %d", track.Ticks)
	}
}
