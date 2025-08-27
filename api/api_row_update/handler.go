package api_row_update

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/dracory/api"
	"github.com/dracory/weebase/shared/driver"
	"github.com/dracory/weebase/shared/session"
	"github.com/dracory/weebase/shared/types"
	"gorm.io/gorm"
)

// RowUpdate handles row update requests
type RowUpdate struct {
	config         types.Config
	safeModeDefault bool
}
// New creates a new RowUpdate handler
func New(config types.Config, safeModeDefault bool) *RowUpdate {
	return &RowUpdate{
		config:         config,
		safeModeDefault: safeModeDefault,
	}
}

// Handle processes the update row request
func (h *RowUpdate) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		api.Respond(w, r, api.Error("update_row must be POST"))
		return
	}

	sess := session.EnsureSession(nil, nil, h.config.SessionSecret)
	if sess == nil || sess.Conn == nil {
		api.Respond(w, r, api.Error("not connected to database"))
		return
	}

	if err := r.ParseForm(); err != nil {
		api.Respond(w, r, api.Error("failed to parse form"))
		return
	}

	// Check for confirmation in safe mode
	if h.safeModeDefault && strings.TrimSpace(r.Form.Get("confirm")) != "yes" {
		api.Respond(w, r, api.Error("confirmation required (set confirm=yes)"))
		return
	}

	table := strings.TrimSpace(r.Form.Get("table"))
	schema := strings.TrimSpace(r.Form.Get("schema"))
	keyColumn := strings.TrimSpace(r.Form.Get("key_column"))
	keyValue := strings.TrimSpace(r.Form.Get("key_value"))
	setColsCSV := strings.TrimSpace(r.Form.Get("set_cols"))
	setValsCSV := strings.TrimSpace(r.Form.Get("set_vals"))

	// Validate required parameters
	if table == "" || keyColumn == "" || keyValue == "" || setColsCSV == "" || setValsCSV == "" {
		api.Respond(w, r, api.Error("table, key_column, key_value, set_cols, and set_vals are required"))
		return
	}

	setCols := splitCSV(setColsCSV)
	setVals := splitCSV(setValsCSV)

	if len(setCols) == 0 || len(setCols) != len(setVals) {
		api.Respond(w, r, api.Error("number of set_cols must equal number of set_vals"))
		return
	}

	// Validate identifiers
	if !sanitizeIdent(table) || (schema != "" && !sanitizeIdent(schema)) || !sanitizeIdent(keyColumn) {
		api.Respond(w, r, api.Error("invalid table, schema or key column identifier"))
		return
	}

	for _, col := range setCols {
		if !sanitizeIdent(col) {
			api.Respond(w, r, api.Error("invalid column name: "+col))
			return
		}
	}

	// Get database connection
	db, err := driver.OpenDBWithDSN(sess.Conn.Driver, sess.Conn.DSN)
	if err != nil {
		api.Respond(w, r, api.Error(fmt.Sprintf("failed to connect to database: %v", err)))
		return
	}
	sqlDB, err := db.DB()
	if err != nil {
		api.Respond(w, r, api.Error(fmt.Sprintf("failed to get database instance: %v", err)))
		return
	}
	defer sqlDB.Close()

	driverName := normalizeDriver(sess.Conn.Driver)
	qtable := table
	if schema != "" {
		qtable = schema + "." + table
	}
	qtable = quoteIdent(driverName, qtable)
	qkey := quoteIdent(driverName, keyColumn)

	// Execute the update in a transaction
	err = db.Transaction(func(tx *gorm.DB) error {
		// Safety check: ensure exactly one row will be updated
		countSQL := "SELECT COUNT(*) FROM " + qtable + " WHERE " + qkey + " = ?"
		var cnt int64
		if err := tx.Raw(countSQL, keyValue).Scan(&cnt).Error; err != nil {
			return fmt.Errorf("safety check failed: %w", err)
		}

		if cnt != 1 {
			return fmt.Errorf("refusing to update: match count (%d) != 1", cnt)
		}

		// Build the SET clause
		sets := make([]string, len(setCols))
		args := make([]any, 0, len(setCols)+1)

		for i, c := range setCols {
			sets[i] = quoteIdent(driverName, c) + " = ?"
			args = append(args, setVals[i])
		}
		args = append(args, keyValue)

		// Build and execute the UPDATE query
		sqlStr := "UPDATE " + qtable + " SET " + strings.Join(sets, ", ") + " WHERE " + qkey + " = ?"
		if err := tx.Exec(sqlStr, args...).Error; err != nil {
			return fmt.Errorf("update failed: %w", err)
		}

		return nil
	})

	if err != nil {
		api.Respond(w, r, api.Error(err.Error()))
		return
	}

	api.Respond(w, r, api.Success("row updated"))
}

// Helper functions

// splitCSV splits a comma-separated string into a slice of strings
func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

// sanitizeIdent checks if an identifier contains only safe characters
func sanitizeIdent(s string) bool {
	return s != "" && !strings.ContainsAny(s, " ;'\"`")
}

// quoteIdent quotes an identifier for the given SQL dialect
func quoteIdent(driver, ident string) string {
	switch normalizeDriver(driver) {
	case "mysql":
		return "`" + strings.ReplaceAll(ident, "`", "``") + "`"
	case "postgres", "sqlite":
		return `"` + strings.ReplaceAll(ident, `"`, `""`) + `"`
	case "sqlserver":
		return "[" + strings.ReplaceAll(ident, "]", "]]") + "]"
	default:
		return ident
	}
}

// normalizeDriver normalizes the database driver name
func normalizeDriver(driver string) string {
	switch driver {
	case "postgres", "postgresql":
		return "postgres"
	case "mysql", "mariadb":
		return "mysql"
	case "sqlserver", "mssql":
		return "sqlserver"
	case "sqlite", "sqlite3":
		return "sqlite"
	default:
		return driver
	}
}
