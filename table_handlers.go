package weebase

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/dracory/api"
)

// handleTablesList handles the request to list all tables in a database
func (w *Weebase) handleTablesList(rw http.ResponseWriter, r *http.Request) {
	if w.db == nil {
		api.Respond(rw, r, api.Error("Not connected to any database"))
		return
	}

	tables, err := w.listTables()
	if err != nil {
		api.Respond(rw, r, api.Error(err.Error()))
		return
	}

	api.Respond(rw, r, api.SuccessWithData("Tables list", map[string]interface{}{
		"tables": tables,
	}))
}

// listTables returns a list of all tables in the current database
func (w *Weebase) listTables() ([]TableInfo, error) {
	var tables []TableInfo
	var rows *sql.Rows
	var err error

	switch w.db.Dialector.Name() {
	case "mysql":
		rows, err = w.db.Raw("SHOW TABLES").Rows()
	case "postgres":
		rows, err = w.db.Raw(
			`SELECT table_name 
			FROM information_schema.tables 
			WHERE table_schema = 'public'`).Rows()
	case "sqlite", "sqlite3":
		rows, err = w.db.Raw(
			`SELECT name 
			FROM sqlite_master 
			WHERE type='table' AND name NOT LIKE 'sqlite_%'`).Rows()
	case "sqlserver":
		rows, err = w.db.Raw(
			`SELECT table_name 
			FROM information_schema.tables 
			WHERE table_type = 'BASE TABLE'`).Rows()
	default:
		return nil, fmt.Errorf("unsupported database driver")
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		tables = append(tables, TableInfo{Name: name})
	}

	return tables, nil
}
