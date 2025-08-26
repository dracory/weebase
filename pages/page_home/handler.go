package page_home

import (
	"embed"
	"html/template"

	"github.com/dracory/weebase/shared"
	layout "github.com/dracory/weebase/shared/layout"
	"github.com/dracory/weebase/shared/urls"
	"github.com/gouniverse/cdn"
	hb "github.com/gouniverse/hb"
)

//go:embed view.html script.js styles.css
var embeddedFS embed.FS

// Handle renders the Home page and returns full HTML.
func Handle(
	basePath, actionParam string,
	enabledDrivers []string,
	allowAdHocConnections bool,
	safeModeDefault bool,
	csrfToken string,
	connInfo map[string]any,
) (
	template.HTML,
	error,
) {
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
		hb.NewTag("li").Child(hb.A().Class("text-slate-700 hover:underline dark:text-slate-200").Href(urls.URL(basePath, "sql_execute", nil)).Text("SQL command")).Attr("title", "Open SQL console"),
		hb.NewTag("li").Child(hb.A().Class("text-slate-700 hover:underline dark:text-slate-200 opacity-60").Href(urls.URL(basePath, "import", nil)).Text("Import")),
		hb.NewTag("li").Child(hb.A().Class("text-slate-700 hover:underline dark:text-slate-200 opacity-60").Href(urls.URL(basePath, "export", nil)).Text("Export")),
		hb.NewTag("li").Child(hb.A().Class("text-slate-700 hover:underline dark:text-slate-200 opacity-60").Href(urls.URL(basePath, "ddl_create_table", nil)).Text("Create table")),
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
	listURL := urls.ListTables(basePath)
	tableViewURL := urls.TableView(basePath)
	sqlURL := urls.SQLExecute(basePath)
	createTableURL := urls.PageTableCreate(basePath)
	// importURL := urls.Import(basePath)
	// exportURL := urls.Export(basePath)
	// For BrowseRows, we'll need a table name which we'll handle in the frontend
	browseBase := urls.BrowseRows(basePath, "")

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
		BasePath:        basePath,
		SafeModeDefault: safeModeDefault,
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
