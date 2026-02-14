package kafka

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(brokers []string, topic string) *Producer {

	writer := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}

	return &Producer{writer}
}

func (p *Producer) Send(message []byte) error {

	return p.writer.WriteMessages(
		context.Background(),
		kafka.Message{
			Key:   []byte(time.Now().String()),
			Value: message,
		},
	)
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
