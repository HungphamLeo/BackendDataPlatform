package config

import (
	"os"
	"strconv"
	"strings"
)

// Config for data-platform-query service.
type Config struct {
	GRPC      GRPCConfig
	Streaming StreamingConfig
	Batch     BatchConfig
	Redis     RedisConfig
}

type GRPCConfig struct {
	Port string // listen port for DataQuery service
}

type StreamingConfig struct {
	Address string // e.g. "localhost:9092"
}

type BatchConfig struct {
	Address string // e.g. "localhost:9093"
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int // 10 reserved for query cache
}

func Load() *Config {
	return &Config{
		GRPC: GRPCConfig{
			Port: getEnv("GRPC_PORT", "9091"),
		},
		Streaming: StreamingConfig{
			Address: getEnv("STREAMING_GRPC_ADDR", "localhost:9092"),
		},
		Batch: BatchConfig{
			Address: getEnv("BATCH_GRPC_ADDR", "localhost:9093"),
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_CACHE_DB", 10),
		},
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		n, _ := strconv.Atoi(v)
		return n
	}
	return fallback
}

// suppress unused import lint if strings not needed elsewhere
var _ = strings.Split
