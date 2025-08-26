package rows_browse

import (
	"net/http"
	"strconv"

	"github.com/dracory/api"
	"github.com/dracory/weebase/shared/session"
	"gorm.io/gorm"
)

// RowsBrowse handles row browsing operations
type RowsBrowse struct {
	conn *session.ActiveConnection
}

// New creates a new RowsBrowse handler
func New(conn *session.ActiveConnection) *RowsBrowse {
	return &RowsBrowse{conn: conn}
}

// Handle processes the request
func (h *RowsBrowse) Handle(w http.ResponseWriter, r *http.Request) {
	if h.conn == nil || h.conn.DB == nil {
		api.Respond(w, r, api.Error("not connected to database"))
		return
	}

	// Get query parameters
	table := r.URL.Query().Get("table")
	if table == "" {
		api.Respond(w, r, api.Error("table is required"))
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 {
		limit = 50 // Default limit
	}
	offset := (page - 1) * limit

	// Get the database connection
	db, ok := h.conn.DB.(*gorm.DB)
	if !ok {
		http.Error(w, "Invalid database connection type", http.StatusInternalServerError)
		return
	}

	// Query to get rows
	var results []map[string]interface{}
	result := db.Table(table).Offset(offset).Limit(limit).Find(&results)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	// Get total count
	var count int64
	db.Table(table).Count(&count)

	api.Respond(w, r, api.SuccessWithData("rows", map[string]any{
		"rows":  results,
		"total": count,
		"page":  page,
		"limit": limit,
	}))
}
