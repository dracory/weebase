package page_login

import (
	"embed"
	"html/template"
	"net/http"

	"github.com/dracory/weebase/shared"
	"github.com/dracory/weebase/shared/layout"
	"github.com/dracory/weebase/shared/types"
	"github.com/dracory/weebase/shared/urls"
	"github.com/gouniverse/cdn"
	hb "github.com/gouniverse/hb"
)

const (
	// DefaultTitle is the default page title
	DefaultTitle = "Database Manager - Login"
	// DefaultViewport is the default viewport meta tag content
	DefaultViewport = "width=device-width, initial-scale=1.0"
)

//go:embed view.html script.js styles.css
var embeddedFS embed.FS

// Handler handles login page requests
type Handler struct {
	config *types.Config
}

func New(config *types.Config) *Handler {
	return &Handler{
		config: config,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	html, err := h.GenerateHTML()
	if err != nil {
		http.Error(w, "Failed to render login page: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

// GenerateHTML renders the login form page and returns full HTML.
func (h *Handler) GenerateHTML() (
	template.HTML,
	error,
) {
	// Ensure base path has a trailing slash
	basePath := h.config.BasePath
	if basePath != "" && basePath[len(basePath)-1] != '/' {
		basePath += "/"
	}

	// Build action URL for form submission
	actionUrl := urls.Connect(basePath, nil)
	profilesUrl := urls.Profiles(basePath, nil)
	csrfToken := ""

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

	// Page-specific assets
	extraHead := []hb.TagInterface{
		hb.Style(pageCSS),
		hb.Meta().Name("viewport").Attr("content", DefaultViewport),
	}

	// JavaScript dependencies and configuration
	extraBody := []hb.TagInterface{
		// Vue 3 CDN
		hb.ScriptURL(cdn.VueJs_3()),
		// SweetAlert2 for better alerts
		hb.ScriptURL(cdn.Sweetalert2_11()),
		// Configuration for the frontend
		hb.Script(`
			window.appConfig = {
				urls: {
					action: "` + template.JSEscapeString(actionUrl) + `",
					profiles: "` + template.JSEscapeString(profilesUrl) + `",
					redirect: "` + template.JSEscapeString(basePath) + `"
				},
				csrfToken: "` + template.JSEscapeString(csrfToken) + `",
				safeMode: ` + template.JSEscapeString(func() string {
			if h.config.SafeModeDefault {
				return "true"
			}
			return "false"
		}()) + `
			};
		`),
		hb.Script(pageJS), // Our main application script
	}

	// Render the page with layout
	page := layout.RenderWith(layout.Options{
		Title:           DefaultTitle,
		BasePath:        h.config.BasePath,
		SafeModeDefault: h.config.SafeModeDefault,
		MainHTML:        pageHTML,
		ExtraHead:       extraHead,
		ExtraBodyEnd:    extraBody,
	})

	return page, nil
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
