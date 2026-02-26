package model

import "time"

// CharacterSnapshot captures key character state for log entries.
type CharacterSnapshot struct {
	Health   int `json:"health"`
	Spirit   int `json:"spirit"`
	Supply   int `json:"supply"`
	Momentum int `json:"momentum"`
}

// LogEntry is a RAG-ready structured log entry.
type LogEntry struct {
	ID       string    `json:"id"`
	GameID   string    `json:"game_id"`
	Sequence int       `json:"sequence"`
	Timestamp time.Time `json:"timestamp"`

	// Classification (for RAG retrieval)
	EntryType    string `json:"entry_type"`               // "move", "oracle", "narrative", "state_change"
	MoveID       string `json:"move_id,omitempty"`
	MoveCategory string `json:"move_category,omitempty"`
	Tags         []string `json:"tags"`

	// Snapshot for context (RAG needs self-contained entries)
	CharacterBefore *CharacterSnapshot `json:"character_before,omitempty"`
	CharacterAfter  *CharacterSnapshot `json:"character_after,omitempty"`

	// Action details
	Roll         *ActionRoll         `json:"roll,omitempty"`
	ProgressRoll *ProgressRollResult `json:"progress_roll,omitempty"`
	OracleResult *OracleRollResult   `json:"oracle_result,omitempty"`
	Outcome      Outcome             `json:"outcome,omitempty"`

	// Human-readable content (key for RAG embeddings)
	Summary          string   `json:"summary"`
	NarrativeContext string   `json:"narrative_context,omitempty"`
	MechanicalEffects []string `json:"mechanical_effects,omitempty"`

	// Relationships (for graph-based RAG)
	RelatedTrackIDs []string `json:"related_track_ids,omitempty"`
	RelatedAssetIDs []string `json:"related_asset_ids,omitempty"`
}

// LogFilter provides filtering options for log queries.
type LogFilter struct {
	EntryType string   `json:"entry_type,omitempty"`
	Tags      []string `json:"tags,omitempty"`
	Outcome   string   `json:"outcome,omitempty"`
	Limit     int      `json:"limit,omitempty"`
	Offset    int      `json:"offset,omitempty"`
}
