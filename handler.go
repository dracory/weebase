package weebase

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
	"path"
	"strings"

	loginpage "github.com/dracory/weebase/pages/login"
	layout "github.com/dracory/weebase/shared/layout"
)

// Handler implements http.Handler for the single-endpoint router controlled by a query action.
type Handler struct {
	opts     Options
	tmplBase *template.Template
	drivers  *DriverRegistry
	profiles ConnectionStore
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
	w.Header().Set("Content-Security-Policy", "default-src 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline'; script-src 'self' 'unsafe-inline' https://unpkg.com")

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

	// If no active connection, redirect GET requests to login except for public assets/health
	action := r.URL.Query().Get(h.opts.ActionParam)
	if s.Conn == nil && r.Method == http.MethodGet {
		switch action {
		case ActionLogin, ActionAssetCSS, ActionAssetJS, ActionHealthz, ActionReadyz, ActionLoginJS, ActionLoginCSS:
			// allow
		default:
			http.Redirect(w, r, h.opts.BasePath+"?"+h.opts.ActionParam+"="+ActionLogin, http.StatusFound)
			return
		}
	}
	switch action {
	case "", ActionHome:
		h.handleHome(w, r, csrfToken)
		return
	case ActionAssetCSS:
		serveAsset(w, r, AssetPathCSS, ContentTypeCSS)
		return
	case ActionAssetJS:
		serveAsset(w, r, AssetPathJS, ContentTypeJS)
		return
	case ActionLoginCSS:
		serveAsset(w, r, LoginAssetPathCSS, ContentTypeCSS)
		return
	case ActionLoginJS:
		serveAsset(w, r, LoginAssetPathJS, ContentTypeJS)
		return
	case ActionHealthz:
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
		return
	case ActionReadyz:
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
	case ActionLogin:
		// Render login page via pages/login package
		full, err := loginpage.Handle(h.tmplBase, h.opts.BasePath, h.opts.ActionParam, h.drivers.List(), h.opts.AllowAdHocConnections, h.opts.SafeModeDefault, csrfToken)
		if err != nil {
			log.Printf("render login: %v", err)
			h.renderStatus(w, r, http.StatusInternalServerError, "template error")
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(full))
		return
	case ActionConnect:
		h.handleConnect(w, r)
		return
	case ActionDisconnect:
		h.handleDisconnect(w, r)
		return
	case ActionListSchemas:
		h.handleListSchemas(w, r)
		return
	case ActionListTables:
		h.handleListTables(w, r)
		return
	case ActionTableInfo:
		h.handleTableInfo(w, r)
		return
	case ActionBrowseRows:
		h.handleBrowseRows(w, r)
		return
	case ActionRowView:
		h.handleRowView(w, r)
	case ActionDeleteRow:
		h.handleDeleteRow(w, r)
		return
	case ActionInsertRow:
		h.handleInsertRow(w, r)
		return
	case ActionUpdateRow:
		h.handleUpdateRow(w, r)
		return
	case ActionViewDefinition:
		h.handleViewDefinition(w, r)
		return
	case ActionProfiles:
		h.handleProfiles(w, r)
		return
	case ActionProfilesSave:
		h.handleProfilesSave(w, r)
		return
	// --- Action stubs (to be implemented) ---
	case ActionSQLExecute, ActionSQLExplain,
		ActionListSaved, ActionSaveQuery,
		ActionDDLCreateTable, ActionDDLAlterTable, ActionDDLDropTable,
		ActionExport, ActionImport,
		ActionLogout:
		// handle implemented SQL console actions
		if action == ActionSQLExecute {
			h.handleSQLExecute(w, r)
			return
		}
		if action == ActionSQLExplain {
			h.handleSQLExplain(w, r)
			return
		}
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
	// Render content-only template into buffer
	var buf bytes.Buffer
	if err := h.tmplBase.ExecuteTemplate(&buf, "index_content", data); err != nil {
		log.Printf("render home content: %v", err)
		h.renderStatus(w, r, http.StatusInternalServerError, "template error")
		return
	}
	// Wrap with shared layout
	full := layout.Render("Home", h.opts.BasePath, h.opts.SafeModeDefault, template.HTML(buf.String()), nil, nil)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(full))
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

// handleListSchemas is implemented in `handler_list_schemas.go`.

// handleListTables is implemented in `handler_list_tables.go`.
