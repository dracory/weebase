package api_profiles_save

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/dracory/api"
	"github.com/dracory/weebase/shared/types"
)

// ConnectionStore defines the interface for profile storage operations
type ConnectionStore interface {
	Save(profile types.ConnectionProfile) error
}

// DriverValidator is an alias for the driver.Validator interface
type DriverValidator interface {
	Validate(string) error
}

// ProfilesSave handles profile save requests
type ProfilesSave struct {
	store    ConnectionStore
	validate DriverValidator
}

// New creates a new ProfilesSave handler
func New(store ConnectionStore, validator DriverValidator) *ProfilesSave {
	return &ProfilesSave{
		store:    store,
		validate: validator,
	}
}


// Handle processes the profile save request
func (h *ProfilesSave) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		api.Respond(w, r, api.Error("profiles_save must be POST"))
		return
	}

	if err := r.ParseForm(); err != nil {
		api.Respond(w, r, api.Error("failed to parse form"))
		return
	}

	name := strings.TrimSpace(r.Form.Get("name"))
	driver := strings.TrimSpace(r.Form.Get("driver"))
	dsn := strings.TrimSpace(r.Form.Get("dsn"))

	// Validate required parameters
	if name == "" || driver == "" || dsn == "" {
		api.Respond(w, r, api.Error("name, driver and dsn are required"))
		return
	}

	// Validate the driver
	if err := h.validate.Validate(driver); err != nil {
		api.Respond(w, r, api.Error(err.Error()))
		return
	}

	// Create and save the profile
	p := types.ConnectionProfile{
		ID:     newRandomID(),
		Name:   name,
		Driver: driver,
		DSN:    dsn,
	}

	if err := h.store.Save(p); err != nil {
		api.Respond(w, r, api.Error(err.Error()))
		return
	}

	api.Respond(w, r, api.SuccessWithData("profile saved", map[string]any{
		"profile": p,
	}))
}

// newRandomID generates a new random ID for the profile
func newRandomID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
