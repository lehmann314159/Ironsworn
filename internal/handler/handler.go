package handler

import (
	"encoding/json"
	"net/http"

	"github.com/ironsworn/ironsworn/internal/engine"
	"github.com/ironsworn/ironsworn/internal/gamelog"
	"github.com/ironsworn/ironsworn/internal/store"
)

// Handler holds shared dependencies for all HTTP handlers.
type Handler struct {
	Store    store.Store
	Moves    *engine.MoveRegistry
	Roller   engine.Roller
	Logger   *gamelog.Logger
}

// New creates a new Handler with all dependencies.
func New(s store.Store, roller engine.Roller) *Handler {
	return &Handler{
		Store:  s,
		Moves:  engine.NewMoveRegistry(),
		Roller: roller,
		Logger: gamelog.NewLogger(s),
	}
}

// respondJSON writes a JSON response.
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// respondError writes a JSON error response.
func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

// decodeJSON decodes a JSON request body into dst.
func decodeJSON(r *http.Request, dst interface{}) error {
	return json.NewDecoder(r.Body).Decode(dst)
}
