package domain

import (
	"context"
	"time"
)

// Tick đại diện cho một sự kiện biến động giá (tick) từ các sàn
type Tick struct {
	Symbol    string    `json:"symbol"`
	Price     float64   `json:"price"`
	Quantity  float64   `json:"quantity"`
	Timestamp time.Time `json:"timestamp"`
	Source    string    `json:"source"` // Ví dụ: "BINANCE_FUTURES", "BINANCE_SPOT", "XTB"
}

// MarketDataProvider là Interface cốt lõi. Sàn nào muốn tích hợp vào hệ thống đều phải implement cái này.
type MarketDataProvider interface {
	Connect(ctx context.Context) error
	Subscribe(symbols []string) error
	Stream(ctx context.Context, tickCh chan<- *Tick, errCh chan<- error)
	Close() error
}

// MarketDataPublisher là Interface để xuất message lên Event Bus (Kafka)
type MarketDataPublisher interface {
	PublishTick(ctx context.Context, topic string, tick *Tick) error
}