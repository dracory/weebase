package api_disconnect

import (
	"net/http"

	"github.com/dracory/api"
	"github.com/dracory/weebase/shared/session"
)

// Handler handles database disconnection requests
type Handler struct {
	sessionSecret string
}

// New creates a new disconnect handler
func New(sessionSecret string) *Handler {
	return &Handler{
		sessionSecret: sessionSecret,
	}
}

// Handle handles the HTTP request for disconnecting from a database
func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	s := session.EnsureSession(w, r, h.sessionSecret)
	if s == nil {
		api.Respond(w, r, api.Error("no active session"))
		return
	}

	if s.Conn != nil {
		s.Conn = nil
	}

	api.Respond(w, r, api.Success("disconnected"))
}
