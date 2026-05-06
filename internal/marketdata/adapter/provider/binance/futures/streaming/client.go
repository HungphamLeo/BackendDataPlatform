package streaming

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/HungphamLeo/BackendDataPlatform/internal/marketdata/domain"
	"github.com/HungphamLeo/BackendDataPlatform/internal/platform/logging"
)

const maxRetries = 5

// BinanceFuturesClient triển khai domain.MarketDataProvider
type BinanceFuturesClient struct {
	wsURL   string
	conn *websocket.Conn
	logger  logging.Logger
	symbols []string
}

func NewBinanceFuturesClient(wsURL string, logger logging.Logger) *BinanceFuturesClient {
	if wsURL == "" {
		wsURL = "wss://fstream.binance.com/ws" // Default fallback, nhưng thực tế sẽ lấy từ config YAML
	}
	return &BinanceFuturesClient{
		wsURL:  wsURL,
		logger: logger,
	}
}

func (c *BinanceFuturesClient) Connect(ctx context.Context) error {
	// FIXME: Ở Production cần add Transport layer logic (dialer timeout, TLS setup)
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, c.wsURL, nil)
	if err != nil {
		return fmt.Errorf("binance futures ws dial error: %w", err)
	}
	c.conn = conn
	c.logger.Info("Connected to Binance Futures WebSocket")
	return nil
}

func (c *BinanceFuturesClient) Subscribe(symbols []string) error {
	c.symbols = symbols // Lưu lại để dùng khi reconnect
	var streams []string
	for _, s := range symbols {
		streams = append(streams, FuturesStreamAggTrade(s)) // Tái sử dụng generator có sẵn
	}

	// Chuyển params []string thành []any để dùng chung cấu trúc Request
	params := make([]any, len(streams))
	for i, v := range streams {
		params[i] = v
	}

	req := FuturesWSRequest{
		Method: futWSMethodSubscribe,
		Params: params,
		ID:     uint64(time.Now().UnixNano()), // Randomize Request ID
	}

	return c.conn.WriteJSON(req)
}

func (c *BinanceFuturesClient) Stream(ctx context.Context, tickCh chan<- *domain.Tick, errCh chan<- error) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			_, message, err := c.conn.ReadMessage()
			if err != nil {
				c.logger.Warn(fmt.Sprintf("Websocket read error: %v. Attempting to reconnect...", err))
				if reconnErr := c.reconnect(ctx); reconnErr != nil {
					errCh <- fmt.Errorf("failed to reconnect: %w", reconnErr)
					return
				}
				continue
			}

			// Cắt json chỉ lấy field "e" để định tuyến cho nhanh
			var baseEvent struct {
				EventType string `json:"e"`
			}
			if err := json.Unmarshal(message, &baseEvent); err != nil {
				continue
			}

			if baseEvent.EventType == "aggTrade" {
				var aggTrade FuturesAggTradeEvent
				if err := json.Unmarshal(message, &aggTrade); err == nil {
					price, _ := strconv.ParseFloat(aggTrade.Price, 64)
					qty, _ := strconv.ParseFloat(aggTrade.Qty, 64)

					tickCh <- &domain.Tick{
						Symbol:    aggTrade.Symbol,
						Price:     price,
						Quantity:  qty,
						Timestamp: time.UnixMilli(aggTrade.TradeTime),
						Source:    "BINANCE_FUTURES",
					}
				}
			}
		}
	}
}

func (c *BinanceFuturesClient) reconnect(ctx context.Context) error {
	for i := 0; i < maxRetries; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		time.Sleep(time.Duration(math.Pow(2, float64(i))) * time.Second) // Exponential backoff

		c.logger.Info(fmt.Sprintf("Reconnecting... Attempt %d/%d", i+1, maxRetries))
		
		_ = c.Close()
		if err := c.Connect(ctx); err == nil {
			if err := c.Subscribe(c.symbols); err == nil { // Resubscribe lại những cặp coin đang nghe
				c.logger.Info("Reconnected and resubscribed successfully")
				return nil
			}
		}
	}
	return fmt.Errorf("exceeded max retries for websocket reconnection")
}

func (c *BinanceFuturesClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}