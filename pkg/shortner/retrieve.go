package shortner

import (
	"context"
	"log/slog"

	"github.com/Gilmardealcantara/shortener/db"
)

func Retrieve(code string) (string, error) {
	var longURL string
	err := db.PG.QueryRow(context.Background(), "SELECT long_url FROM shortener WHERE code = $1", code).Scan(&longURL)
	if err != nil {
		return "", err
	}
	slog.Info("Retrieve", "code", code, "longURL", longURL)
	return longURL, nil
}
