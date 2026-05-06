package repository

import (
	"context"

	"github.com/HungphamLeo/BackendDataPlatform/internal/marketdata/domain/aggregate"
	"github.com/HungphamLeo/BackendDataPlatform/internal/marketdata/domain/value_object"
)

// TickerRepository định nghĩa interface để lưu trữ và truy xuất Ticker
type TickerRepository interface {
	// FindByID tìm Ticker theo ID
	FindByID(ctx context.Context, id aggregate.TickerID) (*aggregate.Ticker, error)
	
	// FindBySymbol tìm Ticker theo Symbol và Exchange
	FindBySymbol(ctx context.Context, symbol value_object.Symbol, exchange string) (*aggregate.Ticker, error)
	
	// FindAll tìm tất cả Ticker
	FindAll(ctx context.Context) ([]*aggregate.Ticker, error)
	
	// FindByExchange tìm tất cả Ticker theo Exchange
	FindByExchange(ctx context.Context, exchange string) ([]*aggregate.Ticker, error)
	
	// Save lưu Ticker
	Save(ctx context.Context, ticker *aggregate.Ticker) error
	
	// Delete xóa Ticker
	Delete(ctx context.Context, id aggregate.TickerID) error
}
