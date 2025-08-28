package api_schemas_list

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/dracory/api"
	"github.com/dracory/weebase/shared/session"
	"github.com/dracory/weebase/shared/types"
)

// SchemasList handles schema listing operations
type SchemasList struct {
	config types.Config
}

// New creates a new SchemasList handler
func New(config types.Config) *SchemasList {
	return &SchemasList{config: config}
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

// Handle processes the request
func (h *SchemasList) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		api.Respond(w, r, api.Error("list_schemas must be GET"))
		return
	}

	sess := session.EnsureSession(nil, nil, h.config.SessionSecret)
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

	// Get schemas based on database type
	var schemas []string
	switch normalizeDriver(sess.Conn.Driver) {
	case "mysql":
		rows, err := db.Query("SHOW DATABASES")
		if err != nil {
			api.Respond(w, r, api.Error(fmt.Sprintf("failed to list databases: %v", err)))
			return
		}
		defer rows.Close()

		for rows.Next() {
			var dbName string
			if err := rows.Scan(&dbName); err != nil {
				api.Respond(w, r, api.Error(fmt.Sprintf("failed to scan database name: %v", err)))
				return
			}
			schemas = append(schemas, dbName)
		}

		if err = rows.Err(); err != nil {
			api.Respond(w, r, api.Error(fmt.Sprintf("error iterating databases: %v", err)))
			return
		}

	case "postgres":
		rows, err := db.Query(`
			SELECT schema_name 
			FROM information_schema.schemata 
			WHERE schema_name NOT IN ('pg_catalog', 'information_schema', 'pg_toast')
			ORDER BY schema_name`)
		if err != nil {
			api.Respond(w, r, api.Error(fmt.Sprintf("failed to list schemas: %v", err)))
			return
		}
		defer rows.Close()

		for rows.Next() {
			var schema string
			if err := rows.Scan(&schema); err != nil {
				api.Respond(w, r, api.Error(fmt.Sprintf("failed to scan schema name: %v", err)))
				return
			}
			schemas = append(schemas, schema)
		}

		if err = rows.Err(); err != nil {
			api.Respond(w, r, api.Error(fmt.Sprintf("error iterating schemas: %v", err)))
			return
		}

	case "sqlite":
		// SQLite doesn't have schemas, so we'll just return an empty list
		schemas = []string{}

	case "sqlserver":
		rows, err := db.Query(`
			SELECT name 
			FROM sys.schemas 
			WHERE name NOT IN ('sys', 'INFORMATION_SCHEMA', 'guest', 'db_owner', 'db_accessadmin', 'db_securityadmin', 'db_ddladmin', 'db_backupoperator', 'db_datareader', 'db_datawriter', 'db_denydatareader', 'db_denydatawriter')
			ORDER BY name`)
		if err != nil {
			api.Respond(w, r, api.Error(fmt.Sprintf("failed to list schemas: %v", err)))
			return
		}
		defer rows.Close()

		for rows.Next() {
			var schema string
			if err := rows.Scan(&schema); err != nil {
				api.Respond(w, r, api.Error(fmt.Sprintf("failed to scan schema name: %v", err)))
				return
			}
			schemas = append(schemas, schema)
		}

		if err = rows.Err(); err != nil {
			api.Respond(w, r, api.Error(fmt.Sprintf("error iterating schemas: %v", err)))
			return
		}

	default:
		api.Respond(w, r, api.Error(fmt.Sprintf("database type %s is not supported for schema listing", sess.Conn.Driver)))
		return
	}

	// Return the schemas as JSON
	api.Respond(w, r, api.SuccessWithData("schemas", map[string]any{
		"schemas": schemas,
	}))
}
