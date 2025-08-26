package weebase

import (
	"html/template"
	"log"
	"net/http"
	"path"
	"strings"

	apipkg "github.com/dracory/weebase/api"
	homepage "github.com/dracory/weebase/pages/home"
	loginpage "github.com/dracory/weebase/pages/login"
	layout "github.com/dracory/weebase/shared/layout"
	"github.com/dracory/weebase/shared/urls"
	hb "github.com/gouniverse/hb"
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
	w.Header().Set("Content-Security-Policy", "default-src 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline'; script-src 'self' 'unsafe-inline' 'unsafe-eval' https://unpkg.com https://cdn.jsdelivr.net https://cdn.tailwindcss.com")

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

	handlers := h.actionHandlers(r, s, csrfToken)
	// empty action maps to home
	if action == "" {
		action = ActionHome
	}
	if handler, ok := handlers[action]; ok {
		handler(w, r)
		return
	}
	// Default 404 within layout
	h.renderStatus(w, r, http.StatusNotFound, "Unknown action: "+action)
}

// actionHandlers assembles the request-scoped action map.
func (h *Handler) actionHandlers(r *http.Request, s *Session, csrfToken string) map[string]func(http.ResponseWriter, *http.Request) {
	return map[string]func(http.ResponseWriter, *http.Request){
		ActionAssetCSS: func(w http.ResponseWriter, r *http.Request) { serveAsset(w, r, AssetPathCSS, ContentTypeCSS) },
		ActionAssetJS:  func(w http.ResponseWriter, r *http.Request) { serveAsset(w, r, AssetPathJS, ContentTypeJS) },
		ActionLoginCSS: func(w http.ResponseWriter, r *http.Request) { serveAsset(w, r, LoginAssetPathCSS, ContentTypeCSS) },
		ActionLoginJS:  func(w http.ResponseWriter, r *http.Request) { serveAsset(w, r, LoginAssetPathJS, ContentTypeJS) },
		ActionHealthz:  func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK); _, _ = w.Write([]byte("ok")) },
		ActionReadyz: func(w http.ResponseWriter, r *http.Request) {
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
		},
		ActionHome: func(w http.ResponseWriter, r *http.Request) {
			var connInfo map[string]any
			if s.Conn != nil {
				connInfo = map[string]any{"driver": s.Conn.Driver}
			}
			full, err := homepage.Handle(h.opts.BasePath, h.opts.ActionParam, h.drivers.List(), h.opts.AllowAdHocConnections, h.opts.SafeModeDefault, csrfToken, connInfo)
			if err != nil {
				log.Printf("render home: %v", err)
				h.renderStatus(w, r, http.StatusInternalServerError, "template error")
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write([]byte(full))
		},
		ActionLogin: func(w http.ResponseWriter, r *http.Request) {
			full, err := loginpage.Handle(h.tmplBase, h.opts.BasePath, h.opts.ActionParam, h.drivers.List(), h.opts.AllowAdHocConnections, h.opts.SafeModeDefault, csrfToken)
			if err != nil {
				log.Printf("render login: %v", err)
				h.renderStatus(w, r, http.StatusInternalServerError, "template error")
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write([]byte(full))
		},
		ActionLogout:       func(w http.ResponseWriter, r *http.Request) { h.handleLogout(w, r, csrfToken) },
		ActionConnect:      func(w http.ResponseWriter, r *http.Request) { h.handleConnect(w, r) },
		ActionDisconnect:   func(w http.ResponseWriter, r *http.Request) { h.handleDisconnect(w, r) },
		ActionListSchemas:  func(w http.ResponseWriter, r *http.Request) { apipkg.SchemasList(h, w, r) },
		ActionListTables:   func(w http.ResponseWriter, r *http.Request) { apipkg.TablesList(h, w, r) },
		ActionSchemasList:  func(w http.ResponseWriter, r *http.Request) { apipkg.SchemasList(h, w, r) },
		ActionTablesList:   func(w http.ResponseWriter, r *http.Request) { apipkg.TablesList(h, w, r) },
		ActionTableInfo:    func(w http.ResponseWriter, r *http.Request) { h.handleTableInfo(w, r) },
		ActionBrowseRows:   func(w http.ResponseWriter, r *http.Request) { apipkg.RowsBrowse(h, w, r) },
		ActionRowsBrowse:   func(w http.ResponseWriter, r *http.Request) { apipkg.RowsBrowse(h, w, r) },
		ActionRowView:      func(w http.ResponseWriter, r *http.Request) { h.handleRowView(w, r) },
		ActionDeleteRow:    func(w http.ResponseWriter, r *http.Request) { h.handleDeleteRow(w, r) },
		ActionRowDelete:    func(w http.ResponseWriter, r *http.Request) { h.handleDeleteRow(w, r) },
		ActionInsertRow:    func(w http.ResponseWriter, r *http.Request) { h.handleInsertRow(w, r) },
		ActionRowInsert:    func(w http.ResponseWriter, r *http.Request) { h.handleInsertRow(w, r) },
		ActionUpdateRow:    func(w http.ResponseWriter, r *http.Request) { h.handleUpdateRow(w, r) },
		ActionRowUpdate:    func(w http.ResponseWriter, r *http.Request) { h.handleUpdateRow(w, r) },
		ActionViewDefinition: func(w http.ResponseWriter, r *http.Request) { h.handleViewDefinition(w, r) },
		ActionProfiles:     func(w http.ResponseWriter, r *http.Request) { h.handleProfiles(w, r) },
		ActionProfilesList: func(w http.ResponseWriter, r *http.Request) { h.handleProfiles(w, r) },
		ActionProfilesSave: func(w http.ResponseWriter, r *http.Request) { h.handleProfilesSave(w, r) },
		ActionProfileSave:  func(w http.ResponseWriter, r *http.Request) { h.handleProfilesSave(w, r) },
		ActionSQLExecute:   func(w http.ResponseWriter, r *http.Request) { h.handleSQLExecute(w, r) },
		ActionSQLExplain:   func(w http.ResponseWriter, r *http.Request) { h.handleSQLExplain(w, r) },
		ActionDDLCreateTable: func(w http.ResponseWriter, r *http.Request) { h.handleDDLCreateTable(w, r) },
		ActionTableCreate:    func(w http.ResponseWriter, r *http.Request) { h.handleDDLCreateTable(w, r) },
		ActionTableEdit:      func(w http.ResponseWriter, r *http.Request) { JSONNotImplemented(w, ActionDDLAlterTable) },
		ActionTableDrop:      func(w http.ResponseWriter, r *http.Request) { JSONNotImplemented(w, ActionDDLDropTable) },
		ActionExport:       func(w http.ResponseWriter, r *http.Request) { JSONNotImplemented(w, ActionExport) },
		ActionImport:       func(w http.ResponseWriter, r *http.Request) { JSONNotImplemented(w, ActionImport) },
		ActionTableView: func(w http.ResponseWriter, r *http.Request) {
			// temporary redirect to browse_rows while richer view is being built
			q := r.URL.Query()
			params := map[string]string{}
			if schema := q.Get("schema"); schema != "" { params["schema"] = schema }
			if table := q.Get("table"); table != "" { params["table"] = table }
			if limit := q.Get("limit"); limit != "" { params["limit"] = limit }
			if offset := q.Get("offset"); offset != "" { params["offset"] = offset }
			dest := urls.URL(h.opts.BasePath, ActionBrowseRows, params)
			http.Redirect(w, r, dest, http.StatusFound)
		},
	}
}

func (h *Handler) renderStatus(w http.ResponseWriter, r *http.Request, code int, msg string) {
	w.WriteHeader(code)
	// Build a minimal status page using shared layout
	main := hb.Div().Children([]hb.TagInterface{
		hb.Heading2().Text(http.StatusText(code)),
		hb.Paragraph().Text(msg),
	}).ToHTML()
	full := layout.RenderWith(layout.Options{
		Title:           http.StatusText(code),
		BasePath:        h.opts.BasePath,
		SafeModeDefault: h.opts.SafeModeDefault,
		MainHTML:        main,
	})
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(full))
}

func serveAsset(w http.ResponseWriter, r *http.Request, assetPath, contentType string) {
	// Set content type and cache headers
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	w.Header().Set("Content-Disposition", "inline; filename="+path.Base(assetPath))
	http.ServeFileFS(w, r, embeddedFS, assetPath)
}
