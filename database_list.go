package weebase

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/dracory/api"
)

// handleDatabasesList lists all databases on the server
func (w *Weebase) handleDatabasesList(rw http.ResponseWriter, r *http.Request) {
	if w.db == nil {
		api.Respond(rw, r, api.Error("Not connected to any database"))
		return
	}

	databases, err := w.listDatabases()
	if err != nil {
		api.Respond(rw, r, api.Error(err.Error()))
		return
	}

	api.Respond(rw, r, api.SuccessWithData("Databases list", map[string]interface{}{
		"databases": databases,
	}))
}

// listDatabases returns a list of all databases
func (w *Weebase) listDatabases() ([]DatabaseInfo, error) {
	var databases []DatabaseInfo
	var rows *sql.Rows
	var err error

	switch w.db.Dialector.Name() {
	case "mysql":
		rows, err = w.db.Raw("SHOW DATABASES").Rows()
	case "postgres":
		rows, err = w.db.Raw("SELECT datname FROM pg_database WHERE datistemplate = false").Rows()
	case "sqlite", "sqlite3":
		return []DatabaseInfo{{Name: "main"}}, nil
	case "sqlserver":
		rows, err = w.db.Raw("SELECT name FROM sys.databases WHERE name NOT IN ('master', 'tempdb', 'model', 'msdb')").Rows()
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
		databases = append(databases, DatabaseInfo{Name: name})
	}

	return databases, nil
}
