package aggregate

import (
	"fmt"
	"time"

	"github.com/HungphamLeo/BackendDataPlatform/internal/marketdata/domain/entity"
	"github.com/HungphamLeo/BackendDataPlatform/internal/marketdata/domain/event"
	"github.com/HungphamLeo/BackendDataPlatform/internal/marketdata/domain/value_object"
)

// ✅ AGGREGATE ROOT: Ticker
// Chứa tất cả các candles, ticks của một symbol
type Ticker struct {
	id           TickerID
	symbol       value_object.Symbol       // EUR/USD, AAPL, etc
	exchange     string                    // FOREX, NASDAQ, etc
	candles      map[string]*entity.Candle // Key: timeframe (1m, 5m, 1h)
	lastPrice    value_object.Quote
	lastUpdate   time.Time
	domainEvents []event.DomainEvent // Events to publish later
}

// TickerID là value object
type TickerID string

func NewTicker(symbol value_object.Symbol, exchange string) *Ticker {
	return &Ticker{
		id:       TickerID(fmt.Sprintf("%s:%s", symbol, exchange)),
		symbol:   symbol,
		exchange: exchange,
		candles:  make(map[string]*entity.Candle),
	}
}

// 🎯 Business Logic: Update price
func (t *Ticker) UpdatePrice(quote value_object.Quote) error {
	// ❌ KHÔNG gọi database ở đây!
	// ❌ KHÔNG gọi HTTP API ở đây!
	// ✅ CHỈ kiểm tra business rules

	if err := quote.Validate(); err != nil {
		return fmt.Errorf("invalid quote: %w", err)
	}

	// Check if price changed significantly (business rule)
	if !t.isPriceMoved(quote) {
		return nil // Ignore minor fluctuations
	}

	t.lastPrice = quote
	t.lastUpdate = time.Now()

	// Raise domain event (to be published later)
	t.raiseEvent(&event.PriceUpdated{
		Symbol:    t.symbol,
		NewPrice:  quote,
		Timestamp: time.Now(),
	})

	return nil
}

// 🎯 Business Logic: Close candle
func (t *Ticker) CloseCandle(tf value_object.Timeframe) error {
	candle, exists := t.candles[tf.String()]
	if !exists {
		return fmt.Errorf("candle not found for timeframe: %s", tf.String())
	}

	// Validate OHLC data
	if err := candle.Validate(); err != nil {
		return fmt.Errorf("invalid candle: %w", err)
	}

	candle.Close()

	// Raise domain event
	t.raiseEvent(&event.CandleClosed{
		Symbol:    t.symbol,
		Timeframe: tf,
		OHLC:      candle.OHLC(),
		Timestamp: time.Now(),
	})

	return nil
}

// Helper: Check if price moved enough to trigger update
func (t *Ticker) isPriceMoved(newQuote value_object.Quote) bool {
	if t.lastPrice.Bid == 0 {
		return true // First price
	}

	// Business rule: Only trigger if moved > 0.01 pips
	change := (newQuote.Bid - t.lastPrice.Bid) / t.lastPrice.Bid
	return change > 0.0001 || change < -0.0001
}

// Internal: Add event to aggregate
func (t *Ticker) raiseEvent(evt event.DomainEvent) {
	t.domainEvents = append(t.domainEvents, evt)
}

// Public: Get uncommitted events (to be published)
func (t *Ticker) DomainEvents() []event.DomainEvent {
	return t.domainEvents
}

func (t *Ticker) ClearDomainEvents() {
	t.domainEvents = []event.DomainEvent{}
}
