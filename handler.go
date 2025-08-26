package weebase

import (
	"crypto/rand"
	"encoding/hex"
	"html/template"
	"log"
	"net/http"
	"path"
	"strings"
	"time"

	"gorm.io/gorm"

	apiConnect "github.com/dracory/weebase/api/api_connect"
	apiDisconnect "github.com/dracory/weebase/api/api_disconnect"
	apiProfilesList "github.com/dracory/weebase/api/api_profiles_list"
	apiProfilesSave "github.com/dracory/weebase/api/api_profiles_save"
	apiRowDelete "github.com/dracory/weebase/api/api_row_delete"
	apiRowInsert "github.com/dracory/weebase/api/api_row_insert"
	apiRowUpdate "github.com/dracory/weebase/api/api_row_update"
	apiRowView "github.com/dracory/weebase/api/api_row_view"
	apiRowsBrowse "github.com/dracory/weebase/api/api_rows_browse"
	apiSchemasList "github.com/dracory/weebase/api/api_schemas_list"
	apiSQLExecute "github.com/dracory/weebase/api/api_sql_execute"
	apiSQLExplain "github.com/dracory/weebase/api/api_sql_explain"
	apiTableCreate "github.com/dracory/weebase/api/api_table_create"
	apiTableInfo "github.com/dracory/weebase/api/api_table_info"
	apiTablesList "github.com/dracory/weebase/api/api_tables_list"
	pageHome "github.com/dracory/weebase/pages/page_home"
	pageLogin "github.com/dracory/weebase/pages/page_login"
	pageLogout "github.com/dracory/weebase/pages/page_logout"
	pageTableCreate "github.com/dracory/weebase/pages/page_table_create"
	"github.com/dracory/weebase/shared/constants"
	"github.com/dracory/weebase/shared/driver"
	"github.com/dracory/weebase/shared/layout"
	"github.com/dracory/weebase/shared/session"
	"github.com/dracory/weebase/shared/types"
	"github.com/dracory/weebase/shared/urls"
	"github.com/gouniverse/hb"
)

// newSessionID generates a new random session ID
func newSessionID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic(err) // This should never happen with crypto/rand
	}
	return hex.EncodeToString(b)
}

// execAdapter adapts a function to the Exec interface required by api/table_create.
type execAdapter struct{ exec func(string) error }

func (e execAdapter) Exec(sql string) error { return e.exec(sql) }

// Handler implements http.Handler for the single-endpoint router controlled by a query action.
type Handler struct {
	opts     Options
	tmplBase *template.Template
	drivers  *DriverRegistry
	profiles types.ConnectionStore
}

// tryAutoConnect opens and pings a DB, then stores it into the session.
func (h *Handler) tryAutoConnect(s *session.Session, driver, dsn string) error {
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

	// Create a new session.ActiveConnection with the correct type
	s.Conn = &session.ActiveConnection{
		ID:       newSessionID(),
		Driver:   driver,
		DB:       db,
		LastUsed: time.Now(),
	}
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
			p.ID = newSessionID()
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
	s := session.EnsureSession(w, r, h.opts.SessionSecret)

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
		case constants.ActionPageLogin,
			constants.ActionAssetCSS,
			constants.ActionAssetJS,
			constants.ActionHealthz,
			constants.ActionReadyz,
			constants.ActionLoginJS,
			constants.ActionLoginCSS:
			// allow
		default:
			http.Redirect(w, r, urls.Login(h.opts.BasePath), http.StatusFound)
			return
		}
	}

	// Choose handler set based on method: GET -> page, POST -> API
	var handlers map[string]func(http.ResponseWriter, *http.Request)
	switch r.Method {
	case http.MethodGet:
		handlers = h.pageHandlers(s, csrfToken)
	case http.MethodPost:
		handlers = h.apiHandlers(s, csrfToken)
	default:
		h.renderStatus(w, r, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	// empty action maps to home
	if action == "" {
		action = constants.ActionHome
	}
	if handler, ok := handlers[action]; ok {
		handler(w, r)
		return
	}
	// Default 404 within layout
	h.renderStatus(w, r, http.StatusNotFound, "Unknown action: "+action)
}

// pageHandlers assembles the request-scoped action map for GET requests.
func (h *Handler) pageHandlers(s *session.Session, csrfToken string) map[string]func(http.ResponseWriter, *http.Request) {
	// GET-only handlers that render pages or public assets
	return map[string]func(http.ResponseWriter, *http.Request){
		constants.ActionAssetCSS: func(w http.ResponseWriter, r *http.Request) { serveAsset(w, r, AssetPathCSS, ContentTypeCSS) },
		constants.ActionAssetJS:  func(w http.ResponseWriter, r *http.Request) { serveAsset(w, r, AssetPathJS, ContentTypeJS) },
		constants.ActionLoginCSS: func(w http.ResponseWriter, r *http.Request) { serveAsset(w, r, LoginAssetPathCSS, ContentTypeCSS) },
		constants.ActionLoginJS:  func(w http.ResponseWriter, r *http.Request) { serveAsset(w, r, LoginAssetPathJS, ContentTypeJS) },
		constants.ActionHealthz: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		},
		constants.ActionReadyz: func(w http.ResponseWriter, r *http.Request) {
			if s.Conn != nil {
				if gormDB, ok := s.Conn.DB.(*gorm.DB); ok {
					if sqlDB, err := gormDB.DB(); err == nil {
						if err := sqlDB.Ping(); err != nil {
							http.Error(w, "not ready", http.StatusServiceUnavailable)
							return
						}
					}
				}
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ready"))
		},
		constants.ActionHome: func(w http.ResponseWriter, r *http.Request) {
			var connInfo map[string]any
			if s.Conn != nil {
				connInfo = map[string]any{"driver": s.Conn.Driver}
			}
			full, err := pageHome.Handle(h.opts.BasePath, h.opts.ActionParam, h.drivers.List(), h.opts.AllowAdHocConnections, h.opts.SafeModeDefault, csrfToken, connInfo)
			if err != nil {
				log.Printf("render home: %v", err)
				h.renderStatus(w, r, http.StatusInternalServerError, "template error")
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write([]byte(full))
		},
		constants.ActionPageLogin: func(w http.ResponseWriter, r *http.Request) {
			full, err := pageLogin.Handle(h.tmplBase, h.opts.BasePath, h.opts.ActionParam, h.drivers.List(), h.opts.AllowAdHocConnections, h.opts.SafeModeDefault, csrfToken)
			if err != nil {
				log.Printf("render login: %v", err)
				h.renderStatus(w, r, http.StatusInternalServerError, "template error")
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write([]byte(full))
		},
		constants.ActionPageLogout: func(w http.ResponseWriter, r *http.Request) {
			full, err := pageLogout.Handle(
				h.tmplBase, h.opts.BasePath, h.opts.ActionParam, h.opts.SafeModeDefault, csrfToken)
			if err != nil {
				log.Printf("render logout: %v", err)
				h.renderStatus(w, r, http.StatusInternalServerError, "template error")
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write([]byte(full))
		},
		// New explicit page handler for table create
		constants.ActionPageTableCreate: func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				WriteError(w, r, "method not allowed")
				return
			}
			full, err := pageTableCreate.Handle(h.opts.BasePath, h.opts.ActionParam, EnsureCSRFCookie(w, r, h.opts.SessionSecret), h.opts.SafeModeDefault)
			if err != nil {
				h.renderStatus(w, r, http.StatusInternalServerError, err.Error())
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write([]byte(full))
		},
		constants.ActionPageTableView: func(w http.ResponseWriter, r *http.Request) {
			// temporary redirect to browse_rows while richer view is being built
			q := r.URL.Query()
			params := map[string]string{}
			if schema := q.Get("schema"); schema != "" {
				params["schema"] = schema
			}
			if table := q.Get("table"); table != "" {
				params["table"] = table
			}
			if limit := q.Get("limit"); limit != "" {
				params["limit"] = limit
			}
			if offset := q.Get("offset"); offset != "" {
				params["offset"] = offset
			}
			dest := urls.URL(h.opts.BasePath, constants.ActionPageTableView, params)
			http.Redirect(w, r, dest, http.StatusFound)
		},
	}
}

// apiHandlers are POST-only handlers that perform API operations
// createProfilesSaveHandler creates a new ProfilesSave handler with the correct dependencies
func (h *Handler) createProfilesSaveHandler() http.HandlerFunc {
	validator := driver.NewValidator(&driverRegistryWrapper{h.drivers})
	handler := apiProfilesSave.New(h.profiles, validator)
	return handler.Handle
}

func (h *Handler) apiHandlers(s *session.Session, csrfToken string) map[string]func(http.ResponseWriter, *http.Request) {
	// Create driver validator
	validator := driver.NewValidator(&driverRegistryWrapper{h.drivers})

	// Create disconnect handler
	disconnectHandler := apiDisconnect.New(h.opts.SessionSecret).Handle

	return map[string]func(http.ResponseWriter, *http.Request){
		// Connection management
		constants.ActionConnect:    apiConnect.New(h.opts.SessionSecret, h.profiles, validator).Handle,
		constants.ActionDisconnect: disconnectHandler,
		// Schema and table operations
		constants.ActionListSchemas: apiSchemasList.New(s.Conn).Handle,
		constants.ActionSchemasList: apiSchemasList.New(s.Conn).Handle,
		constants.ActionListTables:  apiTablesList.New(s.Conn).Handle,
		constants.ActionTablesList:  apiTablesList.New(s.Conn).Handle,

		// Row operations
		constants.ActionBrowseRows: apiRowsBrowse.New(s.Conn).Handle,
		constants.ActionRowsBrowse: apiRowsBrowse.New(s.Conn).Handle,
		constants.ActionRowView:    apiRowView.New(s.Conn).Handle,
		constants.ActionDeleteRow:  apiRowDelete.New(s.Conn, h.opts.SafeModeDefault, h.opts.SessionSecret).Handle,
		constants.ActionRowDelete:  apiRowDelete.New(s.Conn, h.opts.SafeModeDefault, h.opts.SessionSecret).Handle,
		constants.ActionInsertRow:  apiRowInsert.New(s.Conn, h.opts.SafeModeDefault).Handle,
		constants.ActionRowInsert:  apiRowInsert.New(s.Conn, h.opts.SafeModeDefault).Handle,
		constants.ActionUpdateRow:  apiRowUpdate.New(s.Conn, h.opts.SafeModeDefault).Handle,
		constants.ActionRowUpdate:  apiRowUpdate.New(s.Conn, h.opts.SafeModeDefault).Handle,

		// Profiles
		constants.ActionProfiles:     apiProfilesList.New(h.profiles).Handle,
		constants.ActionProfilesList: apiProfilesList.New(h.profiles).Handle,
		constants.ActionProfilesSave: h.createProfilesSaveHandler(),
		constants.ActionProfileSave:  h.createProfilesSaveHandler(),

		// SQL operations
		constants.ActionSQLExecute: apiSQLExecute.New(s.Conn, h.opts.SafeModeDefault, h.opts.ReadOnlyMode).Handle,
		constants.ActionSQLExplain: apiSQLExplain.New(s.Conn).Handle,

		// Table operations
		constants.ActionTableInfo:      apiTableInfo.New(s.Conn).Handle,
		constants.ActionApiTableCreate: apiTableCreate.New(s.Conn).Handle,

		// Import/Export (not implemented yet)
		constants.ActionExport: func(w http.ResponseWriter, r *http.Request) { JSONNotImplemented(w, constants.ActionExport) },
		constants.ActionImport: func(w http.ResponseWriter, r *http.Request) { JSONNotImplemented(w, constants.ActionImport) },
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
