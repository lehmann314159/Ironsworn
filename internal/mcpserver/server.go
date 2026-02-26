package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/ironsworn/ironsworn/internal/engine"
	"github.com/ironsworn/ironsworn/internal/gamelog"
	"github.com/ironsworn/ironsworn/internal/model"
	"github.com/ironsworn/ironsworn/internal/oracle"
	"github.com/ironsworn/ironsworn/internal/store"
)

// Server wraps the MCP server with Ironsworn game logic.
type Server struct {
	mcp    *server.MCPServer
	store  store.Store
	moves  *engine.MoveRegistry
	roller engine.Roller
	logger *gamelog.Logger
}

// New creates a new Ironsworn MCP server.
func New(s store.Store) *Server {
	srv := &Server{
		store:  s,
		moves:  engine.NewMoveRegistry(),
		roller: engine.CryptoRoller{},
		logger: gamelog.NewLogger(s),
	}

	srv.mcp = server.NewMCPServer(
		"ironsworn",
		"1.0.0",
		server.WithToolCapabilities(true),
		server.WithResourceCapabilities(true, false),
	)

	srv.registerTools()
	srv.registerResources()

	return srv
}

// MCPServer returns the underlying MCP server for use with ServeStdio.
func (s *Server) MCPServer() *server.MCPServer {
	return s.mcp
}

func jsonResult(v interface{}) (*mcp.CallToolResult, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, err
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{Type: "text", Text: string(data)},
		},
	}, nil
}

func errorResult(msg string) (*mcp.CallToolResult, error) {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{Type: "text", Text: msg},
		},
		IsError: true,
	}, nil
}

func (s *Server) registerTools() {
	// create_game
	s.mcp.AddTool(mcp.NewTool("create_game",
		mcp.WithDescription("Create a new Ironsworn game/campaign"),
		mcp.WithString("name", mcp.Description("Name of the campaign"), mcp.Required()),
	), s.handleCreateGame)

	// create_character
	s.mcp.AddTool(mcp.NewTool("create_character",
		mcp.WithDescription("Create a character with name and stat distribution (3,2,2,1,1)"),
		mcp.WithString("game_id", mcp.Description("Game ID"), mcp.Required()),
		mcp.WithString("name", mcp.Description("Character name"), mcp.Required()),
		mcp.WithNumber("edge", mcp.Description("Edge stat (1-3)"), mcp.Required()),
		mcp.WithNumber("heart", mcp.Description("Heart stat (1-3)"), mcp.Required()),
		mcp.WithNumber("iron", mcp.Description("Iron stat (1-3)"), mcp.Required()),
		mcp.WithNumber("shadow", mcp.Description("Shadow stat (1-3)"), mcp.Required()),
		mcp.WithNumber("wits", mcp.Description("Wits stat (1-3)"), mcp.Required()),
	), s.handleCreateCharacter)

	// get_character
	s.mcp.AddTool(mcp.NewTool("get_character",
		mcp.WithDescription("Get current character state (stats, tracks, momentum, debilities)"),
		mcp.WithString("game_id", mcp.Description("Game ID"), mcp.Required()),
	), s.handleGetCharacter)

	// make_move
	s.mcp.AddTool(mcp.NewTool("make_move",
		mcp.WithDescription("Execute any Ironsworn move (face_danger, strike, swear_an_iron_vow, etc.)"),
		mcp.WithString("game_id", mcp.Description("Game ID"), mcp.Required()),
		mcp.WithString("move_id", mcp.Description("Move ID (e.g., face_danger, strike, swear_an_iron_vow)"), mcp.Required()),
		mcp.WithString("stat", mcp.Description("Stat to roll (edge, heart, iron, shadow, wits)")),
		mcp.WithNumber("adds", mcp.Description("Bonus adds to the roll")),
		mcp.WithString("track_id", mcp.Description("Progress track ID (for progress moves)")),
		mcp.WithNumber("amount", mcp.Description("Amount for suffer moves (harm/stress)")),
		mcp.WithString("narrative", mcp.Description("Player narrative context for the fiction")),
	), s.handleMakeMove)

	// burn_momentum
	s.mcp.AddTool(mcp.NewTool("burn_momentum",
		mcp.WithDescription("Burn momentum to improve the last roll's outcome"),
		mcp.WithString("game_id", mcp.Description("Game ID"), mcp.Required()),
	), s.handleBurnMomentum)

	// list_vows
	s.mcp.AddTool(mcp.NewTool("list_vows",
		mcp.WithDescription("List active vows with progress"),
		mcp.WithString("game_id", mcp.Description("Game ID"), mcp.Required()),
		mcp.WithBoolean("include_completed", mcp.Description("Include completed vows")),
	), s.handleListVows)

	// create_vow
	s.mcp.AddTool(mcp.NewTool("create_vow",
		mcp.WithDescription("Swear an iron vow (creates progress track)"),
		mcp.WithString("game_id", mcp.Description("Game ID"), mcp.Required()),
		mcp.WithString("name", mcp.Description("Vow description"), mcp.Required()),
		mcp.WithString("rank", mcp.Description("Rank: troublesome, dangerous, formidable, extreme, epic"), mcp.Required()),
	), s.handleCreateVow)

	// mark_progress
	s.mcp.AddTool(mcp.NewTool("mark_progress",
		mcp.WithDescription("Mark progress on a track"),
		mcp.WithString("track_id", mcp.Description("Progress track ID"), mcp.Required()),
		mcp.WithNumber("count", mcp.Description("Number of times to mark progress (default 1)")),
	), s.handleMarkProgress)

	// ask_oracle
	s.mcp.AddTool(mcp.NewTool("ask_oracle",
		mcp.WithDescription("Ask a yes/no question with likelihood"),
		mcp.WithString("game_id", mcp.Description("Game ID"), mcp.Required()),
		mcp.WithString("question", mcp.Description("The yes/no question"), mcp.Required()),
		mcp.WithString("likelihood", mcp.Description("almost_certain, likely, fifty_fifty, unlikely, small_chance")),
	), s.handleAskOracle)

	// roll_oracle_table
	s.mcp.AddTool(mcp.NewTool("roll_oracle_table",
		mcp.WithDescription("Roll on a named oracle table"),
		mcp.WithString("game_id", mcp.Description("Game ID"), mcp.Required()),
		mcp.WithString("table_id", mcp.Description("Oracle table ID"), mcp.Required()),
	), s.handleRollOracleTable)

	// get_game_log
	s.mcp.AddTool(mcp.NewTool("get_game_log",
		mcp.WithDescription("Retrieve recent game log entries"),
		mcp.WithString("game_id", mcp.Description("Game ID"), mcp.Required()),
		mcp.WithNumber("limit", mcp.Description("Number of entries to return (default 20)")),
	), s.handleGetGameLog)

	// add_narrative
	s.mcp.AddTool(mcp.NewTool("add_narrative",
		mcp.WithDescription("Add narrative/fiction text to the game log"),
		mcp.WithString("game_id", mcp.Description("Game ID"), mcp.Required()),
		mcp.WithString("text", mcp.Description("Narrative text"), mcp.Required()),
	), s.handleAddNarrative)

	// list_moves
	s.mcp.AddTool(mcp.NewTool("list_moves",
		mcp.WithDescription("List all available Ironsworn moves with descriptions"),
	), s.handleListMoves)

	// get_move_details
	s.mcp.AddTool(mcp.NewTool("get_move_details",
		mcp.WithDescription("Get detailed rules text for a specific move"),
		mcp.WithString("move_id", mcp.Description("Move ID"), mcp.Required()),
	), s.handleGetMoveDetails)
}

func (s *Server) registerResources() {
	// Character state resource
	s.mcp.AddResourceTemplate(
		mcp.NewResourceTemplate("ironsworn://character/{gameID}", "Current character state"),
		s.handleCharacterResource,
	)

	// Vows resource
	s.mcp.AddResourceTemplate(
		mcp.NewResourceTemplate("ironsworn://vows/{gameID}", "Active vows and progress"),
		s.handleVowsResource,
	)

	// Game log resource
	s.mcp.AddResourceTemplate(
		mcp.NewResourceTemplate("ironsworn://log/{gameID}", "Recent game log"),
		s.handleLogResource,
	)

	// Move rules resource
	s.mcp.AddResourceTemplate(
		mcp.NewResourceTemplate("ironsworn://rules/moves/{moveID}", "Move rules reference"),
		s.handleMoveRulesResource,
	)

	// Oracle tables resource
	s.mcp.AddResource(
		mcp.NewResource("ironsworn://oracle/tables", "Available oracle tables"),
		s.handleOracleTablesResource,
	)
}

// --- Tool Handlers ---

func (s *Server) handleCreateGame(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name := req.GetString("name", "")
	game := &model.Game{ID: gamelog.GenerateID(), Name: name}
	if err := s.store.CreateGame(ctx, game); err != nil {
		return errorResult(err.Error())
	}
	return jsonResult(game)
}

func (s *Server) handleCreateCharacter(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	gameID := req.GetString("game_id", "")
	name := req.GetString("name", "")
	stats := model.Stats{
		Edge:   int(req.GetFloat("edge", 0)),
		Heart:  int(req.GetFloat("heart", 0)),
		Iron:   int(req.GetFloat("iron", 0)),
		Shadow: int(req.GetFloat("shadow", 0)),
		Wits:   int(req.GetFloat("wits", 0)),
	}

	if err := engine.ValidateStatDistribution(stats); err != nil {
		return errorResult(err.Error())
	}

	ch, err := engine.CreateCharacter(gamelog.GenerateID(), gameID, name, stats)
	if err != nil {
		return errorResult(err.Error())
	}
	if err := s.store.CreateCharacter(ctx, ch); err != nil {
		return errorResult(err.Error())
	}
	return jsonResult(ch)
}

func (s *Server) handleGetCharacter(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	gameID := req.GetString("game_id", "")
	ch, err := s.store.GetCharacter(ctx, gameID)
	if err != nil {
		return errorResult(err.Error())
	}
	return jsonResult(ch)
}

func (s *Server) handleMakeMove(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	gameID := req.GetString("game_id", "")
	moveReq := model.MoveRequest{
		MoveID:    req.GetString("move_id", ""),
		Stat:      req.GetString("stat", ""),
		Adds:      int(req.GetFloat("adds", 0)),
		TrackID:   req.GetString("track_id", ""),
		Amount:    int(req.GetFloat("amount", 0)),
		Narrative: req.GetString("narrative", ""),
	}

	ch, err := s.store.GetCharacter(ctx, gameID)
	if err != nil {
		return errorResult("Character not found: " + err.Error())
	}

	before := ch.Snapshot()

	tracks, err := s.store.ListTracks(ctx, gameID, "", nil)
	if err != nil {
		return errorResult(err.Error())
	}

	result, err := s.moves.Execute(s.roller, ch, moveReq, tracks)
	if err != nil {
		return errorResult(err.Error())
	}

	after := ch.Snapshot()

	if err := s.store.UpdateCharacter(ctx, ch); err != nil {
		return errorResult(err.Error())
	}
	for _, t := range tracks {
		s.store.UpdateTrack(ctx, t)
	}

	var trackIDs []string
	if moveReq.TrackID != "" {
		trackIDs = []string{moveReq.TrackID}
	}
	s.logger.LogMove(ctx, gameID, before, after, result, moveReq.Narrative, trackIDs)

	return jsonResult(result)
}

func (s *Server) handleBurnMomentum(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	gameID := req.GetString("game_id", "")
	ch, err := s.store.GetCharacter(ctx, gameID)
	if err != nil {
		return errorResult(err.Error())
	}

	logs, err := s.store.GetLogs(ctx, gameID, model.LogFilter{EntryType: "move", Limit: 1})
	if err != nil || len(logs) == 0 {
		return errorResult("No previous move to burn momentum on")
	}

	lastLog := logs[len(logs)-1]
	if lastLog.Roll == nil {
		return errorResult("Last move has no action roll")
	}

	before := ch.Snapshot()
	newOutcome, err := engine.BurnMomentum(ch, lastLog.Roll)
	if err != nil {
		return errorResult(err.Error())
	}

	s.store.UpdateCharacter(ctx, ch)
	after := ch.Snapshot()
	s.logger.LogStateChange(ctx, gameID, before, after, "Burned momentum: "+string(newOutcome))

	return jsonResult(map[string]interface{}{
		"new_outcome":    newOutcome,
		"momentum_reset": ch.Momentum,
		"character":      ch,
	})
}

func (s *Server) handleListVows(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	gameID := req.GetString("game_id", "")
	includeCompleted := req.GetBool("include_completed", false)

	var completed *bool
	if !includeCompleted {
		f := false
		completed = &f
	}

	tracks, err := s.store.ListTracks(ctx, gameID, "vow", completed)
	if err != nil {
		return errorResult(err.Error())
	}

	type vowInfo struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Rank  string `json:"rank"`
		Ticks int    `json:"ticks"`
		Score int    `json:"score"`
		Done  bool   `json:"completed"`
	}
	vows := make([]vowInfo, len(tracks))
	for i, t := range tracks {
		vows[i] = vowInfo{
			ID: t.ID, Name: t.Name, Rank: string(t.Rank),
			Ticks: t.Ticks, Score: t.Score(), Done: t.Completed,
		}
	}
	return jsonResult(vows)
}

func (s *Server) handleCreateVow(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	gameID := req.GetString("game_id", "")
	name := req.GetString("name", "")
	rank := model.ProgressRank(req.GetString("rank", ""))

	track, err := engine.CreateProgressTrack(gamelog.GenerateID(), gameID, name, model.TrackVow, rank)
	if err != nil {
		return errorResult(err.Error())
	}
	if err := s.store.CreateTrack(ctx, track); err != nil {
		return errorResult(err.Error())
	}

	// Also execute swear an iron vow move
	ch, err := s.store.GetCharacter(ctx, gameID)
	if err != nil {
		return jsonResult(track) // Return track even if character doesn't exist yet
	}

	before := ch.Snapshot()
	result, err := s.moves.Execute(s.roller, ch, model.MoveRequest{MoveID: "swear_an_iron_vow"}, nil)
	if err != nil {
		return jsonResult(track)
	}
	after := ch.Snapshot()
	s.store.UpdateCharacter(ctx, ch)
	s.logger.LogMove(ctx, gameID, before, after, result, "", []string{track.ID})

	return jsonResult(map[string]interface{}{
		"track":       track,
		"move_result": result,
	})
}

func (s *Server) handleMarkProgress(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	trackID := req.GetString("track_id", "")
	count := int(req.GetFloat("count", 1))
	if count <= 0 {
		count = 1
	}

	track, err := s.store.GetTrack(ctx, trackID)
	if err != nil {
		return errorResult(err.Error())
	}

	totalTicks := 0
	for i := 0; i < count; i++ {
		ticks, err := engine.MarkProgress(track)
		if err != nil {
			return errorResult(err.Error())
		}
		totalTicks += ticks
	}

	s.store.UpdateTrack(ctx, track)

	return jsonResult(map[string]interface{}{
		"track":       track,
		"ticks_added": totalTicks,
		"score":       track.Score(),
	})
}

func (s *Server) handleAskOracle(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	gameID := req.GetString("game_id", "")
	question := req.GetString("question", "")
	likelihood := model.Likelihood(req.GetString("likelihood", "fifty_fifty"))

	result := oracle.AskYesNo(s.roller, likelihood)

	answer := "No"
	if result.Answer {
		answer = "Yes"
	}
	summary := fmt.Sprintf("Asked: %s (%s) → %s (roll: %d, threshold: %d)",
		question, likelihood, answer, result.Roll, result.Threshold)

	oracleResult := &model.OracleRollResult{
		Roll:   result.Roll,
		Result: answer,
	}
	s.logger.LogOracle(ctx, gameID, oracleResult, summary)

	return jsonResult(map[string]interface{}{
		"question":   question,
		"likelihood": likelihood,
		"roll":       result.Roll,
		"threshold":  result.Threshold,
		"answer":     result.Answer,
		"summary":    summary,
	})
}

func (s *Server) handleRollOracleTable(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	gameID := req.GetString("game_id", "")
	tableID := req.GetString("table_id", "")

	result, err := oracle.RollTable(s.roller, tableID)
	if err != nil {
		return errorResult(err.Error())
	}

	summary := fmt.Sprintf("Rolled on %s: %s (roll: %d)", result.TableName, result.Result, result.Roll)
	s.logger.LogOracle(ctx, gameID, result, summary)

	return jsonResult(result)
}

func (s *Server) handleGetGameLog(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	gameID := req.GetString("game_id", "")
	limit := int(req.GetFloat("limit", 20))

	entries, err := s.store.GetLogs(ctx, gameID, model.LogFilter{Limit: limit})
	if err != nil {
		return errorResult(err.Error())
	}
	return jsonResult(entries)
}

func (s *Server) handleAddNarrative(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	gameID := req.GetString("game_id", "")
	text := req.GetString("text", "")

	if err := s.logger.LogNarrative(ctx, gameID, text, nil); err != nil {
		return errorResult(err.Error())
	}
	return jsonResult(map[string]string{"status": "logged"})
}

func (s *Server) handleListMoves(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	defs := s.moves.ListDefinitions()
	type moveSummary struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Category string `json:"category"`
		Desc     string `json:"description"`
	}
	summaries := make([]moveSummary, len(defs))
	for i, d := range defs {
		summaries[i] = moveSummary{
			ID: d.ID, Name: d.Name,
			Category: string(d.Category), Desc: d.Description,
		}
	}
	return jsonResult(summaries)
}

func (s *Server) handleGetMoveDetails(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	moveID := req.GetString("move_id", "")
	def, ok := s.moves.GetDefinition(moveID)
	if !ok {
		return errorResult("Move not found: " + moveID)
	}
	return jsonResult(def)
}

// --- Resource Handlers ---

func (s *Server) handleCharacterResource(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	gameID := extractURIParam(req.Params.URI, "ironsworn://character/")
	ch, err := s.store.GetCharacter(ctx, gameID)
	if err != nil {
		return nil, err
	}
	data, _ := json.MarshalIndent(ch, "", "  ")
	return []mcp.ResourceContents{
		mcp.TextResourceContents{URI: req.Params.URI, MIMEType: "application/json", Text: string(data)},
	}, nil
}

func (s *Server) handleVowsResource(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	gameID := extractURIParam(req.Params.URI, "ironsworn://vows/")
	f := false
	tracks, err := s.store.ListTracks(ctx, gameID, "vow", &f)
	if err != nil {
		return nil, err
	}
	data, _ := json.MarshalIndent(tracks, "", "  ")
	return []mcp.ResourceContents{
		mcp.TextResourceContents{URI: req.Params.URI, MIMEType: "application/json", Text: string(data)},
	}, nil
}

func (s *Server) handleLogResource(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	gameID := extractURIParam(req.Params.URI, "ironsworn://log/")
	entries, err := s.store.GetLogs(ctx, gameID, model.LogFilter{Limit: 20})
	if err != nil {
		return nil, err
	}
	data, _ := json.MarshalIndent(entries, "", "  ")
	return []mcp.ResourceContents{
		mcp.TextResourceContents{URI: req.Params.URI, MIMEType: "application/json", Text: string(data)},
	}, nil
}

func (s *Server) handleMoveRulesResource(_ context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	moveID := extractURIParam(req.Params.URI, "ironsworn://rules/moves/")
	def, ok := s.moves.GetDefinition(moveID)
	if !ok {
		return nil, fmt.Errorf("move not found: %s", moveID)
	}
	data, _ := json.MarshalIndent(def, "", "  ")
	return []mcp.ResourceContents{
		mcp.TextResourceContents{URI: req.Params.URI, MIMEType: "application/json", Text: string(data)},
	}, nil
}

func (s *Server) handleOracleTablesResource(_ context.Context, _ mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	tables := oracle.ListTables()
	type tableSummary struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Category string `json:"category"`
		Entries  int    `json:"entry_count"`
	}
	summaries := make([]tableSummary, len(tables))
	for i, t := range tables {
		summaries[i] = tableSummary{
			ID: t.ID, Name: t.Name, Category: t.Category, Entries: len(t.Entries),
		}
	}
	data, _ := json.MarshalIndent(summaries, "", "  ")
	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      "ironsworn://oracle/tables",
			MIMEType: "application/json",
			Text:     string(data),
		},
	}, nil
}

func extractURIParam(uri, prefix string) string {
	if len(uri) > len(prefix) {
		return uri[len(prefix):]
	}
	return ""
}
