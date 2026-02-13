package domain

import (
    "context"
)

// OrderRepository: Interface để lưu/lấy Order từ database
// Lưu ý: Interface định nghĩa ở domain (rules), implement ở adapter
type OrderRepository interface {
    Save(ctx context.Context, order *Order) error
    GetByID(ctx context.Context, id string) (*Order, error)
    GetByUserID(ctx context.Context, userID string) ([]*Order, error)
    Update(ctx context.Context, order *Order) error
    Delete(ctx context.Context, id string) error
}

// EventPublisher: Interface để gửi event (sự kiện)
type EventPublisher interface {
    Publish(ctx context.Context, topic string, event interface{}) error
}