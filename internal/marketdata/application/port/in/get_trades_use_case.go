package in

import (
	"context"

	"github.com/HungphamLeo/BackendDataPlatform/internal/marketdata/application/dto"
)

// GetTradesUseCase định nghĩa port đầu vào cho use case lấy dữ liệu giao dịch
type GetTradesUseCase interface {
	// Execute thực thi use case
	Execute(ctx context.Context, request *dto.GetTradesRequest) (*dto.GetTradesResponse, error)
}

// GetHistoricalTradesUseCase định nghĩa port đầu vào cho use case lấy dữ liệu giao dịch lịch sử
type GetHistoricalTradesUseCase interface {
	// Execute thực thi use case
	Execute(ctx context.Context, request *dto.GetHistoricalTradesRequest) (*dto.GetHistoricalTradesResponse, error)
}

// StreamTradesUseCase định nghĩa port đầu vào cho use case stream dữ liệu giao dịch
type StreamTradesUseCase interface {
	// Execute thực thi use case
	Execute(ctx context.Context, request *dto.StreamTradesRequest) (<-chan *dto.StreamTradesResponse, error)
}
