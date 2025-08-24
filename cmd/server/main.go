package main

import (
	"log"
	"net/http"

	weebase "github.com/dracory/weebase"
)

func main() {
	h := weebase.NewHandler(weebase.Options{
		EnabledDrivers:        []string{"postgres", "mysql", "sqlite"},
		SafeModeDefault:       true,
		AllowAdHocConnections: true,
		BasePath:              "/db",
	})

	mux := http.NewServeMux()
	weebase.Register(mux, "/db", h)

	addr := ":8080"
	log.Printf("WeeBase listening on %s (mount %s)", addr, "/db")
	log.Fatal(http.ListenAndServe(addr, mux))
}
