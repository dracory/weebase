package weebase

import (
    "fmt"
    "net/http"
    "strings"

    createpage "github.com/dracory/weebase/pages/table_create"
)

// handleDDLCreateTable serves an Adminer-like Create Table page and executes the create on POST.
func (h *Handler) handleDDLCreateTable(w http.ResponseWriter, r *http.Request) {
    s := EnsureSession(w, r, h.opts.SessionSecret)
    if s.Conn == nil || s.Conn.DB == nil {
        WriteError(w, r, "not connected")
        return
    }
    switch r.Method {
    case http.MethodGet:
        // Render via pages/table_create to follow the pages/login structure
        full, err := createpage.Handle(h.opts.BasePath, h.opts.ActionParam, EnsureCSRFCookie(w, r, h.opts.SessionSecret), h.opts.SafeModeDefault)
        if err != nil {
            WriteError(w, r, err.Error())
            return
        }
        w.Header().Set("Content-Type", "text/html; charset=utf-8")
        _, _ = w.Write([]byte(full))
        return
    case http.MethodPost:
        h.executeCreateTable(w, r, s)
        return
    default:
        WriteError(w, r, "method not allowed")
        return
    }
}

func (h *Handler) executeCreateTable(w http.ResponseWriter, r *http.Request, s *Session) {
    if !VerifyCSRF(r, h.opts.SessionSecret) {
        WriteError(w, r, "invalid CSRF token")
        return
    }
    _ = r.ParseForm()
    schema := strings.TrimSpace(r.Form.Get("schema"))
    table := strings.TrimSpace(r.Form.Get("table"))
    if table == "" {
        WriteError(w, r, "table required")
        return
    }
    names := r.Form["col_name[]"]
    types := r.Form["col_type[]"]
    lens := r.Form["col_length[]"]
    // checkboxes return subset; mark via map of indices present
    nullable := indexSet(r.Form["col_nullable[]"]) 
    pkset := indexSet(r.Form["col_pk[]"]) 
    aiset := indexSet(r.Form["col_ai[]"]) 

    // Build column definitions
    driver := normalizeDriver(s.Conn.Driver)
    var defs []string
    var pks []string
    for i := range names {
        name := strings.TrimSpace(names[i])
        if name == "" { continue }
        typ := strings.TrimSpace(get(types, i))
        ln := strings.TrimSpace(get(lens, i))
        def, pk := buildColumnDef(driver, name, typ, ln, nullable[fmt.Sprint(i+1)], pkset[fmt.Sprint(i+1)], aiset[fmt.Sprint(i+1)])
        defs = append(defs, def)
        if pk != "" { pks = append(pks, pk) }
    }
    if len(defs) == 0 {
        WriteError(w, r, "at least one column required")
        return
    }
    // Add PK constraint if needed (for engines where not embedded)
    if len(pks) > 0 && driver != "sqlite" { // sqlite PK often embedded
        defs = append(defs, fmt.Sprintf("PRIMARY KEY (%s)", strings.Join(pks, ", ")))
    }

    // Build qualified and quoted table name using existing quoteIdent helper
    id := table
    if strings.TrimSpace(schema) != "" {
        id = schema + "." + table
    }
    stmt := fmt.Sprintf("CREATE TABLE %s (\n  %s\n)", quoteIdent(driver, id), strings.Join(defs, ",\n  "))

    if err := s.Conn.DB.Exec(stmt).Error; err != nil {
        WriteError(w, r, err.Error())
        return
    }
    WriteSuccessWithData(w, r, "created", map[string]any{"sql": stmt})
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
            // prefer serial type for AI
            if strings.Contains(strings.ToLower(typ), "big") {
                t = "bigserial"
            } else {
                t = "serial"
            }
        }
        def = fmt.Sprintf("%s %s", qt, t)
        if !nullable && !ai { // serial is NOT NULL implicitly
            def += " NOT NULL"
        }
        if pk { pkOut = qt }
    case "mysql":
        def = fmt.Sprintf("%s %s", qt, t)
        if !nullable { def += " NOT NULL" }
        if ai { def += " AUTO_INCREMENT" }
        if pk { pkOut = qt }
    case "sqlite":
        // In SQLite, AI only works with INTEGER PRIMARY KEY
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
        if ai { def += " IDENTITY(1,1)" }
        if !nullable { def += " NOT NULL" }
        if pk { pkOut = qt }
    default:
        def = fmt.Sprintf("%s %s", qt, t)
        if !nullable { def += " NOT NULL" }
        if pk { pkOut = qt }
    }
    return def, pkOut
}

func indexSet(vals []string) map[string]bool {
    m := map[string]bool{}
    for _, v := range vals { m[v] = true }
    return m
}

func get(ss []string, i int) string { if i < len(ss) { return ss[i] } ; return "" }
