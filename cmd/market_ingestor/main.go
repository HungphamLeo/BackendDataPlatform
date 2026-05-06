package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/HungphamLeo/BackendDataPlatform/internal/marketdata/adapter/messaging/kafka"
	pgRepo "github.com/HungphamLeo/BackendDataPlatform/internal/marketdata/adapter/persistence/postgres"
	"github.com/HungphamLeo/BackendDataPlatform/internal/marketdata/adapter/provider/binance/spot/rest"
	"github.com/HungphamLeo/BackendDataPlatform/internal/marketdata/application/port/in"
	"github.com/HungphamLeo/BackendDataPlatform/internal/marketdata/application/service"
	httpTransport "github.com/HungphamLeo/BackendDataPlatform/internal/marketdata/transport/http"
	grpcTransport "github.com/HungphamLeo/BackendDataPlatform/internal/marketdata/transport/grpc"
	pb "github.com/HungphamLeo/BackendDataPlatform/api/proto/marketdata"
	"github.com/HungphamLeo/BackendDataPlatform/internal/platform/config"
	"github.com/HungphamLeo/BackendDataPlatform/internal/platform/logging"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "config/marketdata.yaml", "Path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger, err := logging.NewLogger(&cfg.Logging)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	logger.Info("Starting market data service", logging.Fields{
		"app":         cfg.App.Name,
		"version":     cfg.App.Version,
		"environment": cfg.App.Environment,
	})

	// Initialize database connection
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.Database,
		cfg.Database.SSLMode,
	)
	
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Fatal("Failed to connect to database", logging.Fields{
			"error": err.Error(),
		})
	}
	
	// Set connection pool settings
	sqlDB, err := db.DB()
	if err != nil {
		logger.Fatal("Failed to get database connection", logging.Fields{
			"error": err.Error(),
		})
	}
	
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	// Initialize repositories
	tickerRepository := pgRepo.NewTickerRepository(db)

	// Initialize event publisher
	eventPublisher, err := kafka.NewEventPublisher(&cfg.Kafka, logger)
	if err != nil {
		logger.Fatal("Failed to initialize event publisher", logging.Fields{
			"error": err.Error(),
		})
	}
	defer eventPublisher.Close()

	// Initialize market data providers
	binanceSpotRestClient := rest.NewBinanceSpotRestClient(&cfg.Binance, logger)

	// Initialize use cases
	var getTickerUseCase in.GetTickerUseCase = service.NewGetTickerService(tickerRepository, binanceSpotRestClient)

	// Initialize HTTP handlers
	tickerHandler := httpTransport.NewTickerHandler(getTickerUseCase, nil, logger)

	// Initialize HTTP router
	router := mux.NewRouter()
	tickerHandler.RegisterRoutes(router)

	// Initialize HTTP server
	httpServer := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.HTTP.Host, cfg.HTTP.Port),
		Handler:      router,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}

	// Initialize gRPC server
	grpcServer := grpc.NewServer(
		grpc.MaxConcurrentStreams(cfg.GRPC.MaxConcurrentStreams),
	)
	
	// Register gRPC services
	tickerService := grpcTransport.NewTickerService(getTickerUseCase, nil, logger)
	pb.RegisterTickerServiceServer(grpcServer, tickerService)
	
	// Enable reflection for grpcurl
	reflection.Register(grpcServer)

	// Start HTTP server in a goroutine
	go func() {
		logger.Info("Starting HTTP server", logging.Fields{
			"address": httpServer.Addr,
		})
		
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start HTTP server", logging.Fields{
				"error": err.Error(),
			})
		}
	}()

	// Start gRPC server in a goroutine
	go func() {
		addr := fmt.Sprintf("%s:%d", cfg.GRPC.Host, cfg.GRPC.Port)
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			logger.Fatal("Failed to listen for gRPC server", logging.Fields{
				"error": err.Error(),
			})
		}
		
		logger.Info("Starting gRPC server", logging.Fields{
			"address": addr,
		})
		
		if err := grpcServer.Serve(listener); err != nil {
			logger.Fatal("Failed to start gRPC server", logging.Fields{
				"error": err.Error(),
			})
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the servers
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	logger.Info("Shutting down servers", nil)

	// Shutdown gRPC server
	grpcServer.GracefulStop()
	logger.Info("gRPC server stopped", nil)

	// Shutdown HTTP server
	ctx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()
	
	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Fatal("Failed to shutdown HTTP server gracefully", logging.Fields{
			"error": err.Error(),
		})
	}
	
	logger.Info("HTTP server stopped", nil)
	
	// Close database connection
	if err := sqlDB.Close(); err != nil {
		logger.Error("Failed to close database connection", logging.Fields{
			"error": err.Error(),
		})
	}
	
	logger.Info("Market data service stopped", nil)
}
