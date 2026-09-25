package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/Gilmardealcantara/shortener/db"
	"github.com/Gilmardealcantara/shortener/pkg/config"
	"github.com/Gilmardealcantara/shortener/pkg/shortner"
)

func main() {
	ctx := context.Background()
	cfg := config.New()

	db.InitPostgres(ctx, cfg.PostgresDSN)
	defer db.ClosePostgres(ctx)

	db.InitRedis(ctx)
	defer db.CloseRedis(ctx)

	mux := getServer(cfg)

	slog.Info("ListenAndServe", "port", "8080")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		slog.Error("ListenAndServe", "error", err)
	}
}

type ShortenerRequest struct {
	LongURL string `json:"long_url"`
}

type ShortenerResponse struct {
	ShortURL string `json:"short_url"`
}

func getServer(cfg *config.Config) *http.ServeMux {
	shortner := shortner.New(cfg)
	mux := http.NewServeMux()

	mux.HandleFunc("GET /{code}", func(w http.ResponseWriter, r *http.Request) {
		code := r.PathValue("code")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(code))
	})

	mux.HandleFunc("POST /shorten", func(w http.ResponseWriter, r *http.Request) {
		var payload ShortenerRequest
		err := json.NewDecoder(r.Body).Decode(&payload)
		if err != nil {
			slog.Error("POST /shorten", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		slog.Info("POST /shorten", "body", payload.LongURL)
		shortURL, err := shortner.Create(payload.LongURL)
		if err != nil {
			slog.Error("POST /shorten", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(&ShortenerResponse{ShortURL: shortURL})
	})

	return mux
}
