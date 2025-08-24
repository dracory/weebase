package weebase

import (
	"net/http"
	"strings"
)

// handleViewDefinition returns the SQL definition of a view.
func (h *Handler) handleViewDefinition(w http.ResponseWriter, r *http.Request) {
	s := EnsureSession(w, r, h.opts.SessionSecret)
	if s.Conn == nil || s.Conn.DB == nil {
		WriteError(w, r, "not connected")
		return
	}
	_ = r.ParseForm()
	schema := strings.TrimSpace(r.Form.Get("schema"))
	view := strings.TrimSpace(r.Form.Get("view"))
	if view == "" {
		WriteError(w, r, "view is required")
		return
	}
	if !sanitizeIdent(view) || (schema != "" && !sanitizeIdent(schema)) {
		WriteError(w, r, "invalid identifier")
		return
	}
	driver := normalizeDriver(s.Conn.Driver)
	var sqlStr string
	var def string
	var err error
	switch driver {
	case "postgres":
		// Prefer pg_get_viewdef; schema optional (default public)
		if schema == "" {
			schema = "public"
		}
		err = s.Conn.DB.Raw(
			"SELECT pg_get_viewdef(format('%%s.%%s', ?, ?), true)", schema, view,
		).Scan(&def).Error
		if err == nil && def == "" {
			// Fallback via information_schema
			err = s.Conn.DB.Raw(
				"SELECT view_definition FROM information_schema.views WHERE table_schema = ? AND table_name = ?",
				schema, view,
			).Scan(&def).Error
		}
	case "mysql":
		// MySQL requires database (schema). SHOW CREATE VIEW returns two columns: View, Create View
		if schema == "" {
			WriteError(w, r, "schema required for mysql")
			return
		}
		rows, qerr := s.Conn.DB.Raw("SHOW CREATE VIEW `" + schema + "`.`" + view + "`").Rows()
		if qerr != nil {
			err = qerr
			break
		}
		defer rows.Close()
		if rows.Next() {
			var viewName, createSQL string
			if scanErr := rows.Scan(&viewName, &createSQL); scanErr != nil {
				err = scanErr
				break
			}
			def = createSQL
		}
	case "sqlite":
		// sqlite_master holds the SQL
		err = s.Conn.DB.Raw(
			"SELECT sql FROM sqlite_master WHERE type='view' AND name = ?",
			view,
		).Scan(&def).Error
	case "sqlserver":
		// Join sys.views to sys.sql_modules; default schema dbo
		if schema == "" {
			schema = "dbo"
		}
		err = s.Conn.DB.Raw(
			"SELECT m.definition FROM sys.views v JOIN sys.schemas s ON v.schema_id=s.schema_id JOIN sys.sql_modules m ON v.object_id=m.object_id WHERE s.name = ? AND v.name = ?",
			schema, view,
		).Scan(&def).Error
	default:
		WriteError(w, r, "unsupported driver")
		return
	}
	if err != nil {
		WriteError(w, r, err.Error())
		return
	}
	// If still empty, surface a friendly message
	if strings.TrimSpace(def) == "" {
		def = "<empty>"
	}
	WriteSuccessWithData(w, r, "ok", map[string]any{"definition": def, "sql": sqlStr})
}
