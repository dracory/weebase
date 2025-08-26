package api_table_create

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/dracory/api"
	"github.com/dracory/weebase/shared/session"
	"gorm.io/gorm"
)

type TableCreate struct {
	conn *session.ActiveConnection
}

func New(conn *session.ActiveConnection) *TableCreate {
	return &TableCreate{conn: conn}
}

// Handle validates, builds SQL, and executes using only injected deps.
func (tc *TableCreate) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		api.Respond(w, r, api.Error("method not allowed"))
		return
	}

	if tc.conn == nil {
		api.Respond(w, r, api.Error("not connected"))
		return
	}

	_ = r.ParseForm()
	schema := strings.TrimSpace(r.Form.Get("schema"))
	table := strings.TrimSpace(r.Form.Get("table"))
	if table == "" {
		api.Respond(w, r, api.Error("table required"))
		return
	}
	names := r.Form["col_name[]"]
	types := r.Form["col_type[]"]
	lens := r.Form["col_length[]"]
	nullable := indexSet(r.Form["col_nullable[]"])
	pkset := indexSet(r.Form["col_pk[]"])
	aiset := indexSet(r.Form["col_ai[]"])

	d := normalizeDriver(tc.conn.Driver)
	stmt, errMsg := buildSQL(d, schema, table, names, types, lens, nullable, pkset, aiset)
	if errMsg != "" {
		api.Respond(w, r, api.Error(errMsg))
		return
	}

	// Try to get *sql.DB from gorm.DB first
	if gormDB, ok := tc.conn.DB.(*gorm.DB); ok {
		sqlDB, err := gormDB.DB()
		if err != nil {
			api.Respond(w, r, api.Error("failed to get database connection: "+err.Error()))
			return
		}
		if _, err := sqlDB.ExecContext(r.Context(), stmt); err != nil {
			api.Respond(w, r, api.Error(err.Error()))
			return
		}
	} else if sqlDB, ok := tc.conn.DB.(*sql.DB); ok {
		// Handle direct *sql.DB
		if _, err := sqlDB.ExecContext(r.Context(), stmt); err != nil {
			api.Respond(w, r, api.Error(err.Error()))
			return
		}
	} else if execCtx, ok := tc.conn.DB.(interface {
		ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	}); ok {
		// Handle any type that implements ExecContext
		if _, err := execCtx.ExecContext(r.Context(), stmt); err != nil {
			api.Respond(w, r, api.Error(err.Error()))
			return
		}
	} else {
		api.Respond(w, r, api.Error("database connection does not support execution"))
		return
	}
	api.Respond(w, r, api.SuccessWithData("created", map[string]any{"sql": stmt}))
}

func buildSQL(driver, schema, table string, names, types, lens []string, nullable, pkset, aiset map[string]bool) (string, string) {
	var defs []string
	var pks []string
	for i := range names {
		name := strings.TrimSpace(names[i])
		if name == "" {
			continue
		}
		typ := strings.TrimSpace(get(types, i))
		ln := strings.TrimSpace(get(lens, i))
		def, pk := buildColumnDef(driver, name, typ, ln, nullable[fmt.Sprint(i+1)], pkset[fmt.Sprint(i+1)], aiset[fmt.Sprint(i+1)])
		defs = append(defs, def)
		if pk != "" {
			pks = append(pks, pk)
		}
	}
	if len(defs) == 0 {
		return "", "at least one column required"
	}
	if len(pks) > 0 && driver != "sqlite" {
		defs = append(defs, fmt.Sprintf("PRIMARY KEY (%s)", strings.Join(pks, ", ")))
	}
	id := table
	if strings.TrimSpace(schema) != "" {
		id = schema + "." + table
	}
	stmt := fmt.Sprintf("CREATE TABLE %s (\n  %s\n)", quoteIdent(driver, id), strings.Join(defs, ",\n  "))
	return stmt, ""
}

// Helpers (implement locally to avoid importing root)
func indexSet(vals []string) map[string]bool {
	m := map[string]bool{}
	for _, v := range vals {
		m[v] = true
	}
	return m
}

func get(ss []string, i int) string {
	if i < len(ss) {
		return ss[i]
	}
	return ""
}

func normalizeDriver(d string) string {
	switch strings.ToLower(d) {
	case "pg", "postgresql":
		return "postgres"
	case "mariadb":
		return "mysql"
	case "sqlite3":
		return "sqlite"
	case "mssql":
		return "sqlserver"
	default:
		return strings.ToLower(d)
	}
}

func quoteIdent(driver, ident string) string {
	d := normalizeDriver(driver)
	parts := strings.Split(ident, ".")
	for i, p := range parts {
		switch d {
		case "postgres":
			p = strings.ReplaceAll(p, "\"", "\"\"")
			parts[i] = "\"" + p + "\""
		case "mysql", "sqlite":
			p = strings.ReplaceAll(p, "`", "``")
			parts[i] = "`" + p + "`"
		case "sqlserver":
			p = strings.ReplaceAll(p, "]", "]]")
			parts[i] = "[" + p + "]"
		default:
			parts[i] = p
		}
	}
	return strings.Join(parts, ".")
}

func buildColumnDef(driver, name, typ, length string, nullable, pk, ai bool) (def string, pkOut string) {
	d := normalizeDriver(driver)
	qt := quoteIdent(driver, name)
	t := typ
	if length != "" && !strings.Contains(strings.ToLower(typ), "(") {
		t = fmt.Sprintf("%s(%s)", typ, length)
	}
	switch d {
	case "postgres":
		if ai {
			if strings.Contains(strings.ToLower(typ), "big") {
				t = "bigserial"
			} else {
				t = "serial"
			}
		}
		def = fmt.Sprintf("%s %s", qt, t)
		if !nullable && !ai {
			def += " NOT NULL"
		}
		if pk {
			pkOut = qt
		}
	case "mysql":
		def = fmt.Sprintf("%s %s", qt, t)
		if !nullable {
			def += " NOT NULL"
		}
		if ai {
			def += " AUTO_INCREMENT"
		}
		if pk {
			pkOut = qt
		}
	case "sqlite":
		def = fmt.Sprintf("%s %s", qt, t)
		if pk {
			def += " PRIMARY KEY"
			if ai {
				def += " AUTOINCREMENT"
			}
		} else if !nullable {
			def += " NOT NULL"
		}
	case "sqlserver":
		def = fmt.Sprintf("%s %s", qt, t)
		if ai {
			def += " IDENTITY(1,1)"
		}
		if !nullable {
			def += " NOT NULL"
		}
		if pk {
			pkOut = qt
		}
	default:
		def = fmt.Sprintf("%s %s", qt, t)
		if !nullable {
			def += " NOT NULL"
		}
		if pk {
			pkOut = qt
		}
	}
	return def, pkOut
}
