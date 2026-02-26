package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/ironsworn/ironsworn-backend/internal/model"
)

func (h *Handler) GetLogs(w http.ResponseWriter, r *http.Request) {
	gameID := chi.URLParam(r, "gameID")

	filter := model.LogFilter{}
	if t := r.URL.Query().Get("type"); t != "" {
		filter.EntryType = t
	}
	if tags := r.URL.Query().Get("tags"); tags != "" {
		filter.Tags = strings.Split(tags, ",")
	}
	if o := r.URL.Query().Get("outcome"); o != "" {
		filter.Outcome = o
	}
	if l := r.URL.Query().Get("limit"); l != "" {
		filter.Limit, _ = strconv.Atoi(l)
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		filter.Offset, _ = strconv.Atoi(o)
	}

	if filter.Limit <= 0 {
		filter.Limit = 50
	}

	entries, err := h.Store.GetLogs(r.Context(), gameID, filter)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if entries == nil {
		entries = []*model.LogEntry{}
	}
	respondJSON(w, http.StatusOK, entries)
}

func (h *Handler) ExportLogs(w http.ResponseWriter, r *http.Request) {
	gameID := chi.URLParam(r, "gameID")

	// Export all entries as JSONL (one JSON object per line)
	entries, err := h.Store.GetLogs(r.Context(), gameID, model.LogFilter{Limit: 0})
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/x-ndjson")
	w.Header().Set("Content-Disposition", "attachment; filename=game_log.jsonl")
	w.WriteHeader(http.StatusOK)

	enc := json.NewEncoder(w)
	for _, entry := range entries {
		enc.Encode(entry)
	}
}

type addNarrativeRequest struct {
	Text string   `json:"text"`
	Tags []string `json:"tags,omitempty"`
}

func (h *Handler) AddNarrative(w http.ResponseWriter, r *http.Request) {
	gameID := chi.URLParam(r, "gameID")

	var req addNarrativeRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Text == "" {
		respondError(w, http.StatusBadRequest, "text is required")
		return
	}

	if err := h.Logger.LogNarrative(r.Context(), gameID, req.Text, req.Tags); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, map[string]string{"status": "logged"})
}
