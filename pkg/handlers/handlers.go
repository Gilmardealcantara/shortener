package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/Gilmardealcantara/shortener/db"
	"github.com/Gilmardealcantara/shortener/pkg/shortner"
)

type ShortenerRequest struct {
	LongURL string `json:"long_url"`
}

type ShortenerResponse struct {
	ShortURL string `json:"short_url"`
}

func Redirect(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	code := r.PathValue("code")
	longURL, err := shortner.Retrieve(r.Context(), code)
	if err != nil {
		slog.ErrorContext(ctx, "GET /{code}", "error", err)
		if err == db.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Location", longURL)
	w.WriteHeader(http.StatusFound) // 302
}

func Create(shortnerSrv *shortner.Shortener) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var payload ShortenerRequest
		err := json.NewDecoder(r.Body).Decode(&payload)
		if err != nil {
			slog.ErrorContext(ctx, "POST /shorten", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		slog.InfoContext(ctx, "POST /shorten", "body", payload.LongURL, "host", r.Host)
		shortURL, err := shortnerSrv.Create(r.Context(), payload.LongURL)
		if err != nil {
			slog.ErrorContext(ctx, "POST /shorten", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(&ShortenerResponse{ShortURL: shortURL})
	}
}
