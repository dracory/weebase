package layout

import (
	"html/template"

	hb "github.com/gouniverse/hb"
)

// Options bundles parameters for rendering the full HTML layout.
type Options struct {
	Title           string
	BasePath        string
	SafeModeDefault bool
	MainHTML        string
	// SidebarHTML, when provided, renders on the left similar to Adminer
	// and Tailwind's "w-64" style width.
	SidebarHTML  string
	ExtraHead    []hb.TagInterface
	ExtraBodyEnd []hb.TagInterface
}

// Render builds the full HTML page using hb and returns it as a safe HTML string.
// - title: page title text
// - basePath: base path for asset links and navigation
// - safeModeDefault: whether safe mode is on (footer indicator)
// - mainHTML: the pre-rendered inner HTML for the main content area
// - extraHead: optional extra <head> tags for page-specific assets
// - extraBodyEnd: optional extra tags right before </body> (e.g., scripts)
func Render(title, basePath string, safeModeDefault bool, mainHTML string, extraHead []hb.TagInterface, extraBodyEnd []hb.TagInterface) template.HTML {
	return RenderWith(Options{
		Title:           title,
		BasePath:        basePath,
		SafeModeDefault: safeModeDefault,
		MainHTML:        mainHTML,
		ExtraHead:       extraHead,
		ExtraBodyEnd:    extraBodyEnd,
	})
}

// RenderWith builds the full HTML using the provided options struct.
func RenderWith(o Options) template.HTML {
	// Head
	headChildren := []hb.TagInterface{
		hb.NewTag("meta").Attr("charset", "utf-8"),
		hb.NewTag("meta").Attr("name", "viewport").Attr("content", "width=device-width, initial-scale=1"),
		hb.NewTag("title").Text(o.Title + " · WeeBase"),
		// Tailwind via CDN for rapid Adminer-like styling
		hb.ScriptURL("https://cdn.tailwindcss.com"),
		hb.StyleURL(o.BasePath + "?action=asset_css"),
	}
	if len(o.ExtraHead) > 0 {
		headChildren = append(headChildren, o.ExtraHead...)
	}

	// Header/nav
	nav := hb.Nav().Class("wb-nav").Children([]hb.TagInterface{
		hb.A().Href(o.BasePath).Text("Home"),
		hb.A().Href(o.BasePath + "?action=healthz").Text("Health"),
		hb.A().Href(o.BasePath + "?action=readyz").Text("Ready"),
	})

	header := hb.Header().
		Class("wb-header").
		Child(
			hb.Div().
				Class("wb-container").
				Children([]hb.TagInterface{
					hb.Heading1().
						Class("wb-title").
						Child(hb.A().Href(o.BasePath).Text("WeeBase")),
					nav,
				}),
		)

	// Shell: sidebar + main content
	// Sidebar (optional)
	var sidebar hb.TagInterface
	if o.SidebarHTML != "" {
		sidebar = hb.Aside().Class("wb-sidebar shrink-0 border-r border-gray-200 bg-gray-50 dark:bg-slate-900 dark:border-slate-800 p-3 w-60").
			Child(hb.Raw(o.SidebarHTML))
	}

	// Main content wraps the provided HTML; container max-width applies to content area only
	main := hb.Main().Class("wb-main grow p-4").
		Child(hb.Div().Class("wb-container").
			Child(hb.Raw(string(o.MainHTML))))

	// Footer
	footer := hb.Footer().Class("wb-footer wb-container").Child(
		hb.NewTag("small").Child(hb.Text("Safe mode: ")).ChildIf(o.SafeModeDefault, hb.Text("ON")).ChildIf(!o.SafeModeDefault, hb.Text("OFF")),
	)

	// Body base
	bodyChildren := []hb.TagInterface{
		header,
		hb.Div().Class("wb-shell flex min-h-[60vh]").Children([]hb.TagInterface{
			sidebar,
			main,
		}),
		footer,
		hb.ScriptURL(o.BasePath + "?action=asset_js"),
	}
	if len(o.ExtraBodyEnd) > 0 {
		bodyChildren = append(bodyChildren, o.ExtraBodyEnd...)
	}

	html := hb.NewTag("html").
		Attr("lang", "en").
		Children([]hb.TagInterface{
			hb.NewTag("head").
				Children(headChildren),
			hb.NewTag("body").
				Children(bodyChildren),
		})

	// Wrap in <!doctype html>
	return template.HTML("<!doctype html>" + html.ToHTML())
}
