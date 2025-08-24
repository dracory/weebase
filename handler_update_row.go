package weebase

import (
	"fmt"
	"net/http"
	"strings"

	"gorm.io/gorm"
)

// handleUpdateRow updates columns on a single row identified by key_column=value.
// Params: schema (optional), table (required), key_column, key_value, set_cols (csv), set_vals (csv), confirm ("yes" when SafeMode)
func (h *Handler) handleUpdateRow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, r, "update_row must be POST")
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
	keyCol := strings.TrimSpace(r.Form.Get("key_column"))
	keyVal := strings.TrimSpace(r.Form.Get("key_value"))
	setColsCSV := strings.TrimSpace(r.Form.Get("set_cols"))
	setValsCSV := strings.TrimSpace(r.Form.Get("set_vals"))
	if table == "" || keyCol == "" || keyVal == "" || setColsCSV == "" || setValsCSV == "" {
		WriteError(w, r, "table, key_column, key_value, set_cols, set_vals are required")
		return
	}
	setCols := splitCSV(setColsCSV)
	setVals := splitCSV(setValsCSV)
	if len(setCols) == 0 || len(setCols) != len(setVals) {
		WriteError(w, r, "number of set_cols must equal number of set_vals")
		return
	}
	if !sanitizeIdent(table) || (schema != "" && !sanitizeIdent(schema)) || !sanitizeIdent(keyCol) {
		WriteError(w, r, "invalid identifier")
		return
	}
	for _, c := range setCols {
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
	qkey := quoteIdent(driver, keyCol)
	// Transactional safety check + update
	if err := s.Conn.DB.Transaction(func(tx *gorm.DB) error {
		// Safety: ensure exactly one row will be updated
		countSQL := "SELECT COUNT(*) FROM " + qtable + " WHERE " + qkey + " = ?"
		var cnt int64
		if err := tx.Raw(countSQL, keyVal).Scan(&cnt).Error; err != nil {
			return err
		}
		if cnt != 1 {
			return fmt.Errorf("refusing to update: match count != 1")
		}
		sets := make([]string, len(setCols))
		args := make([]any, 0, len(setCols)+1)
		for i, c := range setCols {
			sets[i] = quoteIdent(driver, c) + " = ?"
			args = append(args, setVals[i])
		}
		args = append(args, keyVal)
		sqlStr := "UPDATE " + qtable + " SET " + strings.Join(sets, ", ") + " WHERE " + qkey + " = ?"
		if err := tx.Exec(sqlStr, args...).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		WriteError(w, r, err.Error())
		return
	}
	WriteSuccess(w, r, http.StatusOK, "updated")
}
