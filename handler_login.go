package weebase

import (
	"log"
	"net/http"
)

// handleLogin renders the Adminer-style login/connect form page.
func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request, csrfToken string) {
	data := map[string]any{
		"Title":                 "Login",
		"BasePath":              h.opts.BasePath,
		"ActionParam":           h.opts.ActionParam,
		"EnabledDrivers":        h.drivers.List(),
		"AllowAdHocConnections": h.opts.AllowAdHocConnections,
		"SafeModeDefault":       h.opts.SafeModeDefault,
		"CSRFToken":             csrfToken,
	}
	if err := h.tmplBase.ExecuteTemplate(w, "login.gohtml", data); err != nil {
		log.Printf("render login: %v", err)
		h.renderStatus(w, r, http.StatusInternalServerError, "template error")
		return
	}
}
