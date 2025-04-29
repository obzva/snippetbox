package main

import (
	"net/http"

	"github.com/justinas/alice"
	"github.com/obzva/snippetbox/ui"
)

func routes(app *application) *http.ServeMux {
	mux := http.NewServeMux()

	// ping
	mux.HandleFunc("GET /ping", ping(app))

	// static
	fs := http.FileServerFS(ui.Files)
	mux.Handle("GET /static/", fs)

	// session manager middleware
	smMW := alice.New(app.sessionManager.LoadAndSave, preventCSRF, authenticate(app))

	// middleware for routes that require user authentication
	reqAuth := smMW.Append(requireAuthentication(app))

	// get
	mux.Handle("GET /{$}", smMW.ThenFunc(getHome(app)))
	mux.Handle("GET /snippet/view/{id}", smMW.ThenFunc(getSnippetView(app)))
	mux.Handle("GET /snippet/create", reqAuth.ThenFunc(getSnippetCreate(app)))
	mux.Handle("GET /user/signup", smMW.ThenFunc(getUserSignup(app)))
	mux.Handle("GET /user/login", smMW.ThenFunc(getUserLogin(app)))

	// post
	mux.Handle("POST /snippet/create", reqAuth.ThenFunc(postSnippetCreate(app)))
	mux.Handle("POST /user/signup", smMW.ThenFunc(postUserSignup(app)))
	mux.Handle("POST /user/login", smMW.ThenFunc(postUserLogin(app)))
	mux.Handle("POST /user/logout", reqAuth.ThenFunc(postUserLogout(app)))

	return mux
}
