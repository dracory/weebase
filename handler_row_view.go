package weebase

import (
	"net/http"
	"strings"
)

// handleRowView returns a single row from a table by an equality filter on one column.
// Params: schema (optional per dialect), table (required), key_column (required), key_value (required)
func (h *Handler) handleRowView(w http.ResponseWriter, r *http.Request) {
	s := EnsureSession(w, r, h.opts.SessionSecret)
	if s.Conn == nil || s.Conn.DB == nil {
		WriteError(w, r, "not connected")
		return
	}
	_ = r.ParseForm()
	schema := strings.TrimSpace(r.Form.Get("schema"))
	table := strings.TrimSpace(r.Form.Get("table"))
	col := strings.TrimSpace(r.Form.Get("key_column"))
	val := strings.TrimSpace(r.Form.Get("key_value"))
	if table == "" || col == "" || val == "" {
		WriteError(w, r, "table, key_column and key_value are required")
		return
	}
	// Validate identifiers to prevent injection in identifier positions
	if !sanitizeIdent(table) || (schema != "" && !sanitizeIdent(schema)) || !sanitizeIdent(col) {
		WriteError(w, r, "invalid identifier")
		return
	}
	driver := normalizeDriver(s.Conn.Driver)
	// Build SQL with identifier quoting and single parameter
	qtable := table
	if schema != "" {
		qtable = schema + "." + table
	}
	qtable = quoteIdent(driver, qtable)
	qcol := quoteIdent(driver, col)
	var sqlStr string
	var args []any
	switch driver {
	case "sqlserver":
		sqlStr = "SELECT TOP (1) * FROM " + qtable + " WHERE " + qcol + " = ?"
		args = []any{val}
	default:
		sqlStr = "SELECT * FROM " + qtable + " WHERE " + qcol + " = ? LIMIT 1"
		args = []any{val}
	}
	rows, err := s.Conn.DB.Raw(sqlStr, args...).Rows()
	if err != nil {
		WriteError(w, r, err.Error())
		return
	}
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		WriteError(w, r, err.Error())
		return
	}
	if !rows.Next() {
		WriteSuccessWithData(w, r, "ok", map[string]any{"row": nil})
		return
	}
	vals := make([]any, len(cols))
	ptrs := make([]any, len(cols))
	for i := range vals {
		ptrs[i] = &vals[i]
	}
	if err := rows.Scan(ptrs...); err != nil {
		WriteError(w, r, err.Error())
		return
	}
	out := make(map[string]any, len(cols))
	for i, c := range cols {
		out[c] = vals[i]
	}
	WriteSuccessWithData(w, r, "ok", map[string]any{"row": out})
}
