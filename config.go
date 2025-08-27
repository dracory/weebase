package weebase

import (
	"flag"
	"fmt"

	"github.com/dracory/env"
	"github.com/dracory/weebase/shared/types"
)

// // Config holds all configuration for the Weebase instance
// type Config struct {
// 	// Server settings
// 	HTTPPort      int    // Port to listen on (default: 8080)
// 	BasePath      string // Base URL path (default: "/")
// 	SessionSecret string // Secret for session encryption

// 	// Security settings
// 	AllowAdHocConnections bool // Whether to allow ad-hoc database connections
// 	SafeModeDefault       bool // Default safe mode setting for database operations

// 	// List of supported database drivers (default: all)
// 	Drivers []string

// 	// Query parameter name for actions (default: "action")
// 	ActionParam string
// }

// func (c *Config) toWebConfig() *types.Config {
// 	webConfig := &types.Config{
// 		BasePath:              c.BasePath,
// 		ActionParam:           c.ActionParam,
// 		EnabledDrivers:        c.Drivers,
// 		AllowAdHocConnections: c.AllowAdHocConnections,
// 		SafeModeDefault:       c.SafeModeDefault,
// 		SessionSecret:         c.SessionSecret,
// 	}
// 	return webConfig
// }

// LoadConfig reads flags/env with sensible defaults.
// Flags take precedence over env.
func LoadConfig() (types.Config, error) {
	var cfg types.Config

	// Optionally load from .env files (missing files are ignored inside the lib)
	env.Load(".env")

	// Defaults via env package
	cfg.HTTPPort = env.GetIntOrDefault("HTTP_PORT", 8080)
	cfg.BasePath = env.GetStringOrDefault("BASE_URL", "/")
	cfg.SessionSecret = env.GetStringOrDefault("SESSION_SECRET", "dev-insecure-change-me")
	cfg.AllowAdHocConnections = env.GetBoolOrDefault("ALLOW_ADHOC_CONNECTIONS", true)
	cfg.SafeModeDefault = env.GetBoolOrDefault("SAFE_MODE_DEFAULT", true)
	cfg.ActionParam = env.GetStringOrDefault("ACTION_PARAM", "action")

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
