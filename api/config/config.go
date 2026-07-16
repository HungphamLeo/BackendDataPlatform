package config

import (
	"os"
	"strconv"
)

// Config for data-platform-api service.
type Config struct {
	REST  RESTConfig
	GRPC  GRPCConfig
	Query QueryConfig
	JWT   JWTConfig
}

type RESTConfig struct {
	Port string // default 8090
}

type GRPCConfig struct {
	Port string // default 9090
}

type QueryConfig struct {
	Address string // data-platform-query gRPC address
}

type JWTConfig struct {
	Secret string
}

func Load() *Config {
	return &Config{
		REST: RESTConfig{
			Port: getEnv("REST_PORT", "8090"),
		},
		GRPC: GRPCConfig{
			Port: getEnv("GRPC_PORT", "9090"),
		},
		Query: QueryConfig{
			Address: getEnv("QUERY_GRPC_ADDR", "localhost:9091"),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", "change-me-in-production"),
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
