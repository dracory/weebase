package api_row_view

import (
	"net/http"
	"strings"

	"github.com/dracory/api"
	"github.com/dracory/weebase/shared/session"
	"gorm.io/gorm"
)

// RowView handles viewing a single row from a table
type RowView struct {
	conn *session.ActiveConnection
}

// New creates a new RowView handler
func New(conn *session.ActiveConnection) *RowView {
	return &RowView{conn: conn}
}

// Handle processes the request
func (h *RowView) Handle(w http.ResponseWriter, r *http.Request) {
	if h.conn == nil || h.conn.DB == nil {
		api.Respond(w, r, api.Error("not connected to database"))
		return
	}

	_ = r.ParseForm()
	schema := strings.TrimSpace(r.Form.Get("schema"))
	table := strings.TrimSpace(r.Form.Get("table"))
	col := strings.TrimSpace(r.Form.Get("key_column"))
	val := strings.TrimSpace(r.Form.Get("key_value"))

	if table == "" || col == "" || val == "" {
		api.Respond(w, r, api.Error("table, key_column and key_value are required"))
		return
	}

	// Validate identifiers to prevent injection in identifier positions
	if !sanitizeIdent(table) || (schema != "" && !sanitizeIdent(schema)) || !sanitizeIdent(col) {
		api.Respond(w, r, api.Error("invalid identifier"))
		return
	}

	// Get the database connection
	db, ok := h.conn.DB.(*gorm.DB)
	if !ok {
		api.Respond(w, r, api.Error("invalid database connection type"))
		return
	}

	// Build query
	qtable := table
	if schema != "" {
		qtable = schema + "." + table
	}
	qtable = quoteIdent(h.conn.Driver, qtable)
	qcol := quoteIdent(h.conn.Driver, col)

	var sqlStr string
	var args []any

	switch normalizeDriver(h.conn.Driver) {
	case "sqlserver":
		sqlStr = "SELECT TOP (1) * FROM " + qtable + " WHERE " + qcol + " = ?"
		args = []any{val}
	default:
		sqlStr = "SELECT * FROM " + qtable + " WHERE " + qcol + " = ? LIMIT 1"
		args = []any{val}
	}

	// Execute query
	rows, err := db.Raw(sqlStr, args...).Rows()
	if err != nil {
		api.Respond(w, r, api.Error(err.Error()))
		return
	}
	defer rows.Close()

	// Get column names
	cols, err := rows.Columns()
	if err != nil {
		api.Respond(w, r, api.Error(err.Error()))
		return
	}

	// No rows found
	if !rows.Next() {
		api.Respond(w, r, api.SuccessWithData("row", map[string]any{"row": nil}))
		return
	}

	// Read row data
	vals := make([]any, len(cols))
	ptrs := make([]any, len(cols))
	for i := range vals {
		ptrs[i] = &vals[i]
	}

	if err := rows.Scan(ptrs...); err != nil {
		api.Respond(w, r, api.Error(err.Error()))
		return
	}

	// Convert to map
	out := make(map[string]any, len(cols))
	for i, c := range cols {
		out[c] = vals[i]
	}

	api.Respond(w, r, api.SuccessWithData("row", map[string]any{"row": out}))
}

// Helper functions from the original handler
func sanitizeIdent(s string) bool {
	// Simple check for now - should be extended with proper validation
	return s != "" && !strings.ContainsAny(s, " ;'\"`")
}

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
