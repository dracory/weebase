package login

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

// Handle renders the Adminer-style login/connect form page and returns full HTML.
func Handle(
	tmpl *template.Template,
	basePath, actionParam string,
	enabledDrivers []string,
	allowAdHocConnections bool,
	safeModeDefault bool,
	csrfToken string,
) (
	template.HTML,
	error,
) {
	// ActionConnect is handled as its own action; build direct connect URL
	actionUrl := urls.Connect(basePath, nil)

	// Precompute profiles URL using urls.Profiles helper
	profilesUrl := urls.Profiles(basePath, nil)

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
	}

	extraBody := []hb.TagInterface{
		// Vue 3 CDN
		hb.ScriptURL(cdn.VueJs_3()),
		hb.ScriptURL(cdn.Sweetalert2_11()),
		hb.Script(`window.urlAction = "` + template.JSEscapeString(actionUrl) + `"`),
		hb.Script(`window.urlProfiles = "` + template.JSEscapeString(profilesUrl) + `"`),
		hb.Script(`window.urlRedirect = "` + template.JSEscapeString(basePath) + `"`),
		hb.Script(`window.csrfToken = "` + template.JSEscapeString(csrfToken) + `"`),
		hb.Script(pageJS),
	}

	full := layout.RenderWith(layout.Options{
		Title:           "Login",
		BasePath:        basePath,
		SafeModeDefault: safeModeDefault,
		MainHTML:        pageHTML,
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
