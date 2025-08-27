package api_tables_list

import (
	"net/http"

	"github.com/dracory/api"
	"github.com/dracory/weebase/shared/session"
	"github.com/dracory/weebase/shared/types"
	"gorm.io/gorm"
)

// TablesList handles table listing operations
type TablesList struct {
	cfg types.Config
}

// New creates a new TablesList handler
func New(cfg types.Config) *TablesList {
	return &TablesList{cfg: cfg}
}

// Handle processes the request
func (h *TablesList) Handle(w http.ResponseWriter, r *http.Request) {
	conn := session.EnsureSession(w, r, h.cfg.SessionSecret)
	if conn == nil || conn.Conn == nil || conn.Conn.DB == nil {
		w.WriteHeader(http.StatusBadRequest)
		api.Respond(w, r, api.Error("not connected to database"))
		return
	}

	// Get the database connection
	db, ok := conn.Conn.DB.(*gorm.DB)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		api.Respond(w, r, api.Error("invalid database connection type"))
		return
	}

	// Try to get the database dialect
	dialect := db.Dialector.Name()
	var tables []string
	var result *gorm.DB

	switch dialect {
	case "sqlite":
		// SQLite specific query
		type sqliteTable struct {
			Name string `gorm:"column:name"`
		}
		var sqliteTables []sqliteTable
		result = db.Raw("SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' AND name != 'migrations' ORDER BY name").Scan(&sqliteTables)
		for _, t := range sqliteTables {
			tables = append(tables, t.Name)
		}
	default:
		// Default to MySQL/PostgreSQL style SHOW TABLES
		result = db.Raw("SHOW TABLES").Scan(&tables)
	}

	if result.Error != nil {
		api.Respond(w, r, api.Error(result.Error.Error()))
		return
	}

	api.Respond(w, r, api.SuccessWithData("Tables listed successfully", map[string]any{
		"tables": tables,
	}))
}
