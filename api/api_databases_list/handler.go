package api_databases_list

import (
	"encoding/json"
	"net/http"

	"github.com/dracory/weebase/shared/types"
)

type Handler struct {
	config types.Config
}

func New(config types.Config) *Handler {
	return &Handler{
		config: config,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement actual database listing logic
	// For now, return a mock response
	response := map[string]interface{}{
		"success": true,
		"data": []map[string]interface{}{
			{
				"name":   "database1",
				"tables": 10,
				"size":   1024 * 1024 * 5, // 5MB
			},
			{
				"name":   "database2",
				"tables": 5,
				"size":   1024 * 1024 * 2, // 2MB
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
