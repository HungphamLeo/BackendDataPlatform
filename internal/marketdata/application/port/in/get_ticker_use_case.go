package in

import (
	"context"

	"github.com/HungphamLeo/BackendDataPlatform/internal/marketdata/application/dto"
)

// GetTickerUseCase định nghĩa port đầu vào cho use case lấy dữ liệu ticker
type GetTickerUseCase interface {
	// Execute thực thi use case
	Execute(ctx context.Context, request *dto.GetTickerRequest) (*dto.GetTickerResponse, error)
}

// GetTickersUseCase định nghĩa port đầu vào cho use case lấy dữ liệu ticker cho nhiều symbol
type GetTickersUseCase interface {
	// Execute thực thi use case
	Execute(ctx context.Context, request *dto.GetTickersRequest) (*dto.GetTickersResponse, error)
}

// StreamTickerUseCase định nghĩa port đầu vào cho use case stream dữ liệu ticker
type StreamTickerUseCase interface {
	// Execute thực thi use case
	Execute(ctx context.Context, request *dto.GetTickerRequest) (<-chan *dto.GetTickerResponse, error)
}
