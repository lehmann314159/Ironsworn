package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ironsworn/ironsworn-backend/internal/engine"
	"github.com/ironsworn/ironsworn-backend/internal/model"
)

func (h *Handler) ExecuteMove(w http.ResponseWriter, r *http.Request) {
	gameID := chi.URLParam(r, "gameID")

	var req model.MoveRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.MoveID == "" {
		respondError(w, http.StatusBadRequest, "move_id is required")
		return
	}

	// Load character
	ch, err := h.Store.GetCharacter(r.Context(), gameID)
	if err != nil {
		respondError(w, http.StatusNotFound, "character not found")
		return
	}

	before := ch.Snapshot()

	// Load tracks (some moves need them)
	tracks, err := h.Store.ListTracks(r.Context(), gameID, "", nil)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Execute the move
	result, err := h.Moves.Execute(h.Roller, ch, req, tracks)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	after := ch.Snapshot()

	// Persist character changes
	if err := h.Store.UpdateCharacter(r.Context(), ch); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Persist any track changes
	for _, t := range tracks {
		if err := h.Store.UpdateTrack(r.Context(), t); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	// Log the move
	var trackIDs []string
	if req.TrackID != "" {
		trackIDs = []string{req.TrackID}
	}
	h.Logger.LogMove(r.Context(), gameID, before, after, result, req.Narrative, trackIDs)

	respondJSON(w, http.StatusOK, result)
}

func (h *Handler) BurnMomentum(w http.ResponseWriter, r *http.Request) {
	gameID := chi.URLParam(r, "gameID")

	ch, err := h.Store.GetCharacter(r.Context(), gameID)
	if err != nil {
		respondError(w, http.StatusNotFound, "character not found")
		return
	}

	// Get the last move from the log to find the roll
	logs, err := h.Store.GetLogs(r.Context(), gameID, model.LogFilter{EntryType: "move", Limit: 1, Offset: 0})
	if err != nil || len(logs) == 0 {
		respondError(w, http.StatusBadRequest, "no previous move to burn momentum on")
		return
	}

	lastLog := logs[len(logs)-1]
	if lastLog.Roll == nil {
		respondError(w, http.StatusBadRequest, "last move has no action roll (cannot burn momentum on progress rolls)")
		return
	}

	before := ch.Snapshot()
	newOutcome, err := engine.BurnMomentum(ch, lastLog.Roll)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.Store.UpdateCharacter(r.Context(), ch); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	after := ch.Snapshot()
	h.Logger.LogStateChange(r.Context(), gameID, before, after, "Burned momentum: outcome changed to "+string(newOutcome))

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"new_outcome":    newOutcome,
		"momentum_reset": ch.Momentum,
		"character":      ch,
	})
}

func (h *Handler) ListMoves(w http.ResponseWriter, r *http.Request) {
	defs := h.Moves.ListDefinitions()
	respondJSON(w, http.StatusOK, defs)
}

func (h *Handler) GetMoveDetails(w http.ResponseWriter, r *http.Request) {
	moveID := chi.URLParam(r, "moveID")
	def, ok := h.Moves.GetDefinition(moveID)
	if !ok {
		respondError(w, http.StatusNotFound, "move not found: "+moveID)
		return
	}
	respondJSON(w, http.StatusOK, def)
}
