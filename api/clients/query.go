package clients

import (
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// QueryClient wraps a gRPC connection to data-platform-query.
// After proto stubs are generated, add the generated client as a field.
type QueryClient struct {
	conn    *grpc.ClientConn
	address string
}

// NewQueryClient dials the data-platform-query gRPC server.
func NewQueryClient(address string) (*QueryClient, error) {
	conn, err := grpc.NewClient(address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("dial query gRPC %s: %w", address, err)
	}
	return &QueryClient{conn: conn, address: address}, nil
}

// Close shuts down the connection.
func (c *QueryClient) Close() error {
	return c.conn.Close()
}

// Conn returns the raw grpc.ClientConn (for generated code).
func (c *QueryClient) Conn() *grpc.ClientConn {
	return c.conn
}
