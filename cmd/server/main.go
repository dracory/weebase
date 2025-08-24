package main

import (
	"fmt"
	"log"
	"net/http"

	weebase "github.com/dracory/weebase"
)

func main() {
	// Load configuration (flags override env)
	cfg, err := weebase.LoadConfig()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	cfg.AllowAdHocConnections = true

	h := weebase.NewHandler(weebase.Options{
		EnabledDrivers:        []string{"postgres", "mysql", "sqlite"},
		SafeModeDefault:       cfg.SafeModeDefault,
		AllowAdHocConnections: cfg.AllowAdHocConnections,
		BasePath:              cfg.BasePath,
		SessionSecret:         cfg.SessionSecret,
	})

	mux := http.NewServeMux()
	weebase.Register(mux, cfg.BasePath, h)

	// Wrap with request logging middleware
	handler := weebase.RequestLogger(mux)

	addr := fmt.Sprintf(":%d", cfg.HTTPPort)
	log.Printf("WeeBase listening on %s (mount %s)", addr, cfg.BasePath)
	log.Fatal(http.ListenAndServe(addr, handler))
}
