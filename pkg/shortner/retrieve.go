package shortner

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Gilmardealcantara/shortener/db"
)

func Retrieve(ctx context.Context, code string) (string, error) {
	redisKey := fmt.Sprintf(db.RedisCodeURLKey, code)
	longURL, err := db.Redis.Get(ctx, redisKey).Result()
	if err == nil {
		slog.InfoContext(ctx, "Retrieve from redis", "code", code, "longURL", longURL)
		return longURL, nil
	}

	err = db.PG.QueryRow(ctx, "SELECT long_url FROM shortener WHERE code = $1", code).Scan(&longURL)
	if err != nil {
		return "", err
	}
	slog.InfoContext(ctx, "Retrieve from db", "code", code, "longURL", longURL)
	err = db.Redis.Set(ctx, redisKey, longURL, 0).Err()
	return longURL, err
}
