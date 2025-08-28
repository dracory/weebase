package page_home

import (
	"embed"
	"html/template"
	"net/http"

	"github.com/dracory/weebase/shared"
	layout "github.com/dracory/weebase/shared/layout"
	"github.com/dracory/weebase/shared/session"
	"github.com/dracory/weebase/shared/types"
	"github.com/dracory/weebase/shared/urls"
	hb "github.com/gouniverse/hb"
)

//go:embed view.html script.js styles.css
var embeddedFS embed.FS

type pageHomeController struct {
	cfg types.Config
}

// New creates a new page home controller
func New(cfg types.Config) *pageHomeController {
	return &pageHomeController{
		cfg: cfg,
	}
}

// ServeHTTP handles the HTTP request for the Home page
func (h *pageHomeController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sess := session.EnsureSession(w, r, h.cfg.SessionSecret)

	// Get enabled drivers
	enabledDrivers := h.cfg.EnabledDrivers
	if len(enabledDrivers) == 0 {
		enabledDrivers = []string{"mysql", "postgres", "sqlite", "sqlserver"}
	}

	// Check if we have an active connection in the session
	if sess.Conn == nil || sess.Conn.Driver == "" {
		urlLogin := urls.PageLogin(h.cfg.BasePath)
		// No active connection, redirect to login
		http.Redirect(w, r, urlLogin, http.StatusFound)
		return
	}

	// Connection is valid, continue processing the home page

	html, err := h.Handle()
	if err != nil {
		http.Error(w, "Failed to render home page: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

// buildSidebar creates the sidebar HTML with navigation links
func (h *pageHomeController) buildSidebar() string {
	urlSQLExecute := urls.PageSQLExecute(h.cfg.BasePath)
	urlImport := urls.PageImport(h.cfg.BasePath)
	urlExport := urls.PageExport(h.cfg.BasePath)
	urlPageTableCreate := urls.PageTableCreate(h.cfg.BasePath)

	linkSQLExecute := hb.A().Class("nav-link text-dark").Href(urlSQLExecute).Text("SQL command").Attr("title", "Open SQL console")
	linkImport := hb.A().Class("nav-link text-dark").Href(urlImport).Text("Import").Attr("title", "Import data")
	linkExport := hb.A().Class("nav-link text-dark").Href(urlExport).Text("Export").Attr("title", "Export data")
	linkTableCreate := hb.A().Class("nav-link text-dark").Href(urlPageTableCreate).Attr("title", "Create table").Text("Create table")

	// Quick actions links
	quickLinks := hb.UL().
		Class("nav flex-column mb-4").
		Children([]hb.TagInterface{
			hb.LI().Class("nav-item").Child(linkSQLExecute),
			hb.LI().Class("nav-item").Child(linkImport),
			hb.LI().Class("nav-item").Child(linkExport),
			hb.LI().Class("nav-item").Child(linkTableCreate),
		})

	// Database objects section
	objects := hb.Div().Children([]hb.TagInterface{
		hb.Paragraph().Class("small text-uppercase text-muted mb-2").Text("Objects"),
		hb.UL().
			ID("wb-objects").
			Class("nav flex-column").
			Children([]hb.TagInterface{
				hb.LI().
					Class("nav-item").
					Attr("data-placeholder", "1").
					Child(hb.Span().
						Class("nav-link text-muted").
						Text("loading...")),
			}),
	})

	// Combine all sidebar sections
	return hb.Div().Class("p-3").Children([]hb.TagInterface{
		hb.Div().Class("mb-4").Children([]hb.TagInterface{
			hb.Paragraph().Class("small text-uppercase text-muted mb-2").Text("Quick actions"),
			quickLinks,
		}),
		objects,
	}).ToHTML()
}

// Handle renders the Home page and returns full HTML.
func (h *pageHomeController) Handle() (template.HTML, error) {
	// Load page assets
	pageCSS, err := css()
	if err != nil {
		return "", err
	}

	// Generate CSRF token
	csrfToken := session.GenerateCSRFToken(h.cfg.SessionSecret)

	// Generate URLs using URL builder functions
	listURL := urls.ApiTablesList(h.cfg.BasePath)
	tableURL := urls.PageTable(h.cfg.BasePath)
	sqlURL := urls.PageSQLExecute(h.cfg.BasePath)
	browseBase := urls.BrowseRows(h.cfg.BasePath, "")

	// Build the page using the layout with sidebar
	page := layout.RenderWith(layout.Options{
		Title:           "WeeBase - Home",
		BasePath:        h.cfg.BasePath,
		SafeModeDefault: false,
		MainHTML:        "<div id='main-app'></div>",
		SidebarHTML:     "<div id='sidebar-app'></div>",
		ExtraHead: []hb.TagInterface{
			hb.NewTag("style").Child(hb.Text(pageCSS)),
		},
		ExtraBodyEnd: []hb.TagInterface{
			// Global variables for the frontend
			hb.Script(`
				window.csrfToken = '` + csrfToken + `';
				window.urlListTables = '` + listURL + `';
				window.urlTable = '` + tableURL + `';
				window.urlSQLExecute = '` + sqlURL + `';
				window.urlBrowseBase = '` + browseBase + `';
			`),
			// Vue 3 from CDN
			hb.NewTag("script").
				Attr("src", "https://unpkg.com/vue@3.3.4/dist/vue.global.js"),
			// Main app
			hb.Script(`
				document.addEventListener('DOMContentLoaded', function() {
					// Initialize sidebar app
					const { createApp } = Vue;
					import('/page_home/sidebar.js').then(module => {
						createApp(module.default).mount('#sidebar-app');
					});

					// Initialize main app
					import('/page_home/main.js').then(module => {
						createApp(module.default).mount('#main-app');
					});
				});
			`),
		},
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
