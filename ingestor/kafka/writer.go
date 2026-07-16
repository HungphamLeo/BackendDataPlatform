package kafka

import (
	"context"
	"fmt"

	kafkago "github.com/segmentio/kafka-go"

	"github.com/HungphamLeo/BackendDataPlatform/ingestor/config"
)

// Writer is a thin wrapper around segmentio/kafka-go writer.
type Writer struct {
	writers map[string]*kafkago.Writer // keyed by topic
}

// New creates a kafka Writer for all data-platform topics.
func New(cfg config.KafkaConfig) *Writer {
	topics := []string{
		"market.ticker",
		"market.orderbook",
		"market.kline.1m",
		"market.kline.5m",
		"market.kline.1h",
	}

	writers := make(map[string]*kafkago.Writer, len(topics))
	for _, topic := range topics {
		writers[topic] = &kafkago.Writer{
			Addr:                   kafkago.TCP(cfg.Brokers...),
			Topic:                  topic,
			Balancer:               &kafkago.Hash{},
			AllowAutoTopicCreation: true,
		}
	}

	return &Writer{writers: writers}
}

// Publish sends a message to the given topic, using key as the partition key.
func (w *Writer) Publish(ctx context.Context, topic, key string, payload []byte) error {
	wr, ok := w.writers[topic]
	if !ok {
		return fmt.Errorf("unknown topic: %s", topic)
	}

	return wr.WriteMessages(ctx, kafkago.Message{
		Key:   []byte(key),
		Value: payload,
	})
}

// Close closes all underlying kafka writers.
func (w *Writer) Close() {
	for _, wr := range w.writers {
		_ = wr.Close()
	}
}
