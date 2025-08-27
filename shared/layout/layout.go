package layout

import (
	"html/template"

	hb "github.com/gouniverse/hb"
)

// buildNavbar creates a responsive Bootstrap 5 navbar component
func buildNavbar(basePath string, safeMode bool) *hb.Tag {
	// Navbar toggler button
	toggleButton := hb.NewTag("button").
		Class("navbar-toggler").
		Attr("type", "button").
		Attr("data-bs-toggle", "collapse").
		Attr("data-bs-target", "#navbarNav").
		Attr("aria-controls", "navbarNav").
		Attr("aria-expanded", "false").
		Attr("aria-label", "Toggle navigation").
		Child(hb.NewTag("span").Class("navbar-toggler-icon"))

	// Navbar brand
	brandLink := hb.NewTag("a").
		Class("navbar-brand").
		Attr("href", basePath).
		Child(hb.Text("WeeBase"))

	// Navbar menu items
	navMenuItems := hb.NewTag("ul").Class("navbar-nav me-auto").
		Child(hb.NewTag("li").Class("nav-item").
			Child(hb.NewTag("a").
				Class("nav-link").
				Attr("href", basePath).
				Child(hb.Text("Home")),
			),
		).
		Child(hb.NewTag("li").Class("nav-item").
			Child(hb.NewTag("a").
				Class("nav-link").
				Attr("href", basePath+"?action=healthz").
				Child(hb.Text("Health")),
			),
		).
		Child(hb.NewTag("li").Class("nav-item").
			Child(hb.NewTag("a").
				Class("nav-link").
				Attr("href", basePath+"?action=readyz").
				Child(hb.Text("Ready")),
			),
		)

	// Safe mode indicator
	safeModeText := "OFF"
	safeModeClass := "text-warning"
	if safeMode {
		safeModeText = "ON"
		safeModeClass = "text-success"
	}

	safeModeIndicator := hb.NewTag("div").Class("d-flex").
		Child(hb.NewTag("span").Class("navbar-text me-3").
			Child(hb.NewTag("small").Child(hb.Text("Safe mode: "))).
			Child(hb.NewTag("code").
				Child(hb.Text(safeModeText)).
				Class(safeModeClass),
			),
		)

	// Navbar collapse container
	navbarCollapse := hb.NewTag("div").
		Class("collapse navbar-collapse").
		ID("navbarNav").
		Children([]hb.TagInterface{
			navMenuItems,
			safeModeIndicator,
		})

	// Navbar container
	navbarContainer := hb.NewTag("div").Class("container-fluid").
		Children([]hb.TagInterface{
			brandLink,
			toggleButton,
			navbarCollapse,
		})

	// Return the complete navbar
	return hb.NewTag("nav").
		Class("navbar navbar-expand-lg navbar-dark bg-primary mb-4").
		Child(navbarContainer)
}

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
		hb.NewTag("link").Attr("rel", "icon").Attr("type", "image/png").Attr("href", "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/x8AAwMCAO+ip1sAAAAASUVORK5CYII="),
		// Bootstrap 5 CSS
		hb.NewTag("link").Attr("rel", "stylesheet").Attr("href", "https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css"),
		// Bootstrap Icons
		hb.NewTag("link").Attr("rel", "stylesheet").Attr("href", "https://cdn.jsdelivr.net/npm/bootstrap-icons@1.11.1/font/bootstrap-icons.css"),
		// hb.NewTag("link").Attr("rel", "stylesheet").Attr("href", o.BasePath+"?action=asset_css"),
		// Bootstrap 5 JS Bundle with Popper
		hb.NewTag("script").Attr("src", "https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/js/bootstrap.bundle.min.js"),
	}
	if len(o.ExtraHead) > 0 {
		headChildren = append(headChildren, o.ExtraHead...)
	}

	// Build the navbar
	navbar := buildNavbar(o.BasePath, o.SafeModeDefault)
	header := hb.NewTag("header").Child(navbar)

	// Shell: sidebar + main content
	// Sidebar (optional)
	var sidebar hb.TagInterface
	if o.SidebarHTML != "" {
		sidebar = hb.NewTag("aside").Class("wb-sidebar shrink-0 border-r border-gray-200 bg-gray-50 dark:bg-slate-900 dark:border-slate-800 p-3 w-60").
			Child(hb.Raw(o.SidebarHTML))
	}

	// Main content wraps the provided HTML
	main := hb.NewTag("main").Class("container py-3").Child(hb.Raw(o.MainHTML))

	// Footer
	footer := hb.NewTag("footer").Class("container-fluid py-3 mt-5 border-top").
		Child(hb.NewTag("div").Class("container").
			Child(hb.NewTag("div").Class("text-muted small").
				Child(hb.Text("WeeBase - A simple database management tool")),
			),
		)

	// Body base
	bodyChildren := []hb.TagInterface{
		header,
		hb.NewTag("div").Class("wb-shell flex min-h-[60vh]").Children([]hb.TagInterface{
			sidebar,
			main,
		}),
		footer,
		// hb.NewTag("script").Attr("src", o.BasePath+"?action=asset_js"),
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
