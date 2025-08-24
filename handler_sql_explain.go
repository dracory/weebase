package weebase

import (
	"fmt"
	"net/http"
	"strings"
)

// handleSQLExplain runs an EXPLAIN (or dialect equivalent) for the provided SQL.
// Params: sql (required)
func (h *Handler) handleSQLExplain(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, r, "sql_explain must be POST")
		return
	}
	if !VerifyCSRF(r, h.opts.SessionSecret) {
		WriteError(w, r, "csrf invalid")
		return
	}
	s := EnsureSession(w, r, h.opts.SessionSecret)
	if s.Conn == nil || s.Conn.DB == nil {
		WriteError(w, r, "not connected")
		return
	}
	_ = r.ParseForm()
	sqlText := strings.TrimSpace(r.Form.Get("sql"))
	if sqlText == "" {
		WriteError(w, r, "sql is required")
		return
	}
	driver := normalizeDriver(s.Conn.Driver)
	var explainSQL string
	switch driver {
	case "postgres", "mysql", "sqlserver":
		explainSQL = "EXPLAIN " + sqlText
	case "sqlite":
		explainSQL = "EXPLAIN QUERY PLAN " + sqlText
	default:
		explainSQL = "EXPLAIN " + sqlText
	}
	rows, err := s.Conn.DB.Raw(explainSQL).Rows()
	if err != nil {
		WriteError(w, r, fmt.Sprintf("explain error: %v", err))
		return
	}
	defer rows.Close()
	if err := writeRowsResult(w, r, rows); err != nil {
		WriteError(w, r, err.Error())
		return
	}
}
