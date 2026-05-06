package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/HungphamLeo/BackendDataPlatform/internal/marketdata/application/dto"
	"github.com/HungphamLeo/BackendDataPlatform/internal/marketdata/application/port/in"
	pb "github.com/HungphamLeo/BackendDataPlatform/api/proto/marketdata"
	"github.com/HungphamLeo/BackendDataPlatform/internal/platform/logging"
)

// TickerService triển khai gRPC service cho ticker
type TickerService struct {
	pb.UnimplementedTickerServiceServer
	getTickerUseCase  in.GetTickerUseCase
	getTickersUseCase in.GetTickersUseCase
	logger            logging.Logger
}

// NewTickerService tạo mới TickerService
func NewTickerService(
	getTickerUseCase in.GetTickerUseCase,
	getTickersUseCase in.GetTickersUseCase,
	logger logging.Logger,
) *TickerService {
	return &TickerService{
		getTickerUseCase:  getTickerUseCase,
		getTickersUseCase: getTickersUseCase,
		logger:            logger,
	}
}

// GetTicker xử lý request lấy dữ liệu ticker
func (s *TickerService) GetTicker(ctx context.Context, req *pb.GetTickerRequest) (*pb.GetTickerResponse, error) {
	// Tạo request DTO
	request := &dto.GetTickerRequest{
		Symbol:   req.Symbol,
		Exchange: req.Exchange,
	}
	
	// Gọi use case
	response, err := s.getTickerUseCase.Execute(ctx, request)
	if err != nil {
		s.logger.Error("Failed to get ticker", logging.Fields{
			"error":    err.Error(),
			"symbol":   req.Symbol,
			"exchange": req.Exchange,
		})
		
		// Trả về lỗi
		return nil, status.Errorf(codes.Internal, "failed to get ticker: %v", err)
	}
	
	// Chuyển đổi từ DTO sang protobuf
	ticker := &pb.Ticker{
		Symbol:      response.Ticker.Symbol,
		Exchange:    response.Ticker.Exchange,
		Price:       response.Ticker.Price,
		Bid:         response.Ticker.Bid,
		Ask:         response.Ticker.Ask,
		BidSize:     response.Ticker.BidSize,
		AskSize:     response.Ticker.AskSize,
		High24H:     response.Ticker.High24h,
		Low24H:      response.Ticker.Low24h,
		Volume24H:   response.Ticker.Volume24h,
		QuoteVolume: response.Ticker.QuoteVolume,
		Timestamp:   response.Ticker.Timestamp.UnixNano(),
	}
	
	// Trả về kết quả
	return &pb.GetTickerResponse{
		Ticker: ticker,
	}, nil
}

// GetTickers xử lý request lấy dữ liệu ticker cho nhiều symbol
func (s *TickerService) GetTickers(ctx context.Context, req *pb.GetTickersRequest) (*pb.GetTickersResponse, error) {
	// Kiểm tra tham số
	if len(req.Symbols) == 0 {
		return nil, status.Error(codes.InvalidArgument, "at least one symbol is required")
	}
	
	// Tạo request DTO
	request := &dto.GetTickersRequest{
		Symbols:  req.Symbols,
		Exchange: req.Exchange,
	}
	
	// Gọi use case
	response, err := s.getTickersUseCase.Execute(ctx, request)
	if err != nil {
		s.logger.Error("Failed to get tickers", logging.Fields{
			"error":    err.Error(),
			"symbols":  req.Symbols,
			"exchange": req.Exchange,
		})
		
		// Trả về lỗi
		return nil, status.Errorf(codes.Internal, "failed to get tickers: %v", err)
	}
	
	// Chuyển đổi từ DTO sang protobuf
	tickers := make([]*pb.Ticker, len(response.Tickers))
	for i, tickerDTO := range response.Tickers {
		tickers[i] = &pb.Ticker{
			Symbol:      tickerDTO.Symbol,
			Exchange:    tickerDTO.Exchange,
			Price:       tickerDTO.Price,
			Bid:         tickerDTO.Bid,
			Ask:         tickerDTO.Ask,
			BidSize:     tickerDTO.BidSize,
			AskSize:     tickerDTO.AskSize,
			High24H:     tickerDTO.High24h,
			Low24H:      tickerDTO.Low24h,
			Volume24H:   tickerDTO.Volume24h,
			QuoteVolume: tickerDTO.QuoteVolume,
			Timestamp:   tickerDTO.Timestamp.UnixNano(),
		}
	}
	
	// Trả về kết quả
	return &pb.GetTickersResponse{
		Tickers: tickers,
	}, nil
}

// StreamTicker xử lý request stream dữ liệu ticker
func (s *TickerService) StreamTicker(req *pb.StreamTickerRequest, stream pb.TickerService_StreamTickerServer) error {
	// Tạo request DTO
	request := &dto.GetTickerRequest{
		Symbol:   req.Symbol,
		Exchange: req.Exchange,
	}
	
	// Gọi use case để lấy channel dữ liệu
	tickerCh, err := s.getTickerUseCase.Execute(stream.Context(), request)
	if err != nil {
		s.logger.Error("Failed to stream ticker", logging.Fields{
			"error":    err.Error(),
			"symbol":   req.Symbol,
			"exchange": req.Exchange,
		})
		
		// Trả về lỗi
		return status.Errorf(codes.Internal, "failed to stream ticker: %v", err)
	}
	
	// Lặp qua channel và gửi dữ liệu đến client
	for {
		select {
		case <-stream.Context().Done():
			return nil
		case tickerResponse, ok := <-tickerCh:
			if !ok {
				return nil
			}
			
			// Chuyển đổi từ DTO sang protobuf
			ticker := &pb.Ticker{
				Symbol:      tickerResponse.Ticker.Symbol,
				Exchange:    tickerResponse.Ticker.Exchange,
				Price:       tickerResponse.Ticker.Price,
				Bid:         tickerResponse.Ticker.Bid,
				Ask:         tickerResponse.Ticker.Ask,
				BidSize:     tickerResponse.Ticker.BidSize,
				AskSize:     tickerResponse.Ticker.AskSize,
				High24H:     tickerResponse.Ticker.High24h,
				Low24H:      tickerResponse.Ticker.Low24h,
				Volume24H:   tickerResponse.Ticker.Volume24h,
				QuoteVolume: tickerResponse.Ticker.QuoteVolume,
				Timestamp:   tickerResponse.Ticker.Timestamp.UnixNano(),
			}
			
			// Gửi dữ liệu đến client
			if err := stream.Send(&pb.StreamTickerResponse{
				Ticker: ticker,
			}); err != nil {
				s.logger.Error("Failed to send ticker to client", logging.Fields{
					"error":    err.Error(),
					"symbol":   req.Symbol,
					"exchange": req.Exchange,
				})
				return err
			}
		}
	}
}
