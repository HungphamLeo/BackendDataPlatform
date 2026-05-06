package in

import (
	"context"

	"github.com/HungphamLeo/BackendDataPlatform/internal/marketdata/application/dto"
)

// GetCandlesUseCase định nghĩa port đầu vào cho use case lấy dữ liệu nến
type GetCandlesUseCase interface {
	// Execute thực thi use case
	Execute(ctx context.Context, request *dto.GetCandlesRequest) (*dto.GetCandlesResponse, error)
}

// StreamCandlesUseCase định nghĩa port đầu vào cho use case stream dữ liệu nến
type StreamCandlesUseCase interface {
	// Execute thực thi use case
	Execute(ctx context.Context, request *dto.StreamCandlesRequest) (<-chan *dto.StreamCandlesResponse, error)
}
