package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Kafka    KafkaConfig
	Database DatabaseConfig
	GRPC     GRPCConfig
}

type KafkaConfig struct {
	Brokers       []string
	ConsumerGroup string
}

type DatabaseConfig struct {
	DSN string
}

type GRPCConfig struct {
	Port string
}

func Load() *Config {
	return &Config{
		Kafka: KafkaConfig{
			Brokers:       strings.Split(getEnv("KAFKA_BROKERS", "localhost:9092"), ","),
			ConsumerGroup: getEnv("KAFKA_GROUP", "data-platform-streaming"),
		},
		Database: DatabaseConfig{
			DSN: getEnv("MYSQL_DSN", "root:password@tcp(localhost:3306)/data_platform?parseTime=true"),
		},
		GRPC: GRPCConfig{
			Port: getEnv("GRPC_PORT", "9092"),
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
