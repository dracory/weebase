package weebase

import (
	"database/sql"
	"net/http"
	"strings"

	"gorm.io/gorm"
)

// handleSQLExecute executes a single SQL statement.
// Params: sql (required), transactional=[true|false]
func (h *Handler) handleSQLExecute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, r, "sql_execute must be POST")
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
	// Safe mode guard for clearly destructive DDL
	low := strings.ToLower(strings.TrimSpace(sqlText))
	if h.opts.SafeModeDefault && (strings.HasPrefix(low, "drop ") || strings.HasPrefix(low, "alter ") || strings.HasPrefix(low, "truncate ")) {
		WriteError(w, r, "blocked by safe mode: DDL requires explicit feature to be implemented")
		return
	}
	transactional := strings.EqualFold(strings.TrimSpace(r.Form.Get("transactional")), "true")

	run := func(tx *gorm.DB) error {
		// Heuristic: treat as SELECT if it starts with SELECT/WITH/SHOW/PRAGMA
		head := strings.Fields(low)
		stmt := ""
		if len(head) > 0 {
			stmt = head[0]
		}
		isSelect := stmt == "select" || stmt == "with" || stmt == "show" || stmt == "pragma" || stmt == "explain"
		if h.opts.ReadOnlyMode && !isSelect {
			WriteError(w, r, "read-only mode: only SELECT-like statements are allowed")
			return nil
		}
		switch stmt {
		case "select", "with", "show", "pragma", "explain":
			// Query rows (limit results for safety)
			rows, err := tx.Raw(sqlText).Rows()
			if err != nil {
				return err
			}
			defer rows.Close()
			return writeRowsResult(w, r, rows)
		default:
			// Exec non-select
			res := tx.Exec(sqlText)
			if res.Error != nil {
				return res.Error
			}
			affected := res.RowsAffected
			WriteSuccessWithData(w, r, "ok", map[string]any{"rows_affected": affected})
			return nil
		}
	}

	if transactional {
		if err := s.Conn.DB.Transaction(func(tx *gorm.DB) error { return run(tx) }); err != nil {
			WriteError(w, r, err.Error())
			return
		}
		return
	}
	if err := run(s.Conn.DB); err != nil {
		WriteError(w, r, err.Error())
		return
	}
}

// writeRowsResult scans sql.Rows into a JSON-friendly structure with a sane row cap.
func writeRowsResult(w http.ResponseWriter, r *http.Request, rows *sql.Rows) error {
	cols, err := rows.Columns()
	if err != nil {
		return err
	}
	// Cap results to avoid huge payloads
	const maxRows = 200
	out := make([]map[string]any, 0, 32)
	for len(out) < maxRows && rows.Next() {
		vals := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return err
		}
		m := make(map[string]any, len(cols))
		for i, c := range cols {
			m[c] = vals[i]
		}
		out = append(out, m)
	}
	WriteSuccessWithData(w, r, "ok", map[string]any{
		"columns": cols,
		"rows":    out,
		"truncated": !rows.NextResultSet() && rows.Err() == nil && len(out) == maxRows,
	})
	return nil
}
