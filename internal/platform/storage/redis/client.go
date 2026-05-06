package client

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisClient struct {
	client *redis.Client
}

func New(addr string) *RedisClient {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	return &RedisClient{client: rdb}
}

func (r *RedisClient) Close() error {
	return r.client.Close()
}

func (r *RedisClient) SetJSON(ctx context.Context, key string, value string, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *RedisClient) Raw() *redis.Client {
	return r.client
}

// InfoMemory returns INFO memory section
func (r *RedisClient) InfoMemory(ctx context.Context) (string, error) {
	return r.client.Info(ctx, "memory").Result()
}

// DBSize returns number of keys in current DB
func (r *RedisClient) DBSize(ctx context.Context) (int64, error) {
	return r.client.DBSize(ctx).Result()
}

// ScanKeys streams keys by pattern using SCAN cursor
func (r *RedisClient) ScanKeys(ctx context.Context, pattern string, count int64) *redis.ScanIterator {
	return r.client.Scan(ctx, 0, pattern, count).Iterator()
}

// Del deletes keys in a single call (best-effort)
func (r *RedisClient) Del(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	return r.client.Del(ctx, keys...).Err()
}

// Ping quick healthcheck
func (r *RedisClient) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

func (r *RedisClient) SetJSON(ctx context.Context, key string, value string, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}