package weebase

import (
	"net/http"
	"strings"
)

// handleConnect establishes a DB connection and stores it in the session.
func (h *Handler) handleConnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, r, "connect must be POST")
		return
	}
	_ = r.ParseForm()
	profileID := strings.TrimSpace(r.Form.Get("profile_id"))
	driver := strings.TrimSpace(r.Form.Get("driver"))
	dsn := strings.TrimSpace(r.Form.Get("dsn"))
	// If profile_id provided, resolve it
	if profileID != "" {
		if p, ok := h.profiles.Get(profileID); ok {
			driver, dsn = p.Driver, p.DSN
		} else {
			WriteError(w, r, "profile not found")
			return
		}
	}
	if err := h.ValidateDriver(driver); err != nil {
		WriteError(w, r, err.Error())
		return
	}
	if dsn == "" {
		WriteError(w, r, "dsn is required")
		return
	}
	s := EnsureSession(w, r, h.opts.SessionSecret)
	if err := h.tryAutoConnect(s, driver, dsn); err != nil {
		WriteError(w, r, err.Error())
		return
	}
	WriteSuccessWithData(w, r, "connected", map[string]any{"driver": driver})
}
