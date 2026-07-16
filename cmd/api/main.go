package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/HungphamLeo/BackendDataPlatform/api/clients"
	"github.com/HungphamLeo/BackendDataPlatform/api/config"
	"github.com/HungphamLeo/BackendDataPlatform/api/middleware"
	"github.com/HungphamLeo/BackendDataPlatform/api/rest"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync() //nolint:errcheck

	cfg := config.Load()

	// gRPC client → data-platform-query
	queryClient, err := clients.NewQueryClient(cfg.Query.Address)
	if err != nil {
		logger.Fatal("failed to dial query service", zap.Error(err))
	}
	defer queryClient.Close() //nolint:errcheck

	// Build REST handlers
	marketH := rest.NewMarketHandler(queryClient)
	batchH := rest.NewBatchHandler(queryClient)
	signalsH := rest.NewSignalsHandler(queryClient)

	// HTTP router with JWT auth on all /v1/* routes
	mux := http.NewServeMux()
	jwtMiddleware := middleware.JWTAuth(cfg.JWT.Secret)

	mux.Handle("/v1/market/ticker", jwtMiddleware(http.HandlerFunc(marketH.GetTicker)))
	mux.Handle("/v1/market/orderbook", jwtMiddleware(http.HandlerFunc(marketH.GetOrderBook)))
	mux.Handle("/v1/market/kline", jwtMiddleware(http.HandlerFunc(marketH.GetKline)))
	mux.Handle("/v1/batch/ohlcv", jwtMiddleware(http.HandlerFunc(batchH.GetOHLCV)))
	mux.Handle("/v1/batch/indicators", jwtMiddleware(http.HandlerFunc(batchH.GetIndicators)))
	mux.Handle("/v1/signals", jwtMiddleware(http.HandlerFunc(signalsH.GetSignals)))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"status":"ok"}`)
	})

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.REST.Port),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// gRPC server (public, auth via interceptor — TODO: add auth interceptor)
	grpcServer := grpc.NewServer()
	// TODO: register DataQuery service (re-expose from api layer) once stubs generated
	reflection.Register(grpcServer)

	// Start REST
	go func() {
		logger.Info("REST API started", zap.String("port", cfg.REST.Port))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("REST serve failed", zap.Error(err))
		}
	}()

	// Start gRPC
	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.GRPC.Port))
		if err != nil {
			logger.Fatal("gRPC listen failed", zap.Error(err))
		}
		logger.Info("gRPC API started", zap.String("port", cfg.GRPC.Port))
		if err := grpcServer.Serve(lis); err != nil {
			logger.Fatal("gRPC serve failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down api service")
	grpcServer.GracefulStop()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Error("http shutdown error", zap.Error(err))
	}
}
