package config

import (
	"os"
)

// Config holds application configuration.
type Config struct {
	Port   string
	DBPath string
}

// Load returns configuration from environment variables with defaults.
func Load() Config {
	cfg := Config{
		Port:   ":8080",
		DBPath: "ironsworn.db",
	}

	if port := os.Getenv("IRONSWORN_PORT"); port != "" {
		cfg.Port = ":" + port
	}
	if dbPath := os.Getenv("IRONSWORN_DB_PATH"); dbPath != "" {
		cfg.DBPath = dbPath
	}

	return cfg
}
