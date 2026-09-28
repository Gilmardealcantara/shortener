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
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
)

func main() {
	// Handle SIGINT (CTRL+C) gracefully.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	cfg := config.New()

	// Set up OpenTelemetry first — slog must be configured after so the
	// OTel logger provider is already registered when the bridge is created.
	otelShutdown, err := setupOTelSDK(ctx, cfg)
	if err != nil {
		panic(err)
	}
	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
	}()

	// Now configure slog to fan out to both stdout and the OTel log pipeline.
	configSlog()

	// Send a startup probe span to verify connectivity with the OTel backend.
	probeOTelConnectivity(ctx, cfg.OTelEndpoint)

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

	handler := otelhttp.NewHandler(mux, "/")
	return handler
}

// configSlog fans out slog records to both a JSON stdout handler and the OTel
// log bridge. Must be called after setupOTelSDK so the logger provider is
// already registered.
func configSlog() {
	slog.SetDefault(otelslog.NewLogger("github.com/Gilmardealcantara/shortener"))
}

// probeOTelConnectivity creates a single span and forces an immediate flush
// to verify the backend is reachable on startup.
func probeOTelConnectivity(ctx context.Context, endpoint string) {
	tracer := otel.Tracer("startup-probe")
	_, span := tracer.Start(ctx, "otel.connectivity.probe")
	span.End()

	type forceFlush interface {
		ForceFlush(context.Context) error
	}
	if ff, ok := otel.GetTracerProvider().(forceFlush); ok {
		if err := ff.ForceFlush(ctx); err != nil {
			slog.ErrorContext(ctx, "OTel connectivity probe failed — traces may not reach the backend",
				"endpoint", endpoint,
				"error", err,
			)
			return
		}
	}
	slog.InfoContext(ctx, "OTel connectivity probe OK", "endpoint", endpoint)
}
