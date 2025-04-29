package main

import (
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"
	"time"

	"github.com/justinas/nosurf"
	"github.com/obzva/snippetbox/internal/model"
	"github.com/obzva/snippetbox/ui"
)

type templateData struct {
	CurrentYear    int
	Snippet        model.Snippet
	Snippets       []model.Snippet
	Form           any
	FieldErrors    map[string]error
	NonFieldErrors []error
	Flash          string
	Authenticated  bool
	CSRFToken      string
}

func newTemplateData(app *application, r *http.Request) templateData {
	td := templateData{
		CurrentYear:   time.Now().Year(),
		Flash:         app.sessionManager.PopString(r.Context(), sessionKeyFlash),
		Authenticated: app.checkAuthenticated(r.Context()),
		CSRFToken:     nosurf.Token(r),
	}
	return td
}

func newTemplateCache() (map[string]*template.Template, error) {
	cache := make(map[string]*template.Template)

	pages, err := fs.Glob(ui.Files, "html/page/*.tmpl")
	if err != nil {
		return nil, err
	}

	customFuncs := template.FuncMap{
		"prettifyDate": prettifyDate,
	}

	for _, page := range pages {
		name := filepath.Base(page)

		patterns := []string{
			"html/base.tmpl",
			"html/partial/*.tmpl",
			page,
		}

		ts, err := template.New(name).Funcs(customFuncs).ParseFS(ui.Files, patterns...)
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}

	return cache, nil
}

func prettifyDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}

	return t.UTC().Format("02 Jan 2006 at 15:04")
}
