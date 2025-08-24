package weebase

import (
	"net/http"
	"strconv"
	"strings"
)

// handleBrowseRows selects rows from a table with pagination.
func (h *Handler) handleBrowseRows(w http.ResponseWriter, r *http.Request) {
	s := EnsureSession(w, r, h.opts.SessionSecret)
	if s.Conn == nil || s.Conn.DB == nil {
		WriteError(w, r, "not connected")
		return
	}
	_ = r.ParseForm()
	schema := strings.TrimSpace(r.Form.Get("schema"))
	table := strings.TrimSpace(r.Form.Get("table"))
	if table == "" {
		WriteError(w, r, "table is required")
		return
	}
	if !sanitizeIdent(table) || (schema != "" && !sanitizeIdent(schema)) {
		WriteError(w, r, "invalid identifier")
		return
	}
	limit := 50
	offset := 0
	if v := strings.TrimSpace(r.Form.Get("limit")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 500 {
			limit = n
		}
	}
	if v := strings.TrimSpace(r.Form.Get("offset")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}
	driver := normalizeDriver(s.Conn.Driver)
	var sqlStr string
	var args []any
	switch driver {
	case "postgres":
		if schema == "" {
			sqlStr = "SELECT * FROM \"" + table + "\" LIMIT ? OFFSET ?"
			args = []any{limit, offset}
		} else {
			sqlStr = "SELECT * FROM \"" + schema + "\".\"" + table + "\" LIMIT ? OFFSET ?"
			args = []any{limit, offset}
		}
	case "mysql":
		if schema == "" {
			WriteError(w, r, "schema required for mysql")
			return
		}
		sqlStr = "SELECT * FROM `" + schema + "`.`" + table + "` LIMIT ? OFFSET ?"
		args = []any{limit, offset}
	case "sqlite":
		sqlStr = "SELECT * FROM `" + table + "` LIMIT ? OFFSET ?"
		args = []any{limit, offset}
	case "sqlserver":
		// Without ORDER BY, OFFSET is not allowed; use TOP and ignore offset for now.
		if schema == "" {
			sqlStr = "SELECT TOP (?) * FROM [" + table + "]"
			args = []any{limit}
		} else {
			sqlStr = "SELECT TOP (?) * FROM [" + schema + "].[" + table + "]"
			args = []any{limit}
		}
	default:
		WriteError(w, r, "unsupported driver")
		return
	}
	// Execute and materialize rows into []map[string]any
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
	out := make([]map[string]any, 0, limit)
	for rows.Next() {
		vals := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			WriteError(w, r, err.Error())
			return
		}
		m := make(map[string]any, len(cols))
		for i, c := range cols {
			m[c] = vals[i]
		}
		out = append(out, m)
	}
	WriteSuccessWithData(w, r, "ok", map[string]any{"rows": out, "limit": limit, "offset": offset})
}
