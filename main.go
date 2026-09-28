package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/Gilmardealcantara/shortener/db"
	"github.com/Gilmardealcantara/shortener/pkg/config"
	"github.com/Gilmardealcantara/shortener/pkg/handlers"
	"github.com/Gilmardealcantara/shortener/pkg/middlewares"
	"github.com/Gilmardealcantara/shortener/pkg/shortner"
	"github.com/Gilmardealcantara/shortener/pkg/tel"
)

func main() {
	// Handle SIGINT (CTRL+C) gracefully.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	cfg := config.New()

	otelShutdown, err := tel.Setup(ctx, cfg)
	if err != nil {
		panic(err)
	}
	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
	}()

	db.InitPostgres(ctx, cfg.PostgresDSN)
	defer db.ClosePostgres(ctx)
	slog.InfoContext(ctx, "Postgres Started!", "dsn", cfg.PostgresDSN)

	db.InitRedis(ctx, cfg.RedisDSN)
	defer db.CloseRedis(ctx)
	slog.InfoContext(ctx, "Redis startded!", "dsn", cfg.RedisDSN)

	err = run(ctx, cfg, stop)
	if err != nil {
		slog.ErrorContext(ctx, "Server Shutdown Error", "err", err)
		return
	}
}

func run(ctx context.Context, cfg *config.Config, stop context.CancelFunc) error {
	srv := http.Server{
		Addr:         ":8080",
		BaseContext:  func(l net.Listener) context.Context { return ctx },
		ReadTimeout:  time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      newServerHandler(cfg),
	}

	srvErr := make(chan error)

	go func() {
		slog.InfoContext(ctx, "Running Server...", "addr", srv.Addr)
		srvErr <- srv.ListenAndServe()
	}()

	select {
	case err := <-srvErr:
		slog.ErrorContext(ctx, "Server Error", "err", err)
		return err
	case <-ctx.Done():
		slog.InfoContext(ctx, "Server Shutdown")
		stop()
	}

	return srv.Shutdown(ctx)
}

func newServerHandler(cfg *config.Config) http.Handler {
	shortnerSrv := shortner.New(cfg)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{code}", handlers.Redirect)
	mux.Handle("POST /shorten", middlewares.HostContext(http.HandlerFunc(handlers.Create(shortnerSrv))))

	return tel.HTTPHandler(mux)
}
