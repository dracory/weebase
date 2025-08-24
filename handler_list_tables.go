package weebase

import (
	"net/http"
	"strconv"
	"strings"
)

// handleListTables returns table names for a given schema (where applicable).
func (h *Handler) handleListTables(w http.ResponseWriter, r *http.Request) {
	s := EnsureSession(w, r, h.opts.SessionSecret)
	if s.Conn == nil || s.Conn.DB == nil {
		WriteError(w, r, "not connected")
		return
	}
	_ = r.ParseForm()
	schema := strings.TrimSpace(r.Form.Get("schema"))
	q := strings.TrimSpace(r.Form.Get("q"))
	includeViews := strings.TrimSpace(r.Form.Get("include_views")) != "" && strings.TrimSpace(r.Form.Get("include_views")) != "0"
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
	type row struct{ Name string }
	var rows []row
	var err error
	switch driver {
	case "postgres":
		if schema == "" {
			schema = "public"
		}
		// Build query: filter by schema, optionally include views, optional name search via ILIKE, with limit/offset
		base := "SELECT table_name AS name FROM information_schema.tables WHERE table_schema = ?"
		args := []any{schema}
		if !includeViews {
			base += " AND table_type = 'BASE TABLE'"
		}
		if q != "" {
			base += " AND table_name ILIKE ?"
			args = append(args, "%"+q+"%")
		}
		base += " ORDER BY name LIMIT ? OFFSET ?"
		args = append(args, limit, offset)
		err = s.Conn.DB.Raw(base, args...).Scan(&rows).Error
	case "mysql":
		// In MySQL, schema == database
		if schema == "" {
			WriteError(w, r, "schema required")
			return
		}
		base := "SELECT table_name AS name FROM information_schema.tables WHERE table_schema = ?"
		args := []any{schema}
		if !includeViews {
			base += " AND table_type = 'BASE TABLE'"
		}
		if q != "" {
			base += " AND table_name LIKE ?"
			args = append(args, "%"+q+"%")
		}
		base += " ORDER BY name LIMIT ? OFFSET ?"
		args = append(args, limit, offset)
		err = s.Conn.DB.Raw(base, args...).Scan(&rows).Error
	case "sqlite":
		// sqlite_master holds tables and views
		base := "SELECT name AS name FROM sqlite_master WHERE "
		if includeViews {
			base += "type IN ('table','view')"
		} else {
			base += "type = 'table'"
		}
		var args []any
		if q != "" {
			base += " AND name LIKE ?"
			args = append(args, "%"+q+"%")
		}
		base += " ORDER BY name LIMIT ? OFFSET ?"
		args = append(args, limit, offset)
		err = s.Conn.DB.Raw(base, args...).Scan(&rows).Error
	case "sqlserver":
		if schema == "" {
			schema = "dbo"
		}
		// Use sys.objects to allow including views (U=user table, V=view). OFFSET/FETCH requires ORDER BY.
		base := "SELECT o.name AS name FROM sys.objects o JOIN sys.schemas s ON o.schema_id=s.schema_id WHERE s.name = ? AND o.type IN (" + func() string { if includeViews { return "'U','V'" } else { return "'U'" } }() + ")"
		args := []any{schema}
		if q != "" {
			base += " AND o.name LIKE ?"
			args = append(args, "%"+q+"%")
		}
		base += " ORDER BY o.name OFFSET ? ROWS FETCH NEXT ? ROWS ONLY"
		args = append(args, offset, limit)
		err = s.Conn.DB.Raw(base, args...).Scan(&rows).Error
	default:
		WriteError(w, r, "unsupported driver")
		return
	}
	if err != nil {
		WriteError(w, r, err.Error())
		return
	}
	names := make([]string, 0, len(rows))
	for _, x := range rows {
		names = append(names, x.Name)
	}
	WriteSuccessWithData(w, r, "ok", map[string]any{"tables": names, "limit": limit, "offset": offset})
}
