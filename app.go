// Package weebase provides a lightweight database management module for Go web applications.
// It supports MySQL, PostgreSQL, SQLite, and SQL Server databases through a simple HTTP interface.
package weebase

import (
	"net/http"

	"github.com/dracory/api"
	"github.com/dracory/weebase/api/api_connect"
	"github.com/dracory/weebase/api/api_databases_list"
	"github.com/dracory/weebase/api/api_profiles_list"
	"github.com/dracory/weebase/api/api_tables_list"
	"github.com/dracory/weebase/pages/page_database"
	"github.com/dracory/weebase/pages/page_home"
	"github.com/dracory/weebase/pages/page_login"
	"github.com/dracory/weebase/pages/page_logout"
	"github.com/dracory/weebase/pages/page_table"
	"github.com/dracory/weebase/shared/constants"
	"github.com/dracory/weebase/shared/types"
	"github.com/samber/lo"

	"gorm.io/gorm"
)

// driverRegistry implements the driver.Registry interface
type driverRegistry struct {
	enabled map[string]bool
}

func (d *driverRegistry) IsEnabled(name string) bool {
	return d.enabled[name]
}

// Supported database drivers
const (
	MYSQL    = constants.DriverMySQL
	POSTGRES = constants.DriverPostgres
	SQLITE   = constants.DriverSQLite
	SQLSRV   = constants.DriverSQLServer
)

// App represents the main application instance
type App struct {
	config  types.Config
	db      *gorm.DB
	drivers map[string]driverConfig
}

type driverConfig struct {
	Name   string
	Driver string
}

// New creates a new App instance with the given configuration
// The configuration should be loaded using LoadConfig() from config.go
func New(cfg types.Config, options ...func(*types.Config)) *App {
	// Apply any option functions to the config
	for _, option := range options {
		option(&cfg)
	}

	// Initialize default drivers if none provided
	if len(cfg.EnabledDrivers) == 0 {
		cfg.EnabledDrivers = []string{MYSQL, POSTGRES, SQLITE, SQLSRV}
	}

	return &App{
		config:  cfg,
		drivers: make(map[string]driverConfig),
	}
}

// Handler returns an http.Handler that serves the App UI and API
func (g *App) Handler() http.Handler {
	mux := http.NewServeMux()

	// Register API handlers
	mux.HandleFunc(g.config.BasePath, g.handleRequest)

	return g.middleware(mux)
}

// handleRequest routes requests to the appropriate handler
func (g *App) handleRequest(w http.ResponseWriter, r *http.Request) {
	action := r.URL.Query().Get(g.config.ActionParam)
	apiActionMap := g.apiActions()
	pageActionMap := g.pageActions()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		api.Respond(w, r, api.Error("action not found: "+action+""))
	})

	if lo.HasKey(apiActionMap, action) {
		// wrap with API middleware (i.e. CORS, CSRF, etc.)
		handler = apiActionMap[action]
	}

	if lo.HasKey(pageActionMap, action) {
		// wrap with PAGE middleware (i.e. session, CSRF, etc.)
		handler = pageActionMap[action]
	}

	handler(w, r)
}

func (g *App) apiActions() map[string]func(w http.ResponseWriter, r *http.Request) {
	return map[string]func(w http.ResponseWriter, r *http.Request){
		constants.ActionApiConnect:       api_connect.New(g.config).ServeHTTP,
		constants.ActionApiDatabasesList: api_databases_list.New(g.config).ServeHTTP,
		constants.ActionApiProfilesList:  api_profiles_list.New(g.config).ServeHTTP,
		constants.ActionApiTablesList:    api_tables_list.New(g.config).Handle,
	}
}

func (g *App) pageActions() map[string]func(w http.ResponseWriter, r *http.Request) {
	return map[string]func(w http.ResponseWriter, r *http.Request){
		constants.ActionPageHome:     page_home.New(g.config).ServeHTTP,
		constants.ActionPageServer:   page_home.New(g.config).ServeHTTP,
		constants.ActionPageLogin:    page_login.New(g.config).ServeHTTP,
		constants.ActionPageLogout:   page_logout.New(g.config).ServeHTTP,
		constants.ActionPageDatabase: page_database.New(g.config).ServeHTTP,
		constants.ActionPageTable:    page_table.New(g.config).ServeHTTP,
	}
}

// middleware applies common middleware to all handlers
func (g *App) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "same-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval' cdn.jsdelivr.net cdn.tailwindcss.com unpkg.com; style-src 'self' 'unsafe-inline' cdn.jsdelivr.net cdn.tailwindcss.com unpkg.com; font-src 'self' cdn.jsdelivr.net; img-src 'self' data:;")

		next.ServeHTTP(w, r)
	})
}
