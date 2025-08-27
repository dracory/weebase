package page_logout

import (
	"embed"
	"html/template"
	"net/http"

	"github.com/dracory/weebase/shared"
	layout "github.com/dracory/weebase/shared/layout"
	"github.com/dracory/weebase/shared/session"
	"github.com/dracory/weebase/shared/types"
	"github.com/dracory/weebase/shared/urls"
)

//go:embed view.html
var embeddedFS embed.FS

type pageLogoutController struct {
	config types.Config
}

func New(config types.Config) *pageLogoutController {
	return &pageLogoutController{config: config}
}

func (c *pageLogoutController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Clear the session cookie
	session.DeleteSession(w, r)

	// Redirect to login page
	http.Redirect(w, r, urls.Login(c.config.BasePath), http.StatusFound)
}

// Handle renders a simple logout confirmation page and returns full HTML.
func (c *pageLogoutController) pageHTML() (template.HTML, error) {
	pageHTML, err := view()
	if err != nil {
		return "", err
	}

	full := layout.RenderWith(layout.Options{
		Title:           "Logout",
		BasePath:        c.config.BasePath,
		SafeModeDefault: c.config.SafeModeDefault,
		MainHTML:        pageHTML,
		ExtraHead:       nil,
		ExtraBodyEnd:    nil,
	})
	return full, nil
}

func view() (string, error) {
	view, err := shared.EmbeddedFileToString(embeddedFS, "view.html")
	if err != nil {
		return "", err
	}
	return view, nil
}
