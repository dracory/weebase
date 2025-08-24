package weebase

import (
	"net/http"
	"strings"
)

// handleTableInfo returns column metadata for a given table.
func (h *Handler) handleTableInfo(w http.ResponseWriter, r *http.Request) {
	s := EnsureSession(w, r, h.opts.SessionSecret)
	if s.Conn == nil || s.Conn.DB == nil {
		WriteError(w, r, "not connected")
		return
	}
	_ = r.ParseForm()
	schema := strings.TrimSpace(r.Form.Get("schema"))
	table := strings.TrimSpace(r.Form.Get("table"))
	if table == "" {
		WriteError(w, r, "table is required")
		return
	}
	if !sanitizeIdent(table) || (schema != "" && !sanitizeIdent(schema)) {
		WriteError(w, r, "invalid identifier")
		return
	}
	driver := normalizeDriver(s.Conn.Driver)
	type col struct {
		Name          string `json:"name"`
		DataType      string `json:"data_type"`
		IsNullable    string `json:"is_nullable"`
		ColumnDefault any    `json:"column_default"`
	}
	var cols []col
	var err error
	switch driver {
	case "postgres":
		if schema == "" {
			schema = "public"
		}
		err = s.Conn.DB.Raw(
			"SELECT column_name AS name, data_type AS data_type, is_nullable AS is_nullable, column_default AS column_default FROM information_schema.columns WHERE table_schema = ? AND table_name = ? ORDER BY ordinal_position",
			schema, table,
		).Scan(&cols).Error
	case "mysql":
		if schema == "" {
			WriteError(w, r, "schema required for mysql")
			return
		}
		err = s.Conn.DB.Raw(
			"SELECT column_name AS name, data_type AS data_type, is_nullable AS is_nullable, column_default AS column_default FROM information_schema.columns WHERE table_schema = ? AND table_name = ? ORDER BY ordinal_position",
			schema, table,
		).Scan(&cols).Error
	case "sqlite":
		// SQLite PRAGMA cannot bind identifiers; we restrict allowed characters and interpolate safely.
		q := "PRAGMA table_info(" + table + ")"
		type row struct {
			name, ctype string
			notnull     int
			dflt_value  any
		}
		var rs []row
		err = s.Conn.DB.Raw(q).Scan(&rs).Error
		if err == nil {
			cols = make([]col, 0, len(rs))
			for _, r := range rs {
				isNull := "YES"
				if r.notnull == 1 {
					isNull = "NO"
				}
				cols = append(cols, col{Name: r.name, DataType: r.ctype, IsNullable: isNull, ColumnDefault: r.dflt_value})
			}
		}
	case "sqlserver":
		if schema == "" {
			schema = "dbo"
		}
		err = s.Conn.DB.Raw(
			"SELECT c.name AS name, t.name AS data_type, CASE WHEN c.is_nullable=1 THEN 'YES' ELSE 'NO' END AS is_nullable, c.default_object_id AS column_default FROM sys.columns c JOIN sys.types t ON c.user_type_id=t.user_type_id JOIN sys.tables tb ON c.object_id=tb.object_id JOIN sys.schemas s ON tb.schema_id=s.schema_id WHERE s.name = ? AND tb.name = ? ORDER BY c.column_id",
			schema, table,
		).Scan(&cols).Error
	default:
		WriteError(w, r, "unsupported driver")
		return
	}
	if err != nil {
		WriteError(w, r, err.Error())
		return
	}
	WriteSuccessWithData(w, r, "ok", map[string]any{"columns": cols})
}
