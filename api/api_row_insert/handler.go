package api_row_insert

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

// RowInsert handles row insertion requests
type RowInsert struct {
	config         types.Config
	safeModeDefault bool
}

// New creates a new RowInsert handler
func New(config types.Config, safeModeDefault bool) *RowInsert {
	return &RowInsert{
		config:         config,
		safeModeDefault: safeModeDefault,
	}
}

// Handle processes the insert row request
func (h *RowInsert) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		api.Respond(w, r, api.Error("insert_row must be POST"))
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
	columns := splitCSV(r.Form.Get("columns"))
	values := splitCSV(r.Form.Get("values"))

	if table == "" || len(columns) == 0 || len(values) == 0 {
		api.Respond(w, r, api.Error("table, columns and values are required"))
		return
	}

	if len(columns) != len(values) {
		api.Respond(w, r, api.Error("number of columns and values must match"))
		return
	}

	// Validate identifiers
	for _, col := range columns {
		if !sanitizeIdent(col) {
			api.Respond(w, r, api.Error("invalid column identifier: "+col))
			return
		}
	}

	if !sanitizeIdent(table) || (schema != "" && !sanitizeIdent(schema)) {
		api.Respond(w, r, api.Error("invalid table or schema identifier"))
		return
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

	// Prepare columns and values
	qcols := make([]string, len(columns))
	ph := make([]string, len(columns))
	args := make([]any, len(values))

	for i, col := range columns {
		qcols[i] = quoteIdent(driverName, col)
		ph[i] = "?"
		args[i] = values[i]
	}

	sqlStr := "INSERT INTO " + qtable + " (" + strings.Join(qcols, ", ") + ") VALUES (" + strings.Join(ph, ", ") + ")"

	// Execute the insert in a transaction
	err = db.Transaction(func(tx *gorm.DB) error {
		return tx.Exec(sqlStr, args...).Error
	})

	if err != nil {
		api.Respond(w, r, api.Error(err.Error()))
		return
	}

	api.Respond(w, r, api.Success("row inserted"))
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
