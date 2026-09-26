package main

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/Gilmardealcantara/shortener/db"
	"github.com/Gilmardealcantara/shortener/pkg/config"
	"github.com/Gilmardealcantara/shortener/pkg/handlers"
	"github.com/Gilmardealcantara/shortener/pkg/middlewares"
	"github.com/Gilmardealcantara/shortener/pkg/shortner"
)

func main() {
	ctx := context.Background()
	cfg := config.New()

	db.InitPostgres(ctx, cfg.PostgresDSN)
	defer db.ClosePostgres(ctx)

	db.InitRedis(ctx, cfg.RedisDSN)
	defer db.CloseRedis(ctx)

	mux := GetServer(cfg)

	slog.Info("ListenAndServe", "port", "8080")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		slog.Error("ListenAndServe", "error", err)
	}
}

func GetServer(cfg *config.Config) *http.ServeMux {
	shortnerSrv := shortner.New(cfg)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{code}", handlers.Redirect)
	mux.Handle("POST /shorten", middlewares.HostContext(http.HandlerFunc(handlers.Create(shortnerSrv))))

	return mux
}
