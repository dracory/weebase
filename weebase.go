// Package weebase provides a lightweight database management module for Go web applications.
// It supports MySQL, PostgreSQL, SQLite, and SQL Server databases through a simple HTTP interface.
package weebase

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/dracory/weebase/api/api_connect"
	"github.com/dracory/weebase/pages/page_database"
	"github.com/dracory/weebase/pages/page_home"
	"github.com/dracory/weebase/pages/page_login"
	"github.com/dracory/weebase/pages/page_logout"
	page_table "github.com/dracory/weebase/pages/page_table"
	"github.com/dracory/weebase/shared/constants"
	"github.com/dracory/weebase/shared/session"
	"github.com/dracory/weebase/shared/web"
	"gorm.io/gorm"
)

// driverRegistry implements the driver.Registry interface
type driverRegistry struct {
	enabled map[string]bool
}

func (d *driverRegistry) IsEnabled(name string) bool {
	return d.enabled[name]
}

func newDriverRegistry(enabledDrivers []string) *driverRegistry {
	dr := &driverRegistry{
		enabled: make(map[string]bool),
	}
	for _, driver := range enabledDrivers {
		dr.enabled[driver] = true
	}
	return dr
}

// Supported database drivers
const (
	MYSQL    = constants.DriverMySQL
	POSTGRES = constants.DriverPostgres
	SQLITE   = constants.DriverSQLite
	SQLSRV   = constants.DriverSQLServer
)

// Weebase represents the main application instance
type Weebase struct {
	config  Config
	db      *gorm.DB
	drivers map[string]driverConfig
}

type driverConfig struct {
	Name   string
	Driver string
}

// New creates a new Weebase instance with the given configuration
// The configuration should be loaded using LoadConfig() from config.go
func New(cfg Config, options ...func(*Config)) *Weebase {
	// Apply any option functions to the config
	for _, option := range options {
		option(&cfg)
	}

	// Initialize default drivers if none provided
	if len(cfg.Drivers) == 0 {
		cfg.Drivers = []string{MYSQL, POSTGRES, SQLITE, SQLSRV}
	}

	return &Weebase{
		config:  cfg,
		drivers: make(map[string]driverConfig),
	}
}

// Handler returns an http.Handler that serves the Weebase UI and API
func (g *Weebase) Handler() http.Handler {
	mux := http.NewServeMux()

	// Register API handlers
	mux.HandleFunc(g.config.BasePath, g.handleRequest)

	return g.middleware(mux)
}

// handleRequest routes requests to the appropriate handler
func (g *Weebase) handleRequest(w http.ResponseWriter, r *http.Request) {
	action := r.URL.Query().Get(g.config.ActionParam)

	switch action {
	// API Handlers
	case "api_connect":
		api_connect.New(g.config.toWebConfig()).ServeHTTP(w, r)

	case "api_tables_list":
		// Get session
		session.EnsureSession(w, r, g.config.SessionSecret)
		// TODO: Implement api_tables_list handler
		http.Error(w, "Not implemented", http.StatusNotImplemented)

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
func (g *Weebase) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "same-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval' cdn.jsdelivr.net; style-src 'self' 'unsafe-inline' cdn.jsdelivr.net; img-src 'self' data:;")

		next.ServeHTTP(w, r)
	})
}

// serveServerPage renders the server/database list page
func (g *Weebase) serveServerPage(w http.ResponseWriter, r *http.Request) {
	// Get the session
	sess := session.EnsureSession(w, r, g.config.SessionSecret)

	// Check if user is authenticated (has an active connection)
	if sess.Conn == nil {
		http.Redirect(w, r, g.config.BasePath+"?action=page_login", http.StatusFound)
		return
	}

	// Get enabled drivers (default to all if not specified)
	enabledDrivers := g.config.Drivers
	if len(enabledDrivers) == 0 {
		enabledDrivers = []string{MYSQL, POSTGRES, SQLITE, SQLSRV}
	}

	// Prepare template data
	data := map[string]interface{}{
		"BasePath":        g.config.BasePath,
		"Title":           "Database Server",
		"CSRFToken":       session.GenerateCSRFToken(g.config.SessionSecret),
		"SafeMode":        g.config.SafeModeDefault,
		"AllowAdHoc":      g.config.AllowAdHocConnections,
		"EnabledDrivers":  enabledDrivers,
		"CurrentDatabase": "", // Will be set if viewing a specific database
	}

	// Execute the template
	err := g.renderTemplate(w, "server", data)
	if err != nil {
		http.Error(w, "Failed to render server page: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// renderTemplate is a helper function to render templates with common data
func (g *Weebase) renderTemplate(w http.ResponseWriter, name string, data map[string]interface{}) error {
	// Add common template data
	if data == nil {
		data = make(map[string]interface{})
	}

	// Set default values if not provided
	if _, ok := data["BasePath"]; !ok {
		data["BasePath"] = g.config.BasePath
	}

	// TODO: Add more common template data as needed

	// Execute the template
	tmpl, err := template.ParseFiles(fmt.Sprintf("templates/%s.html", name))
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	return tmpl.Execute(w, data)
}

// serveDatabasePage renders the database view page
func (g *Weebase) serveDatabasePage(w http.ResponseWriter, r *http.Request) {
	// Get the session
	sess := session.EnsureSession(w, r, g.config.SessionSecret)

	// Check if user is authenticated
	if sess.Conn == nil {
		http.Redirect(w, r, g.config.BasePath+"?action=page_login", http.StatusFound)
		return
	}

	// Get database name from URL
	dbName := r.URL.Query().Get("db")
	if dbName == "" {
		http.Redirect(w, r, g.config.BasePath+"?action=page_server", http.StatusFound)
		return
	}

	// Get enabled drivers (default to all if not specified)
	enabledDrivers := g.config.Drivers
	if len(enabledDrivers) == 0 {
		enabledDrivers = []string{MYSQL, POSTGRES, SQLITE, SQLSRV}
	}

	// Get tables in the database
	db, ok := sess.Conn.DB.(*gorm.DB)
	if !ok {
		http.Error(w, "Invalid database connection type", http.StatusInternalServerError)
		return
	}

	tables, err := g.getTables(db, dbName)
	if err != nil {
		http.Error(w, "Failed to get tables: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Prepare template data
	data := map[string]interface{}{
		"BasePath":        g.config.BasePath,
		"Title":           "Database: " + dbName,
		"CSRFToken":       session.GenerateCSRFToken(g.config.SessionSecret),
		"SafeMode":        g.config.SafeModeDefault,
		"AllowAdHoc":      g.config.AllowAdHocConnections,
		"EnabledDrivers":  enabledDrivers,
		"CurrentDatabase": dbName,
		"Tables":          tables,
	}

	// Execute the template
	err = g.renderTemplate(w, "database", data)
	if err != nil {
		http.Error(w, "Failed to render database page: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// serveTablePage renders the table view page using the page_table package
func (g *Weebase) serveTablePage(w http.ResponseWriter, r *http.Request) {
	// Create a web config from the handler options
	webConfig := &web.Config{
		BasePath:        g.config.BasePath,
		ActionParam:     "action",
		SafeModeDefault: g.config.SafeModeDefault,
		SessionSecret:   g.config.SessionSecret,
	}

	// Create and use the table handler
	handler := page_table.New(webConfig)
	handler.ServeHTTP(w, r)
}

// getTables retrieves the list of tables in a database
func (g *Weebase) getTables(db *gorm.DB, dbName string) ([]string, error) {
	var tables []string
	// Implementation depends on the database driver
	// This is a simplified example - you'll need to adjust based on your database
	result := db.Raw("SHOW TABLES").Scan(&tables)
	if result.Error != nil {
		return nil, result.Error
	}
	return tables, nil
}

// getTableInfo retrieves the structure of a table
func (g *Weebase) getTableInfo(db *gorm.DB, dbName, tableName string) (interface{}, error) {
	// Implementation depends on the database driver
	// This is a simplified example
	var columns []map[string]interface{}
	result := db.Raw(fmt.Sprintf("SHOW COLUMNS FROM %s.%s", dbName, tableName)).Scan(&columns)
	if result.Error != nil {
		return nil, result.Error
	}
	return columns, nil
}

// getTableDataFromDB retrieves paginated data from a table
func (g *Weebase) getTableDataFromDB(db *gorm.DB, dbName, tableName string, page, pageSize int) ([]map[string]interface{}, int64, error) {
	var rows []map[string]interface{}
	var count int64

	// Get total count
	countResult := db.Table(fmt.Sprintf("%s.%s", dbName, tableName)).Count(&count)
	if countResult.Error != nil {
		return nil, 0, countResult.Error
	}

	// Get paginated data
	offset := (page - 1) * pageSize
	result := db.Table(fmt.Sprintf("%s.%s", dbName, tableName)).
		Offset(offset).
		Limit(pageSize).
		Find(&rows)

	if result.Error != nil {
		return nil, 0, result.Error
	}

	return rows, count, nil
}

// getPaginationParams extracts pagination parameters from the request
func getPaginationParams(r *http.Request) (int, int) {
	page := 1
	pageSize := 50 // Default page size

	if p := r.URL.Query().Get("page"); p != "" {
		if pNum, err := strconv.Atoi(p); err == nil && pNum > 0 {
			page = pNum
		}
	}

	if ps := r.URL.Query().Get("pageSize"); ps != "" {
		if psNum, err := strconv.Atoi(ps); err == nil && psNum > 0 && psNum <= 1000 {
			pageSize = psNum
		}
	}

	return page, pageSize
}

// Option functions for configuring Weebase are defined in options.go
