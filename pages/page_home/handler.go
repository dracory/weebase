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
		urlLogin := urls.Login(h.cfg.BasePath)
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

	// Quick actions links
	quickLinks := hb.UL().
		Class("nav flex-column mb-4").
		Children([]hb.TagInterface{
			hb.LI().Class("nav-item").Child(
				hb.A().Class("nav-link text-dark").
					Href(urlSQLExecute).
					Text("SQL command").
					Attr("title", "Open SQL console"),
			),
			hb.LI().Class("nav-item").Child(
				hb.A().Class("nav-link text-dark").
					Href(urlImport).
					Text("Import"),
			),
			hb.LI().Class("nav-item").Child(
				hb.A().Class("nav-link text-dark").
					Href(urlExport).
					Text("Export"),
			),
			hb.LI().Class("nav-item").Child(
				hb.A().Class("nav-link text-dark").
					Href(urlPageTableCreate).
					Attr("title", "Create table").
					Text("Create table"),
			),
		})

	// Database objects section
	objects := hb.Div().Children([]hb.TagInterface{
		hb.Paragraph().Class("small text-uppercase text-muted mb-2").Text("Objects"),
		hb.NewTag("ul").Attr("id", "wb-objects").Class("nav flex-column").Children([]hb.TagInterface{
			hb.NewTag("li").Class("nav-item").
				Attr("data-placeholder", "1").
				Child(hb.NewTag("span").Class("nav-link text-muted").Text("loading...")),
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
	csrfToken := session.GenerateCSRFToken(h.cfg.SessionSecret)

	// Load page assets
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

	// Build sidebar
	sidebarHTML := h.buildSidebar()

	// Generate URLs using URL builder functions
	listURL := urls.ListTables(h.cfg.BasePath)
	tableViewURL := urls.TableView(h.cfg.BasePath)
	sqlURL := urls.SQLExecute(h.cfg.BasePath)
	_ = urls.PageTableCreate(h.cfg.BasePath) // Will be used in the frontend
	browseBase := urls.BrowseRows(h.cfg.BasePath, "")

	// Build the page using the layout with sidebar
	page := layout.RenderWith(layout.Options{
		Title:           "WeeBase - Home",
		BasePath:        h.cfg.BasePath,
		SafeModeDefault: false,
		MainHTML:        pageHTML,
		SidebarHTML:     sidebarHTML,
		ExtraHead: []hb.TagInterface{
			hb.NewTag("style").Child(hb.Text(pageCSS)),
		},
		ExtraBodyEnd: []hb.TagInterface{
			hb.NewTag("script").Child(hb.Text(pageJS)),
			hb.NewTag("script").Child(hb.Text(`
				window.csrfToken = "` + template.JSEscapeString(csrfToken) + `";
				window.urlListTables = "` + template.JSEscapeString(listURL) + `";
				window.urlTableView = "` + template.JSEscapeString(tableViewURL) + `";
				window.urlSQLExecute = "` + template.JSEscapeString(sqlURL) + `";
				window.urlBrowseBase = "` + template.JSEscapeString(browseBase) + `";
			`)),
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
