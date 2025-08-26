package api_sql_execute

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/dracory/api"
	"github.com/dracory/weebase/shared/session"
	"gorm.io/gorm"
)

// SQLExecute handles SQL statement execution
type SQLExecute struct {
	conn           *session.ActiveConnection
	safeModeDefault bool
	readOnlyMode   bool
}

// New creates a new SQLExecute handler
func New(conn *session.ActiveConnection, safeModeDefault, readOnlyMode bool) *SQLExecute {
	return &SQLExecute{
		conn:           conn,
		safeModeDefault: safeModeDefault,
		readOnlyMode:   readOnlyMode,
	}
}

// Handle processes the request
func (h *SQLExecute) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		api.Respond(w, r, api.Error("sql_execute must be POST"))
		return
	}

	if h.conn == nil || h.conn.DB == nil {
		api.Respond(w, r, api.Error("not connected to database"))
		return
	}

	if err := r.ParseForm(); err != nil {
		api.Respond(w, r, api.Error("failed to parse form"))
		return
	}

	sqlText := strings.TrimSpace(r.Form.Get("sql"))
	if sqlText == "" {
		api.Respond(w, r, api.Error("sql is required"))
		return
	}

	// Safe mode guard for destructive DDL
	low := strings.ToLower(strings.TrimSpace(sqlText))
	if h.safeModeDefault && isDestructiveQuery(low) {
		api.Respond(w, r, api.Error("blocked by safe mode: DDL requires explicit feature to be implemented"))
		return
	}

	transactional := strings.EqualFold(strings.TrimSpace(r.Form.Get("transactional")), "true")

	run := func(tx *gorm.DB) error {
		// Heuristic: treat as SELECT if it starts with SELECT/WITH/SHOW/PRAGMA/EXPLAIN
		head := strings.Fields(low)
		stmt := ""
		if len(head) > 0 {
			stmt = head[0]
		}

		isSelect := stmt == "select" || stmt == "with" || stmt == "show" || stmt == "pragma" || stmt == "explain"
		if h.readOnlyMode && !isSelect {
			api.Respond(w, r, api.Error("read-only mode: only SELECT-like statements are allowed"))
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
			api.Respond(w, r, api.SuccessWithData("rows_affected", map[string]any{
				"rows_affected": affected,
			}))
			return nil
		}
	}

	// Get the gorm DB instance
	db, ok := h.conn.DB.(*gorm.DB)
	if !ok {
		api.Respond(w, r, api.Error("invalid database connection type"))
		return
	}

	// Execute in transaction if requested
	if transactional {
		tx := db.Begin()
		if tx.Error != nil {
			api.Respond(w, r, api.Error(tx.Error.Error()))
			return
		}

		err := run(tx)
		if err != nil {
			tx.Rollback()
			api.Respond(w, r, api.Error(err.Error()))
			return
		}

		if err := tx.Commit().Error; err != nil {
			api.Respond(w, r, api.Error(err.Error()))
			return
		}
		return
	}

	// Execute without transaction
	if err := run(db); err != nil {
		api.Respond(w, r, api.Error(err.Error()))
	}
}

// isDestructiveQuery checks if the query is potentially destructive
func isDestructiveQuery(query string) bool {
	return strings.HasPrefix(query, "drop ") || 
	       strings.HasPrefix(query, "alter ") || 
	       strings.HasPrefix(query, "truncate ")
}

// writeRowsResult scans sql.Rows into a JSON-friendly structure with a sane row cap
func writeRowsResult(w http.ResponseWriter, r *http.Request, rows *sql.Rows) error {
	cols, err := rows.Columns()
	if err != nil {
		return err
	}

	var results []map[string]any
	rowCount := 0
	maxRows := 1000 // Safety limit

	for rows.Next() && rowCount < maxRows {
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}

		if err := rows.Scan(columnPointers...); err != nil {
			return err
		}

		row := make(map[string]any)
		for i, colName := range cols {
			val := columnPointers[i].(*interface{})
			switch v := (*val).(type) {
			case []byte:
				row[colName] = string(v)
			default:
				row[colName] = v
			}
		}

		results = append(results, row)
		rowCount++
	}

	if err := rows.Err(); err != nil {
		return err
	}

	api.Respond(w, r, api.SuccessWithData("rows", map[string]any{
		"rows":       results,
		"row_count":  rowCount,
		"has_more":   rows.Next(),
		"limit":      maxRows,
	}))
	return nil
}
