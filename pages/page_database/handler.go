package page_database

import (
	"embed"
	"encoding/json"
	"html/template"

	"github.com/dracory/weebase/shared"
	layout "github.com/dracory/weebase/shared/layout"
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

// Handle renders the database browser page and returns the full HTML.
func Handle(
	tmpl *template.Template,
	basePath string,
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
		"databases": basePath + "api/databases",
		"tables":    basePath + "api/tables",
		"logout":    basePath + "logout",
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
				safeMode: ` + func() string { if safeModeDefault { return "true" }; return "false" }() + `
			};
		`),
		hb.Script(pageJS), // Our main application script
	}

	// Render the page with layout
	page := layout.RenderWith(layout.Options{
		Title:           DefaultTitle,
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
