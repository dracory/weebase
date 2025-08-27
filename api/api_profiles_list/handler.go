package api_profiles_list

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/dracory/api"
	"github.com/dracory/weebase/shared/constants"
	"github.com/dracory/weebase/shared/session"
	"github.com/dracory/weebase/shared/types"
)

// Profile represents a saved database connection profile
type Profile struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Driver   string `json:"driver"`
	Server   string `json:"server,omitempty"`
	Port     string `json:"port,omitempty"`
	Username string `json:"username,omitempty"`
	Database string `json:"database,omitempty"`
}

// Handler handles the profiles list API requests
type Handler struct {
	config types.Config
}

// New creates a new profiles list handler
func New(config types.Config) *Handler {
	return &Handler{
		config: config,
	}
}

// ServeHTTP handles the HTTP request
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s := session.EnsureSession(w, r, h.config.SessionSecret)
	if s == nil {
		api.Respond(w, r, api.Error("failed to create or retrieve session"))
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.handleGetProfiles(w, r, s)
	default:
		api.Respond(w, r, api.Error("method not allowed"))
	}
}

// handleGetProfiles handles GET /api/profiles
func (h *Handler) handleGetProfiles(w http.ResponseWriter, r *http.Request, s *session.Session) {
	profiles, err := h.getProfilesFromCookie(r)
	if err != nil {
		api.Respond(w, r, api.Error("failed to get profiles: "+err.Error()))
		return
	}

	api.Respond(w, r, api.SuccessWithData("", map[string]interface{}{
		"profiles": profiles,
	}))
}

// getProfilesFromCookie retrieves profiles from the cookie
func (h *Handler) getProfilesFromCookie(r *http.Request) ([]Profile, error) {
	cookie, err := r.Cookie(constants.CookieProfiles)
	if err != nil {
		if err == http.ErrNoCookie {
			return []Profile{}, nil
		}
		return nil, err
	}

	var profiles []Profile
	if err := json.Unmarshal([]byte(cookie.Value), &profiles); err != nil {
		return nil, err
	}

	return profiles, nil
}

// setProfilesCookie sets the profiles cookie
func (h *Handler) setProfilesCookie(w http.ResponseWriter, profiles []Profile) error {
	profilesJSON, err := json.Marshal(profiles)
	if err != nil {
		return err
	}

	secure := false // Default to false for local development
	if h.config.SecureCookies {
		secure = true
	}

	http.SetCookie(w, &http.Cookie{
		Name:     constants.CookieProfiles,
		Value:    string(profilesJSON),
		Path:     "/",
		MaxAge:   30 * 24 * 60 * 60, // 30 days
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(30 * 24 * time.Hour),
	})

	return nil
}
