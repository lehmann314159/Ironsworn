package store

import (
	"context"

	"github.com/ironsworn/ironsworn/internal/model"
)

// GameStore manages game persistence.
type GameStore interface {
	CreateGame(ctx context.Context, game *model.Game) error
	GetGame(ctx context.Context, id string) (*model.Game, error)
	ListGames(ctx context.Context) ([]*model.Game, error)
	DeleteGame(ctx context.Context, id string) error
}

// CharacterStore manages character persistence.
type CharacterStore interface {
	CreateCharacter(ctx context.Context, ch *model.Character) error
	GetCharacter(ctx context.Context, gameID string) (*model.Character, error)
	UpdateCharacter(ctx context.Context, ch *model.Character) error
}

// ProgressStore manages progress track persistence.
type ProgressStore interface {
	CreateTrack(ctx context.Context, track *model.ProgressTrack) error
	GetTrack(ctx context.Context, id string) (*model.ProgressTrack, error)
	ListTracks(ctx context.Context, gameID string, trackType string, completed *bool) ([]*model.ProgressTrack, error)
	UpdateTrack(ctx context.Context, track *model.ProgressTrack) error
	DeleteTrack(ctx context.Context, id string) error
}

// AssetStore manages character asset persistence.
type AssetStore interface {
	AddAsset(ctx context.Context, asset *model.CharacterAsset) error
	GetAsset(ctx context.Context, id string) (*model.CharacterAsset, error)
	ListAssets(ctx context.Context, characterID string) ([]*model.CharacterAsset, error)
	UpdateAsset(ctx context.Context, asset *model.CharacterAsset) error
}

// LogStore manages game log persistence.
type LogStore interface {
	AppendLog(ctx context.Context, entry *model.LogEntry) error
	GetLogs(ctx context.Context, gameID string, filter model.LogFilter) ([]*model.LogEntry, error)
	GetNextSequence(ctx context.Context, gameID string) (int, error)
}

// Store combines all storage interfaces.
type Store interface {
	GameStore
	CharacterStore
	ProgressStore
	AssetStore
	LogStore
	Close() error
}
