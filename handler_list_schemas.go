package weebase

import (
	"net/http"

	"github.com/dracory/weebase/shared/session"
	"gorm.io/gorm"
)

// handleListSchemas returns available schemas/namespaces for the current connection.
func (h *Handler) handleListSchemas(w http.ResponseWriter, r *http.Request) {
	s := session.EnsureSession(w, r, h.opts.SessionSecret)
	if s == nil || s.Conn == nil {
		WriteError(w, r, "not connected")
		return
	}

	db, ok := s.Conn.DB.(*gorm.DB)
	if !ok {
		WriteError(w, r, "invalid database connection")
		return
	}

	driver := normalizeDriver(s.Conn.Driver)
	type row struct{ Name string }
	var rows []row
	var err error

	switch driver {
	case "postgres":
		err = db.Raw("SELECT schema_name AS name FROM information_schema.schemata ORDER BY name").Scan(&rows).Error
	case "mysql":
		err = db.Raw("SELECT schema_name AS name FROM information_schema.schemata ORDER BY name").Scan(&rows).Error
	case "sqlite":
		rows = []row{{Name: "main"}, {Name: "temp"}}
	case "sqlserver":
		err = db.Raw("SELECT name AS name FROM sys.schemas ORDER BY name").Scan(&rows).Error
	default:
		WriteError(w, r, "unsupported driver")
		return
	}
	if err != nil {
		WriteError(w, r, err.Error())
		return
	}
	names := make([]string, 0, len(rows))
	for _, x := range rows {
		names = append(names, x.Name)
	}
	WriteSuccessWithData(w, r, "ok", map[string]any{"schemas": names})
}
