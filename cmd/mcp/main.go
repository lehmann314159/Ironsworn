package main

import (
	"log"
	"os"

	"github.com/mark3labs/mcp-go/server"

	"github.com/ironsworn/ironsworn/internal/mcpserver"
	"github.com/ironsworn/ironsworn/internal/store"
)

func main() {
	dbPath := "ironsworn.db"
	if p := os.Getenv("IRONSWORN_DB_PATH"); p != "" {
		dbPath = p
	}

	db, err := store.NewSQLiteStore(dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	srv := mcpserver.New(db)

	if err := server.ServeStdio(srv.MCPServer()); err != nil {
		log.Fatalf("MCP server error: %v", err)
	}
}
