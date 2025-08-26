package home

import (
	"bytes"
	"html/template"

	"github.com/dracory/weebase/shared/constants"
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

	// Prepare sidebar hydration script: fetch real tables and render links
	listURL := urls.URL(basePath, constants.ActionListTables, nil)
	browseBase := urls.URL(basePath, constants.ActionBrowseRows, nil)
	hydrate := hb.Script("(function(){\n" +
		"  var el=document.getElementById('wb-objects'); if(!el) return;\n" +
		"  fetch('" + template.JSEscapeString(listURL) + "', {credentials:'same-origin'})\n" +
		"    .then(function(r){return r.json()}).then(function(d){\n" +
		"      if(!d||!d.data||!Array.isArray(d.data.tables)) return;\n" +
		"      el.innerHTML='';\n" +
		"      var base='" + template.JSEscapeString(browseBase) + "';\n" +
		"      d.data.tables.forEach(function(t){\n" +
		"        var li=document.createElement('li');\n" +
		"        var a=document.createElement('a');\n" +
		"        a.textContent='select '+t;\n" +
		"        a.href=base + (base.indexOf('?')>-1?'&':'?') + 'table=' + encodeURIComponent(t);\n" +
		"        a.className='hover:underline';\n" +
		"        li.appendChild(a);\n" +
		"        el.appendChild(li);\n" +
		"      });\n" +
		"    }).catch(function(){});\n" +
		"})();")

	// Wrap with shared layout
	full := layout.RenderWith(layout.Options{
		Title:           "Home",
		BasePath:        basePath,
		SafeModeDefault: safeModeDefault,
		MainHTML:        buf.String(),
		SidebarHTML:     sidebarHTML,
		ExtraBodyEnd:    []hb.TagInterface{hydrate},
	})
	return full, nil
}
