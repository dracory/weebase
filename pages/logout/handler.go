package logout

import (
	"bytes"
	"html/template"

	layout "github.com/dracory/weebase/shared/layout"
)

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
	data := map[string]any{
		"Title":       "Logout",
		"BasePath":    basePath,
		"ActionParam": actionParam,
		"CSRFToken":   csrfToken,
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "pages/logout/content", data); err != nil {
		return "", err
	}

	full := layout.Render("Logout", basePath, safeModeDefault, template.HTML(buf.String()), nil, nil)
	return full, nil
}
