package db

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

const (
	RedisGlobalCounterKey = "url:global:counter"
	RedisCodeURLKey       = "url:code:%s"
	InitialValue          = 238_328 // Sets the starting threshold (4-character minimum)
)

var Redis *redis.Client

func InitRedis(ctx context.Context, dsn string) {
	opts, err := redis.ParseURL(dsn)
	if err != nil {
		panic(err)
	}
	Redis = redis.NewClient(opts)

	// Verify the connection works cleanly
	if err := Redis.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to establish Redis handshake: %v", err)
	}

	initializeCounter(ctx)
}

func CloseRedis(ctx context.Context) {
	err := Redis.Close()
	if err != nil {
		panic(err)
	}
}

func initializeCounter(ctx context.Context) {
	// SETNX will return true ONLY if the key does not exist yet
	inserted, err := Redis.SetNX(ctx, RedisGlobalCounterKey, InitialValue, 0).Result()
	if err != nil {
		log.Fatalf("Fatal: Failed to seed Redis counter: %v", err)
	}

	if inserted {
		log.Printf("Successfully seeded global counter with starting offset: %d", InitialValue)
	} else {
		log.Println("Global counter already initialized. Skipping seed.")
	}
}
