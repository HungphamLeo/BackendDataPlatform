package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/HungphamLeo/BackendDataPlatform/query/config"
)

// QueryCache provides a Redis-backed TTL cache for query results.
type QueryCache struct {
	rdb *redis.Client
}

// New connects to Redis and returns a QueryCache.
func New(cfg config.RedisConfig) (*QueryCache, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("query cache redis ping failed: %w", err)
	}
	return &QueryCache{rdb: rdb}, nil
}

// Get returns the cached JSON string for the given key, or ("", nil) on miss.
func (c *QueryCache) Get(ctx context.Context, key string) (string, error) {
	val, err := c.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}

// Set stores value with the given TTL.
func (c *QueryCache) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	return c.rdb.Set(ctx, key, value, ttl).Err()
}

// Close shuts down the Redis client.
func (c *QueryCache) Close() error {
	return c.rdb.Close()
}

// RealtimeTTL is the TTL for real-time (ticker/orderbook/kline) cache entries.
const RealtimeTTL = 5 * time.Second

// BatchTTL is the TTL for batch/historical cache entries.
const BatchTTL = 60 * time.Second
