package db

import (
	"context"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	PG          *pgxpool.Pool
	ErrNotFound = pgx.ErrNoRows
)

func InitPostgres(ctx context.Context, dsn string) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		panic(err)
	}

	// Attach OTel tracer — every query, batch, and prepared statement
	// will emit a child span visible in Jaeger under the request trace.
	cfg.ConnConfig.Tracer = otelpgx.NewTracer()

	PG, err = pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		panic(err)
	}

	if err := PG.Ping(ctx); err != nil {
		panic(err)
	}

	// Record pool stats (active conns, idle conns, etc.) as OTel metrics.
	if err := otelpgx.RecordStats(PG); err != nil {
		panic(err)
	}
}

func ClosePostgres(ctx context.Context) {
	PG.Close()
}
