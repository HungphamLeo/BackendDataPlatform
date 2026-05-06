package service

import (
	"context"
	"fmt"
	"time"

	"github.com/HungphamLeo/BackendDataPlatform/internal/marketdata/application/dto"
	"github.com/HungphamLeo/BackendDataPlatform/internal/marketdata/application/port/in"
	"github.com/HungphamLeo/BackendDataPlatform/internal/marketdata/application/port/out"
	"github.com/HungphamLeo/BackendDataPlatform/internal/marketdata/domain/value_object"
)

// GetTickerService triển khai use case lấy dữ liệu ticker
type GetTickerService struct {
	tickerRepository out.TickerRepositoryPort
	marketDataProvider out.MarketDataProviderPort
}

// NewGetTickerService tạo mới GetTickerService
func NewGetTickerService(
	tickerRepository out.TickerRepositoryPort,
	marketDataProvider out.MarketDataProviderPort,
) *GetTickerService {
	return &GetTickerService{
		tickerRepository: tickerRepository,
		marketDataProvider: marketDataProvider,
	}
}

// Đảm bảo GetTickerService triển khai interface GetTickerUseCase
var _ in.GetTickerUseCase = (*GetTickerService)(nil)

// Execute thực thi use case
func (s *GetTickerService) Execute(ctx context.Context, request *dto.GetTickerRequest) (*dto.GetTickerResponse, error) {
	// Chuyển đổi symbol từ string sang value object
	symbol, err := value_object.NewSymbol(request.Symbol)
	if err != nil {
		return nil, fmt.Errorf("invalid symbol: %w", err)
	}
	
	// Tìm ticker trong repository
	ticker, err := s.tickerRepository.FindBySymbol(ctx, symbol, request.Exchange)
	if err != nil {
		// Nếu không tìm thấy trong repository, lấy từ provider
		quote, err := s.marketDataProvider.GetTicker(ctx, symbol)
		if err != nil {
			return nil, fmt.Errorf("failed to get ticker: %w", err)
		}
		
		// Chuyển đổi từ domain model sang DTO
		tickerDTO := dto.TickerDTO{
			Symbol:    symbol.String(),
			Exchange:  request.Exchange,
			Price:     quote.Last,
			Bid:       quote.Bid,
			Ask:       quote.Ask,
			BidSize:   quote.BidSize,
			AskSize:   quote.AskSize,
			Timestamp: time.Now(),
		}
		
		return &dto.GetTickerResponse{
			Ticker: tickerDTO,
		}, nil
	}
	
	// Chuyển đổi từ domain model sang DTO
	tickerDTO := dto.TickerDTO{
		Symbol:      ticker.Symbol.String(),
		Exchange:    ticker.Exchange,
		Price:       ticker.LastPrice,
		Bid:         ticker.BestBid,
		Ask:         ticker.BestAsk,
		BidSize:     ticker.BestBidSize,
		AskSize:     ticker.BestAskSize,
		High24h:     ticker.High24h,
		Low24h:      ticker.Low24h,
		Volume24h:   ticker.Volume24h,
		QuoteVolume: ticker.QuoteVolume,
		Timestamp:   ticker.UpdatedAt,
	}
	
	return &dto.GetTickerResponse{
		Ticker: tickerDTO,
	}, nil
}
