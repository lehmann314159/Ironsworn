package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ironsworn/ironsworn/internal/gamelog"
	"github.com/ironsworn/ironsworn/internal/model"
)

type addAssetRequest struct {
	AssetID string `json:"asset_id"`
	Name    string `json:"name"`
}

func (h *Handler) AddAsset(w http.ResponseWriter, r *http.Request) {
	gameID := chi.URLParam(r, "gameID")

	ch, err := h.Store.GetCharacter(r.Context(), gameID)
	if err != nil {
		respondError(w, http.StatusNotFound, "character not found")
		return
	}

	var req addAssetRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.AssetID == "" || req.Name == "" {
		respondError(w, http.StatusBadRequest, "asset_id and name are required")
		return
	}

	asset := &model.CharacterAsset{
		ID:          gamelog.GenerateID(),
		CharacterID: ch.ID,
		AssetID:     req.AssetID,
		Name:        req.Name,
		Ability1:    true, // First ability always unlocked
	}

	if err := h.Store.AddAsset(r.Context(), asset); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, asset)
}

func (h *Handler) ListAssets(w http.ResponseWriter, r *http.Request) {
	gameID := chi.URLParam(r, "gameID")

	ch, err := h.Store.GetCharacter(r.Context(), gameID)
	if err != nil {
		respondError(w, http.StatusNotFound, "character not found")
		return
	}

	assets, err := h.Store.ListAssets(r.Context(), ch.ID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if assets == nil {
		assets = []*model.CharacterAsset{}
	}

	respondJSON(w, http.StatusOK, assets)
}

type upgradeAssetRequest struct {
	Ability2 *bool `json:"ability2,omitempty"`
	Ability3 *bool `json:"ability3,omitempty"`
}

func (h *Handler) UpgradeAsset(w http.ResponseWriter, r *http.Request) {
	assetID := chi.URLParam(r, "assetID")

	asset, err := h.Store.GetAsset(r.Context(), assetID)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	var req upgradeAssetRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Ability2 != nil {
		asset.Ability2 = *req.Ability2
	}
	if req.Ability3 != nil {
		asset.Ability3 = *req.Ability3
	}

	if err := h.Store.UpdateAsset(r.Context(), asset); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, asset)
}
