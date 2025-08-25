package weebase

import (
	"net/http"
	"time"

	logoutpage "github.com/dracory/weebase/pages/logout"
)

// handleLogout clears the session and renders a confirmation page.
func (h *Handler) handleLogout(w http.ResponseWriter, r *http.Request, csrfToken string) {
	// Clear any active DB connection
	s := EnsureSession(w, r, h.opts.SessionSecret)
	if s.Conn != nil {
		if sqlDB, err := s.Conn.DB.DB(); err == nil {
			_ = sqlDB.Close()
		}
	}
	// Remove server-side session entry
	if c, err := r.Cookie(sessionCookieName); err == nil && c != nil && c.Value != "" {
		sessionsMu.Lock()
		delete(sessions, c.Value)
		sessionsMu.Unlock()
	}
	// Expire the client cookie
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	})

	// Render logout page
	full, err := logoutpage.Handle(h.tmplBase, h.opts.BasePath, h.opts.ActionParam, h.opts.SafeModeDefault, csrfToken)
	if err != nil {
		// Fallback: simple OK message
		WriteSuccess(w, r, http.StatusOK, "logged out")
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(full))
}
