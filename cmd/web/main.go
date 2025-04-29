package main

import (
	"context"
	"crypto/tls"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/pgxstore"
	"github.com/alexedwards/scs/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/justinas/alice"
	"github.com/lmittmann/tint"
	"github.com/obzva/snippetbox/internal/model"
)

const (
	snippetboxPort = "SNIPPETBOX_PORT"
	databaseURI    = "DATABASE_URI"
)

func main() {
	// create a context that will be used throughout the application
	ctx := context.Background()

	// initialize a logger
	logHandler := tint.NewHandler(os.Stdout, &tint.Options{AddSource: true})
	logger := slog.New(logHandler)

	// get env variables
	dbURI := os.Getenv(databaseURI)
	if dbURI == "" {
		logger.Error("can't find env variable", slog.String("env-var", snippetboxPort))
		os.Exit(1)
	}
	port := os.Getenv(snippetboxPort)
	if port == "" {
		logger.Error("can't find env variable", slog.String("env-var", snippetboxPort))
		os.Exit(1)
	}

	// initialize db connection pool
	dbPool, err := pgxpool.New(ctx, dbURI)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	if err := dbPool.Ping(ctx); err != nil {
		dbPool.Close()
		logger.Error(err.Error())
		os.Exit(1)
	}
	defer dbPool.Close()
	logger.Info("initialized db connection pool", slog.String("URI", dbURI))

	// initialize template cache
	tc, err := newTemplateCache()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	// initialize session manager
	sm := scs.New()
	sm.Store = pgxstore.New(dbPool)
	sm.Lifetime = 12 * time.Hour

	// initialize app struct
	app := &application{
		logger: logger,
		snippetModel: &model.SnippetModel{
			DBPool: dbPool,
		},
		userModel: &model.UserModel{
			DBPool: dbPool,
		},
		templateCache:  tc,
		sessionManager: sm,
	}

	// chaining generalMW and mux
	generalMW := alice.New(recoverPanic(app), logRequest(app), setCommonHeaders)
	mux := routes(app)
	chain := generalMW.Then(mux)

	// start web server
	srv := &http.Server{
		Addr:     port,
		Handler:  chain,
		ErrorLog: slog.NewLogLogger(logHandler, slog.LevelError),
		TLSConfig: &tls.Config{
			CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
		},
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  1 * time.Minute,
	}
	logger.Info("started server", slog.String("port", port))
	if err := srv.ListenAndServeTLS("./tls/localhost.pem", "./tls/localhost-key.pem"); err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
