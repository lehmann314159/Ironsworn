package testutil

import (
	"context"
	"testing"

	"github.com/ironsworn/ironsworn-backend/internal/engine"
	"github.com/ironsworn/ironsworn-backend/internal/model"
	"github.com/ironsworn/ironsworn-backend/internal/store"
)

// TestStore creates an in-memory SQLite store for testing.
func TestStore(t *testing.T) store.Store {
	t.Helper()
	s, err := store.NewSQLiteStore(":memory:")
	if err != nil {
		t.Fatalf("failed to create test store: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

// TestCharacter creates a standard test character.
func TestCharacter(t *testing.T, s store.Store, gameID string) *model.Character {
	t.Helper()
	ctx := context.Background()

	ch, err := model.NewCharacter("test-char-1", gameID, "Kara", model.Stats{
		Edge: 2, Heart: 3, Iron: 1, Shadow: 2, Wits: 1,
	})
	if err != nil {
		t.Fatalf("failed to create character: %v", err)
	}
	if err := s.CreateCharacter(ctx, ch); err != nil {
		t.Fatalf("failed to persist character: %v", err)
	}
	return ch
}

// TestGame creates a standard test game.
func TestGame(t *testing.T, s store.Store) *model.Game {
	t.Helper()
	ctx := context.Background()

	game := &model.Game{ID: "test-game-1", Name: "Test Campaign"}
	if err := s.CreateGame(ctx, game); err != nil {
		t.Fatalf("failed to create game: %v", err)
	}
	return game
}

// FixedRoller returns an engine.FixedRoller with the given values.
func FixedRoller(d6 []int, d10 []int, d100 []int) *engine.FixedRoller {
	return &engine.FixedRoller{
		D6Values:   d6,
		D10Values:  d10,
		D100Values: d100,
	}
}

// StrongHitRoller returns a roller that always produces a strong hit (6 vs 1,2).
func StrongHitRoller() *engine.FixedRoller {
	return &engine.FixedRoller{
		D6Values:  []int{6},
		D10Values: []int{1, 2},
	}
}

// WeakHitRoller returns a roller that produces a weak hit (3 vs 2,8).
func WeakHitRoller() *engine.FixedRoller {
	return &engine.FixedRoller{
		D6Values:  []int{3},
		D10Values: []int{2, 8},
	}
}

// MissRoller returns a roller that always produces a miss (1 vs 8,9).
func MissRoller() *engine.FixedRoller {
	return &engine.FixedRoller{
		D6Values:  []int{1},
		D10Values: []int{8, 9},
	}
}
