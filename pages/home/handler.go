package home

import (
	"bytes"
	"html/template"

	layout "github.com/dracory/weebase/shared/layout"
	"github.com/dracory/weebase/shared/urls"
	hb "github.com/gouniverse/hb"
)

// Handle renders the Home page using the shared index content template and returns full HTML.
func Handle(
	tmpl *template.Template,
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
	data := map[string]any{
		"Title":                 "WeeBase",
		"BasePath":              basePath,
		"ActionParam":           actionParam,
		"EnabledDrivers":        enabledDrivers,
		"AllowAdHocConnections": allowAdHocConnections,
		"SafeModeDefault":       safeModeDefault,
		"CSRFToken":             csrfToken,
		"Conn":                  connInfo,
	}

	// Render content-only template into buffer
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "index_content", data); err != nil {
		return "", err
	}

	// Build a simple Adminer-like sidebar
	// Top: quick actions
	quickLinks := hb.NewTag("ul").Class("space-y-1 text-sm").Children([]hb.TagInterface{
		hb.NewTag("li").Child(hb.A().Class("text-slate-700 hover:underline dark:text-slate-200").Href(urls.URL(basePath, "sql_execute", nil)).Text("SQL command")).Attr("title", "Open SQL console"),
		hb.NewTag("li").Child(hb.A().Class("text-slate-700 hover:underline dark:text-slate-200 opacity-60").Href(urls.URL(basePath, "import", nil)).Text("Import")),
		hb.NewTag("li").Child(hb.A().Class("text-slate-700 hover:underline dark:text-slate-200 opacity-60").Href(urls.URL(basePath, "export", nil)).Text("Export")),
		hb.NewTag("li").Child(hb.A().Class("text-slate-700 hover:underline dark:text-slate-200 opacity-60").Href(urls.URL(basePath, "ddl_create_table", nil)).Text("Create table")),
	})

	// Objects section (placeholder; can be hydrated by JS later)
	objects := hb.Div().Children([]hb.TagInterface{
		hb.Paragraph().Class("text-xs uppercase tracking-wide text-slate-500 mb-1").Text("Objects"),
		hb.NewTag("ul").Class("list-disc list-inside text-sm text-slate-700 dark:text-slate-200").Children([]hb.TagInterface{
			hb.NewTag("li").Child(hb.A().Href("#").Text("select albums")).Attr("data-placeholder", "1"),
			hb.NewTag("li").Child(hb.A().Href("#").Text("select interprets")).Attr("data-placeholder", "1"),
			hb.NewTag("li").Child(hb.A().Href("#").Text("select songs")).Attr("data-placeholder", "1"),
		}),
	})

	sidebarHTML := hb.Div().Children([]hb.TagInterface{
		hb.Div().Class("mb-3").Children([]hb.TagInterface{
			hb.Paragraph().Class("text-xs uppercase tracking-wide text-slate-500 mb-1").Text("Quick actions"),
			quickLinks,
		}),
		objects,
	}).ToHTML()

	// Wrap with shared layout
	full := layout.RenderWith(layout.Options{
		Title:           "Home",
		BasePath:        basePath,
		SafeModeDefault: safeModeDefault,
		MainHTML:        buf.String(),
		SidebarHTML:     sidebarHTML,
	})
	return full, nil
}
