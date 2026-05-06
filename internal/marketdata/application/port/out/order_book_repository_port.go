package out

import (
	"context"

	"github.com/HungphamLeo/BackendDataPlatform/internal/marketdata/domain/entity"
	"github.com/HungphamLeo/BackendDataPlatform/internal/marketdata/domain/value_object"
)

// OrderBookRepositoryPort định nghĩa port đầu ra cho repository lưu trữ và truy xuất OrderBook
type OrderBookRepositoryPort interface {
	// FindBySymbol tìm OrderBook theo Symbol và Exchange
	FindBySymbol(
		ctx context.Context,
		symbol value_object.Symbol,
		exchange string,
	) (*entity.OrderBook, error)
	
	// Save lưu OrderBook
	Save(ctx context.Context, orderBook *entity.OrderBook) error
	
	// Delete xóa OrderBook
	Delete(
		ctx context.Context,
		symbol value_object.Symbol,
		exchange string,
	) error
}
