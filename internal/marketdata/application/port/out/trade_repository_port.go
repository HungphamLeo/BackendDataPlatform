package out

import (
	"context"
	"time"

	"github.com/HungphamLeo/BackendDataPlatform/internal/marketdata/domain/entity"
	"github.com/HungphamLeo/BackendDataPlatform/internal/marketdata/domain/value_object"
)

// TradeRepositoryPort định nghĩa port đầu ra cho repository lưu trữ và truy xuất Trade
type TradeRepositoryPort interface {
	// FindByID tìm Trade theo ID
	FindByID(ctx context.Context, id string) (*entity.Trade, error)
	
	// FindBySymbol tìm Trade theo Symbol, Exchange và thời gian
	FindBySymbol(
		ctx context.Context,
		symbol value_object.Symbol,
		exchange string,
		startTime, endTime time.Time,
		limit int,
	) ([]*entity.Trade, error)
	
	// FindBySymbolFromID tìm Trade theo Symbol, Exchange và ID
	FindBySymbolFromID(
		ctx context.Context,
		symbol value_object.Symbol,
		exchange string,
		fromID string,
		limit int,
	) ([]*entity.Trade, error)
	
	// Save lưu Trade
	Save(ctx context.Context, trade *entity.Trade) error
	
	// SaveBatch lưu nhiều Trade
	SaveBatch(ctx context.Context, trades []*entity.Trade) error
	
	// Delete xóa Trade
	Delete(ctx context.Context, id string) error
}
