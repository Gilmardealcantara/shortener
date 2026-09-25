package tests

import (
	"context"
	"log/slog"
	"net/http"
	"path/filepath"
	"time"

	"github.com/Gilmardealcantara/shortener/pkg/config"
	"github.com/Gilmardealcantara/shortener/pkg/handlers"
	"github.com/Gilmardealcantara/shortener/pkg/shortner"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	redisContainer *redis.RedisContainer
	pgContainer    *postgres.PostgresContainer
)

func SetupContainers(ctx context.Context) config.Config {
	var err error

	redisContainer, err = redis.Run(
		ctx,
		"redis:7",
		redis.WithSnapshotting(10, 1),
		redis.WithLogLevel(redis.LogLevelVerbose),
		// redis.WithConfigFile(filepath.Join("testdata", "redis7.conf")),
	)
	if err != nil {
		panic(err)
	}

	pgContainer, err = postgres.Run(
		ctx,
		"postgres:15.3-alpine",
		postgres.WithInitScripts(filepath.Join("testdata", "init-user-db.sh")),
		postgres.WithDatabase("test-db"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(5*time.Second),
		),
	)
	redisDsn, _ := redisContainer.ConnectionString(ctx) // sslmode=disable
	slog.Info("SetupContainers", "redisDSN", redisDsn)
	pgDsn, _ := pgContainer.ConnectionString(ctx)
	slog.Info("SetupContainers", "pgDSN", pgDsn)

	return config.Config{
		PostgresDSN: pgDsn,
		RedisDSN:    redisDsn,
		BaseURL:     "http://localhost:8080",
	}
}

func TearDownContainers(ctx context.Context) {
	if redisContainer != nil {
		redisContainer.Terminate(ctx)
	}
	if pgContainer != nil {
		pgContainer.Terminate(ctx)
	}
}

func SetupMuxServer() *http.ServeMux {
	cfg := config.New()
	shortnerSrv := shortner.New(cfg)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{code}", handlers.Redirect)
	mux.HandleFunc("POST /shorten", handlers.Create(shortnerSrv))
	return mux
}
