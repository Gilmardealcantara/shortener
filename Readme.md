# URL Shortener

A simple URL shortener written in Go, using PostgreSQL for persistence and Redis for caching and distributed unique ID generation.

## How it works

1. A `POST /shorten` request receives a long URL and returns a short URL.
2. A globally unique integer is generated via a Redis atomic counter (`INCR`), guaranteeing no collisions even across multiple instances.
3. The integer is encoded to base62 (e.g. `100M`) and stored alongside the original URL in PostgreSQL.
4. A `GET /{code}` request first checks the Redis cache; on a miss it queries PostgreSQL, populates the cache, and issues a `302` redirect.

## Tech stack

- **Go** 1.27
- **PostgreSQL** 16 (via [pgx/v5](https://github.com/jackc/pgx))
- **Redis** 7 (via [go-redis/v9](https://github.com/redis/go-redis))
- **golang-migrate** 4.20 for schema migrations

## API

| Method | Path        | Description                          |
|--------|-------------|--------------------------------------|
| POST   | `/shorten`  | Shorten a URL                        |
| GET    | `/{code}`   | Redirect to the original URL         |

### POST /shorten

```
curl -X POST http://localhost:8080/shorten \
  -H "Content-Type: application/json" \
  -d '{"long_url": "http://pudim.com.br"}'
```

Response:
```json
{"short_url": "http://localhost:8080/100M"}
```

### GET /{code}

```
curl -L http://localhost:8080/100M
```

Returns `302 Found` with a `Location` header pointing to the original URL, or `404` if the code doesn't exist.

## Getting started

### Prerequisites

- Go 1.27+
- Docker & Docker Compose
- [golang-migrate CLI](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate)

### 1. Start the dependencies

```sh
docker compose up -d
```

This starts PostgreSQL on port `5432` and Redis on port `6379`.

### 2. Run database migrations

```sh
make migrateup
```

### 3. Run the server

```sh
make run
```

The server listens on `:8080`.

## Development

### Run tests

Tests require the dependencies to be running (`docker compose up -d`).

```sh
make test
```

### Connect to databases

```sh
make pgcon      # opens psql inside the postgres container
make rediscon   # opens redis-cli inside the redis container
```

### Create a new migration

```sh
migrate create -ext sql -dir db/migrations -seq <migration_name>
```

### Roll back migrations

```sh
make migratedown
```

## Project structure

```
.
├── main.go                          # Entry point, HTTP server setup
├── main_test.go                     # Integration tests
├── compose.yml                      # Docker Compose for local dependencies
├── Makefile                         # Common dev tasks
├── db/
│   ├── pg.go                        # PostgreSQL connection pool
│   ├── redis.go                     # Redis client + counter initialization
│   └── migrations/
│       ├── 000001_init_schema.up.sql
│       └── 000001_init_schema.down.sql
└── pkg/
    ├── config/config.go             # App configuration (base URL, DSN)
    ├── handlers/handlers.go         # HTTP handlers
    └── shortner/
        ├── create.go                # URL shortening logic
        └── retrieve.go             # URL lookup with Redis cache
```
