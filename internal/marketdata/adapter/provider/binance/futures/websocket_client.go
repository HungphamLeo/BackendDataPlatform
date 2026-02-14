package binance

import (
	"context"
	"time"

	"github.com/gorilla/websocket"
)

type FuturesWS struct {
	baseURL string
	conn    *websocket.Conn
}

func NewFuturesWS(baseURL string) *FuturesWS {
	return &FuturesWS{baseURL: baseURL}
}

func (f *FuturesWS) Connect(ctx context.Context) error {

	conn, _, err := websocket.DefaultDialer.DialContext(
		ctx,
		f.baseURL,
		nil,
	)

	if err != nil {
		return err
	}

	f.conn = conn
	return nil
}

func (f *FuturesWS) SendRequest(id string, method string, params interface{}) error {

	req := map[string]interface{}{
		"id":     id,
		"method": method,
		"params": params,
	}

	return f.conn.WriteJSON(req)
}

func (f *FuturesWS) ReadLoop(ctx context.Context, handler func([]byte)) error {

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			_, msg, err := f.conn.ReadMessage()
			if err != nil {
				return err
			}
			handler(msg)
		}
	}
}

func (f *FuturesWS) Ping() error {
	return f.conn.WriteControl(
		websocket.PingMessage,
		[]byte{},
		time.Now().Add(time.Second),
	)
}

func (f *FuturesWS) Close() error {
	if f.conn != nil {
		return f.conn.Close()
	}
	return nil
}
