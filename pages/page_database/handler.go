package page_database

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
	DefaultTitle = "Database Browser"
	// DefaultViewport is the default viewport meta tag content
	DefaultViewport = "width=device-width, initial-scale=1.0"
)

//go:embed view.html script.js styles.css
var embeddedFS embed.FS

type pageDatabaseController struct {
	config types.Config
}

func New(config types.Config) *pageDatabaseController {
	return &pageDatabaseController{config: config}
}

// ServeHTTP handles the HTTP request for the database browser page
func (c *pageDatabaseController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	html, err := c.pageHtml()
	if err != nil {
		http.Error(w, "Failed to render database page: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

// Handle renders the database browser page and returns the full HTML.
func (c pageDatabaseController) pageHtml() (template.HTML, error) {
	// Ensure base path has a trailing slash
	if c.config.BasePath != "" && c.config.BasePath[len(c.config.BasePath)-1] != '/' {
		c.config.BasePath += "/"
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
		"databases": c.config.BasePath + "api/databases",
		"tables":    c.config.BasePath + "api/tables",
		"logout":    c.config.BasePath + "logout",
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
				csrfToken: "` + template.JSEscapeString(session.GenerateCSRFToken(c.config.SessionSecret)) + `",
				safeMode: ` + func() string {
			if c.config.SafeModeDefault {
				return "true"
			}
			return "false"
		}() + `
			};
		`),
		hb.Script(pageJS), // Our main application script
	}

	// Render the page with layout
	page := layout.RenderWith(layout.Options{
		Title:           DefaultTitle,
		BasePath:        c.config.BasePath,
		SafeModeDefault: c.config.SafeModeDefault,
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
