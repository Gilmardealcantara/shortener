package main

import (
	"context"
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
)

func main() {
	// Handle SIGINT (CTRL+C) gracefully.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	cfg := config.New()

	db.InitPostgres(ctx, cfg.PostgresDSN)
	defer db.ClosePostgres(ctx)
	slog.Info("Postgres Started!", "dsn", cfg.PostgresDSN)

	db.InitRedis(ctx, cfg.RedisDSN)
	defer db.CloseRedis(ctx)
	slog.Info("Redis startded!", "dsn", cfg.RedisDSN)

	// start hppt server
	err := run(ctx, cfg, stop)
	if err != nil {
		slog.Error("Server Shutdown Error", "err", err)
		return
	}
}

func run(ctx context.Context, cfg *config.Config, stop context.CancelFunc) error {
	srv := http.Server{
		Addr:         ":8080",
		BaseContext:  func(l net.Listener) context.Context { return ctx },
		ReadTimeout:  time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      newServerServer(cfg),
	}

	srvErr := make(chan error)

	go func() {
		slog.Info("Running Server...", "addr", srv.Addr)
		srvErr <- srv.ListenAndServe()
	}()

	// Wait for shutdown signal
	select {
	case err := <-srvErr:
		// Error when starting the server
		slog.Error("Server Error", "err", err)
		return err
	case <-ctx.Done():
		// Shutdown signal received
		slog.Info("Server Shutdown")
		stop()
	}

	// when sutdown signal received
	err := srv.Shutdown(ctx)
	return err
}

func newServerServer(cfg *config.Config) *http.ServeMux {
	shortnerSrv := shortner.New(cfg)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{code}", handlers.Redirect)
	mux.Handle("POST /shorten", middlewares.HostContext(http.HandlerFunc(handlers.Create(shortnerSrv))))

	return mux
}
