package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/HungphamLeo/BackendDataPlatform/ingestor/config"
	"github.com/HungphamLeo/BackendDataPlatform/ingestor/kafka"
	redisclient "github.com/HungphamLeo/BackendDataPlatform/ingestor/redis"
)

// redisKeyPattern: {SYMBOL}_{EXCHANGE}_{DATA_TYPE}
// DATA_TYPE one of: ticker, orderbook, kline_1m, kline_5m, kline_1h

// Ingestor polls Redis and publishes events to Kafka.
type Ingestor struct {
	cfg    config.IngestorConfig
	redis  *redisclient.Client
	kafka  *kafka.Writer
	logger *zap.Logger
}

func New(
	cfg config.IngestorConfig,
	redis *redisclient.Client,
	kafkaWriter *kafka.Writer,
	logger *zap.Logger,
) *Ingestor {
	return &Ingestor{
		cfg:    cfg,
		redis:  redis,
		kafka:  kafkaWriter,
		logger: logger,
	}
}

var dataTypes = []struct {
	suffix string
	topic  string
}{
	{"ticker", "market.ticker"},
	{"orderbook", "market.orderbook"},
	{"kline_1m", "market.kline.1m"},
	{"kline_5m", "market.kline.5m"},
	{"kline_1h", "market.kline.1h"},
}

// Run starts the polling loop; it exits when ctx is cancelled.
func (s *Ingestor) Run(ctx context.Context) {
	interval := time.Duration(s.cfg.PollIntervalMS) * time.Millisecond
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	s.logger.Info("ingestor started",
		zap.Strings("symbols", s.cfg.Symbols),
		zap.Strings("exchanges", s.cfg.Exchanges),
		zap.Duration("interval", interval),
	)

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("ingestor shutting down")
			return
		case <-ticker.C:
			s.poll(ctx)
		}
	}
}

func (s *Ingestor) poll(ctx context.Context) {
	for _, symbol := range s.cfg.Symbols {
		for _, exchange := range s.cfg.Exchanges {
			for _, dt := range dataTypes {
				key := fmt.Sprintf("%s_%s_%s", symbol, exchange, dt.suffix)
				s.processKey(ctx, key, symbol, exchange, dt.topic)
			}
		}
	}
}

func (s *Ingestor) processKey(ctx context.Context, key, symbol, exchange, topic string) {
	raw, err := s.redis.Get(ctx, key)
	if err != nil {
		s.logger.Warn("redis get error", zap.String("key", key), zap.Error(err))
		return
	}
	if raw == "" {
		return // key doesn't exist yet
	}

	// Dedup: only publish if content changed since last publish.
	dedupKey := "ingestor:dedup:" + key
	set, err := s.redis.SetNX(ctx, dedupKey, raw, s.cfg.DeduplicateTTL)
	if err != nil {
		s.logger.Warn("redis setnx error", zap.String("key", dedupKey), zap.Error(err))
	}
	if !set {
		// Value unchanged — check if current value differs.
		prev, _ := s.redis.Get(ctx, dedupKey)
		if prev == raw {
			return // no change
		}
		// Update dedup entry with new value.
		_ = s.redis.Set(ctx, dedupKey, raw, s.cfg.DeduplicateTTL)
	}

	// Wrap in envelope so consumers know origin.
	envelope := map[string]interface{}{
		"symbol":    symbol,
		"exchange":  exchange,
		"timestamp": time.Now().UnixMilli(),
		"payload":   json.RawMessage(raw),
	}
	payload, err := json.Marshal(envelope)
	if err != nil {
		s.logger.Error("marshal envelope failed", zap.Error(err))
		return
	}

	partitionKey := fmt.Sprintf("%s_%s", symbol, exchange)
	if err := s.kafka.Publish(ctx, topic, partitionKey, payload); err != nil {
		s.logger.Error("kafka publish failed",
			zap.String("topic", topic),
			zap.String("key", key),
			zap.Error(err),
		)
	}
}
