package weebase

import (
	"net/http"
	"strings"
)

// handleProfilesSave saves a new profile (POST) in the in-memory store.
func (h *Handler) handleProfilesSave(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, r, "profiles_save must be POST")
		return
	}
	_ = r.ParseForm()
	name := strings.TrimSpace(r.Form.Get("name"))
	driver := strings.TrimSpace(r.Form.Get("driver"))
	dsn := strings.TrimSpace(r.Form.Get("dsn"))
	if name == "" || driver == "" || dsn == "" {
		WriteError(w, r, "name, driver and dsn are required")
		return
	}
	if err := h.ValidateDriver(driver); err != nil {
		WriteError(w, r, err.Error())
		return
	}
	p := ConnectionProfile{ID: newRandomID(), Name: name, Driver: driver, DSN: dsn}
	if err := h.profiles.Save(p); err != nil {
		WriteError(w, r, err.Error())
		return
	}
	WriteSuccessWithData(w, r, "saved", map[string]any{"profile": p})
}
