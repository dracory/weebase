package weebase

import (
	"encoding/json"
	"net/http"

	"github.com/dracory/api"
	"github.com/dracory/weebase/shared/session"
)

// handleConnect handles the database connection request
func (w *App) handleConnect(rw http.ResponseWriter, r *http.Request) {
	// Ensure session exists
	sess := session.EnsureSession(rw, r, w.config.SessionSecret)
	if sess == nil {
		api.Respond(rw, r, api.Error("failed to create or retrieve session"))
		return
	}
	if r.Method != http.MethodPost {
		api.Respond(rw, r, api.Error("Method not allowed"))
		return
	}

	var conn DatabaseConnection
	if err := json.NewDecoder(r.Body).Decode(&conn); err != nil {
		api.Respond(rw, r, api.Error("Invalid request body"))
		return
	}

	db, err := w.connectDB(conn)
	if err != nil {
		api.Respond(rw, r, api.Error(err.Error()))
		return
	}

	w.db = db
	api.Respond(rw, r, api.Success("Connected successfully"))
}
