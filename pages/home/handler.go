package home

import (
	"bytes"
	"html/template"

	layout "github.com/dracory/weebase/shared/layout"
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

	// Wrap with shared layout
	full := layout.Render("Home", basePath, safeModeDefault, buf.String(), nil, nil)
	return full, nil
}
