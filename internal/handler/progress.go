package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	enginepkg "github.com/ironsworn/ironsworn-backend/internal/engine"
	"github.com/ironsworn/ironsworn-backend/internal/gamelog"
	"github.com/ironsworn/ironsworn-backend/internal/model"
)

type createTrackRequest struct {
	Name      string                  `json:"name"`
	TrackType model.ProgressTrackType `json:"track_type"`
	Rank      model.ProgressRank      `json:"rank"`
}

func (h *Handler) CreateTrack(w http.ResponseWriter, r *http.Request) {
	gameID := chi.URLParam(r, "gameID")

	if _, err := h.Store.GetGame(r.Context(), gameID); err != nil {
		respondError(w, http.StatusNotFound, "game not found")
		return
	}

	var req createTrackRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	track, err := enginepkg.CreateProgressTrack(gamelog.GenerateID(), gameID, req.Name, req.TrackType, req.Rank)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.Store.CreateTrack(r.Context(), track); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, track)
}

func (h *Handler) ListTracks(w http.ResponseWriter, r *http.Request) {
	gameID := chi.URLParam(r, "gameID")
	trackType := r.URL.Query().Get("type")

	var completed *bool
	if c := r.URL.Query().Get("completed"); c != "" {
		val := c == "true"
		completed = &val
	}

	tracks, err := h.Store.ListTracks(r.Context(), gameID, trackType, completed)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if tracks == nil {
		tracks = []*model.ProgressTrack{}
	}
	respondJSON(w, http.StatusOK, tracks)
}

type markProgressRequest struct {
	Count int `json:"count"` // How many times to mark progress (default 1)
}

func (h *Handler) MarkProgress(w http.ResponseWriter, r *http.Request) {
	trackID := chi.URLParam(r, "trackID")

	track, err := h.Store.GetTrack(r.Context(), trackID)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	var req markProgressRequest
	if err := decodeJSON(r, &req); err != nil {
		// Default to 1 mark
		req.Count = 1
	}
	if req.Count <= 0 {
		req.Count = 1
	}

	totalTicks := 0
	for i := 0; i < req.Count; i++ {
		ticks, err := enginepkg.MarkProgress(track)
		if err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		totalTicks += ticks
	}

	if err := h.Store.UpdateTrack(r.Context(), track); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"track":       track,
		"ticks_added": totalTicks,
		"score":       track.Score(),
	})
}

func (h *Handler) DeleteTrack(w http.ResponseWriter, r *http.Request) {
	trackID := chi.URLParam(r, "trackID")
	if err := h.Store.DeleteTrack(r.Context(), trackID); err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
