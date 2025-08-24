package weebase

import (
	"embed"
	"html/template"
	"io/fs"
	"log"
)

//go:embed templates/* assets/*
var embeddedFS embed.FS

func parseTemplates() *template.Template {
	// Parse all .gohtml templates from embedded FS
	t := template.New("").Funcs(template.FuncMap{})
	// Walk templates directory
	var files []string
	err := fs.WalkDir(embeddedFS, "templates", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if len(path) >= 7 && path[len(path)-7:] == ".gohtml" {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		log.Printf("walk templates: %v", err)
	}
	if len(files) == 0 {
		// Should not happen; but avoid panic during early scaffold
		return template.Must(t.Parse("{{define \"status.gohtml\"}}<pre>Status</pre>{{end}}"))
	}
	return template.Must(t.ParseFS(embeddedFS, files...))
}
