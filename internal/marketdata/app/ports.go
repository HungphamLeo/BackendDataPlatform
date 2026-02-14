package application

import (
	"context"
	"time"

	"internal/marketdata/domain/aggregate"
	"internal/marketdata/domain/value_object"
)

// ✅ PERSISTENCE PORT
type TickerRepository interface {
	Save(ctx context.Context, ticker *aggregate.Ticker) error
	GetBySymbol(ctx context.Context, symbol value_object.Symbol) (*aggregate.Ticker, error)
}

// ✅ EXTERNAL API PORT
type binanceClient interface {
	// REST API calls
	GetQuote(ctx context.Context, symbol value_object.Symbol) (*value_object.Quote, error)
	GetCandles(ctx context.Context, symbol value_object.Symbol, tf value_object.Timeframe) ([]*entity.Candle, error)

	// WebSocket stream
	StreamPrices(ctx context.Context, symbols []value_object.Symbol) (<-chan *value_object.Quote, error)
}

// ✅ CACHE PORT
type CacheStore interface {
	Get(ctx context.Context, key string) (any, error)
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}

// ✅ LOGGING PORT (from platform)
type Logger interface {
	Info(ctx context.Context, msg string, fields ...any)
	Error(ctx context.Context, msg string, err error)
	Debug(ctx context.Context, msg string, fields ...any)
	Warn(ctx context.Context, msg string, fields ...any)
}

// ✅ METRICS PORT (from platform)
type Metrics interface {
	RecordPrice(symbol string, price float64)
	RecordCacheHit(key string)
	RecordError(operation string, source string)
}

// ✅ EVENT BUS PORT (from platform)
type EventPublisher interface {
	PublishEvent(ctx context.Context, topic string, event any) error
}
