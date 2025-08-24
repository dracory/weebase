package weebase

import (
	"net/http"
	"strings"

	"gorm.io/gorm"
)

// handleInsertRow inserts a row using CSV columns and values.
// Params: schema (optional), table (required), cols (csv), vals (csv), confirm ("yes" when SafeMode)
func (h *Handler) handleInsertRow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, r, "insert_row must be POST")
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
	colsCSV := strings.TrimSpace(r.Form.Get("cols"))
	valsCSV := strings.TrimSpace(r.Form.Get("vals"))
	if table == "" || colsCSV == "" || valsCSV == "" {
		WriteError(w, r, "table, cols, vals are required")
		return
	}
	cols := splitCSV(colsCSV)
	vals := splitCSV(valsCSV)
	if len(cols) == 0 || len(cols) != len(vals) {
		WriteError(w, r, "number of cols must equal number of vals")
		return
	}
	if !sanitizeIdent(table) || (schema != "" && !sanitizeIdent(schema)) {
		WriteError(w, r, "invalid identifier")
		return
	}
	for _, c := range cols {
		if !sanitizeIdent(c) {
			WriteError(w, r, "invalid column name")
			return
		}
	}
	driver := normalizeDriver(s.Conn.Driver)
	qtable := table
	if schema != "" {
		qtable = schema + "." + table
	}
	qtable = quoteIdent(driver, qtable)
	qcols := make([]string, len(cols))
	ph := make([]string, len(cols))
	args := make([]any, len(vals))
	for i := range cols {
		qcols[i] = quoteIdent(driver, cols[i])
		ph[i] = "?"
		args[i] = vals[i]
	}
	sqlStr := "INSERT INTO " + qtable + " (" + strings.Join(qcols, ", ") + ") VALUES (" + strings.Join(ph, ", ") + ")"
	// Transactional insert
	if err := s.Conn.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(sqlStr, args...).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		WriteError(w, r, err.Error())
		return
	}
	WriteSuccess(w, r, http.StatusOK, "inserted")
}
