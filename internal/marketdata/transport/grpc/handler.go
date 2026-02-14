package grpc

import (
	"context"

	pb "api/proto/marketdata/v1"
	"internal/marketdata/application/usecase"
	"internal/marketdata/domain/value_object"
)

// ✅ gRPC SERVICE HANDLER
type MarketDataServiceServer struct {
	pb.UnimplementedMarketDataServiceServer
	getLatestPriceUC *usecase.GetLatestPriceUseCase
}

func NewMarketDataServiceServer(
	getLatestPriceUC *usecase.GetLatestPriceUseCase,
) *MarketDataServiceServer {
	return &MarketDataServiceServer{
		getLatestPriceUC: getLatestPriceUC,
	}
}

// gRPC method: GetLatestPrice
func (s *MarketDataServiceServer) GetLatestPrice(
	ctx context.Context,
	req *pb.GetLatestPriceRequest,
) (*pb.GetLatestPriceResponse, error) {
	// Step 1: Map protobuf to DTO
	useCaseReq := &dto.GetLatestPriceRequest{
		Symbol: req.Symbol,
	}

	// Step 2: Call use case
	useCaseResp, err := s.getLatestPriceUC.Execute(ctx, useCaseReq)
	if err != nil {
		return nil, err
	}

	// Step 3: Map DTO to protobuf response
	return &pb.GetLatestPriceResponse{
		Symbol: useCaseResp.Symbol,
		Bid:    useCaseResp.Bid,
		Ask:    useCaseResp.Ask,
	}, nil
}