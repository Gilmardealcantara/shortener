package shortner

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Gilmardealcantara/shortener/db"
)

func Retrieve(ctx context.Context, code string) (string, error) {
	redisKey := fmt.Sprintf(db.RedisCodeUrlKey, code)
	longURL, err := db.Redis.Get(ctx, redisKey).Result()
	if err == nil {
		return longURL, nil
	}
	err = db.PG.QueryRow(ctx, "SELECT long_url FROM shortener WHERE code = $1", code).Scan(&longURL)
	if err != nil {
		return "", err
	}
	slog.Info("Retrieve", "code", code, "longURL", longURL)
	err = db.Redis.Set(ctx, code, longURL, 0).Err()
	return longURL, err
}
