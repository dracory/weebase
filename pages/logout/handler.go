package logout

import (
	"embed"
	"html/template"

	"github.com/dracory/weebase/shared"
	layout "github.com/dracory/weebase/shared/layout"
)

//go:embed view.html
var embeddedFS embed.FS

// Handle renders a simple logout confirmation page and returns full HTML.
func Handle(
	tmpl *template.Template,
	basePath, actionParam string,
	safeModeDefault bool,
	csrfToken string,
) (
	template.HTML,
	error,
) {
	// data := map[string]any{
	// 	"Title":       "Logout",
	// 	"BasePath":    basePath,
	// 	"ActionParam": actionParam,
	// 	"CSRFToken":   csrfToken,
	// }

	pageHTML, err := view()
	if err != nil {
		return "", err
	}

	full := layout.Render("Logout", basePath, safeModeDefault, pageHTML, nil, nil)
	return full, nil
}

func view() (string, error) {
	view, err := shared.EmbeddedFileToString(embeddedFS, "view.html")
	if err != nil {
		return "", err
	}
	return view, nil
}
