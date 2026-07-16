package kafka

import (
	"context"
	"encoding/json"
	"time"

	kafkago "github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// MessageHandler is called with each decoded message envelope.
type MessageHandler func(ctx context.Context, symbol, exchange string, payload json.RawMessage, ts time.Time)

// Consumer reads from one Kafka topic using a consumer group.
type Consumer struct {
	reader  *kafkago.Reader
	handler MessageHandler
	logger  *zap.Logger
}

// New creates a Consumer for the given topic.
func New(brokers []string, topic, groupID string, handler MessageHandler, logger *zap.Logger) *Consumer {
	reader := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers:  brokers,
		Topic:    topic,
		GroupID:  groupID,
		MinBytes: 1,
		MaxBytes: 10e6,
	})
	return &Consumer{reader: reader, handler: handler, logger: logger}
}

// Run starts consuming messages until ctx is cancelled.
func (c *Consumer) Run(ctx context.Context) {
	for {
		m, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.logger.Error("kafka read error", zap.Error(err))
			continue
		}

		var envelope struct {
			Symbol    string          `json:"symbol"`
			Exchange  string          `json:"exchange"`
			Timestamp int64           `json:"timestamp"`
			Payload   json.RawMessage `json:"payload"`
		}
		if err := json.Unmarshal(m.Value, &envelope); err != nil {
			c.logger.Warn("unmarshal envelope failed", zap.Error(err))
			continue
		}

		ts := time.UnixMilli(envelope.Timestamp)
		c.handler(ctx, envelope.Symbol, envelope.Exchange, envelope.Payload, ts)
	}
}

// Close closes the underlying kafka reader.
func (c *Consumer) Close() {
	_ = c.reader.Close()
}
