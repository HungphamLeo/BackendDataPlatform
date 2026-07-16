package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/HungphamLeo/BackendDataPlatform/ingestor/config"
	"github.com/HungphamLeo/BackendDataPlatform/ingestor/kafka"
	redisclient "github.com/HungphamLeo/BackendDataPlatform/ingestor/redis"
	"github.com/HungphamLeo/BackendDataPlatform/ingestor/service"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync() //nolint:errcheck

	cfg := config.Load()

	redisClient, err := redisclient.New(cfg.Redis)
	if err != nil {
		logger.Fatal("failed to connect to redis", zap.Error(err))
	}
	defer redisClient.Close() //nolint:errcheck

	kafkaWriter := kafka.New(cfg.Kafka)
	defer kafkaWriter.Close()

	svc := service.New(cfg.Ingestor, redisClient, kafkaWriter, logger)

	ctx, cancel := context.WithCancel(context.Background())

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		logger.Info("signal received, stopping")
		cancel()
	}()

	svc.Run(ctx)
}
