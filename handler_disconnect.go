package weebase

import (
	"net/http"

	"github.com/dracory/weebase/shared/session"
)

// handleDisconnect clears the active session connection.
func (h *Handler) handleDisconnect(w http.ResponseWriter, r *http.Request) {
	s := session.EnsureSession(w, r, h.opts.SessionSecret)
	if s == nil {
		WriteError(w, r, "no active session")
		return
	}

	if s.Conn != nil {
		s.Conn = nil
	}

	WriteSuccess(w, r, http.StatusOK, "disconnected")
}
