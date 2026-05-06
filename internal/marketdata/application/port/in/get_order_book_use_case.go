package in

import (
	"context"

	"github.com/HungphamLeo/BackendDataPlatform/internal/marketdata/application/dto"
)

// GetOrderBookUseCase định nghĩa port đầu vào cho use case lấy dữ liệu sổ lệnh
type GetOrderBookUseCase interface {
	// Execute thực thi use case
	Execute(ctx context.Context, request *dto.GetOrderBookRequest) (*dto.GetOrderBookResponse, error)
}

// StreamOrderBookUseCase định nghĩa port đầu vào cho use case stream dữ liệu sổ lệnh
type StreamOrderBookUseCase interface {
	// Execute thực thi use case
	Execute(ctx context.Context, request *dto.StreamOrderBookRequest) (<-chan *dto.StreamOrderBookResponse, error)
}

// StreamOrderBookDiffUseCase định nghĩa port đầu vào cho use case stream dữ liệu thay đổi sổ lệnh
type StreamOrderBookDiffUseCase interface {
	// Execute thực thi use case
	Execute(ctx context.Context, request *dto.StreamOrderBookRequest) (<-chan *dto.StreamOrderBookDiffResponse, error)
}
