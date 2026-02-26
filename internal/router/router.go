package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/ironsworn/ironsworn/internal/handler"
)

// New creates and configures the Chi router with all routes.
func New(h *handler.Handler) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.SetHeader("Content-Type", "application/json"))

	r.Route("/api", func(r chi.Router) {
		// Game management
		r.Route("/games", func(r chi.Router) {
			r.Post("/", h.CreateGame)
			r.Get("/", h.ListGames)

			r.Route("/{gameID}", func(r chi.Router) {
				r.Get("/", h.GetGame)
				r.Delete("/", h.DeleteGame)

				// Character
				r.Post("/character", h.CreateCharacter)
				r.Get("/character", h.GetCharacter)
				r.Patch("/character", h.PatchCharacter)

				// Character assets
				r.Post("/character/assets", h.AddAsset)
				r.Get("/character/assets", h.ListAssets)
				r.Patch("/character/assets/{assetID}", h.UpgradeAsset)

				// Moves
				r.Post("/moves", h.ExecuteMove)
				r.Post("/burn-momentum", h.BurnMomentum)

				// Progress tracks
				r.Route("/progress", func(r chi.Router) {
					r.Post("/", h.CreateTrack)
					r.Get("/", h.ListTracks)
					r.Patch("/{trackID}", h.MarkProgress)
					r.Delete("/{trackID}", h.DeleteTrack)
				})

				// Oracle
				r.Post("/oracle/ask", h.AskOracle)
				r.Post("/oracle/roll", h.RollOracleTable)

				// Game log (RAG)
				r.Get("/log", h.GetLogs)
				r.Get("/log/export", h.ExportLogs)
				r.Post("/log", h.AddNarrative)
			})
		})

		// Move reference (game-independent)
		r.Get("/moves", h.ListMoves)
		r.Get("/moves/{moveID}", h.GetMoveDetails)

		// Oracle tables (game-independent)
		r.Get("/oracle/tables", h.ListOracleTables)
	})

	return r
}
