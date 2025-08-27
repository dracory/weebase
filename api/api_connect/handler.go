package api_connect

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"

	"github.com/dracory/api"
	"github.com/dracory/weebase/shared/session"
	"github.com/dracory/weebase/shared/types"
)

// apiConnectController handles database connection requests
type apiConnectController struct {
	cfg types.Config
}

// New creates a new connection handler
func New(cfg types.Config) *apiConnectController {
	return &apiConnectController{
		cfg: cfg,
	}
}

// ConnectRequest represents a database connection request
type ConnectRequest struct {
	ProfileID string `json:"profile_id"`
	Driver    string `json:"driver"`
	DSN       string `json:"dsn"`
	Server    string `json:"server"`
	Port      string `json:"port"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Database  string `json:"database"`
}

// ConnectResponse represents the response from a connection attempt
type ConnectResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Driver  string `json:"driver,omitempty"`
}

// Handle handles the HTTP request for connecting to a database
func (h *apiConnectController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s := session.EnsureSession(w, r, h.cfg.SessionSecret)
	if s == nil {
		api.Respond(w, r, api.Error("failed to create or retrieve session"))
		return
	}

	if r.Method != http.MethodPost {
		api.Respond(w, r, api.Error("connect must be POST"))
		return
	}

	if err := r.ParseForm(); err != nil {
		api.Respond(w, r, api.Error("failed to parse form"))
		return
	}

	// Get connection parameters
	req := ConnectRequest{
		Driver:   strings.TrimSpace(r.Form.Get("driver")),
		DSN:      strings.TrimSpace(r.Form.Get("dsn")),
		Server:   strings.TrimSpace(r.Form.Get("server")),
		Port:     strings.TrimSpace(r.Form.Get("port")),
		Username: strings.TrimSpace(r.Form.Get("username")),
		Password: strings.TrimSpace(r.Form.Get("password")),
		Database: strings.TrimSpace(r.Form.Get("database")),
	}

	// Validate driver
	if !slices.Contains(h.cfg.EnabledDrivers, req.Driver) {
		api.Respond(w, r, api.Error("unsupported driver"))
		return
	}

	// Build DSN if not provided
	if req.DSN == "" {
		dsn, err := buildDSNFromFields(req.Driver, req.Server, req.Port, req.Username, req.Password, req.Database)
		if err != nil {
			api.Respond(w, r, api.Error(err.Error()))
			return
		}
		req.DSN = dsn
	}

	// Test the connection
	db, err := openDBWithDSN(req.Driver, req.DSN)
	if err != nil {
		api.Respond(w, r, api.Error(fmt.Sprintf("connection failed: %v", err)))
		return
	}

	// Close the test connection
	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.Close()
	}

	// Store the connection in the session
	s.Conn = &session.ActiveConnection{
		ID:       s.ID,
		Driver:   req.Driver,
		DB:       db,
		LastUsed: time.Now(),
	}

	// Ensure the session cookie is set in the response
	http.SetCookie(w, &http.Cookie{
		Name:     session.SessionCookieName,
		Value:    s.ID,
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
	})

	api.Respond(w, r, api.SuccessWithData("connected", map[string]any{
		"driver": req.Driver,
	}))
}

// openDBWithDSN opens a database connection using the specified driver and DSN
func openDBWithDSN(driver, dsn string) (*gorm.DB, error) {
	switch driver {
	case "postgres", "pg", "postgresql":
		return gorm.Open(postgres.Open(dsn), &gorm.Config{})
	case "mysql", "mariadb":
		return gorm.Open(mysql.Open(dsn), &gorm.Config{})
	case "sqlite", "sqlite3":
		return gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	case "sqlserver", "mssql":
		return gorm.Open(sqlserver.Open(dsn), &gorm.Config{})
	default:
		return nil, fmt.Errorf("unsupported driver: %s", driver)
	}
}

// buildDSNFromFields constructs a DSN from discrete connection fields per driver.
// For sqlite, if a file path is provided, ensures parent directory exists.
func buildDSNFromFields(driver, host, port, user, pass, db string) (string, error) {
	switch strings.ToLower(driver) {
	case "postgres", "pg", "postgresql":
		// Example: host=... user=... password=... dbname=... port=... sslmode=disable
		parts := []string{}
		if host != "" {
			parts = append(parts, "host="+host)
		}
		if user != "" {
			parts = append(parts, "user="+user)
		}
		if pass != "" {
			parts = append(parts, "password="+pass)
		}
		if db != "" {
			parts = append(parts, "dbname="+db)
		}
		if port != "" {
			parts = append(parts, "port="+port)
		}
		parts = append(parts, "sslmode=disable")
		return strings.Join(parts, " "), nil

	case "mysql", "mariadb":
		// Example: user:pass@tcp(host:port)/dbname?parseTime=true
		dsn := ""
		if user != "" {
			dsn += user
			if pass != "" {
				dsn += ":" + pass
			}
			dsn += "@"
		}
		dsn += "tcp(" + host
		if port != "" {
			dsn += ":" + port
		}
		dsn += ")"
		if db != "" {
			dsn += "/" + db
		}
		dsn += "?parseTime=true"
		return dsn, nil

	case "sqlite", "sqlite3":
		// Ensure directory exists for SQLite file
		if db != "" && db != ":memory:" {
			dir := filepath.Dir(db)
			if dir != "." {
				if err := os.MkdirAll(dir, 0755); err != nil {
					return "", fmt.Errorf("failed to create directory for SQLite file: %v", err)
				}
			}
		}
		return db, nil

	case "sqlserver", "mssql":
		// Example: sqlserver://username:password@localhost:1433?database=dbname
		u := &url.URL{
			Scheme: "sqlserver",
			Host:   host,
		}
		if port != "" {
			u.Host += ":" + port
		}
		if user != "" {
			if pass != "" {
				u.User = url.UserPassword(user, pass)
			} else {
				u.User = url.User(user)
			}
		}
		if db != "" {
			q := u.Query()
			q.Set("database", db)
			u.RawQuery = q.Encode()
		}
		return u.String(), nil

	default:
		return "", fmt.Errorf("unsupported driver: %s", driver)
	}
}
