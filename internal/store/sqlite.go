package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ironsworn/ironsworn-backend/internal/model"

	_ "modernc.org/sqlite"
)

// SQLiteStore implements Store using SQLite.
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore opens (or creates) a SQLite database and runs migrations.
func NewSQLiteStore(path string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", path+"?_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return &SQLiteStore{db: db}, nil
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

// --- Game ---

func (s *SQLiteStore) CreateGame(ctx context.Context, game *model.Game) error {
	now := time.Now().UTC()
	game.CreatedAt = now
	game.UpdatedAt = now
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO games (id, name, created_at, updated_at) VALUES (?, ?, ?, ?)`,
		game.ID, game.Name, now.Format(time.RFC3339), now.Format(time.RFC3339))
	return err
}

func (s *SQLiteStore) GetGame(ctx context.Context, id string) (*model.Game, error) {
	g := &model.Game{}
	var createdAt, updatedAt string
	err := s.db.QueryRowContext(ctx, `SELECT id, name, created_at, updated_at FROM games WHERE id = ?`, id).
		Scan(&g.ID, &g.Name, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("game not found: %s", id)
	}
	if err != nil {
		return nil, err
	}
	g.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	g.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	return g, nil
}

func (s *SQLiteStore) ListGames(ctx context.Context) ([]*model.Game, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, name, created_at, updated_at FROM games ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var games []*model.Game
	for rows.Next() {
		g := &model.Game{}
		var createdAt, updatedAt string
		if err := rows.Scan(&g.ID, &g.Name, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		g.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		g.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
		games = append(games, g)
	}
	return games, rows.Err()
}

func (s *SQLiteStore) DeleteGame(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM games WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("game not found: %s", id)
	}
	return nil
}

// --- Character ---

func (s *SQLiteStore) CreateCharacter(ctx context.Context, ch *model.Character) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO characters (id, game_id, name, edge, heart, iron, shadow, wits,
		  health, spirit, supply, momentum, momentum_max, momentum_reset,
		  wounded, shaken, unprepared, encumbered, maimed, corrupted, cursed, tormented,
		  experience_earned, experience_spent)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		ch.ID, ch.GameID, ch.Name,
		ch.Stats.Edge, ch.Stats.Heart, ch.Stats.Iron, ch.Stats.Shadow, ch.Stats.Wits,
		ch.Health, ch.Spirit, ch.Supply,
		ch.Momentum, ch.MomentumMax, ch.MomentumReset,
		boolToInt(ch.Debilities.Wounded), boolToInt(ch.Debilities.Shaken),
		boolToInt(ch.Debilities.Unprepared), boolToInt(ch.Debilities.Encumbered),
		boolToInt(ch.Debilities.Maimed), boolToInt(ch.Debilities.Corrupted),
		boolToInt(ch.Debilities.Cursed), boolToInt(ch.Debilities.Tormented),
		ch.ExperienceEarned, ch.ExperienceSpent)
	return err
}

func (s *SQLiteStore) GetCharacter(ctx context.Context, gameID string) (*model.Character, error) {
	ch := &model.Character{}
	var w, sh, u, e, m, co, cu, t int
	err := s.db.QueryRowContext(ctx,
		`SELECT id, game_id, name, edge, heart, iron, shadow, wits,
		  health, spirit, supply, momentum, momentum_max, momentum_reset,
		  wounded, shaken, unprepared, encumbered, maimed, corrupted, cursed, tormented,
		  experience_earned, experience_spent
		 FROM characters WHERE game_id = ?`, gameID).
		Scan(&ch.ID, &ch.GameID, &ch.Name,
			&ch.Stats.Edge, &ch.Stats.Heart, &ch.Stats.Iron, &ch.Stats.Shadow, &ch.Stats.Wits,
			&ch.Health, &ch.Spirit, &ch.Supply,
			&ch.Momentum, &ch.MomentumMax, &ch.MomentumReset,
			&w, &sh, &u, &e, &m, &co, &cu, &t,
			&ch.ExperienceEarned, &ch.ExperienceSpent)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("character not found for game: %s", gameID)
	}
	if err != nil {
		return nil, err
	}
	ch.Debilities = model.Debilities{
		Wounded: w == 1, Shaken: sh == 1, Unprepared: u == 1, Encumbered: e == 1,
		Maimed: m == 1, Corrupted: co == 1, Cursed: cu == 1, Tormented: t == 1,
	}
	return ch, nil
}

func (s *SQLiteStore) UpdateCharacter(ctx context.Context, ch *model.Character) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE characters SET name=?, edge=?, heart=?, iron=?, shadow=?, wits=?,
		  health=?, spirit=?, supply=?, momentum=?, momentum_max=?, momentum_reset=?,
		  wounded=?, shaken=?, unprepared=?, encumbered=?, maimed=?, corrupted=?, cursed=?, tormented=?,
		  experience_earned=?, experience_spent=?
		 WHERE id = ?`,
		ch.Name, ch.Stats.Edge, ch.Stats.Heart, ch.Stats.Iron, ch.Stats.Shadow, ch.Stats.Wits,
		ch.Health, ch.Spirit, ch.Supply,
		ch.Momentum, ch.MomentumMax, ch.MomentumReset,
		boolToInt(ch.Debilities.Wounded), boolToInt(ch.Debilities.Shaken),
		boolToInt(ch.Debilities.Unprepared), boolToInt(ch.Debilities.Encumbered),
		boolToInt(ch.Debilities.Maimed), boolToInt(ch.Debilities.Corrupted),
		boolToInt(ch.Debilities.Cursed), boolToInt(ch.Debilities.Tormented),
		ch.ExperienceEarned, ch.ExperienceSpent,
		ch.ID)
	return err
}

// --- Progress Tracks ---

func (s *SQLiteStore) CreateTrack(ctx context.Context, track *model.ProgressTrack) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO progress_tracks (id, game_id, name, track_type, rank, ticks, completed)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		track.ID, track.GameID, track.Name, string(track.TrackType), string(track.Rank),
		track.Ticks, boolToInt(track.Completed))
	return err
}

func (s *SQLiteStore) GetTrack(ctx context.Context, id string) (*model.ProgressTrack, error) {
	t := &model.ProgressTrack{}
	var completed int
	err := s.db.QueryRowContext(ctx,
		`SELECT id, game_id, name, track_type, rank, ticks, completed FROM progress_tracks WHERE id = ?`, id).
		Scan(&t.ID, &t.GameID, &t.Name, &t.TrackType, &t.Rank, &t.Ticks, &completed)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("progress track not found: %s", id)
	}
	if err != nil {
		return nil, err
	}
	t.Completed = completed == 1
	return t, nil
}

func (s *SQLiteStore) ListTracks(ctx context.Context, gameID string, trackType string, completed *bool) ([]*model.ProgressTrack, error) {
	query := `SELECT id, game_id, name, track_type, rank, ticks, completed FROM progress_tracks WHERE game_id = ?`
	args := []interface{}{gameID}

	if trackType != "" {
		query += ` AND track_type = ?`
		args = append(args, trackType)
	}
	if completed != nil {
		query += ` AND completed = ?`
		args = append(args, boolToInt(*completed))
	}

	query += ` ORDER BY name`

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tracks []*model.ProgressTrack
	for rows.Next() {
		t := &model.ProgressTrack{}
		var comp int
		if err := rows.Scan(&t.ID, &t.GameID, &t.Name, &t.TrackType, &t.Rank, &t.Ticks, &comp); err != nil {
			return nil, err
		}
		t.Completed = comp == 1
		tracks = append(tracks, t)
	}
	return tracks, rows.Err()
}

func (s *SQLiteStore) UpdateTrack(ctx context.Context, track *model.ProgressTrack) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE progress_tracks SET name=?, ticks=?, completed=? WHERE id = ?`,
		track.Name, track.Ticks, boolToInt(track.Completed), track.ID)
	return err
}

func (s *SQLiteStore) DeleteTrack(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM progress_tracks WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("progress track not found: %s", id)
	}
	return nil
}

// --- Assets ---

func (s *SQLiteStore) AddAsset(ctx context.Context, asset *model.CharacterAsset) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO character_assets (id, character_id, asset_id, name, ability1, ability2, ability3, health)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		asset.ID, asset.CharacterID, asset.AssetID, asset.Name,
		boolToInt(asset.Ability1), boolToInt(asset.Ability2), boolToInt(asset.Ability3),
		asset.Health)
	return err
}

func (s *SQLiteStore) GetAsset(ctx context.Context, id string) (*model.CharacterAsset, error) {
	a := &model.CharacterAsset{}
	var a1, a2, a3 int
	err := s.db.QueryRowContext(ctx,
		`SELECT id, character_id, asset_id, name, ability1, ability2, ability3, health
		 FROM character_assets WHERE id = ?`, id).
		Scan(&a.ID, &a.CharacterID, &a.AssetID, &a.Name, &a1, &a2, &a3, &a.Health)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("asset not found: %s", id)
	}
	if err != nil {
		return nil, err
	}
	a.Ability1 = a1 == 1
	a.Ability2 = a2 == 1
	a.Ability3 = a3 == 1
	return a, nil
}

func (s *SQLiteStore) ListAssets(ctx context.Context, characterID string) ([]*model.CharacterAsset, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, character_id, asset_id, name, ability1, ability2, ability3, health
		 FROM character_assets WHERE character_id = ? ORDER BY name`, characterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assets []*model.CharacterAsset
	for rows.Next() {
		a := &model.CharacterAsset{}
		var a1, a2, a3 int
		if err := rows.Scan(&a.ID, &a.CharacterID, &a.AssetID, &a.Name, &a1, &a2, &a3, &a.Health); err != nil {
			return nil, err
		}
		a.Ability1 = a1 == 1
		a.Ability2 = a2 == 1
		a.Ability3 = a3 == 1
		assets = append(assets, a)
	}
	return assets, rows.Err()
}

func (s *SQLiteStore) UpdateAsset(ctx context.Context, asset *model.CharacterAsset) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE character_assets SET ability1=?, ability2=?, ability3=?, health=? WHERE id = ?`,
		boolToInt(asset.Ability1), boolToInt(asset.Ability2), boolToInt(asset.Ability3),
		asset.Health, asset.ID)
	return err
}

// --- Game Log ---

func (s *SQLiteStore) AppendLog(ctx context.Context, entry *model.LogEntry) error {
	details, err := marshalLogDetails(entry)
	if err != nil {
		return fmt.Errorf("marshal log details: %w", err)
	}
	tagsJSON, err := json.Marshal(entry.Tags)
	if err != nil {
		return fmt.Errorf("marshal tags: %w", err)
	}

	_, err = s.db.ExecContext(ctx,
		`INSERT INTO game_log (id, game_id, sequence, timestamp, entry_type, summary, details, tags)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		entry.ID, entry.GameID, entry.Sequence,
		entry.Timestamp.Format(time.RFC3339),
		entry.EntryType, entry.Summary, string(details), string(tagsJSON))
	return err
}

func (s *SQLiteStore) GetLogs(ctx context.Context, gameID string, filter model.LogFilter) ([]*model.LogEntry, error) {
	query := `SELECT id, game_id, sequence, timestamp, entry_type, summary, details, tags FROM game_log WHERE game_id = ?`
	args := []interface{}{gameID}

	if filter.EntryType != "" {
		query += ` AND entry_type = ?`
		args = append(args, filter.EntryType)
	}
	if filter.Outcome != "" {
		query += ` AND json_extract(details, '$.outcome') = ?`
		args = append(args, filter.Outcome)
	}
	if len(filter.Tags) > 0 {
		for _, tag := range filter.Tags {
			query += ` AND EXISTS (SELECT 1 FROM json_each(tags) WHERE json_each.value = ?)`
			args = append(args, tag)
		}
	}

	query += ` ORDER BY sequence ASC`

	if filter.Limit > 0 {
		query += fmt.Sprintf(` LIMIT %d`, filter.Limit)
	}
	if filter.Offset > 0 {
		query += fmt.Sprintf(` OFFSET %d`, filter.Offset)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*model.LogEntry
	for rows.Next() {
		e := &model.LogEntry{}
		var ts, detailsStr, tagsStr string
		if err := rows.Scan(&e.ID, &e.GameID, &e.Sequence, &ts, &e.EntryType, &e.Summary, &detailsStr, &tagsStr); err != nil {
			return nil, err
		}
		e.Timestamp, _ = time.Parse(time.RFC3339, ts)
		_ = json.Unmarshal([]byte(tagsStr), &e.Tags)
		unmarshalLogDetails(e, detailsStr)
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

func (s *SQLiteStore) GetNextSequence(ctx context.Context, gameID string) (int, error) {
	var seq int
	err := s.db.QueryRowContext(ctx,
		`SELECT COALESCE(MAX(sequence), 0) + 1 FROM game_log WHERE game_id = ?`, gameID).
		Scan(&seq)
	return seq, err
}

// --- Helpers ---

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

type logDetails struct {
	MoveID            string                   `json:"move_id,omitempty"`
	MoveCategory      string                   `json:"move_category,omitempty"`
	CharacterBefore   *model.CharacterSnapshot `json:"character_before,omitempty"`
	CharacterAfter    *model.CharacterSnapshot `json:"character_after,omitempty"`
	Roll              *model.ActionRoll        `json:"roll,omitempty"`
	ProgressRoll      *model.ProgressRollResult `json:"progress_roll,omitempty"`
	OracleResult      *model.OracleRollResult  `json:"oracle_result,omitempty"`
	Outcome           model.Outcome            `json:"outcome,omitempty"`
	NarrativeContext  string                   `json:"narrative_context,omitempty"`
	MechanicalEffects []string                 `json:"mechanical_effects,omitempty"`
	RelatedTrackIDs   []string                 `json:"related_track_ids,omitempty"`
	RelatedAssetIDs   []string                 `json:"related_asset_ids,omitempty"`
}

func marshalLogDetails(e *model.LogEntry) ([]byte, error) {
	d := logDetails{
		MoveID:            e.MoveID,
		MoveCategory:      e.MoveCategory,
		CharacterBefore:   e.CharacterBefore,
		CharacterAfter:    e.CharacterAfter,
		Roll:              e.Roll,
		ProgressRoll:      e.ProgressRoll,
		OracleResult:      e.OracleResult,
		Outcome:           e.Outcome,
		NarrativeContext:   e.NarrativeContext,
		MechanicalEffects: e.MechanicalEffects,
		RelatedTrackIDs:   e.RelatedTrackIDs,
		RelatedAssetIDs:   e.RelatedAssetIDs,
	}
	return json.Marshal(d)
}

func unmarshalLogDetails(e *model.LogEntry, data string) {
	if data == "" || data == "{}" {
		return
	}
	var d logDetails
	if err := json.Unmarshal([]byte(data), &d); err != nil {
		return
	}
	e.MoveID = d.MoveID
	e.MoveCategory = d.MoveCategory
	e.CharacterBefore = d.CharacterBefore
	e.CharacterAfter = d.CharacterAfter
	e.Roll = d.Roll
	e.ProgressRoll = d.ProgressRoll
	e.OracleResult = d.OracleResult
	e.Outcome = d.Outcome
	e.NarrativeContext = d.NarrativeContext
	e.MechanicalEffects = d.MechanicalEffects
	e.RelatedTrackIDs = d.RelatedTrackIDs
	e.RelatedAssetIDs = d.RelatedAssetIDs
}

// Ensure SQLiteStore implements Store at compile time.
var _ Store = (*SQLiteStore)(nil)
