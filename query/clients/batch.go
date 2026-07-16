package clients

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// BatchClient is a gRPC client wrapper for the BatchData service (Python server).
type BatchClient struct {
	conn    *grpc.ClientConn
	address string
}

// NewBatchClient dials the batch gRPC server.
func NewBatchClient(address string) (*BatchClient, error) {
	conn, err := grpc.NewClient(address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("dial batch gRPC %s: %w", address, err)
	}
	return &BatchClient{conn: conn, address: address}, nil
}

// Close shuts down the gRPC connection.
func (c *BatchClient) Close() error {
	return c.conn.Close()
}

// Conn exposes the underlying grpc.ClientConn for generated client wrappers.
func (c *BatchClient) Conn() *grpc.ClientConn {
	return c.conn
}

// OHLCVBar is a plain-Go DTO for batch OHLCV data.
type OHLCVBar struct {
	Symbol, Exchange, Interval string
	Open, High, Low, Close     float64
	Volume                     float64
	Timestamp                  time.Time
}

// IndicatorPoint is a plain-Go DTO for batch indicator data.
type IndicatorPoint struct {
	Symbol, Exchange, Interval string
	IndicatorName              string
	Value                      float64
	MetaJSON                   string
	Timestamp                  time.Time
}

// GetOHLCV calls BatchData.GetOHLCV.
func (c *BatchClient) GetOHLCV(ctx context.Context, symbol, exchange, interval string, from, to time.Time, limit int) ([]OHLCVBar, error) {
	// Placeholder — replace with generated client after `make proto`
	return nil, fmt.Errorf("batch.GetOHLCV: stubs not yet generated — run `make proto`")
}

// GetIndicator calls BatchData.GetIndicator.
func (c *BatchClient) GetIndicator(ctx context.Context, symbol, exchange, interval, indicatorName string, from, to time.Time, limit int) ([]IndicatorPoint, error) {
	return nil, fmt.Errorf("batch.GetIndicator: stubs not yet generated — run `make proto`")
}
