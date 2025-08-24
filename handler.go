package weebase

import (
	"html/template"
	"log"
	"net/http"
	"path"
	"strconv"
	"strings"
)

// Handler implements http.Handler for the single-endpoint router controlled by a query action.
type Handler struct {
	opts     Options
	tmplBase *template.Template
	drivers  *DriverRegistry
	profiles ConnectionStore
}

// handleViewDefinition returns the SQL definition of a view.
func (h *Handler) handleViewDefinition(w http.ResponseWriter, r *http.Request) {
	s := EnsureSession(w, r, h.opts.SessionSecret)
	if s.Conn == nil || s.Conn.DB == nil {
		WriteError(w, r, "not connected")
		return
	}
	_ = r.ParseForm()
	schema := strings.TrimSpace(r.Form.Get("schema"))
	view := strings.TrimSpace(r.Form.Get("view"))
	if view == "" {
		WriteError(w, r, "view is required")
		return
	}
	if !sanitizeIdent(view) || (schema != "" && !sanitizeIdent(schema)) {
		WriteError(w, r, "invalid identifier")
		return
	}
	driver := normalizeDriver(s.Conn.Driver)
	var sqlStr string
	var def string
	var err error
	switch driver {
	case "postgres":
		// Prefer pg_get_viewdef; schema optional (default public)
		if schema == "" {
			schema = "public"
		}
		err = s.Conn.DB.Raw(
			"SELECT pg_get_viewdef(format('%%s.%%s', ?, ?), true)", schema, view,
		).Scan(&def).Error
		if err == nil && def == "" {
			// Fallback via information_schema
			err = s.Conn.DB.Raw(
				"SELECT view_definition FROM information_schema.views WHERE table_schema = ? AND table_name = ?",
				schema, view,
			).Scan(&def).Error
		}
	case "mysql":
		// MySQL requires database (schema). SHOW CREATE VIEW returns two columns: View, Create View
		if schema == "" {
			WriteError(w, r, "schema required for mysql")
			return
		}
		rows, qerr := s.Conn.DB.Raw("SHOW CREATE VIEW `" + schema + "`.`" + view + "`").Rows()
		if qerr != nil {
			err = qerr
			break
		}
		defer rows.Close()
		if rows.Next() {
			var viewName, createSQL string
			if scanErr := rows.Scan(&viewName, &createSQL); scanErr != nil {
				err = scanErr
				break
			}
			def = createSQL
		}
	case "sqlite":
		// sqlite_master holds the SQL
		err = s.Conn.DB.Raw(
			"SELECT sql FROM sqlite_master WHERE type='view' AND name = ?",
			view,
		).Scan(&def).Error
	case "sqlserver":
		// Join sys.views to sys.sql_modules; default schema dbo
		if schema == "" {
			schema = "dbo"
		}
		err = s.Conn.DB.Raw(
			"SELECT m.definition FROM sys.views v JOIN sys.schemas s ON v.schema_id=s.schema_id JOIN sys.sql_modules m ON v.object_id=m.object_id WHERE s.name = ? AND v.name = ?",
			schema, view,
		).Scan(&def).Error
	default:
		WriteError(w, r, "unsupported driver")
		return
	}
	if err != nil {
		WriteError(w, r, err.Error())
		return
	}
	// If still empty, surface a friendly message
	if strings.TrimSpace(def) == "" {
		def = "<empty>"
	}
	WriteSuccessWithData(w, r, "ok", map[string]any{"definition": def, "sql": sqlStr})
}

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

// handleBrowseRows selects rows from a table with pagination.
func (h *Handler) handleBrowseRows(w http.ResponseWriter, r *http.Request) {
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
	limit := 50
	offset := 0
	if v := strings.TrimSpace(r.Form.Get("limit")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 500 {
			limit = n
		}
	}
	if v := strings.TrimSpace(r.Form.Get("offset")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}
	driver := normalizeDriver(s.Conn.Driver)
	var sqlStr string
	var args []any
	switch driver {
	case "postgres":
		if schema == "" {
			sqlStr = "SELECT * FROM \"" + table + "\" LIMIT ? OFFSET ?"
			args = []any{limit, offset}
		} else {
			sqlStr = "SELECT * FROM \"" + schema + "\".\"" + table + "\" LIMIT ? OFFSET ?"
			args = []any{limit, offset}
		}
	case "mysql":
		if schema == "" {
			WriteError(w, r, "schema required for mysql")
			return
		}
		sqlStr = "SELECT * FROM `" + schema + "`.`" + table + "` LIMIT ? OFFSET ?"
		args = []any{limit, offset}
	case "sqlite":
		sqlStr = "SELECT * FROM `" + table + "` LIMIT ? OFFSET ?"
		args = []any{limit, offset}
	case "sqlserver":
		// Without ORDER BY, OFFSET is not allowed; use TOP and ignore offset for now.
		if schema == "" {
			sqlStr = "SELECT TOP (?) * FROM [" + table + "]"
			args = []any{limit}
		} else {
			sqlStr = "SELECT TOP (?) * FROM [" + schema + "].[" + table + "]"
			args = []any{limit}
		}
	default:
		WriteError(w, r, "unsupported driver")
		return
	}
	// Execute and materialize rows into []map[string]any
	rows, err := s.Conn.DB.Raw(sqlStr, args...).Rows()
	if err != nil {
		WriteError(w, r, err.Error())
		return
	}
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		WriteError(w, r, err.Error())
		return
	}
	out := make([]map[string]any, 0, limit)
	for rows.Next() {
		vals := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			WriteError(w, r, err.Error())
			return
		}
		m := make(map[string]any, len(cols))
		for i, c := range cols {
			m[c] = vals[i]
		}
		out = append(out, m)
	}
	WriteSuccessWithData(w, r, "ok", map[string]any{"rows": out, "limit": limit, "offset": offset})
}

// sanitizeIdent allows only letters, digits, underscore and dot.
func sanitizeIdent(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '.' {
			continue
		}
		return false
	}
	return true
}

// tryAutoConnect opens and pings a DB, then stores it into the session.
func (h *Handler) tryAutoConnect(s *Session, driver, dsn string) error {
	db, err := OpenGORM(driver, dsn)
	if err != nil {
		return err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	if err := sqlDB.Ping(); err != nil {
		return err
	}
	s.Conn = &ActiveConnection{Driver: driver, DSN: dsn, DB: db}
	return nil
}

// handleConnect establishes a DB connection and stores it in the session.
func (h *Handler) handleConnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, r, "connect must be POST")
		return
	}
	_ = r.ParseForm()
	profileID := strings.TrimSpace(r.Form.Get("profile_id"))
	driver := strings.TrimSpace(r.Form.Get("driver"))
	dsn := strings.TrimSpace(r.Form.Get("dsn"))
	// If profile_id provided, resolve it
	if profileID != "" {
		if p, ok := h.profiles.Get(profileID); ok {
			driver, dsn = p.Driver, p.DSN
		} else {
			WriteError(w, r, "profile not found")
			return
		}
	}
	if err := h.ValidateDriver(driver); err != nil {
		WriteError(w, r, err.Error())
		return
	}
	if dsn == "" {
		WriteError(w, r, "dsn is required")
		return
	}
	s := EnsureSession(w, r, h.opts.SessionSecret)
	if err := h.tryAutoConnect(s, driver, dsn); err != nil {
		WriteError(w, r, err.Error())
		return
	}
	WriteSuccessWithData(w, r, "connected", map[string]any{"driver": driver})
}

// handleDisconnect clears the active session connection.
func (h *Handler) handleDisconnect(w http.ResponseWriter, r *http.Request) {
	s := EnsureSession(w, r, h.opts.SessionSecret)
	if s.Conn != nil {
		if sqlDB, err := s.Conn.DB.DB(); err == nil {
			_ = sqlDB.Close()
		}
	}
	s.Conn = nil
	WriteSuccess(w, r, http.StatusOK, "disconnected")
}

// handleProfiles lists saved profiles (GET) using the in-memory store for now.
func (h *Handler) handleProfiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, r, "profiles must be GET")
		return
	}
	list := h.profiles.List()
	WriteSuccessWithData(w, r, "ok", map[string]any{"profiles": list})
}

// handleProfilesSave saves a new profile (POST) in the in-memory store.
func (h *Handler) handleProfilesSave(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, r, "profiles_save must be POST")
		return
	}
	_ = r.ParseForm()
	name := strings.TrimSpace(r.Form.Get("name"))
	driver := strings.TrimSpace(r.Form.Get("driver"))
	dsn := strings.TrimSpace(r.Form.Get("dsn"))
	if name == "" || driver == "" || dsn == "" {
		WriteError(w, r, "name, driver and dsn are required")
		return
	}
	if err := h.ValidateDriver(driver); err != nil {
		WriteError(w, r, err.Error())
		return
	}
	p := ConnectionProfile{ID: newRandomID(), Name: name, Driver: driver, DSN: dsn}
	if err := h.profiles.Save(p); err != nil {
		WriteError(w, r, err.Error())
		return
	}
	WriteSuccessWithData(w, r, "saved", map[string]any{"profile": p})
}

// NewHandler constructs a new Handler with defaults applied.
func NewHandler(o Options) http.Handler {
	o = o.withDefaults()
	// build templates once
	tmpl := parseTemplates()
	// initialize driver registry and in-memory profile store for now
	reg := NewDriverRegistry(o.EnabledDrivers)
	store := NewMemoryConnectionStore()
	// preload preconfigured profiles
	for _, p := range o.PreconfiguredProfiles {
		if p.ID == "" {
			p.ID = newRandomID()
		}
		_ = store.Save(p)
	}
	return &Handler{opts: o, tmplBase: tmpl, drivers: reg, profiles: store}
}

// Register mounts the handler on the provided mux at path.
func Register(mux *http.ServeMux, path string, h http.Handler) {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	mux.Handle(path, h)
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Basic secure headers
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("Content-Security-Policy", "default-src 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline'; script-src 'self' 'unsafe-inline'")

	// Ensure a session exists (sets cookie if missing)
	s := EnsureSession(w, r, h.opts.SessionSecret)

	// Ensure CSRF cookie and get a token for templates
	csrfToken := EnsureCSRFCookie(w, r, h.opts.SessionSecret)

	// Verify CSRF for POST requests
	if r.Method == http.MethodPost {
		if !VerifyCSRF(r, h.opts.SessionSecret) {
			WriteError(w, r, "invalid or missing CSRF token")
			return
		}
	}

	// Auto-connect on first GET if DefaultConnection is configured and no active session conn
	if r.Method == http.MethodGet && s.Conn == nil && h.opts.DefaultConnection != nil {
		if err := h.tryAutoConnect(s, h.opts.DefaultConnection.Driver, h.opts.DefaultConnection.DSN); err != nil {
			// Do not fail the request; just log via standard logger for now
			log.Printf("auto-connect failed: %v", err)
		}
	}

	action := r.URL.Query().Get(h.opts.ActionParam)
	switch action {
	case "", "home":
		h.handleHome(w, r, csrfToken)
		return
	case "asset_css":
		serveAsset(w, r, "assets/style.css", "text/css; charset=utf-8")
		return
	case "asset_js":
		serveAsset(w, r, "assets/app.js", "application/javascript; charset=utf-8")
	case "healthz":
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
		return
	case "readyz":
		// If there's an active connection, verify we can ping it.
		if s.Conn != nil {
			if sqlDB, err := s.Conn.DB.DB(); err == nil {
				if err := sqlDB.Ping(); err != nil {
					http.Error(w, "not ready", http.StatusServiceUnavailable)
					return
				}
			}
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ready"))
		return
	// --- Implemented actions ---
	case "connect":
		h.handleConnect(w, r)
		return
	case "disconnect":
		h.handleDisconnect(w, r)
		return
	case "list_schemas":
		h.handleListSchemas(w, r)
		return
	case "list_tables":
		h.handleListTables(w, r)
		return
	case "table_info":
		h.handleTableInfo(w, r)
		return
	case "browse_rows":
		h.handleBrowseRows(w, r)
		return
	case "view_definition":
		h.handleViewDefinition(w, r)
		return
	case "profiles":
		h.handleProfiles(w, r)
		return
	case "profiles_save":
		h.handleProfilesSave(w, r)
		return
	// --- Action stubs (to be implemented) ---
	case "insert_row", "update_row", "delete_row",
		"sql_execute", "sql_explain",
		"list_saved_queries", "save_query",
		"ddl_create_table", "ddl_alter_table", "ddl_drop_table",
		"export", "import",
		"login", "logout":
		JSONNotImplemented(w, action)
		return
	default:
		// For now, render 404 within layout
		h.renderStatus(w, r, http.StatusNotFound, "Unknown action: "+action)
		return
	}
}

func (h *Handler) handleHome(w http.ResponseWriter, r *http.Request, csrfToken string) {
	s := EnsureSession(w, r, h.opts.SessionSecret)
	var connInfo map[string]any
	if s.Conn != nil {
		connInfo = map[string]any{"driver": s.Conn.Driver}
	}
	data := map[string]any{
		"Title":                 "WeeBase",
		"BasePath":              h.opts.BasePath,
		"ActionParam":           h.opts.ActionParam,
		"EnabledDrivers":        h.drivers.List(),
		"AllowAdHocConnections": h.opts.AllowAdHocConnections,
		"SafeModeDefault":       h.opts.SafeModeDefault,
		"CSRFToken":             csrfToken,
		"Conn":                  connInfo,
	}
	if err := h.tmplBase.ExecuteTemplate(w, "index.gohtml", data); err != nil {
		log.Printf("render home: %v", err)
		h.renderStatus(w, r, http.StatusInternalServerError, "template error")
		return
	}
}

func (h *Handler) renderStatus(w http.ResponseWriter, r *http.Request, code int, msg string) {
	w.WriteHeader(code)
	data := map[string]any{
		"Title":    http.StatusText(code),
		"Message":  msg,
		"BasePath": h.opts.BasePath,
	}
	_ = h.tmplBase.ExecuteTemplate(w, "status.gohtml", data)
}

func serveAsset(w http.ResponseWriter, r *http.Request, assetPath, contentType string) {
	// Set content type and cache headers
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	w.Header().Set("Content-Disposition", "inline; filename="+path.Base(assetPath))
	http.ServeFileFS(w, r, embeddedFS, assetPath)
}

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

// handleListTables returns table names for a given schema (where applicable).
func (h *Handler) handleListTables(w http.ResponseWriter, r *http.Request) {
	s := EnsureSession(w, r, h.opts.SessionSecret)
	if s.Conn == nil || s.Conn.DB == nil {
		WriteError(w, r, "not connected")
		return
	}
	_ = r.ParseForm()
	schema := strings.TrimSpace(r.Form.Get("schema"))
	q := strings.TrimSpace(r.Form.Get("q"))
	includeViews := strings.TrimSpace(r.Form.Get("include_views")) != "" && strings.TrimSpace(r.Form.Get("include_views")) != "0"
	limit := 50
	offset := 0
	if v := strings.TrimSpace(r.Form.Get("limit")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 500 {
			limit = n
		}
	}
	if v := strings.TrimSpace(r.Form.Get("offset")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}
	driver := normalizeDriver(s.Conn.Driver)
	type row struct{ Name string }
	var rows []row
	var err error
	switch driver {
	case "postgres":
		if schema == "" {
			schema = "public"
		}
		// Build query: filter by schema, optionally include views, optional name search via ILIKE, with limit/offset
		base := "SELECT table_name AS name FROM information_schema.tables WHERE table_schema = ?"
		args := []any{schema}
		if !includeViews {
			base += " AND table_type = 'BASE TABLE'"
		}
		if q != "" {
			base += " AND table_name ILIKE ?"
			args = append(args, "%"+q+"%")
		}
		base += " ORDER BY name LIMIT ? OFFSET ?"
		args = append(args, limit, offset)
		err = s.Conn.DB.Raw(base, args...).Scan(&rows).Error
	case "mysql":
		// In MySQL, schema == database
		if schema == "" {
			WriteError(w, r, "schema required")
			return
		}
		base := "SELECT table_name AS name FROM information_schema.tables WHERE table_schema = ?"
		args := []any{schema}
		if !includeViews {
			base += " AND table_type = 'BASE TABLE'"
		}
		if q != "" {
			base += " AND table_name LIKE ?"
			args = append(args, "%"+q+"%")
		}
		base += " ORDER BY name LIMIT ? OFFSET ?"
		args = append(args, limit, offset)
		err = s.Conn.DB.Raw(base, args...).Scan(&rows).Error
	case "sqlite":
		// sqlite_master holds tables and views
		base := "SELECT name AS name FROM sqlite_master WHERE "
		if includeViews {
			base += "type IN ('table','view')"
		} else {
			base += "type = 'table'"
		}
		var args []any
		if q != "" {
			base += " AND name LIKE ?"
			args = append(args, "%"+q+"%")
		}
		base += " ORDER BY name LIMIT ? OFFSET ?"
		args = append(args, limit, offset)
		err = s.Conn.DB.Raw(base, args...).Scan(&rows).Error
	case "sqlserver":
		if schema == "" {
			schema = "dbo"
		}
		// Use sys.objects to allow including views (U=user table, V=view). OFFSET/FETCH requires ORDER BY.
		base := "SELECT o.name AS name FROM sys.objects o JOIN sys.schemas s ON o.schema_id=s.schema_id WHERE s.name = ? AND o.type IN (" +
			func() string { if includeViews { return "'U','V'" } else { return "'U'" } }() + ")"
		args := []any{schema}
		if q != "" {
			base += " AND o.name LIKE ?"
			args = append(args, "%"+q+"%")
		}
		base += " ORDER BY o.name OFFSET ? ROWS FETCH NEXT ? ROWS ONLY"
		args = append(args, offset, limit)
		err = s.Conn.DB.Raw(base, args...).Scan(&rows).Error
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
	WriteSuccessWithData(w, r, "ok", map[string]any{"tables": names, "limit": limit, "offset": offset})
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

// quoteIdent safely quotes a possibly schema-qualified identifier for the given driver.
// It assumes parts have already passed sanitizeIdent (letters/digits/underscore/dot).
// This is a best-effort helper for building simple SQL strings.
func quoteIdent(driver, ident string) string {
	d := normalizeDriver(driver)
	// split by '.' to support schema.table
	parts := strings.Split(ident, ".")
	for i, p := range parts {
		switch d {
		case "postgres":
			// escape double quotes by doubling them
			p = strings.ReplaceAll(p, "\"", "\"\"")
			parts[i] = "\"" + p + "\""
		case "mysql", "sqlite":
			// escape backticks by doubling them
			p = strings.ReplaceAll(p, "`", "``")
			parts[i] = "`" + p + "`"
		case "sqlserver":
			// escape ']' by doubling it
			p = strings.ReplaceAll(p, "]", "]]")
			parts[i] = "[" + p + "]"
		default:
			parts[i] = p
		}
	}
	return strings.Join(parts, ".")
}
