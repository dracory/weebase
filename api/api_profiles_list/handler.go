package api_profiles_list

import (
	"encoding/json"
	"net/http"
	"net/url"
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
	sess := session.EnsureSession(w, r, h.config.SessionSecret)
	if sess == nil {
		api.Respond(w, r, api.Error("failed to get session"))
		return
	}

	if r.Method != http.MethodGet {
		api.Respond(w, r, api.Error("method not allowed"))
		return
	}

	profiles, err := h.GetProfilesFromCookie(r)
	if err != nil {
		api.Respond(w, r, api.Error("failed to get profiles: "+err.Error()))
		return
	}

	api.Respond(w, r, api.SuccessWithData("", map[string]interface{}{
		"profiles": profiles,
	}))
}

// GetProfilesFromCookie retrieves profiles from the cookie
// This method is exported for testing purposes
func (h *Handler) GetProfilesFromCookie(r *http.Request) ([]Profile, error) {
	cookie, err := r.Cookie(constants.CookieProfiles)
	if err != nil {
		if err == http.ErrNoCookie {
			return []Profile{}, nil
		}
		return nil, err
	}

	// URL decode the cookie value first
	decodedValue, err := url.QueryUnescape(cookie.Value)
	if err != nil {
		return nil, err
	}

	var profiles []Profile
	if err := json.Unmarshal([]byte(decodedValue), &profiles); err != nil {
		return nil, err
	}

	return profiles, nil
}

// SetProfilesCookie sets the profiles cookie
// This method is exported for testing purposes
func (h *Handler) SetProfilesCookie(w http.ResponseWriter, profiles []Profile) error {
	profilesJSON, err := json.Marshal(profiles)
	if err != nil {
		return err
	}

	secure := false // Default to false for local development
	if h.config.SecureCookies {
		secure = true
	}

	// URL encode the JSON string before setting it as a cookie value
	encodedValue := url.QueryEscape(string(profilesJSON))

	http.SetCookie(w, &http.Cookie{
		Name:     constants.CookieProfiles,
		Value:    encodedValue,
		Path:     "/",
		MaxAge:   30 * 24 * 60 * 60, // 30 days
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(30 * 24 * time.Hour),
	})

	return nil
}
