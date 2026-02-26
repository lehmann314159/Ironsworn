package handler

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ironsworn/ironsworn-backend/internal/model"
	"github.com/ironsworn/ironsworn-backend/internal/oracle"
)

type askOracleRequest struct {
	Question   string           `json:"question"`
	Likelihood model.Likelihood `json:"likelihood"`
}

func (h *Handler) AskOracle(w http.ResponseWriter, r *http.Request) {
	gameID := chi.URLParam(r, "gameID")

	var req askOracleRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Likelihood == "" {
		req.Likelihood = model.LikelihoodFiftyFifty
	}

	result := oracle.AskYesNo(h.Roller, req.Likelihood)

	oracleResult := &model.OracleRollResult{
		Roll:   result.Roll,
		Result: fmt.Sprintf("%s (threshold: %d)", boolToYesNo(result.Answer), result.Threshold),
	}
	answer := "No"
	if result.Answer {
		answer = "Yes"
	}
	summary := fmt.Sprintf("Asked the Oracle (%s): %s (roll: %d, threshold: %d)", req.Likelihood, answer, result.Roll, result.Threshold)

	h.Logger.LogOracle(r.Context(), gameID, oracleResult, summary)

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"question":   req.Question,
		"likelihood": req.Likelihood,
		"roll":       result.Roll,
		"threshold":  result.Threshold,
		"answer":     result.Answer,
		"summary":    summary,
	})
}

type rollTableRequest struct {
	TableID string `json:"table_id"`
}

func (h *Handler) RollOracleTable(w http.ResponseWriter, r *http.Request) {
	gameID := chi.URLParam(r, "gameID")

	var req rollTableRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.TableID == "" {
		respondError(w, http.StatusBadRequest, "table_id is required")
		return
	}

	result, err := oracle.RollTable(h.Roller, req.TableID)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	summary := fmt.Sprintf("Rolled on %s: %s (roll: %d)", result.TableName, result.Result, result.Roll)
	h.Logger.LogOracle(r.Context(), gameID, result, summary)

	respondJSON(w, http.StatusOK, result)
}

func (h *Handler) ListOracleTables(w http.ResponseWriter, r *http.Request) {
	tables := oracle.ListTables()
	respondJSON(w, http.StatusOK, tables)
}

func boolToYesNo(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}
