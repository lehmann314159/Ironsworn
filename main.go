package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ironsworn/ironsworn-backend/internal/config"
	"github.com/ironsworn/ironsworn-backend/internal/engine"
	"github.com/ironsworn/ironsworn-backend/internal/handler"
	"github.com/ironsworn/ironsworn-backend/internal/router"
	"github.com/ironsworn/ironsworn-backend/internal/store"
)

func main() {
	cfg := config.Load()

	db, err := store.NewSQLiteStore(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	h := handler.New(db, engine.CryptoRoller{})
	r := router.New(h)

	fmt.Printf("Ironsworn API server starting on %s\n", cfg.Port)
	if err := http.ListenAndServe(cfg.Port, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
