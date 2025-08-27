package main

import (
	"fmt"
	"log"
	"net/http"

	weebase "github.com/dracory/weebase"
	"github.com/dracory/weebase/shared/constants"
)

func main() {
	// Load configuration (flags override env)
	cfg, err := weebase.LoadConfig()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	// Update config with our settings
	cfg.AllowAdHocConnections = true
	if len(cfg.EnabledDrivers) == 0 {
		cfg.EnabledDrivers = []string{
			constants.DriverPostgres,
			constants.DriverMySQL,
			constants.DriverSQLite,
		}
	}

	// Create a new Weebase instance with the config
	app := weebase.New(cfg)

	// Get the HTTP handler
	h := app.Handler()

	addr := fmt.Sprintf(":%d", cfg.HTTPPort)
	log.Printf("WeeBase listening on %s (mount %s)", addr, cfg.BasePath)

	mux := http.NewServeMux()
	mux.Handle(cfg.BasePath, h)

	// Wrap with request logging middleware
	handler := weebase.RequestLogger(mux)

	log.Fatal(http.ListenAndServe(addr, handler))
}
