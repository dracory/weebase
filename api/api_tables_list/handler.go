package api_tables_list

import (
	"encoding/json"
	"net/http"

	"github.com/dracory/weebase/shared/session"
	"gorm.io/gorm"
)

// TablesList handles table listing operations
type TablesList struct {
	conn *session.ActiveConnection
}

// New creates a new TablesList handler
func New(conn *session.ActiveConnection) *TablesList {
	return &TablesList{conn: conn}
}

// Handle processes the request
func (h *TablesList) Handle(w http.ResponseWriter, r *http.Request) {
	if h.conn == nil || h.conn.DB == nil {
		http.Error(w, "Not connected to database", http.StatusBadRequest)
		return
	}

	// Get the database connection
	db, ok := h.conn.DB.(*gorm.DB)
	if !ok {
		http.Error(w, "Invalid database connection type", http.StatusInternalServerError)
		return
	}

	// Query to get tables
	var tables []string
	result := db.Raw("SHOW TABLES").Scan(&tables)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	// Return the tables as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tables)
}
