package weebase

import (
	"fmt"
	"net/http"
	"strings"

	"gorm.io/gorm"
)

// handleDeleteRow deletes a single row by equality on one column.
// Safe mode: requires confirm=yes when h.opts.SafeModeDefault is true.
// Params: schema (optional), table, key_column, key_value, confirm ("yes")
func (h *Handler) handleDeleteRow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, r, "delete_row must be POST")
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
	if h.opts.SafeModeDefault && strings.TrimSpace(r.Form.Get("confirm")) != "yes" {
		WriteError(w, r, "confirmation required (set confirm=yes)")
		return
	}
	schema := strings.TrimSpace(r.Form.Get("schema"))
	table := strings.TrimSpace(r.Form.Get("table"))
	col := strings.TrimSpace(r.Form.Get("key_column"))
	val := strings.TrimSpace(r.Form.Get("key_value"))
	if table == "" || col == "" || val == "" {
		WriteError(w, r, "table, key_column and key_value are required")
		return
	}
	if !sanitizeIdent(table) || (schema != "" && !sanitizeIdent(schema)) || !sanitizeIdent(col) {
		WriteError(w, r, "invalid identifier")
		return
	}
	driver := normalizeDriver(s.Conn.Driver)
	qtable := table
	if schema != "" {
		qtable = schema + "." + table
	}
	qtable = quoteIdent(driver, qtable)
	qcol := quoteIdent(driver, col)

	// Transactional safety check + delete
	if err := s.Conn.DB.Transaction(func(tx *gorm.DB) error {
		// Safety check: ensure exactly one row matches
		countSQL := "SELECT COUNT(*) FROM " + qtable + " WHERE " + qcol + " = ?"
		var cnt int64
		if err := tx.Raw(countSQL, val).Scan(&cnt).Error; err != nil {
			return err
		}
		if cnt != 1 {
			return fmt.Errorf("refusing to delete: match count != 1")
		}
		// Perform delete with dialect-specific single-row hints where available
		var delSQL string
		var args []any = []any{val}
		switch driver {
		case "mysql":
			delSQL = "DELETE FROM " + qtable + " WHERE " + qcol + " = ? LIMIT 1"
		case "sqlserver":
			delSQL = "DELETE TOP (1) FROM " + qtable + " WHERE " + qcol + " = ?"
		default:
			delSQL = "DELETE FROM " + qtable + " WHERE " + qcol + " = ?"
		}
		if err := tx.Exec(delSQL, args...).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		WriteError(w, r, err.Error())
		return
	}
	WriteSuccess(w, r, http.StatusOK, "deleted")
}
