package api_table_info

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/dracory/api"
	"github.com/dracory/weebase/shared/session"
	"github.com/dracory/weebase/shared/types"
)

// TableInfo handles table information requests
// TableInfo provides information about database table structure
// It supports multiple database backends including PostgreSQL, MySQL, SQLite, and SQL Server
type TableInfo struct {
	config types.Config
}

// New creates a new TableInfo handler
func New(config types.Config) *TableInfo {
	return &TableInfo{config: config}
}

// Column represents a column in a database table
type Column struct {
	Name          string `json:"name"`
	DataType      string `json:"data_type"`
	IsNullable    string `json:"is_nullable"`
	ColumnDefault any    `json:"column_default"`
}

// Handle processes the request for table information
func (h *TableInfo) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		api.Respond(w, r, api.Error("method not allowed"))
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

	schema := strings.TrimSpace(r.Form.Get("schema"))
	table := strings.TrimSpace(r.Form.Get("table"))

	if table == "" {
		api.Respond(w, r, api.Error("table name is required"))
		return
	}

	if !sanitizeIdent(table) || (schema != "" && !sanitizeIdent(schema)) {
		api.Respond(w, r, api.Error("invalid table or schema name"))
		return
	}

	// Open database connection
	db, err := sql.Open(sess.Conn.Driver, sess.Conn.DSN)
	if err != nil {
		api.Respond(w, r, api.Error(fmt.Sprintf("failed to connect to database: %v", err)))
		return
	}
	defer db.Close()

	// Get table information based on database type
	driver := normalizeDriver(sess.Conn.Driver)
	var columns []Column

	switch driver {
	case "postgres":
		columns, err = h.handlePostgres(db, schema, table)
	case "mysql":
		columns, err = h.handleMySQL(db, schema, table)
	case "sqlite":
		columns, err = h.handleSQLite(db, table)
	case "sqlserver":
		columns, err = h.handleSQLServer(db, schema, table)
	default:
		api.Respond(w, r, api.Error("unsupported database driver"))
		return
	}

	if err != nil {
		api.Respond(w, r, api.Error(fmt.Sprintf("error getting table info: %v", err)))
		return
	}

	api.Respond(w, r, api.SuccessWithData("columns", map[string]any{
		"columns":    columns,
		"table":      table,
		"schema":     schema,
		"driver":     driver,
		"row_count":  len(columns),
	}))
}

// handlePostgres handles PostgreSQL table info
func (h *TableInfo) handlePostgres(db *sql.DB, schema, table string) ([]Column, error) {
	if schema == "" {
		schema = "public"
	}

	query := `
		SELECT 
			column_name, 
			data_type, 
			is_nullable, 
			column_default
		FROM information_schema.columns 
		WHERE table_schema = $1 AND table_name = $2 
		ORDER BY ordinal_position`

	rows, err := db.Query(query, schema, table)
	if err != nil {
		return nil, fmt.Errorf("query failed: %v", err)
	}
	defer rows.Close()

	var columns []Column
	for rows.Next() {
		var col Column
		if err := rows.Scan(
			&col.Name,
			&col.DataType,
			&col.IsNullable,
			&col.ColumnDefault,
		); err != nil {
			return nil, fmt.Errorf("scan failed: %v", err)
		}
		columns = append(columns, col)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %v", err)
	}

	return columns, nil
}

// handleMySQL handles MySQL table info
func (h *TableInfo) handleMySQL(db *sql.DB, schema, table string) ([]Column, error) {
	if schema == "" {
		return nil, fmt.Errorf("schema is required for MySQL")
	}

	query := `
		SELECT 
			column_name, 
			data_type, 
			is_nullable, 
			column_default
		FROM information_schema.columns 
		WHERE table_schema = ? AND table_name = ? 
		ORDER BY ordinal_position`

	rows, err := db.Query(query, schema, table)
	if err != nil {
		return nil, fmt.Errorf("query failed: %v", err)
	}
	defer rows.Close()

	var columns []Column
	for rows.Next() {
		var col Column
		if err := rows.Scan(
			&col.Name,
			&col.DataType,
			&col.IsNullable,
			&col.ColumnDefault,
		); err != nil {
			return nil, fmt.Errorf("scan failed: %v", err)
		}
		columns = append(columns, col)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %v", err)
	}

	return columns, nil
}

// handleSQLite handles SQLite table info
func (h *TableInfo) handleSQLite(db *sql.DB, table string) ([]Column, error) {
	// SQLite PRAGMA cannot use parameter binding for table names
	// We've already validated the table name with sanitizeIdent
	q := fmt.Sprintf("PRAGMA table_info(%s)", table)
	
	rows, err := db.Query(q)
	if err != nil {
		return nil, fmt.Errorf("query failed: %v", err)
	}
	defer rows.Close()

	var columns []Column
	for rows.Next() {
		var (
			cid        int
			name       string
			colType    string
			notNull   int
			defaultVal sql.NullString
			pk         int
		)

		if err := rows.Scan(&cid, &name, &colType, &notNull, &defaultVal, &pk); err != nil {
			return nil, fmt.Errorf("scan failed: %v", err)
		}

		isNullable := "YES"
		if notNull == 1 {
			isNullable = "NO"
		}

		columns = append(columns, Column{
			Name:          name,
			DataType:      colType,
			IsNullable:    isNullable,
			ColumnDefault: defaultVal.String,
		})
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %v", err)
	}

	return columns, nil
}

// handleSQLServer handles SQL Server table info
func (h *TableInfo) handleSQLServer(db *sql.DB, schema, table string) ([]Column, error) {
	if schema == "" {
		schema = "dbo"
	}

	query := `
		SELECT 
			c.name, 
			t.name, 
			CASE WHEN c.is_nullable=1 THEN 'YES' ELSE 'NO' END,
			OBJECT_DEFINITION(c.default_object_id)
		FROM sys.columns c 
		JOIN sys.types t ON c.user_type_id=t.user_type_id 
		JOIN sys.tables tb ON c.object_id=tb.object_id 
		JOIN sys.schemas s ON tb.schema_id=s.schema_id 
		WHERE s.name = @p1 AND tb.name = @p2 
		ORDER BY c.column_id`

	rows, err := db.QueryContext(context.Background(), query, sql.Named("p1", schema), sql.Named("p2", table))
	if err != nil {
		return nil, fmt.Errorf("query failed: %v", err)
	}
	defer rows.Close()

	var columns []Column
	for rows.Next() {
		var col Column
		if err := rows.Scan(
			&col.Name,
			&col.DataType,
			&col.IsNullable,
			&col.ColumnDefault,
		); err != nil {
			return nil, fmt.Errorf("scan failed: %v", err)
		}
		columns = append(columns, col)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %v", err)
	}

	return columns, nil
}

// sanitizeIdent checks if an identifier contains only safe characters
func sanitizeIdent(s string) bool {
	// Simple check for now - should be extended with proper validation
	return s != "" && !strings.ContainsAny(s, " ;'\"`")
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
