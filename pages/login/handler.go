package login

import (
	"bytes"
	"html/template"

	layout "github.com/dracory/weebase/shared/layout"
	hb "github.com/gouniverse/hb"
)

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
	data := map[string]any{
		"Title":                 "Login",
		"BasePath":              basePath,
		"ActionParam":           actionParam,
		"EnabledDrivers":        enabledDrivers,
		"AllowAdHocConnections": allowAdHocConnections,
		"SafeModeDefault":       safeModeDefault,
		"CSRFToken":             csrfToken,
	}

	pageHTML, err := renderLoginContent(tmpl, data)
	if err != nil {
		return "", err
	}

	// Page-specific assets
	extraHead := []hb.TagInterface{
		hb.StyleURL(basePath + "?" + actionParam + "=login_css"),
	}

	extraBody := []hb.TagInterface{
		// Vue 3 CDN
		hb.ScriptURL("https://unpkg.com/vue@3/dist/vue.global.prod.js"),

		// Page script
		hb.ScriptURL(basePath + "?" + actionParam + "=login_js"),
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

// renderLoginContent renders the login page inner content template and returns safe HTML.
func renderLoginContent(tmpl *template.Template, data map[string]any) (template.HTML, error) {
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "pages/login/content", data); err != nil {
		return "", err
	}
	return template.HTML(buf.String()), nil
}
