package weebase

import (
	"net/http"
)

// handleListSchemas returns available schemas/namespaces for the current connection.
func (h *Handler) handleListSchemas(w http.ResponseWriter, r *http.Request) {
	s := EnsureSession(w, r, h.opts.SessionSecret)
	if s.Conn == nil || s.Conn.DB == nil {
		WriteError(w, r, "not connected")
		return
	}
	driver := normalizeDriver(s.Conn.Driver)
	type row struct{ Name string }
	var rows []row
	var err error
	switch driver {
	case "postgres":
		err = s.Conn.DB.Raw("SELECT schema_name AS name FROM information_schema.schemata ORDER BY name").Scan(&rows).Error
	case "mysql":
		err = s.Conn.DB.Raw("SELECT schema_name AS name FROM information_schema.schemata ORDER BY name").Scan(&rows).Error
	case "sqlite":
		rows = []row{{Name: "main"}, {Name: "temp"}}
	case "sqlserver":
		err = s.Conn.DB.Raw("SELECT name AS name FROM sys.schemas ORDER BY name").Scan(&rows).Error
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
