package api_table_info

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/dracory/api"
	"github.com/dracory/weebase/shared/session"
	"gorm.io/gorm"
)

// TableInfo handles table information requests
type TableInfo struct {
	conn *session.ActiveConnection
}

// New creates a new TableInfo handler
func New(conn *session.ActiveConnection) *TableInfo {
	return &TableInfo{conn: conn}
}

// Column represents a column in a database table
type Column struct {
	Name          string `json:"name"`
	DataType      string `json:"data_type"`
	IsNullable    string `json:"is_nullable"`
	ColumnDefault any    `json:"column_default"`
}

// Handle processes the request
func (h *TableInfo) Handle(w http.ResponseWriter, r *http.Request) {
	if h.conn == nil || h.conn.DB == nil {
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
		api.Respond(w, r, api.Error("table is required"))
		return
	}

	if !sanitizeIdent(table) || (schema != "" && !sanitizeIdent(schema)) {
		api.Respond(w, r, api.Error("invalid identifier"))
		return
	}

	// Get the gorm DB instance
	db, ok := h.conn.DB.(*gorm.DB)
	if !ok {
		api.Respond(w, r, api.Error("invalid database connection type"))
		return
	}

	var columns []Column
	var err error

	switch normalizeDriver(h.conn.Driver) {
	case "postgres":
		err = h.handlePostgres(db, schema, table, &columns)
	case "mysql":
		err = h.handleMySQL(db, schema, table, &columns)
	case "sqlite":
		err = h.handleSQLite(db, table, &columns)
	case "sqlserver":
		err = h.handleSQLServer(db, schema, table, &columns)
	default:
		api.Respond(w, r, api.Error("unsupported driver"))
		return
	}

	if err != nil {
		api.Respond(w, r, api.Error(err.Error()))
		return
	}

	api.Respond(w, r, api.SuccessWithData("columns", map[string]any{
		"columns": columns,
	}))
}

// handlePostgres handles PostgreSQL table info
func (h *TableInfo) handlePostgres(db *gorm.DB, schema, table string, columns *[]Column) error {
	if schema == "" {
		schema = "public"
	}
	return db.Raw(
		`SELECT 
			column_name AS name, 
			data_type AS data_type, 
			is_nullable AS is_nullable, 
			column_default AS column_default 
		FROM information_schema.columns 
		WHERE table_schema = ? AND table_name = ? 
		ORDER BY ordinal_position`,
		schema, table,
	).Scan(columns).Error
}

// handleMySQL handles MySQL table info
func (h *TableInfo) handleMySQL(db *gorm.DB, schema, table string, columns *[]Column) error {
	if schema == "" {
		return fmt.Errorf("schema is required for MySQL")
	}
	return db.Raw(
		`SELECT 
			column_name AS name, 
			data_type AS data_type, 
			is_nullable AS is_nullable, 
			column_default AS column_default 
		FROM information_schema.columns 
		WHERE table_schema = ? AND table_name = ? 
		ORDER BY ordinal_position`,
		schema, table,
	).Scan(columns).Error
}

// handleSQLite handles SQLite table info
func (h *TableInfo) handleSQLite(db *gorm.DB, table string, columns *[]Column) error {
	// SQLite PRAGMA cannot bind identifiers; we restrict allowed characters and interpolate safely
	q := "PRAGMA table_info(" + table + ")"
	
	type sqliteColumn struct {
		Name     string `gorm:"column:name"`
		Type     string `gorm:"column:type"`
		NotNull  int    `gorm:"column:notnull"`
		Default  any    `gorm:"column:dflt_value"`
	}

	var sqliteColumns []sqliteColumn
	if err := db.Raw(q).Scan(&sqliteColumns).Error; err != nil {
		return err
	}

	*columns = make([]Column, 0, len(sqliteColumns))
	for _, c := range sqliteColumns {
		isNull := "YES"
		if c.NotNull == 1 {
			isNull = "NO"
		}
		*columns = append(*columns, Column{
			Name:          c.Name,
			DataType:      c.Type,
			IsNullable:    isNull,
			ColumnDefault: c.Default,
		})
	}
	return nil
}

// handleSQLServer handles SQL Server table info
func (h *TableInfo) handleSQLServer(db *gorm.DB, schema, table string, columns *[]Column) error {
	if schema == "" {
		schema = "dbo"
	}
	return db.Raw(
		`SELECT 
			c.name AS name, 
			t.name AS data_type, 
			CASE WHEN c.is_nullable=1 THEN 'YES' ELSE 'NO' END AS is_nullable, 
			c.default_object_id AS column_default 
		FROM sys.columns c 
		JOIN sys.types t ON c.user_type_id=t.user_type_id 
		JOIN sys.tables tb ON c.object_id=tb.object_id 
		JOIN sys.schemas s ON tb.schema_id=s.schema_id 
		WHERE s.name = ? AND tb.name = ? 
		ORDER BY c.column_id`,
		schema, table,
	).Scan(columns).Error
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
