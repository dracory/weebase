package weebase

import (
	"net/http"
)

// handleDisconnect clears the active session connection.
func (h *Handler) handleDisconnect(w http.ResponseWriter, r *http.Request) {
	s := EnsureSession(w, r, h.opts.SessionSecret)
	if s.Conn != nil {
		if sqlDB, err := s.Conn.DB.DB(); err == nil {
			_ = sqlDB.Close()
		}
	}
	s.Conn = nil
	WriteSuccess(w, r, http.StatusOK, "disconnected")
}
