package store

import (
	"context"
	"testing"
	"time"

	"github.com/ironsworn/ironsworn-backend/internal/model"
)

func testStore(t *testing.T) *SQLiteStore {
	t.Helper()
	s, err := NewSQLiteStore(":memory:")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestGameCRUD(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()

	// Create
	game := &model.Game{ID: "g1", Name: "Test Campaign"}
	if err := s.CreateGame(ctx, game); err != nil {
		t.Fatalf("create: %v", err)
	}

	// Read
	got, err := s.GetGame(ctx, "g1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Name != "Test Campaign" {
		t.Errorf("expected 'Test Campaign', got %q", got.Name)
	}

	// List
	games, err := s.ListGames(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(games) != 1 {
		t.Errorf("expected 1 game, got %d", len(games))
	}

	// Delete
	if err := s.DeleteGame(ctx, "g1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	_, err = s.GetGame(ctx, "g1")
	if err == nil {
		t.Error("expected error after delete")
	}
}

func TestCharacterCRUD(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()

	s.CreateGame(ctx, &model.Game{ID: "g1", Name: "Test"})

	ch := &model.Character{
		ID: "c1", GameID: "g1", Name: "Kara",
		Stats:         model.Stats{Edge: 2, Heart: 3, Iron: 1, Shadow: 2, Wits: 1},
		Health: 5, Spirit: 5, Supply: 5,
		Momentum: 2, MomentumMax: 10, MomentumReset: 2,
	}
	if err := s.CreateCharacter(ctx, ch); err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := s.GetCharacter(ctx, "g1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Name != "Kara" {
		t.Errorf("expected 'Kara', got %q", got.Name)
	}
	if got.Stats.Heart != 3 {
		t.Errorf("expected heart 3, got %d", got.Stats.Heart)
	}

	// Update
	got.Health = 3
	got.Debilities.Wounded = true
	if err := s.UpdateCharacter(ctx, got); err != nil {
		t.Fatalf("update: %v", err)
	}

	got2, _ := s.GetCharacter(ctx, "g1")
	if got2.Health != 3 {
		t.Errorf("expected health 3, got %d", got2.Health)
	}
	if !got2.Debilities.Wounded {
		t.Error("expected wounded to be true")
	}
}

func TestProgressTrackCRUD(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()

	s.CreateGame(ctx, &model.Game{ID: "g1", Name: "Test"})

	track := &model.ProgressTrack{
		ID: "t1", GameID: "g1", Name: "Protect Village",
		TrackType: model.TrackVow, Rank: model.RankDangerous,
	}
	if err := s.CreateTrack(ctx, track); err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := s.GetTrack(ctx, "t1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Name != "Protect Village" {
		t.Errorf("expected 'Protect Village', got %q", got.Name)
	}

	got.Ticks = 16
	if err := s.UpdateTrack(ctx, got); err != nil {
		t.Fatalf("update: %v", err)
	}

	got2, _ := s.GetTrack(ctx, "t1")
	if got2.Ticks != 16 {
		t.Errorf("expected 16 ticks, got %d", got2.Ticks)
	}

	tracks, _ := s.ListTracks(ctx, "g1", "vow", nil)
	if len(tracks) != 1 {
		t.Errorf("expected 1 track, got %d", len(tracks))
	}
}

func TestGameLogAppendAndQuery(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()

	s.CreateGame(ctx, &model.Game{ID: "g1", Name: "Test"})

	entry := &model.LogEntry{
		ID: "log1", GameID: "g1", Sequence: 1,
		Timestamp: time.Now().UTC(),
		EntryType: "move",
		Summary:   "Faced danger: Strong Hit",
		Tags:      []string{"face_danger", "strong_hit"},
		Outcome:   model.OutcomeStrongHit,
	}
	if err := s.AppendLog(ctx, entry); err != nil {
		t.Fatalf("append: %v", err)
	}

	logs, err := s.GetLogs(ctx, "g1", model.LogFilter{})
	if err != nil {
		t.Fatalf("get logs: %v", err)
	}
	if len(logs) != 1 {
		t.Errorf("expected 1 log, got %d", len(logs))
	}

	// Filter by type
	logs, _ = s.GetLogs(ctx, "g1", model.LogFilter{EntryType: "oracle"})
	if len(logs) != 0 {
		t.Errorf("expected 0 oracle logs, got %d", len(logs))
	}

	// Next sequence
	seq, _ := s.GetNextSequence(ctx, "g1")
	if seq != 2 {
		t.Errorf("expected next sequence 2, got %d", seq)
	}
}

func TestCascadeDelete(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()

	s.CreateGame(ctx, &model.Game{ID: "g1", Name: "Test"})
	s.CreateCharacter(ctx, &model.Character{
		ID: "c1", GameID: "g1", Name: "Kara",
		Stats: model.Stats{Edge: 2, Heart: 3, Iron: 1, Shadow: 2, Wits: 1},
		Health: 5, Spirit: 5, Supply: 5,
		Momentum: 2, MomentumMax: 10, MomentumReset: 2,
	})
	s.CreateTrack(ctx, &model.ProgressTrack{
		ID: "t1", GameID: "g1", Name: "Vow",
		TrackType: model.TrackVow, Rank: model.RankDangerous,
	})

	// Delete game should cascade
	s.DeleteGame(ctx, "g1")

	_, err := s.GetCharacter(ctx, "g1")
	if err == nil {
		t.Error("expected character to be deleted with game")
	}

	tracks, _ := s.ListTracks(ctx, "g1", "", nil)
	if len(tracks) != 0 {
		t.Errorf("expected 0 tracks after cascade, got %d", len(tracks))
	}
}
