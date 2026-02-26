package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	enginepkg "github.com/ironsworn/ironsworn-backend/internal/engine"
	"github.com/ironsworn/ironsworn-backend/internal/gamelog"
	"github.com/ironsworn/ironsworn-backend/internal/model"
)

type createCharacterRequest struct {
	Name   string      `json:"name"`
	Stats  model.Stats `json:"stats"`
}

func (h *Handler) CreateCharacter(w http.ResponseWriter, r *http.Request) {
	gameID := chi.URLParam(r, "gameID")

	// Verify game exists
	if _, err := h.Store.GetGame(r.Context(), gameID); err != nil {
		respondError(w, http.StatusNotFound, "game not found")
		return
	}

	var req createCharacterRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := enginepkg.ValidateStatDistribution(req.Stats); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	ch, err := enginepkg.CreateCharacter(gamelog.GenerateID(), gameID, req.Name, req.Stats)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.Store.CreateCharacter(r.Context(), ch); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, ch)
}

func (h *Handler) GetCharacter(w http.ResponseWriter, r *http.Request) {
	gameID := chi.URLParam(r, "gameID")
	ch, err := h.Store.GetCharacter(r.Context(), gameID)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, ch)
}

type patchCharacterRequest struct {
	Health   *int  `json:"health,omitempty"`
	Spirit   *int  `json:"spirit,omitempty"`
	Supply   *int  `json:"supply,omitempty"`
	Momentum *int  `json:"momentum,omitempty"`

	Wounded    *bool `json:"wounded,omitempty"`
	Shaken     *bool `json:"shaken,omitempty"`
	Unprepared *bool `json:"unprepared,omitempty"`
	Encumbered *bool `json:"encumbered,omitempty"`
	Maimed     *bool `json:"maimed,omitempty"`
	Corrupted  *bool `json:"corrupted,omitempty"`
	Cursed     *bool `json:"cursed,omitempty"`
	Tormented  *bool `json:"tormented,omitempty"`
}

func (h *Handler) PatchCharacter(w http.ResponseWriter, r *http.Request) {
	gameID := chi.URLParam(r, "gameID")
	ch, err := h.Store.GetCharacter(r.Context(), gameID)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	before := ch.Snapshot()

	var req patchCharacterRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Health != nil {
		ch.Health = model.Clamp(*req.Health, 0, 5)
	}
	if req.Spirit != nil {
		ch.Spirit = model.Clamp(*req.Spirit, 0, 5)
	}
	if req.Supply != nil {
		ch.Supply = model.Clamp(*req.Supply, 0, 5)
	}
	if req.Momentum != nil {
		ch.Momentum = model.Clamp(*req.Momentum, -6, ch.MomentumMax)
	}

	if req.Wounded != nil {
		ch.Debilities.Wounded = *req.Wounded
	}
	if req.Shaken != nil {
		ch.Debilities.Shaken = *req.Shaken
	}
	if req.Unprepared != nil {
		ch.Debilities.Unprepared = *req.Unprepared
	}
	if req.Encumbered != nil {
		ch.Debilities.Encumbered = *req.Encumbered
	}
	if req.Maimed != nil {
		ch.Debilities.Maimed = *req.Maimed
	}
	if req.Corrupted != nil {
		ch.Debilities.Corrupted = *req.Corrupted
	}
	if req.Cursed != nil {
		ch.Debilities.Cursed = *req.Cursed
	}
	if req.Tormented != nil {
		ch.Debilities.Tormented = *req.Tormented
	}

	ch.UpdateMomentumLimits()

	if err := h.Store.UpdateCharacter(r.Context(), ch); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	after := ch.Snapshot()
	h.Logger.LogStateChange(r.Context(), gameID, before, after, "Manual character adjustment")

	respondJSON(w, http.StatusOK, ch)
}
