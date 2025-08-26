package api_sql_explain

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/dracory/api"
	"github.com/dracory/weebase/shared/session"
	"gorm.io/gorm"
)

// SQLExplain handles SQL EXPLAIN operations
type SQLExplain struct {
	conn *session.ActiveConnection
}

// New creates a new SQLExplain handler
func New(conn *session.ActiveConnection) *SQLExplain {
	return &SQLExplain{conn: conn}
}

// Handle processes the request
func (h *SQLExplain) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		api.Respond(w, r, api.Error("sql_explain must be POST"))
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

	// Get the gorm DB instance
	db, ok := h.conn.DB.(*gorm.DB)
	if !ok {
		api.Respond(w, r, api.Error("invalid database connection type"))
		return
	}

	// Generate the appropriate EXPLAIN statement based on the database driver
	var explainSQL string
	switch normalizeDriver(h.conn.Driver) {
	case "postgres", "mysql", "sqlserver":
		explainSQL = "EXPLAIN " + sqlText
	case "sqlite":
		explainSQL = "EXPLAIN QUERY PLAN " + sqlText
	default:
		explainSQL = "EXPLAIN " + sqlText
	}

	// Execute the EXPLAIN query
	rows, err := db.Raw(explainSQL).Rows()
	if err != nil {
		api.Respond(w, r, api.Error(fmt.Sprintf("explain error: %v", err)))
		return
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		api.Respond(w, r, api.Error(fmt.Sprintf("error getting columns: %v", err)))
		return
	}

	// Read all rows
	var results []map[string]any
	for rows.Next() {
		// Create a slice of interface{} to hold the column values
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		// Scan the row into the values
		if err := rows.Scan(valuePtrs...); err != nil {
			api.Respond(w, r, api.Error(fmt.Sprintf("error scanning row: %v", err)))
			return
		}

		// Convert row to map
		row := make(map[string]any, len(columns))
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}
		results = append(results, row)
	}

	if err = rows.Err(); err != nil {
		api.Respond(w, r, api.Error(fmt.Sprintf("error iterating rows: %v", err)))
		return
	}

	api.Respond(w, r, api.SuccessWithData("explain", map[string]any{
		"plan": results,
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
