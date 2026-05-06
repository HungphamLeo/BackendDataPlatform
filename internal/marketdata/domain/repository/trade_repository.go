package repository

import (
	"context"
	"time"

	"github.com/HungphamLeo/BackendDataPlatform/internal/marketdata/domain/entity"
	"github.com/HungphamLeo/BackendDataPlatform/internal/marketdata/domain/value_object"
)

// TradeRepository định nghĩa interface để lưu trữ và truy xuất Trade
type TradeRepository interface {
	// FindByID tìm Trade theo ID
	FindByID(ctx context.Context, id string) (*entity.Trade, error)
	
	// FindBySymbol tìm Trade theo Symbol và khoảng thời gian
	FindBySymbol(
		ctx context.Context,
		symbol value_object.Symbol,
		startTime, endTime time.Time,
		limit int,
	) ([]*entity.Trade, error)
	
	// FindLatestBySymbol tìm Trade mới nhất theo Symbol
	FindLatestBySymbol(
		ctx context.Context,
		symbol value_object.Symbol,
		limit int,
	) ([]*entity.Trade, error)
	
	// Save lưu Trade
	Save(ctx context.Context, trade *entity.Trade) error
	
	// SaveBatch lưu nhiều Trade
	SaveBatch(ctx context.Context, trades []*entity.Trade) error
	
	// Delete xóa Trade
	Delete(ctx context.Context, id string) error
	
	// DeleteBefore xóa tất cả Trade trước một thời điểm
	DeleteBefore(
		ctx context.Context,
		symbol value_object.Symbol,
		time time.Time,
	) error
}
