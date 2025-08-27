package page_table

import (
	"embed"
	"encoding/json"
	"html/template"
	"net/http"

	"github.com/dracory/weebase/shared"
	layout "github.com/dracory/weebase/shared/layout"
	"github.com/dracory/weebase/shared/session"
	"github.com/dracory/weebase/shared/types"
	"github.com/gouniverse/cdn"
	hb "github.com/gouniverse/hb"
)

const (
	// DefaultTitle is the default page title
	DefaultTitle = "Table Viewer"
	// DefaultViewport is the default viewport meta tag content
	DefaultViewport = "width=device-width, initial-scale=1.0"
)

//go:embed view.html script.js styles.css
var embeddedFS embed.FS

// pageTableController handles HTTP requests for the table page
type pageTableController struct {
	config types.Config
}

// New creates a new pageTableController instance
func New(config types.Config) *pageTableController {
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

	// Ensure session exists (this will set the session cookie if needed)
	session.EnsureSession(w, r, h.config.SessionSecret)

	// Generate CSRF token
	csrfToken := session.GenerateCSRFToken(h.config.SessionSecret)

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

// Handle renders the table viewer page and returns the full HTML.
func Handle(
	tmpl *template.Template,
	basePath string,
	databaseName string,
	tableName string,
	safeModeDefault bool,
	csrfToken string,
) (template.HTML, error) {
	// Ensure base path has a trailing slash
	if basePath != "" && basePath[len(basePath)-1] != '/' {
		basePath += "/"
	}

	// Get embedded assets
	pageCSS, err := css()
	if err != nil {
		return "", err
	}

	pageJS, err := js()
	if err != nil {
		return "", err
	}

	pageHTML, err := view()
	if err != nil {
		return "", err
	}

	// Build API URLs
	apiURLs := map[string]string{
		"tables": basePath + "api/table/" + databaseName + "/" + tableName,
	}

	// Page-specific assets
	extraHead := []hb.TagInterface{
		hb.Style(pageCSS),
		hb.Meta().Attr("name", "viewport").Attr("content", DefaultViewport),
	}

	// JavaScript dependencies and configuration
	extraBody := []hb.TagInterface{
		// Vue 3 CDN
		hb.ScriptURL(cdn.VueJs_3()),
		// Configuration for the frontend
		hb.Script(`
			window.appConfig = {
				api: ` + string(toJSON(apiURLs)) + `,
				csrfToken: "` + template.JSEscapeString(csrfToken) + `",
				safeMode: ` + func() string {
			if safeModeDefault {
				return "true"
			}
			return "false"
		}() + `,
				databaseName: "` + template.JSEscapeString(databaseName) + `",
				tableName: "` + template.JSEscapeString(tableName) + `"
			};
		`),
		hb.Script(pageJS), // Our main application script
	}

	// Set page title
	pageTitle := DefaultTitle
	if tableName != "" {
		pageTitle = tableName + " - " + pageTitle
	}

	// Render the page with layout
	page := layout.RenderWith(layout.Options{
		Title:           pageTitle,
		BasePath:        basePath,
		SafeModeDefault: safeModeDefault,
		MainHTML:        pageHTML,
		ExtraHead:       extraHead,
		ExtraBodyEnd:    extraBody,
	})

	return page, nil
}

// Helper function to convert Go values to JSON for JavaScript
func toJSON(v interface{}) template.JS {
	b, err := json.Marshal(v)
	if err != nil {
		return template.JS("{}")
	}
	return template.JS(b)
}

func css() (string, error) {
	css, err := shared.EmbeddedFileToString(embeddedFS, "styles.css")
	if err != nil {
		return "", err
	}
	return css, nil
}

func js() (string, error) {
	js, err := shared.EmbeddedFileToString(embeddedFS, "script.js")
	if err != nil {
		return "", err
	}
	return js, nil
}

func view() (string, error) {
	html, err := shared.EmbeddedFileToString(embeddedFS, "view.html")
	if err != nil {
		return "", err
	}

	return html, nil
}
