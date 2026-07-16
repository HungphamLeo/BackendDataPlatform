package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/HungphamLeo/BackendDataPlatform/query/cache"
	"github.com/HungphamLeo/BackendDataPlatform/query/clients"
	"github.com/HungphamLeo/BackendDataPlatform/query/config"
	"github.com/HungphamLeo/BackendDataPlatform/query/server"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync() //nolint:errcheck

	cfg := config.Load()

	// Redis cache (DB 10)
	queryCache, err := cache.New(cfg.Redis)
	if err != nil {
		logger.Fatal("redis cache failed", zap.Error(err))
	}
	defer queryCache.Close() //nolint:errcheck

	// gRPC clients
	streamingClient, err := clients.NewStreamingClient(cfg.Streaming.Address)
	if err != nil {
		logger.Fatal("dial streaming failed", zap.Error(err))
	}
	defer streamingClient.Close() //nolint:errcheck

	batchClient, err := clients.NewBatchClient(cfg.Batch.Address)
	if err != nil {
		logger.Fatal("dial batch failed", zap.Error(err))
	}
	defer batchClient.Close() //nolint:errcheck

	// Query server
	_ = server.New(streamingClient, batchClient, queryCache)

	// gRPC listener
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.GRPC.Port))
	if err != nil {
		logger.Fatal("listen failed", zap.Error(err))
	}

	grpcServer := grpc.NewServer()
	// TODO: once proto stubs generated, register the DataQuery service here:
	// dqpb.RegisterDataQueryServer(grpcServer, queryServer)
	reflection.Register(grpcServer)

	go func() {
		logger.Info("query gRPC server started", zap.String("port", cfg.GRPC.Port))
		if err := grpcServer.Serve(lis); err != nil {
			logger.Fatal("grpc serve failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down query service")
	grpcServer.GracefulStop()
}
