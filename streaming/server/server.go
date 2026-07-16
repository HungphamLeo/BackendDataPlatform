package server

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/HungphamLeo/BackendDataPlatform/streaming/repository"
)

// MarketDataServer implements the MarketDataStream gRPC interface.
// The generated protobuf types are referenced symbolically here since
// buf/protoc codegen runs separately; integrate by importing the generated package.
type MarketDataServer struct {
	repo *repository.DB
}

// New creates a new MarketDataServer.
func New(repo *repository.DB) *MarketDataServer {
	return &MarketDataServer{repo: repo}
}

// --- helper to convert time.Time → timestamppb.Timestamp

func toProtoTimestamp(t time.Time) *timestamppb.Timestamp {
	return timestamppb.New(t)
}

// GetTicker fetches the latest ticker for symbol+exchange.
// Returns (symbol, exchange, price, volume, bid, ask, timestamp).
func (s *MarketDataServer) GetTicker(ctx context.Context, symbol, exchange string) (float64, float64, float64, float64, time.Time, error) {
	t, err := s.repo.GetLatestTicker(ctx, symbol, exchange)
	if err != nil {
		return 0, 0, 0, 0, time.Time{}, status.Errorf(codes.NotFound, "ticker not found: %v", err)
	}
	return t.Price, t.Volume, t.Bid, t.Ask, t.Timestamp, nil
}

// GetOrderBook fetches the latest orderbook for symbol+exchange.
func (s *MarketDataServer) GetOrderBook(ctx context.Context, symbol, exchange string) (string, string, time.Time, error) {
	ob, err := s.repo.GetLatestOrderBook(ctx, symbol, exchange)
	if err != nil {
		return "", "", time.Time{}, status.Errorf(codes.NotFound, "orderbook not found: %v", err)
	}
	return ob.BidsJSON, ob.AsksJSON, ob.Timestamp, nil
}

// GetKlines returns klines within the given time range.
func (s *MarketDataServer) GetKlines(
	ctx context.Context,
	symbol, exchange, interval string,
	from, to time.Time,
	limit int,
) ([]KlinePoint, error) {
	klines, err := s.repo.GetKlines(ctx, symbol, exchange, interval, from, to, limit)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "query klines: %v", err)
	}
	points := make([]KlinePoint, len(klines))
	for i, k := range klines {
		points[i] = KlinePoint{
			Symbol:    k.Symbol,
			Exchange:  k.Exchange,
			Interval:  k.Interval,
			Open:      k.Open,
			High:      k.High,
			Low:       k.Low,
			Close:     k.Close,
			Volume:    k.Volume,
			Timestamp: k.Timestamp,
		}
	}
	return points, nil
}

// KlinePoint is a DTO returned by GetKlines.
type KlinePoint struct {
	Symbol, Exchange, Interval string
	Open, High, Low, Close     float64
	Volume                     float64
	Timestamp                  time.Time
}

// validateSymbolExchange checks that required fields are provided.
func validateSymbolExchange(symbol, exchange string) error {
	if symbol == "" || exchange == "" {
		return fmt.Errorf("symbol and exchange are required")
	}
	return nil
}
