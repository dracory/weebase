package weebase

import (
	"html/template"
	"log"
	"net/http"
	"path"
	"strings"
)

// Handler implements http.Handler for the single-endpoint router controlled by a query action.
type Handler struct {
	opts     Options
	tmplBase *template.Template
	drivers  *DriverRegistry
	profiles ConnectionStore
}

// handleConnect establishes a DB connection and stores it in the session.
func (h *Handler) handleConnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, r, "connect must be POST")
		return
	}
	_ = r.ParseForm()
	driver := strings.TrimSpace(r.Form.Get("driver"))
	dsn := strings.TrimSpace(r.Form.Get("dsn"))
	if err := h.ValidateDriver(driver); err != nil {
		WriteError(w, r, err.Error())
		return
	}
	if dsn == "" {
		WriteError(w, r, "dsn is required")
		return
	}
	db, err := OpenGORM(driver, dsn)
	if err != nil {
		WriteError(w, r, "open: "+err.Error())
		return
	}
	if sqlDB, err2 := db.DB(); err2 != nil {
		WriteError(w, r, "db handle: "+err2.Error())
		return
	} else if pingErr := sqlDB.Ping(); pingErr != nil {
		WriteError(w, r, "ping: "+pingErr.Error())
		return
	}
	s := EnsureSession(w, r, h.opts.SessionSecret)
	s.Conn = &ActiveConnection{Driver: driver, DSN: dsn, DB: db}
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
	_ = EnsureSession(w, r, h.opts.SessionSecret)

	// Ensure CSRF cookie and get a token for templates
	csrfToken := EnsureCSRFCookie(w, r, h.opts.SessionSecret)

	// Verify CSRF for POST requests
	if r.Method == http.MethodPost {
		if !VerifyCSRF(r, h.opts.SessionSecret) {
			WriteError(w, r, "invalid or missing CSRF token")
			return
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
		// Ready when the handler is constructed; deeper checks added once a DB connection is active.
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
	case "profiles":
		h.handleProfiles(w, r)
		return
	case "profiles_save":
		h.handleProfilesSave(w, r)
		return
	// --- Action stubs (to be implemented) ---
	case "list_schemas", "list_tables", "table_info", "view_definition",
		"browse_rows",
		"insert_row", "update_row", "delete_row",
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
	data := map[string]any{
		"Title":           "WeeBase",
		"BasePath":        h.opts.BasePath,
		"ActionParam":     h.opts.ActionParam,
		"EnabledDrivers":  h.drivers.List(),
		"SafeModeDefault": h.opts.SafeModeDefault,
		"CSRFToken":       csrfToken,
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
