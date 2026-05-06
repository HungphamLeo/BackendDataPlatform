package out

import (
	"context"

	"github.com/HungphamLeo/BackendDataPlatform/internal/marketdata/domain/event"
)

// EventPublisherPort định nghĩa port đầu ra cho việc publish các sự kiện domain
type EventPublisherPort interface {
	// Publish đăng một sự kiện domain
	Publish(ctx context.Context, event event.DomainEvent) error
	
	// PublishBatch đăng nhiều sự kiện domain
	PublishBatch(ctx context.Context, events []event.DomainEvent) error
}
