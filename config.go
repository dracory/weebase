package weebase

import (
	"flag"
	"fmt"

	"github.com/dracory/env"
)

// Config holds runtime configuration sourced from env and flags.
type Config struct {
	HTTPPort              int
	BasePath              string
	SessionSecret         string
	AllowAdHocConnections bool
	SafeModeDefault       bool
}

// LoadConfig reads flags/env with sensible defaults.
// Flags take precedence over env.
func LoadConfig() (Config, error) {
	var cfg Config

	// Optionally load from .env files (missing files are ignored inside the lib)
	env.Load(".env")

	// Defaults via env package
	cfg.HTTPPort = env.GetIntOrDefault("HTTP_PORT", 8080)
	cfg.BasePath = env.GetStringOrDefault("BASE_URL", "/db")
	cfg.SessionSecret = env.GetStringOrDefault("SESSION_SECRET", "dev-insecure-change-me")
	cfg.AllowAdHocConnections = env.GetBoolOrDefault("ALLOW_ADHOC_CONNECTIONS", true)
	cfg.SafeModeDefault = env.GetBoolOrDefault("SAFE_MODE_DEFAULT", true)

	// Flags
	port := flag.Int("port", cfg.HTTPPort, "HTTP port to listen on")
	base := flag.String("base", cfg.BasePath, "Base path to mount handler under (e.g. /db)")
	safe := flag.Bool("safe", cfg.SafeModeDefault, "Safe mode default (block destructive ops)")
	adhoc := flag.Bool("adhoc", cfg.AllowAdHocConnections, "Allow ad-hoc connections via UI")
	flag.Parse()

	cfg.HTTPPort = *port
	cfg.BasePath = *base
	cfg.SafeModeDefault = *safe
	cfg.AllowAdHocConnections = *adhoc

	if cfg.SessionSecret == "" {
		return cfg, fmt.Errorf("SESSION_SECRET is required")
	}
	return cfg, nil
}
