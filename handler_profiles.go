package weebase

import (
	"net/http"
)

// handleProfiles lists saved profiles (GET) using the in-memory store for now.
func (h *Handler) handleProfiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, r, "profiles must be GET")
		return
	}
	profiles := h.profiles.List()
	WriteSuccessWithData(w, r, "ok", map[string]any{"profiles": profiles})
}
