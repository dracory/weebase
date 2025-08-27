// Package weebase provides a lightweight database management module for Go web applications.
// It supports MySQL, PostgreSQL, SQLite, and SQL Server databases through a simple HTTP interface.
package weebase

import (
	"html/template"
	"net/http"

	"github.com/dracory/weebase/api/api_connect"
	"github.com/dracory/weebase/api/api_tables_list"
	"github.com/dracory/weebase/pages/page_database"
	"github.com/dracory/weebase/pages/page_home"
	"github.com/dracory/weebase/pages/page_login"
	"github.com/dracory/weebase/pages/page_logout"
	page_table "github.com/dracory/weebase/pages/page_table"
	"github.com/dracory/weebase/shared/constants"
	"github.com/dracory/weebase/shared/session"
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
	config  Config
	db      *gorm.DB
	drivers map[string]driverConfig
}

type driverConfig struct {
	Name   string
	Driver string
}

// New creates a new App instance with the given configuration
// The configuration should be loaded using LoadConfig() from config.go
func New(cfg Config, options ...func(*Config)) *App {
	// Apply any option functions to the config
	for _, option := range options {
		option(&cfg)
	}

	// Initialize default drivers if none provided
	if len(cfg.Drivers) == 0 {
		cfg.Drivers = []string{MYSQL, POSTGRES, SQLITE, SQLSRV}
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

	switch action {
	// API Handlers
	case "api_connect":
		api_connect.New(g.config.toWebConfig()).ServeHTTP(w, r)

	case "api_tables_list":
		api_tables_list.New(g.config.toWebConfig()).Handle(w, r)

	// Page Handlers
	case "page_home", "page_server":
		// Get session
		sess := session.EnsureSession(w, r, g.config.SessionSecret)
		// Get enabled drivers
		enabledDrivers := g.config.Drivers
		if len(enabledDrivers) == 0 {
			enabledDrivers = []string{"mysql", "postgres", "sqlite", "sqlserver"}
		}

		// Create connection info map
		connInfo := make(map[string]interface{})
		if conn := sess.Conn; conn != nil {
			connInfo["driver"] = conn.Driver
			// Add any additional connection info needed by the home page
			// Note: ActiveConnection only has ID, Driver, and DB fields
		}

		// Call the handler function directly
		html, err := page_home.Handle(
			g.config.BasePath,
			g.config.ActionParam,
			enabledDrivers,
			g.config.AllowAdHocConnections,
			g.config.SafeModeDefault,
			session.GenerateCSRFToken(g.config.SessionSecret),
			connInfo,
		)
		if err != nil {
			http.Error(w, "Failed to render home page: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(html))

	case "page_login":
		page_login.New(g.config.toWebConfig()).ServeHTTP(w, r)

	case "page_logout":
		// Call the logout handler function directly
		html, err := page_logout.Handle(
			nil, // No template
			g.config.BasePath,
			g.config.ActionParam,
			g.config.SafeModeDefault,
			"", // No CSRF token needed for logout
		)
		if err != nil {
			http.Error(w, "Failed to render logout page: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(html))

	case "page_database":
		// Get session
		session.EnsureSession(w, r, g.config.SessionSecret)

		// Call the handler function directly
		html, err := page_database.Handle(
			template.New("database"),
			g.config.BasePath,
			g.config.SafeModeDefault,
			session.GenerateCSRFToken(g.config.SessionSecret),
		)
		if err != nil {
			http.Error(w, "Failed to render database page: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(html))

	case "page_table":
		page_table.New(g.config.toWebConfig()).ServeHTTP(w, r)

	// Default to login page
	default:
		http.Redirect(w, r, g.config.BasePath+"?action=page_login", http.StatusFound)
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
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval' cdn.jsdelivr.net cdn.tailwindcss.com unpkg.com; style-src 'self' 'unsafe-inline' cdn.jsdelivr.net cdn.tailwindcss.com unpkg.com; img-src 'self' data:;")

		next.ServeHTTP(w, r)
	})
}
