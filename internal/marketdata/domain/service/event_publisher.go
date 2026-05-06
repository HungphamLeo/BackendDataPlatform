package service

import (
	"context"

	"github.com/HungphamLeo/BackendDataPlatform/internal/marketdata/domain/event"
)

// EventPublisher định nghĩa interface để publish các sự kiện domain
type EventPublisher interface {
	// Publish đăng một sự kiện domain
	Publish(ctx context.Context, event event.DomainEvent) error
	
	// PublishBatch đăng nhiều sự kiện domain
	PublishBatch(ctx context.Context, events []event.DomainEvent) error
}
