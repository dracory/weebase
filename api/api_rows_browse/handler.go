package api_rows_browse

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/dracory/api"
	"github.com/dracory/weebase/shared/session"
	"github.com/dracory/weebase/shared/types"
)

// RowsBrowse handles row browsing operations
type RowsBrowse struct {
	config types.Config
}

// New creates a new RowsBrowse handler
func New(config types.Config) *RowsBrowse {
	return &RowsBrowse{config: config}
}

// sanitizeIdent checks if an identifier contains only safe characters
func sanitizeIdent(s string) bool {
	return s != "" && !strings.ContainsAny(s, " ;'\"`")
}

// normalizeDriver normalizes the driver name for consistent comparison
func normalizeDriver(driver string) string {
	switch driver {
	case "postgres", "postgresql":
		return "postgres"
	case "mysql":
		return "mysql"
	case "sqlite3", "sqlite":
		return "sqlite"
	case "mssql", "sqlserver":
		return "sqlserver"
	default:
		return driver
	}
}

// quoteIdent quotes an identifier based on the SQL dialect
func quoteIdent(driver, ident string) string {
	switch normalizeDriver(driver) {
	case "mysql":
		return "`" + strings.ReplaceAll(ident, "`", "``") + "`"
	case "postgres", "sqlite":
		parts := strings.Split(ident, ".")
		for i, part := range parts {
			parts[i] = "\"" + strings.ReplaceAll(part, "\"", "\"\"") + "\""
		}
		return strings.Join(parts, ".")
	case "sqlserver":
		parts := strings.Split(ident, ".")
		for i, part := range parts {
			parts[i] = "[" + strings.ReplaceAll(part, "]", "]]") + "]"
		}
		return strings.Join(parts, ".")
	default:
		return ident
	}
}

// Handle processes the request
func (h *RowsBrowse) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		api.Respond(w, r, api.Error("browse_rows must be GET"))
		return
	}

	sess := session.EnsureSession(nil, nil, h.config.SessionSecret)
	if sess == nil || sess.Conn == nil {
		api.Respond(w, r, api.Error("not connected to database"))
		return
	}

	// Get query parameters
	table := strings.TrimSpace(r.URL.Query().Get("table"))
	schema := strings.TrimSpace(r.URL.Query().Get("schema"))
	
	// Validate required parameters
	if table == "" {
		api.Respond(w, r, api.Error("table is required"))
		return
	}

	// Validate identifiers
	if !sanitizeIdent(table) || (schema != "" && !sanitizeIdent(schema)) {
		api.Respond(w, r, api.Error("invalid table or schema identifier"))
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 {
		limit = 50 // Default limit
	}
	offset := (page - 1) * limit

	// Open database connection
	db, err := sql.Open(sess.Conn.Driver, sess.Conn.DSN)
	if err != nil {
		api.Respond(w, r, api.Error(fmt.Sprintf("failed to connect to database: %v", err)))
		return
	}
	defer db.Close()

	// Build table name with schema if provided
	tableName := table
	if schema != "" {
		tableName = schema + "." + table
	}
	tableName = quoteIdent(sess.Conn.Driver, tableName)

	// Get total count
	var count int
	countQuery := "SELECT COUNT(*) FROM " + tableName
	err = db.QueryRow(countQuery).Scan(&count)
	if err != nil {
		api.Respond(w, r, api.Error(fmt.Sprintf("failed to get row count: %v", err)))
		return
	}

	// Build and execute the query
	query := "SELECT * FROM " + tableName
	
	// Add pagination
	switch normalizeDriver(sess.Conn.Driver) {
	case "mysql", "sqlite":
		query += fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)
	case "postgres":
		query += fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)
	case "sqlserver":
		// SQL Server uses different syntax for pagination
		query = fmt.Sprintf("SELECT * FROM (SELECT ROW_NUMBER() OVER (ORDER BY (SELECT NULL)) as row_num, * FROM %s) AS t WHERE row_num BETWEEN %d AND %d",
			tableName, offset+1, offset+limit)
	}

	// Execute query
	rows, err := db.Query(query)
	if err != nil {
		api.Respond(w, r, api.Error(fmt.Sprintf("query failed: %v", err)))
		return
	}
	defer rows.Close()

	// Get column names
	cols, err := rows.Columns()
	if err != nil {
		api.Respond(w, r, api.Error(fmt.Sprintf("failed to get columns: %v", err)))
		return
	}

	// Process rows
	var results []map[string]interface{}
	for rows.Next() {
		// Create slice to hold column values
		values := make([]interface{}, len(cols))
		for i := range values {
			var v interface{}
			values[i] = &v
		}

		// Scan row into values
		if err := rows.Scan(values...); err != nil {
			api.Respond(w, r, api.Error(fmt.Sprintf("failed to scan row: %v", err)))
			return
		}

		// Convert values to proper types
		row := make(map[string]interface{})
		for i, col := range cols {
			val := *(values[i].(*interface{}))
			// Convert []byte to string for better JSON serialization
			if b, ok := val.([]byte); ok {
				val = string(b)
			}
			row[col] = val
		}

		results = append(results, row)
	}

	if err = rows.Err(); err != nil {
		api.Respond(w, r, api.Error(fmt.Sprintf("error iterating rows: %v", err)))
		return
	}

	api.Respond(w, r, api.SuccessWithData("rows", map[string]any{
		"rows":  results,
		"total": count,
		"page":  page,
		"limit": limit,
	}))
}
