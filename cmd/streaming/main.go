package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	kafkago "github.com/segmentio/kafka-go"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/HungphamLeo/BackendDataPlatform/streaming/config"
	"github.com/HungphamLeo/BackendDataPlatform/streaming/kafka"
	"github.com/HungphamLeo/BackendDataPlatform/streaming/models"
	"github.com/HungphamLeo/BackendDataPlatform/streaming/repository"
	"github.com/HungphamLeo/BackendDataPlatform/streaming/server"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync() //nolint:errcheck

	cfg := config.Load()

	repo, err := repository.New(cfg.Database.DSN)
	if err != nil {
		logger.Fatal("failed to connect to mysql", zap.Error(err))
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Start consumers for each market data topic.
	topics := map[string]func(context.Context, string, string, json.RawMessage, time.Time){
		"market.ticker":    makeTickerHandler(repo, logger),
		"market.orderbook": makeOrderBookHandler(repo, logger),
		"market.kline.1m":  makeKlineHandler("1m", repo, logger),
		"market.kline.5m":  makeKlineHandler("5m", repo, logger),
		"market.kline.1h":  makeKlineHandler("1h", repo, logger),
	}

	for topic, handler := range topics {
		c := kafka.New(cfg.Kafka.Brokers, topic, cfg.Kafka.ConsumerGroup, handler, logger)
		defer c.Close()
		go c.Run(ctx)
	}

	// Start gRPC server.
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.GRPC.Port))
	if err != nil {
		logger.Fatal("failed to listen", zap.Error(err))
	}

	grpcServer := grpc.NewServer()
	_ = server.New(repo) // register with generated protobuf service once stubs are generated
	reflection.Register(grpcServer)

	go func() {
		logger.Info("streaming gRPC server started", zap.String("port", cfg.GRPC.Port))
		if err := grpcServer.Serve(lis); err != nil {
			logger.Fatal("grpc serve failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down streaming service")
	grpcServer.GracefulStop()
	cancel()
}

// --- Kafka message handlers ---

type tickerPayload struct {
	Price  float64 `json:"price"`
	Volume float64 `json:"volume"`
	Bid    float64 `json:"bid"`
	Ask    float64 `json:"ask"`
}

func makeTickerHandler(repo *repository.DB, logger *zap.Logger) kafka.MessageHandler {
	return func(ctx context.Context, symbol, exchange string, payload json.RawMessage, ts time.Time) {
		var p tickerPayload
		if err := json.Unmarshal(payload, &p); err != nil {
			logger.Warn("parse ticker payload failed", zap.Error(err))
			return
		}
		t := &models.MarketTicker{
			Symbol: symbol, Exchange: exchange,
			Price: p.Price, Volume: p.Volume,
			Bid: p.Bid, Ask: p.Ask,
			Timestamp: ts,
		}
		if err := repo.SaveTicker(ctx, t); err != nil {
			logger.Error("save ticker failed", zap.Error(err))
		}
	}
}

type orderbookPayload struct {
	Bids json.RawMessage `json:"bids"`
	Asks json.RawMessage `json:"asks"`
}

func makeOrderBookHandler(repo *repository.DB, logger *zap.Logger) kafka.MessageHandler {
	return func(ctx context.Context, symbol, exchange string, payload json.RawMessage, ts time.Time) {
		var p orderbookPayload
		if err := json.Unmarshal(payload, &p); err != nil {
			logger.Warn("parse orderbook payload failed", zap.Error(err))
			return
		}
		ob := &models.MarketOrderBook{
			Symbol: symbol, Exchange: exchange,
			BidsJSON: string(p.Bids),
			AsksJSON: string(p.Asks),
			Timestamp: ts,
		}
		if err := repo.SaveOrderBook(ctx, ob); err != nil {
			logger.Error("save orderbook failed", zap.Error(err))
		}
	}
}

type klinePayload struct {
	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Close  float64 `json:"close"`
	Volume float64 `json:"volume"`
}

func makeKlineHandler(interval string, repo *repository.DB, logger *zap.Logger) kafka.MessageHandler {
	return func(ctx context.Context, symbol, exchange string, payload json.RawMessage, ts time.Time) {
		var p klinePayload
		if err := json.Unmarshal(payload, &p); err != nil {
			logger.Warn("parse kline payload failed", zap.Error(err), zap.String("interval", interval))
			return
		}
		k := &models.MarketKline{
			Symbol: symbol, Exchange: exchange, Interval: interval,
			Open: p.Open, High: p.High, Low: p.Low, Close: p.Close,
			Volume: p.Volume, Timestamp: ts,
		}
		if err := repo.SaveKline(ctx, k); err != nil {
			logger.Error("save kline failed", zap.Error(err))
		}
	}
}

// ensure kafkago is used (alias import)
var _ = kafkago.TCP
