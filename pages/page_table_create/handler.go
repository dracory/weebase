package page_table_create

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

// Handle renders the Create Table page following the pages/login pattern and returns full HTML.
func Handle(basePath, actionParam, csrfToken string, safeModeDefault bool) (template.HTML, error) {
	pageCSS, err := shared.EmbeddedFileToString(embeddedFS, "styles.css")
	if err != nil {
		return "", err
	}
	pageJS, err := shared.EmbeddedFileToString(embeddedFS, "script.js")
	if err != nil {
		return "", err
	}
	pageHTML, err := shared.EmbeddedFileToString(embeddedFS, "view.html")
	if err != nil {
		return "", err
	}

	extraHead := []hb.TagInterface{
		hb.Style(pageCSS),
	}
	// Compute URLs: page renders here, POST goes to API action
	actionUrl := urls.ApiTableCreate(basePath, nil)
	redirectUrl := urls.Home(basePath, nil)

	extraBody := []hb.TagInterface{
		// Optional SweetAlert for nicer errors
		hb.ScriptURL(cdn.VueJs_3()),
		hb.ScriptURL(cdn.Sweetalert2_11()),
		hb.Script(`window.urlAction = "` + template.JSEscapeString(actionUrl) + `"`),
		hb.Script(`window.csrfToken = "` + template.JSEscapeString(csrfToken) + `"`),
		hb.Script(`window.urlRedirect = "` + template.JSEscapeString(redirectUrl) + `"`),
		hb.Script(pageJS),
	}

	full := layout.RenderWith(layout.Options{
		Title:           "Create table",
		BasePath:        basePath,
		SafeModeDefault: safeModeDefault,
		MainHTML:        pageHTML,
		ExtraHead:       extraHead,
		ExtraBodyEnd:    extraBody,
	})
	return full, nil
}
