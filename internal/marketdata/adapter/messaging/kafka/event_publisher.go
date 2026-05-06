package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"

	"github.com/HungphamLeo/BackendDataPlatform/internal/marketdata/application/port/out"
	"github.com/HungphamLeo/BackendDataPlatform/internal/marketdata/domain/event"
	"github.com/HungphamLeo/BackendDataPlatform/internal/platform/config"
	"github.com/HungphamLeo/BackendDataPlatform/internal/platform/logging"
)

// EventPublisher triển khai EventPublisherPort sử dụng Kafka
type EventPublisher struct {
	producer *kafka.Producer
	logger   logging.Logger
	config   *config.KafkaConfig
}

// NewEventPublisher tạo mới EventPublisher
func NewEventPublisher(
	config *config.KafkaConfig,
	logger logging.Logger,
) (*EventPublisher, error) {
	// Tạo Kafka producer
	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers":   config.BootstrapServers,
		"client.id":           config.ClientID,
		"acks":                "all",
		"retries":             5,
		"retry.backoff.ms":    500,
		"delivery.timeout.ms": 10000,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka producer: %w", err)
	}
	
	// Khởi tạo EventPublisher
	publisher := &EventPublisher{
		producer: producer,
		logger:   logger,
		config:   config,
	}
	
	// Khởi động goroutine để xử lý các sự kiện delivery report
	go publisher.handleDeliveryReports()
	
	return publisher, nil
}

// Đảm bảo EventPublisher triển khai interface EventPublisherPort
var _ out.EventPublisherPort = (*EventPublisher)(nil)

// Publish đăng một sự kiện domain
func (p *EventPublisher) Publish(ctx context.Context, event event.DomainEvent) error {
	// Lấy topic từ tên sự kiện
	topic := p.getTopicForEvent(event)
	
	// Chuyển đổi sự kiện thành JSON
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}
	
	// Tạo Kafka message
	message := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: kafka.PartitionAny,
		},
		Value: payload,
		Key:   []byte(event.GetAggregateID()),
		Headers: []kafka.Header{
			{
				Key:   "event_type",
				Value: []byte(event.GetEventType()),
			},
			{
				Key:   "event_id",
				Value: []byte(event.GetEventID()),
			},
			{
				Key:   "timestamp",
				Value: []byte(fmt.Sprintf("%d", event.GetTimestamp().UnixNano())),
			},
		},
		Timestamp: time.Now(),
	}
	
	// Gửi message đến Kafka
	err = p.producer.Produce(message, nil)
	if err != nil {
		return fmt.Errorf("failed to produce message: %w", err)
	}
	
	// Flush để đảm bảo message được gửi đi
	remaining := p.producer.Flush(1000)
	if remaining > 0 {
		p.logger.Warn(fmt.Sprintf("%d messages still in queue after flush", remaining), nil)
	}
	
	return nil
}

// PublishBatch đăng nhiều sự kiện domain
func (p *EventPublisher) PublishBatch(ctx context.Context, events []event.DomainEvent) error {
	for _, e := range events {
		if err := p.Publish(ctx, e); err != nil {
			return err
		}
	}
	return nil
}

// getTopicForEvent trả về tên topic cho sự kiện
func (p *EventPublisher) getTopicForEvent(event event.DomainEvent) string {
	// Lấy topic từ cấu hình dựa trên loại sự kiện
	switch event.GetEventType() {
	case "price_updated":
		return p.config.Topics.PriceUpdated
	case "new_trade":
		return p.config.Topics.NewTrade
	case "alert_triggered":
		return p.config.Topics.AlertTriggered
	default:
		// Fallback to default topic
		return p.config.Topics.Default
	}
}

// handleDeliveryReports xử lý các sự kiện delivery report từ Kafka
func (p *EventPublisher) handleDeliveryReports() {
	for e := range p.producer.Events() {
		switch ev := e.(type) {
		case *kafka.Message:
			if ev.TopicPartition.Error != nil {
				p.logger.Error("Failed to deliver message", logging.Fields{
					"topic":     *ev.TopicPartition.Topic,
					"partition": ev.TopicPartition.Partition,
					"error":     ev.TopicPartition.Error.Error(),
				})
			} else {
				p.logger.Debug("Message delivered", logging.Fields{
					"topic":     *ev.TopicPartition.Topic,
					"partition": ev.TopicPartition.Partition,
					"offset":    ev.TopicPartition.Offset,
				})
			}
		}
	}
}

// Close đóng Kafka producer
func (p *EventPublisher) Close() {
	p.producer.Close()
}
