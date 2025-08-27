package weebase

import (
	"encoding/json"
	"net/http"

	"github.com/dracory/api"
)

// handleConnect handles the database connection request
func (w *Weebase) handleConnect(rw http.ResponseWriter, r *http.Request) {
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
