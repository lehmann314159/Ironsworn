package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ironsworn/ironsworn/internal/gamelog"
	"github.com/ironsworn/ironsworn/internal/model"
)

type createGameRequest struct {
	Name string `json:"name"`
}

func (h *Handler) CreateGame(w http.ResponseWriter, r *http.Request) {
	var req createGameRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "name is required")
		return
	}

	game := &model.Game{
		ID:   gamelog.GenerateID(),
		Name: req.Name,
	}
	if err := h.Store.CreateGame(r.Context(), game); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, game)
}

func (h *Handler) ListGames(w http.ResponseWriter, r *http.Request) {
	games, err := h.Store.ListGames(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if games == nil {
		games = []*model.Game{}
	}
	respondJSON(w, http.StatusOK, games)
}

func (h *Handler) GetGame(w http.ResponseWriter, r *http.Request) {
	gameID := chi.URLParam(r, "gameID")
	game, err := h.Store.GetGame(r.Context(), gameID)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, game)
}

func (h *Handler) DeleteGame(w http.ResponseWriter, r *http.Request) {
	gameID := chi.URLParam(r, "gameID")
	if err := h.Store.DeleteGame(r.Context(), gameID); err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
