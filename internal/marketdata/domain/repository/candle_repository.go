package repository

import (
	"context"
	"time"

	"github.com/HungphamLeo/BackendDataPlatform/internal/marketdata/domain/entity"
	"github.com/HungphamLeo/BackendDataPlatform/internal/marketdata/domain/value_object"
)

// CandleRepository định nghĩa interface để lưu trữ và truy xuất Candle
type CandleRepository interface {
	// FindBySymbolAndTimeframe tìm Candle theo Symbol, Timeframe và khoảng thời gian
	FindBySymbolAndTimeframe(
		ctx context.Context,
		symbol value_object.Symbol,
		timeframe value_object.Timeframe,
		startTime, endTime time.Time,
		limit int,
	) ([]*entity.Candle, error)
	
	// FindLatestBySymbolAndTimeframe tìm Candle mới nhất theo Symbol và Timeframe
	FindLatestBySymbolAndTimeframe(
		ctx context.Context,
		symbol value_object.Symbol,
		timeframe value_object.Timeframe,
		limit int,
	) ([]*entity.Candle, error)
	
	// FindBySymbolAndTimeframeAndOpenTime tìm Candle theo Symbol, Timeframe và OpenTime
	FindBySymbolAndTimeframeAndOpenTime(
		ctx context.Context,
		symbol value_object.Symbol,
		timeframe value_object.Timeframe,
		openTime time.Time,
	) (*entity.Candle, error)
	
	// Save lưu Candle
	Save(ctx context.Context, candle *entity.Candle) error
	
	// SaveBatch lưu nhiều Candle
	SaveBatch(ctx context.Context, candles []*entity.Candle) error
	
	// Delete xóa Candle
	Delete(
		ctx context.Context,
		symbol value_object.Symbol,
		timeframe value_object.Timeframe,
		openTime time.Time,
	) error
	
	// DeleteBefore xóa tất cả Candle trước một thời điểm
	DeleteBefore(
		ctx context.Context,
		symbol value_object.Symbol,
		timeframe value_object.Timeframe,
		time time.Time,
	) error
}
