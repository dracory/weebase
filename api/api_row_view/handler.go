package api_row_view

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/dracory/api"
	"github.com/dracory/weebase/shared/session"
	"github.com/dracory/weebase/shared/types"
)

// RowView handles viewing a single row from a table
type RowView struct {
	config types.Config
}

// New creates a new RowView handler
func New(config types.Config) *RowView {
	return &RowView{config: config}
}

// Handle processes the request
func (h *RowView) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		api.Respond(w, r, api.Error("view_row must be GET"))
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

	table := strings.TrimSpace(r.Form.Get("table"))
	schema := strings.TrimSpace(r.Form.Get("schema"))
	keyColumn := strings.TrimSpace(r.Form.Get("key_column"))
	keyValue := strings.TrimSpace(r.Form.Get("key_value"))

	// Validate required parameters
	if table == "" || keyColumn == "" || keyValue == "" {
		api.Respond(w, r, api.Error("table, key_column and key_value are required"))
		return
	}

	// Validate identifiers
	if !sanitizeIdent(table) || (schema != "" && !sanitizeIdent(schema)) || !sanitizeIdent(keyColumn) {
		api.Respond(w, r, api.Error("invalid table, schema or key column identifier"))
		return
	}

	// Open database connection
	db, err := sql.Open(sess.Conn.Driver, sess.Conn.DSN)
	if err != nil {
		api.Respond(w, r, api.Error(fmt.Sprintf("failed to connect to database: %v", err)))
		return
	}
	defer db.Close()

	// Build query
	qtable := table
	if schema != "" {
		qtable = schema + "." + table
	}
	qtable = quoteIdent(sess.Conn.Driver, qtable)
	qcol := quoteIdent(sess.Conn.Driver, keyColumn)

	// Execute query
	query := "SELECT * FROM " + qtable + " WHERE " + qcol + " = ?"
	rows, err := db.Query(query, keyValue)
	if err != nil {
		api.Respond(w, r, api.Error(fmt.Sprintf("query failed: %v", err)))
		return
	}
	defer rows.Close()

	// Get column names
	cols, err := rows.Columns()
	if err != nil {
		api.Respond(w, r, api.Error(fmt.Sprintf("failed to get columns: %v", err)))
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
		api.Respond(w, r, api.Error(fmt.Sprintf("failed to scan row: %v", err)))
		return
	}

	// Convert to map
	out := make(map[string]any, len(cols))
	for i, c := range cols {
		out[c] = vals[i]
	}

	api.Respond(w, r, api.SuccessWithData("row", map[string]any{"row": out}))
}

// sanitizeIdent checks if an identifier contains only safe characters
func sanitizeIdent(s string) bool {
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
