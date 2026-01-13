package ticker

import "context"

// Repository interface for persisting domain objects. This keeps domain separate from infra.
type Repository interface {
	SaveTicker(ctx context.Context, t *Ticker) error
	SaveCandle(ctx context.Context, symbol string, timeframe string, c *Candle) error
	// other repo methods ...
}