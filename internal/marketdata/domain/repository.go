package domain

import (
	"context"

	"github.com/HungphamLeo/BackendDataPlatform/internal/marketdata/domain/aggregate"
	"github.com/HungphamLeo/BackendDataPlatform/internal/marketdata/domain/value_object"
)

// ✅ REPOSITORY INTERFACE (Persistence Port)
// Implementation sẽ ở infrastructure/
type TickerRepository interface {
	Save(ctx context.Context, ticker *aggregate.Ticker) error
	GetBySymbol(ctx context.Context, symbol value_object.Symbol) (*aggregate.Ticker, error)
	GetAllSymbols(ctx context.Context) ([]value_object.Symbol, error)
}

// ✅ PRICE CALCULATOR SERVICE (Domain Service)
// No persistence, chỉ tính toán
type PriceCalculator interface {
	CalculateMovingAverage(prices []float64, period int) float64
	CalculateRSI(prices []float64, period int) float64
	CalculateTrend(prices []float64) string // "UP", "DOWN", "NEUTRAL"
}