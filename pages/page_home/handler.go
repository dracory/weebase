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
	"github.com/gouniverse/cdn"
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

	// Get session
	// Get enabled drivers
	enabledDrivers := h.cfg.EnabledDrivers
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

	// Check if we have an active connection
	if connInfo["driver"] == "" {
		urlLogin := urls.Login(h.cfg.BasePath)
		// No active connection, redirect to login
		http.Redirect(w, r, urlLogin, http.StatusFound)
		return
	}

	html, err := h.Handle()
	if err != nil {
		http.Error(w, "Failed to render home page: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
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

	mainHTML := template.HTML(pageHTML)

	// Build sidebar
	quickLinks := hb.NewTag("ul").Class("space-y-1 text-sm").Children([]hb.TagInterface{
		hb.NewTag("li").Child(hb.A().Class("text-slate-700 hover:underline dark:text-slate-200").Href(urls.URL(h.cfg.BasePath, "sql_execute", nil)).Text("SQL command")).Attr("title", "Open SQL console"),
		hb.NewTag("li").Child(hb.A().Class("text-slate-700 hover:underline dark:text-slate-200 opacity-60").Href(urls.URL(h.cfg.BasePath, "import", nil)).Text("Import")),
		hb.NewTag("li").Child(hb.A().Class("text-slate-700 hover:underline dark:text-slate-200 opacity-60").Href(urls.URL(h.cfg.BasePath, "export", nil)).Text("Export")),
		hb.NewTag("li").Child(hb.A().Class("text-slate-700 hover:underline dark:text-slate-200 opacity-60").Href(urls.URL(h.cfg.BasePath, "ddl_create_table", nil)).Text("Create table")),
	})

	objects := hb.Div().Children([]hb.TagInterface{
		hb.Paragraph().Class("text-xs uppercase tracking-wide text-slate-500 mb-1").Text("Objects"),
		hb.NewTag("ul").Attr("id", "wb-objects").Class("list-disc list-inside text-sm text-slate-700 dark:text-slate-200").Children([]hb.TagInterface{
			hb.NewTag("li").Attr("data-placeholder", "1").Child(hb.Text("loading...")),
		}),
	})

	sidebarHTML := hb.Div().Children([]hb.TagInterface{
		hb.Div().Class("mb-3").Children([]hb.TagInterface{
			hb.Paragraph().Class("text-xs uppercase tracking-wide text-slate-500 mb-1").Text("Quick actions"),
			quickLinks,
		}),
		objects,
	}).ToHTML()

	// Generate URLs using URL builder functions
	listURL := urls.ListTables(h.cfg.BasePath)
	tableViewURL := urls.TableView(h.cfg.BasePath)
	sqlURL := urls.SQLExecute(h.cfg.BasePath)
	createTableURL := urls.PageTableCreate(h.cfg.BasePath)
	// importURL := urls.Import(basePath)
	// exportURL := urls.Export(basePath)
	// For BrowseRows, we'll need a table name which we'll handle in the frontend
	browseBase := urls.BrowseRows(h.cfg.BasePath, "")

	extraHead := []hb.TagInterface{
		hb.Style(pageCSS),
	}

	extraBody := []hb.TagInterface{
		hb.ScriptURL(cdn.VueJs_3()),
		hb.ScriptURL(cdn.Sweetalert2_11()),
		hb.Script(`window.urlListTables = "` + template.JSEscapeString(listURL) + `"`),
		hb.Script(`window.urlBrowseRows = "` + template.JSEscapeString(browseBase) + `"`),
		hb.Script(`window.urlTableView = "` + template.JSEscapeString(tableViewURL) + `"`),
		hb.Script(`window.urlSqlExecute = "` + template.JSEscapeString(sqlURL) + `"`),
		hb.Script(`window.urlCreateTable = "` + template.JSEscapeString(createTableURL) + `"`),
		// hb.Script(`window.urlImport = "` + template.JSEscapeString(importURL) + `"`),
		// hb.Script(`window.urlExport = "` + template.JSEscapeString(exportURL) + `"`),
		hb.Script(`window.csrfToken = "` + template.JSEscapeString(csrfToken) + `"`),
		hb.Script(pageJS),
	}

	// Wrap with shared layout
	full := layout.RenderWith(layout.Options{
		Title:           "Home",
		BasePath:        h.cfg.BasePath,
		SafeModeDefault: h.cfg.SafeModeDefault,
		MainHTML:        string(mainHTML),
		SidebarHTML:     sidebarHTML,
		ExtraHead:       extraHead,
		ExtraBodyEnd:    extraBody,
	})
	return full, nil
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
