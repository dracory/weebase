package api_profiles_list

import (
	"net/http"

	"github.com/dracory/api"
	"github.com/dracory/weebase/shared/types"
)

// ConnectionStore defines the interface for profile storage operations
type ConnectionStore interface {
	List() []types.ConnectionProfile
}

// ProfilesList handles profile list requests
type ProfilesList struct {
	store ConnectionStore
}

// New creates a new ProfilesList handler
func New(store ConnectionStore) *ProfilesList {
	return &ProfilesList{
		store: store,
	}
}

// Handle handles the HTTP request
func (h *ProfilesList) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		api.Respond(w, r, api.Error("profiles must be GET"))
		return
	}
	
	profiles := h.store.List()
	api.Respond(w, r, api.SuccessWithData("ok", map[string]any{"profiles": profiles}))
}
