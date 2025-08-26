package page_home

import (
	"html/template"

	"embed"

	"github.com/dracory/weebase/shared"
	"github.com/dracory/weebase/shared/constants"
	layout "github.com/dracory/weebase/shared/layout"
	"github.com/dracory/weebase/shared/urls"
	"github.com/gouniverse/cdn"
	hb "github.com/gouniverse/hb"
)

//go:embed script.js styles.css
var embeddedFS embed.FS

// Handle renders the Home page using the shared index content template and returns full HTML.
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
	// Load page assets (styles + JS)
	pageCSS, _ := shared.EmbeddedFileToString(embeddedFS, "styles.css")
	pageJS, _ := shared.EmbeddedFileToString(embeddedFS, "script.js")

	// Minimal main content; Vue app will enhance center area.
	main := hb.Div().Children([]hb.TagInterface{
		hb.Heading2().Text("Welcome"),
		hb.Paragraph().Text("This is the WeeBase admin UI. Use the sidebar to navigate."),
	}).ToHTML()

	// Build a simple Adminer-like sidebar
	// Top: quick actions
	quickLinks := hb.NewTag("ul").Class("space-y-1 text-sm").Children([]hb.TagInterface{
		hb.NewTag("li").Child(hb.A().Class("text-slate-700 hover:underline dark:text-slate-200").Href(urls.URL(basePath, "sql_execute", nil)).Text("SQL command")).Attr("title", "Open SQL console"),
		hb.NewTag("li").Child(hb.A().Class("text-slate-700 hover:underline dark:text-slate-200 opacity-60").Href(urls.URL(basePath, "import", nil)).Text("Import")),
		hb.NewTag("li").Child(hb.A().Class("text-slate-700 hover:underline dark:text-slate-200 opacity-60").Href(urls.URL(basePath, "export", nil)).Text("Export")),
		hb.NewTag("li").Child(hb.A().Class("text-slate-700 hover:underline dark:text-slate-200 opacity-60").Href(urls.URL(basePath, "ddl_create_table", nil)).Text("Create table")),
	})

	// Objects section (will be hydrated by JS)
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

	// Extra head and body similar to login page
	extraHead := []hb.TagInterface{hb.Style(pageCSS)}

	listURL := urls.URL(basePath, constants.ActionListTables, nil)
	browseBase := urls.URL(basePath, constants.ActionBrowseRows, nil)
	tableViewURL := urls.URL(basePath, constants.ActionPageTableView, nil)
	sqlURL := urls.URL(basePath, constants.ActionSQLExecute, nil)
	createTableURL := urls.URL(basePath, constants.ActionPageTableCreate, nil)
	importURL := urls.URL(basePath, constants.ActionImport, nil)
	exportURL := urls.URL(basePath, constants.ActionExport, nil)

	extraBody := []hb.TagInterface{
		hb.ScriptURL(cdn.VueJs_3()),
		hb.ScriptURL(cdn.Sweetalert2_11()),
		hb.Script(`window.urlListTables = "` + template.JSEscapeString(listURL) + `"`),
		hb.Script(`window.urlBrowseRows = "` + template.JSEscapeString(browseBase) + `"`),
		hb.Script(`window.urlTableView = "` + template.JSEscapeString(tableViewURL) + `"`),
		hb.Script(`window.urlSqlExecute = "` + template.JSEscapeString(sqlURL) + `"`),
		hb.Script(`window.urlCreateTable = "` + template.JSEscapeString(createTableURL) + `"`),
		hb.Script(`window.urlImport = "` + template.JSEscapeString(importURL) + `"`),
		hb.Script(`window.urlExport = "` + template.JSEscapeString(exportURL) + `"`),
		hb.Script(`window.csrfToken = "` + template.JSEscapeString(csrfToken) + `"`),
		hb.Script(pageJS),
	}

	// Wrap with shared layout
	full := layout.RenderWith(layout.Options{
		Title:           "Home",
		BasePath:        basePath,
		SafeModeDefault: safeModeDefault,
		MainHTML:        main,
		SidebarHTML:     sidebarHTML,
		ExtraHead:       extraHead,
		ExtraBodyEnd:    extraBody,
	})
	return full, nil
}
