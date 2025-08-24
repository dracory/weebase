package weebase

import (
	"encoding/json"
	"net/http"

	api "github.com/dracory/api"
)

// WriteSuccess writes a success envelope with a message and status code using api.Respond.
func WriteSuccess(w http.ResponseWriter, r *http.Request, status int, msg string) {
	// The upstream API provides a RespondWithStatusCode helper; prefer it when status != 200.
	if status == http.StatusOK {
		api.Respond(w, r, api.Success(msg))
		return
	}
	api.RespondWithStatusCode(w, r, api.Success(msg), status)
}

// WriteSuccessWithData writes a success envelope with message and data.
func WriteSuccessWithData(w http.ResponseWriter, r *http.Request, status int, msg string, data map[string]any) {
	if status == http.StatusOK {
		api.Respond(w, r, api.SuccessWithData(msg, data))
		return
	}
	api.RespondWithStatusCode(w, r, api.SuccessWithData(msg, data), status)
}

// WriteError writes an error envelope with message and status code.
func WriteError(w http.ResponseWriter, r *http.Request, status int, msg string) {
	// Ensure JSON content type in case upstream doesn't set it.
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	// Use the helper with explicit status code.
	api.RespondWithStatusCode(w, r, api.Error(msg), status)
}

// JSONNotImplemented is a helper to return 501 for unimplemented actions using the api envelope.
func JSONNotImplemented(w http.ResponseWriter, action string) {
	// No request available in this signature; keep compatibility by writing directly.
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusNotImplemented)
	_ = json.NewEncoder(w).Encode(api.Error("action '" + action + "' not implemented yet"))
}
