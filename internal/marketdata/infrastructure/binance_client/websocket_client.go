package binance_client

import (
	"context"
	"fmt"
	"time"

	"internal/marketdata/domain/value_object"
	"internal/platform/observability"

	"github.com/gorilla/websocket"
)

// ✅ binance WEBSOCKET ADAPTER - Stream real-time prices
type WebSocketClient struct {
	conn      *websocket.Conn
	serverURL string
	apiKey    string
	logger    observability.Logger
	metrics   observability.Metrics
}

func NewWebSocketClient(
	serverURL string,
	apiKey string,
	logger observability.Logger,
	metrics observability.Metrics,
) *WebSocketClient {
	return &WebSocketClient{
		serverURL: serverURL,
		apiKey:    apiKey,
		logger:    logger,
		metrics:   metrics,
	}
}

// StreamPrices - Connect to binance WebSocket and stream prices
func (c *WebSocketClient) StreamPrices(
	ctx context.Context,
	symbols []value_object.Symbol,
) (<-chan *value_object.Quote, error) {
	// ✅ Connect to WebSocket (streaming data)
	conn, _, err := websocket.DefaultDialer.Dial(c.serverURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to WebSocket: %w", err)
	}

	c.conn = conn
	c.logger.Info(ctx, "Connected to binance WebSocket", "url", c.serverURL)

	// Subscribe to symbols
	for _, symbol := range symbols {
		subscribeMsg := map[string]interface{}{
			"command": "subscribe",
			"symbol":  symbol.String(),
		}
		_ = c.conn.WriteJSON(subscribeMsg)
	}

	// Create channel for price updates
	priceChan := make(chan *value_object.Quote)

	// Background goroutine to read from WebSocket
	go func() {
		defer close(priceChan)
		defer c.conn.Close()

		for {
			select {
			case <-ctx.Done():
				return
			default:
				var msg map[string]interface{}
				if err := c.conn.ReadJSON(&msg); err != nil {
					c.logger.Error(ctx, "WebSocket read error", err)
					return
				}

				// ✅ Map WebSocket message to domain value object
				if bid, ok := msg["bid"].(float64); ok {
					if ask, ok := msg["ask"].(float64); ok {
						quote := &value_object.Quote{
							Bid:       bid,
							Ask:       ask,
							Spread:    ask - bid,
							Timestamp: time.Now().UnixNano(),
						}
						priceChan <- quote

						// Record metric
						c.metrics.RecordStreamData("binance", 1)
					}
				}
			}
		}
	}()

	return priceChan, nil
}
