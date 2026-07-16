// Package marketdata provides stub Go types for the market data gRPC service.
// These stubs should be replaced with generated code from protoc/buf after
// running: make proto-go
//
// Stub is intentionally minimal — just enough to make the existing
// internal/marketdata/transport/grpc code compile.
package marketdata

import (
	"context"

	"google.golang.org/grpc"
)

// --- Message types ---

// Ticker holds real-time ticker data.
type Ticker struct {
	Symbol      string  `json:"symbol"`
	Exchange    string  `json:"exchange"`
	Price       float64 `json:"price"`
	Bid         float64 `json:"bid"`
	Ask         float64 `json:"ask"`
	BidSize     float64 `json:"bid_size"`
	AskSize     float64 `json:"ask_size"`
	High24H     float64 `json:"high_24h"`
	Low24H      float64 `json:"low_24h"`
	Volume24H   float64 `json:"volume_24h"`
	QuoteVolume float64 `json:"quote_volume"`
	Timestamp   int64   `json:"timestamp"`
}

// GetTickerRequest is the request for GetTicker RPC.
type GetTickerRequest struct {
	Symbol   string `json:"symbol"`
	Exchange string `json:"exchange"`
}

// GetTickerResponse is the response for GetTicker RPC.
type GetTickerResponse struct {
	Ticker *Ticker `json:"ticker"`
}

// GetTickersRequest is the request for GetTickers RPC.
type GetTickersRequest struct {
	Symbols  []string `json:"symbols"`
	Exchange string   `json:"exchange"`
}

// GetTickersResponse is the response for GetTickers RPC.
type GetTickersResponse struct {
	Tickers []*Ticker `json:"tickers"`
}

// StreamTickerRequest is the request for StreamTicker RPC.
type StreamTickerRequest struct {
	Symbol   string `json:"symbol"`
	Exchange string `json:"exchange"`
}

// StreamTickerResponse is the response message for StreamTicker server-streaming RPC.
type StreamTickerResponse struct {
	Ticker *Ticker `json:"ticker"`
}

// --- gRPC server interface ---

// TickerServiceServer is the server API for TickerService service.
type TickerServiceServer interface {
	GetTicker(context.Context, *GetTickerRequest) (*GetTickerResponse, error)
	GetTickers(context.Context, *GetTickersRequest) (*GetTickersResponse, error)
	StreamTicker(*StreamTickerRequest, TickerService_StreamTickerServer) error
	mustEmbedUnimplementedTickerServiceServer()
}

// UnimplementedTickerServiceServer must be embedded to have forward-compatible implementations.
type UnimplementedTickerServiceServer struct{}

func (UnimplementedTickerServiceServer) GetTicker(context.Context, *GetTickerRequest) (*GetTickerResponse, error) {
	return nil, nil
}
func (UnimplementedTickerServiceServer) GetTickers(context.Context, *GetTickersRequest) (*GetTickersResponse, error) {
	return nil, nil
}
func (UnimplementedTickerServiceServer) StreamTicker(*StreamTickerRequest, TickerService_StreamTickerServer) error {
	return nil
}
func (UnimplementedTickerServiceServer) mustEmbedUnimplementedTickerServiceServer() {}

// TickerService_StreamTickerServer is the interface for the server side of streaming.
type TickerService_StreamTickerServer interface {
	Send(*StreamTickerResponse) error
	Context() context.Context
	grpc.ServerStream
}

// RegisterTickerServiceServer registers the service with the gRPC server.
func RegisterTickerServiceServer(s *grpc.Server, srv TickerServiceServer) {
	// TODO: replace with generated code after running make proto-go
}
