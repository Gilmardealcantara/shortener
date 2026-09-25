package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

var PG *pgxpool.Pool

func InitPostgres(ctx context.Context, dsn string) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		panic(err)
	}
	PG, err = pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		panic(err)
	}

	if err := PG.Ping(ctx); err != nil {
		panic(err)
	}
}

func ClosePostgres(ctx context.Context) {
	PG.Close()
}
