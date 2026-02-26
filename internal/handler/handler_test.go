package handler_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ironsworn/ironsworn-backend/internal/engine"
	"github.com/ironsworn/ironsworn-backend/internal/handler"
	"github.com/ironsworn/ironsworn-backend/internal/model"
	"github.com/ironsworn/ironsworn-backend/internal/router"
	"github.com/ironsworn/ironsworn-backend/internal/store"
)

func setup(t *testing.T, roller engine.Roller) (*httptest.Server, store.Store) {
	t.Helper()
	s, err := store.NewSQLiteStore(":memory:")
	if err != nil {
		t.Fatalf("store: %v", err)
	}
	t.Cleanup(func() { s.Close() })

	h := handler.New(s, roller)
	r := router.New(h)
	ts := httptest.NewServer(r)
	t.Cleanup(ts.Close)
	return ts, s
}

func postJSON(t *testing.T, url string, body interface{}) *http.Response {
	t.Helper()
	data, _ := json.Marshal(body)
	resp, err := http.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		t.Fatalf("POST %s: %v", url, err)
	}
	return resp
}

func getJSON(t *testing.T, url string) *http.Response {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	return resp
}

func patchJSON(t *testing.T, url string, body interface{}) *http.Response {
	t.Helper()
	data, _ := json.Marshal(body)
	req, _ := http.NewRequest("PATCH", url, bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PATCH %s: %v", url, err)
	}
	return resp
}

func decodeBody(t *testing.T, resp *http.Response, dst interface{}) {
	t.Helper()
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(data, dst); err != nil {
		t.Fatalf("decode: %v (body: %s)", err, string(data))
	}
}

func TestFullVowLoop(t *testing.T) {
	// Use a fixed roller for deterministic testing
	roller := &engine.FixedRoller{
		D6Values:  []int{5, 5, 5, 5, 5, 5}, // All action rolls get 5
		D10Values: []int{2, 3, 2, 3, 2, 3, 2, 3, 2, 3, 2, 3}, // All challenge dice are 2, 3
	}
	ts, _ := setup(t, roller)
	baseURL := ts.URL + "/api"

	// 1. Create a game
	resp := postJSON(t, baseURL+"/games", map[string]string{"name": "Test Campaign"})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create game: %d", resp.StatusCode)
	}
	var game model.Game
	decodeBody(t, resp, &game)
	gameURL := baseURL + "/games/" + game.ID

	// 2. Create a character
	resp = postJSON(t, gameURL+"/character", map[string]interface{}{
		"name": "Kara",
		"stats": map[string]int{
			"edge": 2, "heart": 3, "iron": 1, "shadow": 2, "wits": 1,
		},
	})
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("create character: %d (%s)", resp.StatusCode, string(body))
	}
	var ch model.Character
	decodeBody(t, resp, &ch)

	// 3. Create a vow (progress track)
	resp = postJSON(t, gameURL+"/progress", map[string]interface{}{
		"name": "Protect the Village", "track_type": "vow", "rank": "troublesome",
	})
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("create vow: %d (%s)", resp.StatusCode, string(body))
	}
	var track model.ProgressTrack
	decodeBody(t, resp, &track)

	// 4. Swear an iron vow (move)
	resp = postJSON(t, gameURL+"/moves", map[string]interface{}{
		"move_id":   "swear_an_iron_vow",
		"narrative": "I swear to protect the people of Thornwall.",
	})
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("swear vow: %d (%s)", resp.StatusCode, string(body))
	}
	var moveResult model.MoveResult
	decodeBody(t, resp, &moveResult)
	if !moveResult.Outcome.IsStrongHit() {
		t.Errorf("expected strong hit, got %s", moveResult.Outcome)
	}

	// 5. Reach milestones
	for i := 0; i < 3; i++ {
		resp = postJSON(t, gameURL+"/moves", map[string]interface{}{
			"move_id":  "reach_a_milestone",
			"track_id": track.ID,
		})
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("milestone %d: %d (%s)", i+1, resp.StatusCode, string(body))
		}
	}

	// 6. Verify progress
	resp = getJSON(t, gameURL+"/progress?type=vow")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("list tracks: %d", resp.StatusCode)
	}
	var tracks []*model.ProgressTrack
	decodeBody(t, resp, &tracks)
	if len(tracks) != 1 {
		t.Fatalf("expected 1 track, got %d", len(tracks))
	}
	if tracks[0].Ticks != 36 { // 3 * 12 = 36 for troublesome
		t.Errorf("expected 36 ticks, got %d", tracks[0].Ticks)
	}

	// 7. Fulfill the vow
	resp = postJSON(t, gameURL+"/moves", map[string]interface{}{
		"move_id":  "fulfill_your_vow",
		"track_id": track.ID,
	})
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("fulfill: %d (%s)", resp.StatusCode, string(body))
	}

	// 8. Verify game log has entries
	resp = getJSON(t, gameURL+"/log")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("get log: %d", resp.StatusCode)
	}
	var logs []*model.LogEntry
	decodeBody(t, resp, &logs)
	if len(logs) == 0 {
		t.Error("expected log entries")
	}

	// 9. Check character XP
	resp = getJSON(t, gameURL+"/character")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("get character: %d", resp.StatusCode)
	}
	decodeBody(t, resp, &ch)
	if ch.ExperienceEarned < 1 {
		t.Errorf("expected at least 1 XP, got %d", ch.ExperienceEarned)
	}
}

func TestCreateGame_ListGames(t *testing.T) {
	roller := &engine.FixedRoller{}
	ts, _ := setup(t, roller)
	baseURL := ts.URL + "/api"

	// Create two games
	postJSON(t, baseURL+"/games", map[string]string{"name": "Campaign 1"})
	postJSON(t, baseURL+"/games", map[string]string{"name": "Campaign 2"})

	resp := getJSON(t, baseURL+"/games")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("list games: %d", resp.StatusCode)
	}
	var games []*model.Game
	decodeBody(t, resp, &games)
	if len(games) != 2 {
		t.Errorf("expected 2 games, got %d", len(games))
	}
}

func TestInvalidStatDistribution(t *testing.T) {
	roller := &engine.FixedRoller{}
	ts, _ := setup(t, roller)
	baseURL := ts.URL + "/api"

	resp := postJSON(t, baseURL+"/games", map[string]string{"name": "Test"})
	var game model.Game
	decodeBody(t, resp, &game)

	// Try to create character with invalid stats
	resp = postJSON(t, baseURL+"/games/"+game.ID+"/character", map[string]interface{}{
		"name": "Bad Stats",
		"stats": map[string]int{
			"edge": 3, "heart": 3, "iron": 1, "shadow": 1, "wits": 1,
		},
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestPatchCharacter(t *testing.T) {
	roller := &engine.FixedRoller{}
	ts, _ := setup(t, roller)
	baseURL := ts.URL + "/api"

	resp := postJSON(t, baseURL+"/games", map[string]string{"name": "Test"})
	var game model.Game
	decodeBody(t, resp, &game)

	postJSON(t, baseURL+"/games/"+game.ID+"/character", map[string]interface{}{
		"name": "Kara",
		"stats": map[string]int{
			"edge": 2, "heart": 3, "iron": 1, "shadow": 2, "wits": 1,
		},
	})

	health := 3
	wounded := true
	resp = patchJSON(t, baseURL+"/games/"+game.ID+"/character", map[string]interface{}{
		"health":  health,
		"wounded": wounded,
	})
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("patch: %d (%s)", resp.StatusCode, string(body))
	}
	var ch model.Character
	decodeBody(t, resp, &ch)
	if ch.Health != 3 {
		t.Errorf("expected health 3, got %d", ch.Health)
	}
	if !ch.Debilities.Wounded {
		t.Error("expected wounded")
	}
	// Momentum max should decrease by 1
	if ch.MomentumMax != 9 {
		t.Errorf("expected momentum max 9, got %d", ch.MomentumMax)
	}
}

func TestListMoves(t *testing.T) {
	roller := &engine.FixedRoller{}
	ts, _ := setup(t, roller)

	resp := getJSON(t, ts.URL+"/api/moves")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("list moves: %d", resp.StatusCode)
	}
	var moves []model.MoveDefinition
	decodeBody(t, resp, &moves)
	if len(moves) == 0 {
		t.Error("expected moves")
	}
}

func TestOracleTables(t *testing.T) {
	roller := &engine.FixedRoller{}
	ts, _ := setup(t, roller)

	resp := getJSON(t, ts.URL+"/api/oracle/tables")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("list oracle tables: %d", resp.StatusCode)
	}
	var tables []model.OracleTable
	decodeBody(t, resp, &tables)
	if len(tables) == 0 {
		t.Error("expected oracle tables")
	}
}

func TestNarrativeLog(t *testing.T) {
	roller := &engine.FixedRoller{}
	ts, _ := setup(t, roller)
	baseURL := ts.URL + "/api"

	resp := postJSON(t, baseURL+"/games", map[string]string{"name": "Test"})
	var game model.Game
	decodeBody(t, resp, &game)

	resp = postJSON(t, baseURL+"/games/"+game.ID+"/log", map[string]interface{}{
		"text": "The wind howled across the Hinterlands as Kara set out on her journey.",
		"tags": []string{"narrative", "journey_start"},
	})
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("add narrative: %d (%s)", resp.StatusCode, string(body))
	}

	resp = getJSON(t, baseURL+"/games/"+game.ID+"/log")
	var logs []*model.LogEntry
	decodeBody(t, resp, &logs)
	if len(logs) != 1 {
		t.Errorf("expected 1 log, got %d", len(logs))
	}
	if logs[0].EntryType != "narrative" {
		t.Errorf("expected narrative type, got %s", logs[0].EntryType)
	}
}

func TestJSONLExport(t *testing.T) {
	roller := &engine.FixedRoller{D6Values: []int{5}, D10Values: []int{2, 3}}
	ts, _ := setup(t, roller)
	baseURL := ts.URL + "/api"

	resp := postJSON(t, baseURL+"/games", map[string]string{"name": "Test"})
	var game model.Game
	decodeBody(t, resp, &game)

	postJSON(t, baseURL+"/games/"+game.ID+"/character", map[string]interface{}{
		"name": "Kara", "stats": map[string]int{"edge": 2, "heart": 3, "iron": 1, "shadow": 2, "wits": 1},
	})

	// Make a move to generate a log entry
	postJSON(t, baseURL+"/games/"+game.ID+"/moves", map[string]interface{}{
		"move_id": "face_danger", "stat": "edge",
	})

	resp = getJSON(t, baseURL+"/games/"+game.ID+"/log/export")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("export: %d", resp.StatusCode)
	}
	contentType := resp.Header.Get("Content-Type")
	if contentType != "application/x-ndjson" {
		t.Errorf("expected ndjson content type, got %s", contentType)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if len(body) == 0 {
		t.Error("expected non-empty JSONL export")
	}
}
