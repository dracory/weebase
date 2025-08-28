package api_sql_explain

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/dracory/api"
	"github.com/dracory/weebase/shared/session"
	"github.com/dracory/weebase/shared/types"
)

// SQLExplain handles SQL EXPLAIN operations
// SQLExplain handles SQL EXPLAIN operations
// It provides execution plans for SQL queries to help with query optimization
type SQLExplain struct {
	config types.Config
}

// New creates a new SQLExplain handler
func New(config types.Config) *SQLExplain {
	return &SQLExplain{config: config}
}

// getExplainQuery returns the appropriate EXPLAIN query for the given SQL and driver
func getExplainQuery(sqlText, driver string) string {
	switch driver {
	case "postgres":
		return "EXPLAIN (FORMAT JSON) " + sqlText
	case "mysql":
		return "EXPLAIN FORMAT=JSON " + sqlText
	case "sqlite":
		return "EXPLAIN QUERY PLAN " + sqlText
	case "sqlserver":
		return "SET SHOWPLAN_XML ON;\n" + sqlText + "\nSET SHOWPLAN_XML OFF;"
	default:
		return "EXPLAIN " + sqlText
	}
}

// scanRow scans a single row into a map of column names to values
func scanRow(rows *sql.Rows) (map[string]any, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("error getting columns: %v", err)
	}

	// Create a slice of interface{} to hold the column values
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	// Scan the row into the value pointers
	if err := rows.Scan(valuePtrs...); err != nil {
		return nil, fmt.Errorf("error scanning row: %v", err)
	}

	// Convert row to map
	row := make(map[string]any, len(columns))
	for i, col := range columns {
		val := values[i]
		switch v := val.(type) {
		case []byte:
			row[col] = string(v)
		case nil:
			row[col] = nil
		default:
			row[col] = v
		}
	}

	return row, nil
}

// Handle processes the request
func (h *SQLExplain) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		api.Respond(w, r, api.Error("sql_explain must be POST"))
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

	// Open database connection
	db, err := sql.Open(sess.Conn.Driver, sess.Conn.DSN)
	if err != nil {
		api.Respond(w, r, api.Error(fmt.Sprintf("failed to connect to database: %v", err)))
		return
	}
	defer db.Close()

	// Get the database driver name
	driver := normalizeDriver(sess.Conn.Driver)

	// Prepare the EXPLAIN query based on the database driver
	explainSQL := getExplainQuery(sqlText, driver)

	// Execute the EXPLAIN query
	rows, err := db.Query(explainSQL)
	if err != nil {
		api.Respond(w, r, api.Error(fmt.Sprintf("error executing explain: %v", err)))
		return
	}
	defer rows.Close()

	// Process results
	var results []map[string]any
	for rows.Next() {
		row, err := scanRow(rows)
		if err != nil {
			api.Respond(w, r, api.Error(err.Error()))
			return
		}
		results = append(results, row)
	}

	if err = rows.Err(); err != nil {
		api.Respond(w, r, api.Error(fmt.Sprintf("error iterating rows: %v", err)))
		return
	}

	api.Respond(w, r, api.SuccessWithData("explain", map[string]any{
		"plan":    results,
		"driver":  driver,
		"message": "Query plan generated successfully",
	}))
}

// normalizeDriver normalizes the database driver name
func normalizeDriver(driver string) string {
	switch driver {
	case "postgres", "postgresql":
		return "postgres"
	case "mysql", "mariadb":
		return "mysql"
	case "sqlserver", "mssql":
		return "sqlserver"
	case "sqlite", "sqlite3":
		return "sqlite"
	default:
		return driver
	}
}
