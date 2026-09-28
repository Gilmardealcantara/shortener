package shortner

import (
	"context"
	"log/slog"

	"github.com/Gilmardealcantara/shortener/db"
	"github.com/Gilmardealcantara/shortener/pkg/config"
	"github.com/Gilmardealcantara/shortener/pkg/middlewares"
	"github.com/mattheath/base62"
)

type Shortener struct {
	config *config.Config
}

func New(config *config.Config) *Shortener {
	return &Shortener{
		config: config,
	}
}

func (s Shortener) Create(ctx context.Context, longURL string) (string, error) {
	counter, err := generateUniqueID(ctx)
	if err != nil {
		return "", err
	}
	code := base62.EncodeInt64(counter)
	slog.InfoContext(ctx, "Creating for code: "+code, "counter", counter, "code", code)
	_, err = db.PG.Exec(ctx, "INSERT INTO shortener (code, long_url) VALUES ($1, $2)", code, longURL)
	if err != nil {
		return "", err
	}

	baseURL, ok := middlewares.GetBaseURL(ctx)
	if !ok {
		return s.config.BaseURL + "/" + code, nil
	}
	url := baseURL + "/" + code
	slog.InfoContext(ctx, "Create url:"+url, "baseUrl", baseURL, "code", code)
	return url, nil
}

// Every single call to this function across 10 different services
// is guaranteed to return a completely unique number.
func generateUniqueID(ctx context.Context) (int64, error) {
	// If 3 services call this at the exact same millisecond:
	// Service 1 gets 101, Service 2 gets 102, Service 3 gets 103.
	return db.Redis.Incr(ctx, db.RedisGlobalCounterKey).Result()
}
