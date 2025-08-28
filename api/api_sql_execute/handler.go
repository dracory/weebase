package api_sql_execute

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/dracory/api"
	"github.com/dracory/weebase/shared/session"
	"github.com/dracory/weebase/shared/types"
)

// SQLExecute handles SQL statement execution
type SQLExecute struct {
	config         types.Config
	safeModeDefault bool
	readOnlyMode   bool
}

// New creates a new SQLExecute handler
func New(config types.Config, safeModeDefault, readOnlyMode bool) *SQLExecute {
	return &SQLExecute{
		config:         config,
		safeModeDefault: safeModeDefault,
		readOnlyMode:   readOnlyMode,
	}
}

// normalizeSQL normalizes the SQL query for analysis
func normalizeSQL(query string) string {
	return strings.ToLower(strings.TrimSpace(query))
}

// isReadOnlyQuery checks if the query is a read-only query
func isReadOnlyQuery(query string) bool {
	normalized := normalizeSQL(query)
	return strings.HasPrefix(normalized, "select") ||
		strings.HasPrefix(normalized, "show") ||
		strings.HasPrefix(normalized, "explain")
}

// isDestructiveQuery checks if the query is potentially destructive
func isDestructiveQuery(query string) bool {
	normalized := normalizeSQL(query)
	return strings.HasPrefix(normalized, "drop ") ||
		strings.HasPrefix(normalized, "alter ") ||
		strings.HasPrefix(normalized, "truncate ") ||
		strings.HasPrefix(normalized, "delete ") ||
		strings.HasPrefix(normalized, "update ") ||
		strings.HasPrefix(normalized, "insert ")
}

// Handle processes the request
func (h *SQLExecute) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		api.Respond(w, r, api.Error("sql_execute must be POST"))
		return
	}

	sess := session.EnsureSession(nil, nil, h.config.SessionSecret)
	if sess == nil || sess.Conn == nil {
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
	if h.safeModeDefault && isDestructiveQuery(sqlText) && r.Form.Get("confirm") != "yes" {
		api.Respond(w, r, api.Error("confirmation required for destructive operation (add confirm=yes to force)"))
		return
	}

	// Read-only mode guard
	if h.readOnlyMode && !isReadOnlyQuery(sqlText) {
		api.Respond(w, r, api.Error("write operations are not allowed in read-only mode"))
		return
	}

	transactional := r.Form.Get("transactional") == "true"

	// Open database connection
	db, err := sql.Open(sess.Conn.Driver, sess.Conn.DSN)
	if err != nil {
		api.Respond(w, r, api.Error(fmt.Sprintf("failed to connect to database: %v", err)))
		return
	}
	defer db.Close()

	// Execute in transaction if requested
	if transactional {
		tx, err := db.Begin()
		if err != nil {
			api.Respond(w, r, api.Error(fmt.Sprintf("failed to begin transaction: %v", err)))
			return
		}

		// Defer rollback in case anything fails
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()

		// Execute the query in the transaction
		result, err := h.executeQuery(tx, sqlText, w, r)
		if err != nil {
			tx.Rollback()
			api.Respond(w, r, api.Error(err.Error()))
			return
		}

		// If we got a result (for SELECT queries), commit and return
		if result != nil {
			if err := tx.Commit(); err != nil {
				api.Respond(w, r, api.Error(fmt.Sprintf("failed to commit transaction: %v", err)))
				return
			}
			return
		}

		// For non-SELECT queries, commit and return the result
		if err := tx.Commit(); err != nil {
			api.Respond(w, r, api.Error(fmt.Sprintf("failed to commit transaction: %v", err)))
			return
		}

		api.Respond(w, r, api.SuccessWithData("result", map[string]any{
			"message": "Query executed successfully in transaction",
		}))
		return
	}

	// Execute without transaction
	_, err = h.executeQuery(db, sqlText, w, r)
	if err != nil {
		api.Respond(w, r, api.Error(err.Error()))
	}
}

// executeQuery executes the SQL query and handles the result
func (h *SQLExecute) executeQuery(db SQLExecutor, query string, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	// Check if it's a SELECT query
	if isReadOnlyQuery(query) {
		rows, err := db.Query(query)
		if err != nil {
			return nil, fmt.Errorf("query failed: %v", err)
		}
		defer rows.Close()

		// For SELECT queries, return the results
		if err := writeRowsResult(w, r, rows); err != nil {
			return nil, fmt.Errorf("failed to write query results: %v", err)
		}
		return struct{}{}, nil // Return a non-nil value to indicate we've handled the response
	}

	// For non-SELECT queries, execute and return the result
	result, err := db.Exec(query)
	if err != nil {
		return nil, fmt.Errorf("execution failed: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		// Some databases/drivers might not support RowsAffected
		rowsAffected = -1
	}

	api.Respond(w, r, api.SuccessWithData("result", map[string]any{
		"rows_affected": rowsAffected,
		"message":      "Query executed successfully",
	}))
	return nil, nil
}

// SQLExecutor is an interface that matches both *sql.DB and *sql.Tx
type SQLExecutor interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	Exec(query string, args ...interface{}) (sql.Result, error)
}

// writeRowsResult scans sql.Rows into a JSON-friendly structure with a sane row cap
func writeRowsResult(w http.ResponseWriter, r *http.Request, rows *sql.Rows) error {
	cols, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("failed to get columns: %v", err)
	}

	var results []map[string]any
	rowCount := 0
	maxRows := 1000 // Safety limit

	for rows.Next() && rowCount < maxRows {
		// Create slices to hold column values and pointers
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}

		// Scan the row into the column pointers
		if err := rows.Scan(columnPointers...); err != nil {
			return fmt.Errorf("failed to scan row: %v", err)
		}

		// Convert the row to a map with proper type handling
		row := make(map[string]any)
		for i, colName := range cols {
			val := columnPointers[i].(*interface{})
			switch v := (*val).(type) {
			case []byte:
				// Convert []byte to string for better JSON serialization
				row[colName] = string(v)
			case nil:
				row[colName] = nil
			default:
				row[colName] = v
			}
		}

		results = append(results, row)
		rowCount++
	}

	// Check for errors during iteration
	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating rows: %v", err)
	}

	// Check if there are more rows than the limit
	hasMore := false
	if rowCount >= maxRows && rows.Next() {
		hasMore = true
	}

	// Return the results
	api.Respond(w, r, api.SuccessWithData("rows", map[string]any{
		"rows":       results,
		"row_count":  rowCount,
		"has_more":   hasMore,
		"limit":      maxRows,
		"message":    "Query executed successfully",
	}))
	return nil
}
