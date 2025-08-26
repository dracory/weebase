package api_row_insert

import (
	"net/http"
	"strings"

	"github.com/dracory/api"
	"github.com/dracory/weebase/shared/session"
	"gorm.io/gorm"
)

// RowInsert handles row insertion requests
type RowInsert struct {
	conn           *session.ActiveConnection
	safeModeDefault bool
}

// New creates a new RowInsert handler
func New(conn *session.ActiveConnection, safeModeDefault bool) *RowInsert {
	return &RowInsert{
		conn:           conn,
		safeModeDefault: safeModeDefault,
	}
}

// Handle processes the insert row request
func (h *RowInsert) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		api.Respond(w, r, api.Error("insert_row must be POST"))
		return
	}

	if h.conn == nil || h.conn.DB == nil {
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

	schema := strings.TrimSpace(r.Form.Get("schema"))
	table := strings.TrimSpace(r.Form.Get("table"))
	colsCSV := strings.TrimSpace(r.Form.Get("cols"))
	valsCSV := strings.TrimSpace(r.Form.Get("vals"))

	// Validate required parameters
	if table == "" || colsCSV == "" || valsCSV == "" {
		api.Respond(w, r, api.Error("table, cols, and vals are required"))
		return
	}

	// Parse CSV values
	cols := splitCSV(colsCSV)
	vals := splitCSV(valsCSV)

	if len(cols) == 0 || len(cols) != len(vals) {
		api.Respond(w, r, api.Error("number of cols must equal number of vals"))
		return
	}

	// Validate identifiers
	if !sanitizeIdent(table) || (schema != "" && !sanitizeIdent(schema)) {
		api.Respond(w, r, api.Error("invalid table or schema identifier"))
		return
	}

	for _, c := range cols {
		if !sanitizeIdent(c) {
			api.Respond(w, r, api.Error("invalid column name"))
			return
		}
	}

	// Get the gorm DB instance
	db, ok := h.conn.DB.(*gorm.DB)
	if !ok {
		api.Respond(w, r, api.Error("invalid database connection type"))
		return
	}

	// Prepare table and column names with proper quoting
	qtable := table
	if schema != "" {
		qtable = schema + "." + table
	}
	qtable = quoteIdent(h.conn.Driver, qtable)

	// Prepare the SQL query
	qcols := make([]string, len(cols))
	ph := make([]string, len(cols))
	args := make([]any, len(vals))

	for i := range cols {
		qcols[i] = quoteIdent(h.conn.Driver, cols[i])
		ph[i] = "?"
		args[i] = vals[i]
	}

	sqlStr := "INSERT INTO " + qtable + " (" + strings.Join(qcols, ", ") + ") VALUES (" + strings.Join(ph, ", ") + ")"

	// Execute the insert in a transaction
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(sqlStr, args...).Error; err != nil {
			return err
		}
		return nil
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
