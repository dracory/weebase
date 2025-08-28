package api_tables_list

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

// TablesList handles table listing operations
// It provides functionality to list all tables in the connected database
// with support for multiple database backends including PostgreSQL, MySQL, SQLite, and SQL Server
// TablesList provides methods to list database tables
// It supports multiple database backends and handles connection management
// through the session system
type TablesList struct {
	config types.Config
}

// New creates a new TablesList handler
// It initializes the handler with the provided configuration
func New(cfg types.Config) *TablesList {
	return &TablesList{config: cfg}
}

// Handle processes the request to list database tables
// It supports multiple database backends including PostgreSQL, MySQL, SQLite, and SQL Server
func (h *TablesList) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		api.Respond(w, r, api.Error("method not allowed"))
		return
	}

	sess := session.EnsureSession(w, r, h.config.SessionSecret)
	if sess == nil || sess.Conn == nil {
		api.Respond(w, r, api.Error("not connected to database"))
		return
	}

	// Open database connection
	db, err := sql.Open(sess.Conn.Driver, sess.Conn.DSN)
	if err != nil {
		api.Respond(w, r, api.Error(fmt.Sprintf("failed to connect to database: %v", err)))
		return
	}
	defer db.Close()

	// Get tables based on database type
	driver := normalizeDriver(sess.Conn.Driver)
	var tables []string

	ctx := context.Background()
	switch driver {
	case "postgres":
		tables, err = listPostgresTables(ctx, db)
	case "mysql":
		tables, err = listMySQLTables(ctx, db)
	case "sqlite":
		tables, err = listSQLiteTables(ctx, db)
	case "sqlserver":
		tables, err = listSQLServerTables(ctx, db)
	default:
		api.Respond(w, r, api.Error("unsupported database driver"))
		return
	}

	if err != nil {
		api.Respond(w, r, api.Error(fmt.Sprintf("error listing tables: %v", err)))
		return
	}

	api.Respond(w, r, api.SuccessWithData("tables_listed", map[string]any{
		"tables":    tables,
		"count":     len(tables),
		"driver":    driver,
		"message":   "Tables listed successfully",
	}))
}

// listPostgresTables lists all tables in a PostgreSQL database
func listPostgresTables(ctx context.Context, db *sql.DB) ([]string, error) {
	query := `
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		ORDER BY table_name`

	return queryStringList(ctx, db, query)
}

// listMySQLTables lists all tables in a MySQL database
func listMySQLTables(ctx context.Context, db *sql.DB) ([]string, error) {
	query := `SHOW TABLES`
	return queryStringList(ctx, db, query)
}

// listSQLiteTables lists all tables in a SQLite database
func listSQLiteTables(ctx context.Context, db *sql.DB) ([]string, error) {
	query := `
		SELECT name 
		FROM sqlite_master 
		WHERE type='table' 
		AND name NOT LIKE 'sqlite_%' 
		AND name != 'migrations' 
		ORDER BY name`

	return queryStringList(ctx, db, query)
}

// listSQLServerTables lists all tables in a SQL Server database
func listSQLServerTables(ctx context.Context, db *sql.DB) ([]string, error) {
	query := `
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_type = 'BASE TABLE' 
		AND table_catalog = DB_NAME()
		ORDER BY table_name`

	return queryStringList(ctx, db, query)
}

// queryStringList executes a query that returns a single column of strings
func queryStringList(ctx context.Context, db *sql.DB, query string, args ...interface{}) ([]string, error) {
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query failed: %v", err)
	}
	defer rows.Close()

	var result []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("scan failed: %v", err)
		}
		result = append(result, name)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %v", err)
	}

	return result, nil
}

// normalizeDriver normalizes the database driver name
func normalizeDriver(driver string) string {
	switch strings.ToLower(driver) {
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
