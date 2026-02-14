package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	"internal/domain"
)

type SpotClient struct {

	baseURL string

	conn *websocket.Conn
}

func (c *SpotClient) Connect() error {

	conn, _, err :=
		websocket.DefaultDialer.Dial(
			c.baseURL,
			nil,
		)

	if err != nil {
		return err
	}

	c.conn = conn

	return nil
}

func (c *SpotClient) SubscribeBookTicker(
	symbols []string,
) error {

	streams :=
		make([]string, 0)

	for _, s := range symbols {

		streams =
			append(
				streams,
				strings.ToLower(s)+"@bookTicker",
			)
	}

	req := map[string]interface{}{

		"id": time.Now().Unix(),

		"method": "SUBSCRIBE",

		"params": streams,
	}

	return c.conn.WriteJSON(req)
}
package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type SpotWS struct {
	baseURL string
	conn    *websocket.Conn
}

func NewSpotWS(baseURL string) *SpotWS {
	return &SpotWS{baseURL: baseURL}
}

func (s *SpotWS) Connect(ctx context.Context) error {

	conn, _, err := websocket.DefaultDialer.DialContext(
		ctx,
		s.baseURL,
		nil,
	)

	if err != nil {
		return err
	}

	s.conn = conn
	return nil
}

func (s *SpotWS) SubscribeBookTicker(symbols []string) error {

	params := make([]string, 0)

	for _, sym := range symbols {
		params = append(
			params,
			strings.ToLower(sym)+"@bookTicker",
		)
	}

	req := map[string]interface{}{
		"id":     time.Now().Unix(),
		"method": "SUBSCRIBE",
		"params": params,
	}

	return s.conn.WriteJSON(req)
}

func (s *SpotWS) ReadLoop(ctx context.Context, handler func([]byte)) error {

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			_, msg, err := s.conn.ReadMessage()
			if err != nil {
				return err
			}
			handler(msg)
		}
	}
}

func (s *SpotWS) Close() error {
	if s.conn != nil {
		return s.conn.Close()
	}
	return nil
}
