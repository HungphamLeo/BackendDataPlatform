package out

import (
	"context"
	"time"

	"github.com/HungphamLeo/BackendDataPlatform/internal/marketdata/domain/entity"
	"github.com/HungphamLeo/BackendDataPlatform/internal/marketdata/domain/value_object"
)

// CandleRepositoryPort định nghĩa port đầu ra cho repository lưu trữ và truy xuất Candle
type CandleRepositoryPort interface {
	// FindByID tìm Candle theo ID
	FindByID(ctx context.Context, id string) (*entity.Candle, error)
	
	// FindBySymbolAndTimeframe tìm Candle theo Symbol, Exchange, Timeframe và thời gian
	FindBySymbolAndTimeframe(
		ctx context.Context,
		symbol value_object.Symbol,
		exchange string,
		timeframe value_object.Timeframe,
		startTime, endTime time.Time,
		limit int,
	) ([]*entity.Candle, error)
	
	// FindLatestBySymbolAndTimeframe tìm Candle mới nhất theo Symbol, Exchange và Timeframe
	FindLatestBySymbolAndTimeframe(
		ctx context.Context,
		symbol value_object.Symbol,
		exchange string,
		timeframe value_object.Timeframe,
	) (*entity.Candle, error)
	
	// Save lưu Candle
	Save(ctx context.Context, candle *entity.Candle) error
	
	// SaveBatch lưu nhiều Candle
	SaveBatch(ctx context.Context, candles []*entity.Candle) error
	
	// Delete xóa Candle
	Delete(ctx context.Context, id string) error
}
