package middlewares

import (
	"context"
	"fmt"
	"net/http"
)

type contxtKey string

const hostKey contxtKey = "baseUrl"

func HostContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		scheme := "http"
		// Check for native TLS or proxy headers (Nginx, Cloudflare, AWS, etc.)
		if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}

		// Construct the base URL (e.g., "http://localhost:8080" or "https://example.com")
		baseURL := fmt.Sprintf("%s://%s", scheme, r.Host)

		ctx = context.WithValue(ctx, hostKey, baseURL)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetBaseURL(ctx context.Context) (string, bool) {
	baseURL, ok := ctx.Value(hostKey).(string)
	return baseURL, ok
}
