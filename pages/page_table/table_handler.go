package page_table

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/dracory/weebase/shared/session"
	"github.com/dracory/weebase/shared/web"
)

// pageTableController handles HTTP requests for the table page
type pageTableController struct {
	config *web.Config
}

// New creates a new pageTableController instance
func New(config *web.Config) *pageTableController {
	return &pageTableController{
		config: config,
	}
}

// ServeHTTP handles HTTP requests for the table page
func (h *pageTableController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Get database and table names from URL
	dbName := r.URL.Query().Get("db")
	tableName := r.URL.Query().Get("table")

	if dbName == "" || tableName == "" {
		http.Redirect(w, r, h.config.BasePath+"?action=server", http.StatusFound)
		return
	}

	// Get session and ensure CSRF token exists
	sess := session.FromContext(r.Context())
	if sess == nil {
		http.Error(w, "session not found", http.StatusInternalServerError)
		return
	}

	// Get or create CSRF token
	csrfToken := sess.GetString("csrf_token")
	if csrfToken == "" {
		csrfToken = generateRandomString(32)
		sess.Set("csrf_token", csrfToken)
	}

	// Render the page
	html, err := Handle(nil, h.config.BasePath, dbName, tableName, h.config.SafeModeDefault, csrfToken)
	if err != nil {
		http.Error(w, "Failed to render table page: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

// generateRandomString generates a secure random string of the specified length
func generateRandomString(n int) string {
	b := make([]byte, (n+1)/2) // Can be simplified to n/2 after the first line
	if _, err := rand.Read(b); err != nil {
		panic("failed to generate random string: " + err.Error())
	}
	return hex.EncodeToString(b)[:n]
}
