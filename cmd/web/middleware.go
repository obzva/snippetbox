package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/justinas/nosurf"
)

func setCommonHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' fonts.googleapis.com; font-src fonts.gstatic.com")
		w.Header().Set("Referrer-Policy", "origin-when-cross-origin")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "deny")
		w.Header().Set("X-XSS-Protection", "0")

		next.ServeHTTP(w, r)
	})
}

func logRequest(app *application) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			app.logger.Info("recieved a request", slog.String("ip", r.RemoteAddr), slog.String("protocol version", r.Proto), slog.String("method", r.Method), slog.String("uri", r.URL.RequestURI()))

			next.ServeHTTP(w, r)
		})
	}
}

func recoverPanic(app *application) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					w.Header().Set("Connection", "close")
					app.serverError(w, r, "recovered panic", slog.String("panic", fmt.Sprintf("%v", err)))
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

func requireAuthentication(app *application) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !app.checkAuthenticated(r.Context()) {
				http.Redirect(w, r, "/user/login", http.StatusSeeOther)
				return
			}

			w.Header().Set("Cache-Control", "no-store")

			next.ServeHTTP(w, r)
		})
	}
}

func preventCSRF(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)
	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   true,
	})
	return csrfHandler
}

func authenticate(app *application) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := app.sessionManager.GetInt(r.Context(), sessionKeyAuth)
			if id == 0 { // no "authenticatedUserID" value is in the current session
				next.ServeHTTP(w, r)
				return
			}

			ok, err := app.userModel.Check(r.Context(), id)
			if err != nil {
				app.serverError(w, r, err.Error())
				return
			}
			if ok { // the current user is valid and authenticated
				ctx := r.Context()
				ctx = context.WithValue(ctx, ctxKeyAuth, true)
				r = r.WithContext(ctx)
			}

			next.ServeHTTP(w, r)
		})
	}
}
