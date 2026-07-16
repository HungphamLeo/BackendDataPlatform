package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all configuration for the ingestor service.
type Config struct {
	Redis    RedisConfig
	Kafka    KafkaConfig
	Ingestor IngestorConfig
}

type RedisConfig struct {
	Addr     string // e.g. "localhost:6379"
	Password string
	// DB 11 reserved for ingestor dedup tracking
	DB int
}

type KafkaConfig struct {
	Brokers []string // e.g. ["localhost:9092"]
}

type IngestorConfig struct {
	Symbols         []string // e.g. ["BTC_USDT","ETH_USDT"]
	Exchanges       []string // e.g. ["binance","bybit"]
	PollIntervalMS  int      // default 500
	DeduplicateTTL  time.Duration
}

// Load reads config from environment variables with sensible defaults.
func Load() *Config {
	return &Config{
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 11),
		},
		Kafka: KafkaConfig{
			Brokers: strings.Split(getEnv("KAFKA_BROKERS", "localhost:9092"), ","),
		},
		Ingestor: IngestorConfig{
			Symbols:        strings.Split(getEnv("SYMBOLS", "BTC_USDT,ETH_USDT"), ","),
			Exchanges:      strings.Split(getEnv("EXCHANGES", "binance"), ","),
			PollIntervalMS: getEnvInt("POLL_INTERVAL_MS", 500),
			DeduplicateTTL: time.Duration(getEnvInt("DEDUP_TTL_S", 60)) * time.Second,
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
		n, err := strconv.Atoi(v)
		if err == nil {
			return n
		}
	}
	return fallback
}
