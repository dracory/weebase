package weebase

import (
	"net/http"
	"net/url"
	"strings"
)

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
	// Discrete fields (preferred when dsn is omitted by the client)
	host := strings.TrimSpace(r.Form.Get("server"))
	port := strings.TrimSpace(r.Form.Get("port"))
	user := strings.TrimSpace(r.Form.Get("username"))
	pass := strings.TrimSpace(r.Form.Get("password"))
	db := strings.TrimSpace(r.Form.Get("database"))
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
		// Build DSN on the server from discrete fields
		switch strings.ToLower(driver) {
		case "postgres", "pg", "postgresql":
			// Example: host=... user=... password=... dbname=... port=... sslmode=disable
			parts := []string{}
			if host != "" { parts = append(parts, "host="+host) }
			if user != "" { parts = append(parts, "user="+user) }
			if pass != "" { parts = append(parts, "password="+pass) }
			if db != "" { parts = append(parts, "dbname="+db) }
			if port != "" { parts = append(parts, "port="+port) }
			parts = append(parts, "sslmode=disable")
			dsn = strings.Join(parts, " ")
		case "mysql", "mariadb":
			// Example: user:pass@tcp(host:port)/db?parseTime=true
			hostPort := host
			if port != "" { hostPort = host + ":" + port }
			auth := user
			if user != "" || pass != "" { auth = user + ":" + pass }
			dbpart := ""
			if db != "" { dbpart = "/" + db }
			dsn = auth + "@tcp(" + hostPort + ")" + dbpart + "?parseTime=true"
		case "sqlite", "sqlite3":
			// database path or :memory:
			if db == "" { dsn = ":memory:" } else { dsn = db }
		case "sqlserver", "mssql":
			// Example: sqlserver://user:pass@host:port?database=db
			hostPort := host
			if port != "" { hostPort = host + ":" + port }
			u := url.URL{ Scheme: "sqlserver", Host: hostPort }
			if user != "" || pass != "" { u.User = url.UserPassword(user, pass) }
			q := url.Values{}
			if db != "" { q.Set("database", db) }
			u.RawQuery = q.Encode()
			dsn = u.String()
		default:
			WriteError(w, r, "unsupported driver")
			return
		}
	}
	s := EnsureSession(w, r, h.opts.SessionSecret)
	if err := h.tryAutoConnect(s, driver, dsn); err != nil {
		WriteError(w, r, err.Error())
		return
	}
	WriteSuccessWithData(w, r, "connected", map[string]any{"driver": driver})
}
