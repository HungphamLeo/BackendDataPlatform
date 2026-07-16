package clients

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// StreamingClient is a gRPC client wrapper for the MarketDataStream service.
// Once proto stubs are generated, embed the generated client here.
type StreamingClient struct {
	conn    *grpc.ClientConn
	address string
}

// NewStreamingClient dials the streaming gRPC server.
func NewStreamingClient(address string) (*StreamingClient, error) {
	conn, err := grpc.NewClient(address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("dial streaming gRPC %s: %w", address, err)
	}
	return &StreamingClient{conn: conn, address: address}, nil
}

// Close shuts down the gRPC connection.
func (c *StreamingClient) Close() error {
	return c.conn.Close()
}

// Conn exposes the underlying grpc.ClientConn for generated client wrappers.
func (c *StreamingClient) Conn() *grpc.ClientConn {
	return c.conn
}

// TickerSummary is a plain-Go result type returned before generated stubs exist.
type TickerSummary struct {
	Symbol    string
	Exchange  string
	Price     float64
	Volume    float64
	Bid       float64
	Ask       float64
	Timestamp time.Time
}

// GetTicker calls the MarketDataStream.GetTicker RPC.
// Replace this stub with generated code after `make proto`.
func (c *StreamingClient) GetTicker(ctx context.Context, symbol, exchange string) (*TickerSummary, error) {
	// Placeholder: real implementation uses generated client
	return nil, fmt.Errorf("streaming.GetTicker: stubs not yet generated — run `make proto`")
}
