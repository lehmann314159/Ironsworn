package gamelog

import (
	"context"
	"crypto/rand"
	"time"

	"github.com/ironsworn/ironsworn/internal/model"
	"github.com/ironsworn/ironsworn/internal/store"
)

// Logger creates structured, RAG-ready log entries.
type Logger struct {
	store store.LogStore
}

// NewLogger creates a new game logger.
func NewLogger(s store.LogStore) *Logger {
	return &Logger{store: s}
}

// LogMove creates a log entry for a move execution.
func (l *Logger) LogMove(ctx context.Context, gameID string, before, after model.CharacterSnapshot, result *model.MoveResult, narrative string, trackIDs []string) error {
	seq, err := l.store.GetNextSequence(ctx, gameID)
	if err != nil {
		return err
	}

	tags := buildMoveTags(result)

	entry := &model.LogEntry{
		ID:                GenerateID(),
		GameID:            gameID,
		Sequence:          seq,
		Timestamp:         time.Now().UTC(),
		EntryType:         "move",
		MoveID:            result.Move.ID,
		MoveCategory:      string(result.Move.Category),
		Tags:              tags,
		CharacterBefore:   &before,
		CharacterAfter:    &after,
		Roll:              result.Roll,
		ProgressRoll:      result.ProgressRoll,
		Outcome:           result.Outcome,
		Summary:           result.Summary,
		NarrativeContext:  narrative,
		MechanicalEffects: result.MechanicalEffects,
		RelatedTrackIDs:   trackIDs,
	}

	return l.store.AppendLog(ctx, entry)
}

// LogOracle creates a log entry for an oracle query.
func (l *Logger) LogOracle(ctx context.Context, gameID string, oracleResult *model.OracleRollResult, summary string) error {
	seq, err := l.store.GetNextSequence(ctx, gameID)
	if err != nil {
		return err
	}

	entry := &model.LogEntry{
		ID:           GenerateID(),
		GameID:       gameID,
		Sequence:     seq,
		Timestamp:    time.Now().UTC(),
		EntryType:    "oracle",
		Tags:         []string{"oracle"},
		OracleResult: oracleResult,
		Summary:      summary,
	}

	return l.store.AppendLog(ctx, entry)
}

// LogNarrative creates a log entry for player narrative text.
func (l *Logger) LogNarrative(ctx context.Context, gameID, text string, tags []string) error {
	seq, err := l.store.GetNextSequence(ctx, gameID)
	if err != nil {
		return err
	}

	if tags == nil {
		tags = []string{"narrative"}
	}

	entry := &model.LogEntry{
		ID:               GenerateID(),
		GameID:           gameID,
		Sequence:         seq,
		Timestamp:        time.Now().UTC(),
		EntryType:        "narrative",
		Tags:             tags,
		Summary:          text,
		NarrativeContext: text,
	}

	return l.store.AppendLog(ctx, entry)
}

// LogStateChange creates a log entry for a manual state change.
func (l *Logger) LogStateChange(ctx context.Context, gameID string, before, after model.CharacterSnapshot, summary string) error {
	seq, err := l.store.GetNextSequence(ctx, gameID)
	if err != nil {
		return err
	}

	entry := &model.LogEntry{
		ID:              GenerateID(),
		GameID:          gameID,
		Sequence:        seq,
		Timestamp:       time.Now().UTC(),
		EntryType:       "state_change",
		Tags:            []string{"state_change"},
		CharacterBefore: &before,
		CharacterAfter:  &after,
		Summary:         summary,
	}

	return l.store.AppendLog(ctx, entry)
}

func buildMoveTags(result *model.MoveResult) []string {
	tags := []string{result.Move.ID}
	if result.Move.Category != "" {
		tags = append(tags, string(result.Move.Category))
	}
	if result.Outcome != "" {
		tags = append(tags, string(result.Outcome))
	}
	if result.Outcome.IsMatch() {
		tags = append(tags, "match")
	}
	return tags
}

// GenerateID produces a random 16-character alphanumeric ID.
func GenerateID() string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 16)
	rand.Read(b)
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return string(b)
}
