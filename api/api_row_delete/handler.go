package api_row_delete

import (
	"fmt"
	"net/http"
	"strings"

	"gorm.io/gorm"

	"github.com/dracory/api"
	"github.com/dracory/weebase/shared/driver"
	"github.com/dracory/weebase/shared/session"
	"github.com/dracory/weebase/shared/types"
)

// Handler handles row deletion operations
type rowDeleteController struct {
	config types.Config
	// sessionConn   *session.ActiveConnection
	// safeMode      bool
	// sessionSecret string
}

// New creates a new row delete handler
func New(config types.Config) *rowDeleteController {
	return &rowDeleteController{
		config: config,
		// sessionConn:   sessionConn,
		// safeMode:      safeMode,
		// sessionSecret: sessionSecret,
	}
}

// Handle handles the HTTP request for deleting a row
func (h *rowDeleteController) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		api.Respond(w, r, api.Error("delete_row must be POST"))
		return
	}

	if err := r.ParseForm(); err != nil {
		api.Respond(w, r, api.Error("failed to parse form"))
		return
	}

	// Validate required parameters
	schema := strings.TrimSpace(r.Form.Get("schema"))
	table := strings.TrimSpace(r.Form.Get("table"))
	col := strings.TrimSpace(r.Form.Get("key_column"))
	val := strings.TrimSpace(r.Form.Get("key_value"))

	if table == "" || col == "" || val == "" {
		api.Respond(w, r, api.Error("table, key_column and key_value are required"))
		return
	}

	// Validate identifiers
	if !sanitizeIdent(table) || (schema != "" && !sanitizeIdent(schema)) || !sanitizeIdent(col) {
		api.Respond(w, r, api.Error("invalid identifier"))
		return
	}

	// Check safe mode confirmation
	if h.config.SafeModeDefault && strings.TrimSpace(r.Form.Get("confirm")) != "yes" {
		api.Respond(w, r, api.Error("confirmation required (set confirm=yes)"))
		return
	}

	// Execute the delete operation
	if err := h.deleteRow(schema, table, col, val); err != nil {
		api.Respond(w, r, api.Error(err.Error()))
		return
	}

	api.Respond(w, r, api.Success("deleted"))
}

// deleteRow performs the actual row deletion with safety checks
func (h *rowDeleteController) deleteRow(schema, table, col, val string) error {
	sess := session.EnsureSession(nil, nil, h.config.SessionSecret)
	if sess == nil || sess.Conn == nil {
		return fmt.Errorf("not connected to database")
	}

	db, err := driver.OpenDBWithDSN(sess.Conn.Driver, sess.Conn.DSN)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %v", err)
	}
	defer sqlDB.Close()

	driverName := normalizeDriver(sess.Conn.Driver)
	qtable := table
	if schema != "" {
		qtable = schema + "." + table
	}
	qtable = quoteIdent(driverName, qtable)
	qcol := quoteIdent(driverName, col)

	// Transactional safety check + delete
	return db.Transaction(func(tx *gorm.DB) error {
		// Safety check: ensure exactly one row matches
		countSQL := "SELECT COUNT(*) FROM " + qtable + " WHERE " + qcol + " = ?"
		var cnt int64
		if err := tx.Raw(countSQL, val).Scan(&cnt).Error; err != nil {
			return err
		}
		if cnt != 1 {
			return fmt.Errorf("refusing to delete: match count != 1")
		}

		// Perform delete with dialect-specific single-row hints where available
		var delSQL string
		switch driverName {
		case "mysql":
			delSQL = "DELETE FROM " + qtable + " WHERE " + qcol + " = ? LIMIT 1"
		case "sqlserver":
			delSQL = "DELETE TOP (1) FROM " + qtable + " WHERE " + qcol + " = ?"
		default:
			delSQL = "DELETE FROM " + qtable + " WHERE " + qcol + " = ?"
		}

		return tx.Exec(delSQL, val).Error
	})
}

// sanitizeIdent checks if an identifier contains only safe characters
func sanitizeIdent(ident string) bool {
	// Allow alphanumeric and underscore, but not starting with a number
	if len(ident) == 0 || (ident[0] >= '0' && ident[0] <= '9') {
		return false
	}
	for _, c := range ident {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '_') {
			return false
		}
	}
	return true
}

// quoteIdent quotes an identifier based on the database driver
func quoteIdent(driver, ident string) string {
	switch driver {
	case "mysql":
		return "`" + strings.ReplaceAll(ident, "`", "``") + "`"
	case "postgres", "postgresql":
		return `"` + strings.ReplaceAll(ident, `"`, `""`) + `"`
	case "sqlite", "sqlite3":
		return `"` + strings.ReplaceAll(ident, `"`, `""`) + `"`
	case "sqlserver":
		return `[` + strings.ReplaceAll(ident, `]`, `]]`) + `]`
	default:
		return ident
	}
}

// normalizeDriver normalizes the driver name for consistent comparison
func normalizeDriver(driver string) string {
	switch driver {
	case "postgres", "postgresql", "pgx":
		return "postgres"
	case "mysql":
		return "mysql"
	case "sqlite", "sqlite3":
		return "sqlite"
	case "sqlserver", "mssql":
		return "sqlserver"
	default:
		return driver
	}
}
