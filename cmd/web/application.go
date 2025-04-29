package main

import (
	"bytes"
	"context"
	"html/template"
	"log/slog"
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/obzva/snippetbox/internal/model"
)

const (
	keyMethod = "method"
	keyURI    = "uri"
	// keyTrace  = "trace"
)

// application holds the application-wide dependencies and configuration
// and provides some useful helpers
type application struct {
	logger         *slog.Logger
	snippetModel   *model.SnippetModel
	userModel      *model.UserModel
	templateCache  map[string]*template.Template
	sessionManager *scs.SessionManager
}

func (app *application) serverError(w http.ResponseWriter, r *http.Request, msg string, attrs ...any) {
	method := r.Method
	uri := r.URL.RequestURI()
	// trace := string(debug.Stack())

	// attrs = append(attrs, slog.String(keyMethod, method), slog.String(keyURI, uri), slog.String(keyTrace, trace))
	attrs = append(attrs, slog.String(keyMethod, method), slog.String(keyURI, uri))
	app.logger.Error(msg, attrs...)

	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (app *application) clientError(w http.ResponseWriter, statusCode int) {
	http.Error(w, http.StatusText(statusCode), statusCode)
}

func (app *application) render(w http.ResponseWriter, r *http.Request, statusCode int, page string, data templateData) {
	ts, ok := app.templateCache[page]
	if !ok {
		app.serverError(w, r, "template doesn't exist", slog.String("page", page))
		return
	}

	// buffer for trial render
	// it helps us to catch runtime error
	// if trial render onto this buffer succeed then we redner the content onto http.ResponseWriter
	var b bytes.Buffer

	if err := ts.ExecuteTemplate(&b, "base", data); err != nil {
		app.serverError(w, r, err.Error())
		return
	}

	w.WriteHeader(statusCode)

	if _, err := b.WriteTo(w); err != nil {
		app.logger.Error(err.Error())
	}
}

func (app *application) checkAuthenticated(ctx context.Context) bool {
	authenticated, ok := ctx.Value(ctxKeyAuth).(bool)
	if !ok {
		return false
	}
	return authenticated
}
