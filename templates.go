package weebase

import (
	"embed"
	"html/template"
	"io/fs"
)

//go:embed templates/* assets/* pages/**
var embeddedFS embed.FS

func parseTemplates() *template.Template {
    // Parse all .gohtml templates from embedded FS
    t := template.New("").Funcs(template.FuncMap{})
    // Collect template files from both templates/ and pages/
    var files []string
    walkers := []string{"templates", "pages"}
    for _, root := range walkers {
        _ = fs.WalkDir(embeddedFS, root, func(p string, d fs.DirEntry, err error) error {
            if err != nil {
                return err
            }
            if d.IsDir() {
                return nil
            }
            if hasSuffix(p, ".gohtml") || hasSuffix(p, ".html") {
                files = append(files, p)
            }
            return nil
        })
    }
    if len(files) == 0 {
        // Should not happen; but avoid panic during early scaffold
        return template.Must(t.Parse("{{define \"status.gohtml\"}}<pre>Status</pre>{{end}}"))
    }
    return template.Must(t.ParseFS(embeddedFS, files...))
}

// small helper: strings.HasSuffix without importing strings in this file
func hasSuffix(s, suf string) bool {
    if len(s) < len(suf) { return false }
    return s[len(s)-len(suf):] == suf
}
