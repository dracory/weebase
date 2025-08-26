package api_schemas_list

import (
	"net/http"

	"github.com/dracory/api"
	"github.com/dracory/weebase/shared/session"
	"gorm.io/gorm"
)

// SchemasList handles schema listing operations
type SchemasList struct {
	conn *session.ActiveConnection
}

// New creates a new SchemasList handler
func New(conn *session.ActiveConnection) *SchemasList {
	return &SchemasList{conn: conn}
}

// Handle processes the request
func (h *SchemasList) Handle(w http.ResponseWriter, r *http.Request) {
	if h.conn == nil || h.conn.DB == nil {
		api.Respond(w, r, api.Error("not connected to database"))
		return
	}

	// Get the database connection
	db, ok := h.conn.DB.(*gorm.DB)
	if !ok {
		api.Respond(w, r, api.Error("invalid database connection type"))
		return
	}

	// Query to get schemas (using SHOW DATABASES for MySQL/MariaDB)
	var schemas []string
	result := db.Raw("SHOW DATABASES").Scan(&schemas)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	// Return the schemas as JSON
	api.Respond(w, r, api.SuccessWithData("schemas", map[string]any{
		"schemas": schemas,
	}))
}
