package xtb

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
)

// Credentials holds login parameters for XTB.
type Credentials struct {
	UserID  string
	Password string
	AppName string
}

// XTBClient is the minimal interface consumed by the market ingestor.
// Keeping it small helps preserve clean boundaries (app depends on ports).
type XTBClient interface {
	Connect(ctx context.Context) error
	Close() error
	Receive() (<-chan string, error)
	Send(obj any) error

	// Convenience subscription helpers (mirrors xtb_connection.txt)
	SubscribePrice(symbol string) error
	SubscribeBalance() error
	SubscribeTradeStatus() error
	SubscribeNews() error
	SubscribeProfits() error
	UnsubscribePrice(symbol string) error
	UnsubscribeBalance() error
	UnsubscribeTradeStatus() error
	UnsubscribeNews() error
	UnsubscribeProfits() error
}

// Client implements both the API socket and streaming socket.
//
// Behavior mirrors xtb_connection.txt:
// - connect API socket
// - execute login -> streamSessionId
// - connect streaming socket
// - send subscription commands to streaming socket
// - continuously decode incoming JSON objects
//
// Note: this is intentionally a low-dependency client to keep the repo portable.
type Client struct {
	cfg   Config
	creds Credentials

	api    *jsonSocket
	stream *jsonSocket

	mu             sync.RWMutex
	streamSessionId string
	connected      bool

	recvOnce sync.Once
	recvCh   chan string
	doneCh   chan struct{}
}

func NewClient(cfg Config, creds Credentials) *Client {
	return &Client{
		cfg:    cfg,
		creds:  creds,
		api:    newJSONSocket(cfg, cfg.APIPort),
		stream: newJSONSocket(cfg, cfg.StreamingPort),
		recvCh: make(chan string, 1024),
		doneCh: make(chan struct{}),
	}
}

func (c *Client) Connect(ctx context.Context) error {
	// 1) Connect API socket
	if err := c.api.Connect(ctx); err != nil {
		return err
	}

	// 2) Login to get streamSessionId
	ssid, err := c.login(ctx)
	if err != nil {
		_ = c.api.Close()
		return err
	}

	c.mu.Lock()
	c.streamSessionId = ssid
	c.mu.Unlock()

	// 3) Connect streaming socket
	if err := c.stream.Connect(ctx); err != nil {
		_ = c.api.Close()
		return err
	}

	// 4) Start read loop (once)
	c.recvOnce.Do(func() {
		go c.readStreamLoop()
	})

	c.mu.Lock()
	c.connected = true
	c.mu.Unlock()

	return nil
}

func (c *Client) Close() error {
	c.mu.Lock()
	if !c.connected {
		c.mu.Unlock()
		_ = c.api.Close()
		_ = c.stream.Close()
		return nil
	}
	c.connected = false
	c.mu.Unlock()

	close(c.doneCh)
	_ = c.api.Close()
	_ = c.stream.Close()
	return nil
}

func (c *Client) Receive() (<-chan string, error) {
	c.mu.RLock()
	ok := c.connected
	c.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("xtb client not connected")
	}
	return c.recvCh, nil
}

func (c *Client) Send(obj any) error {
	c.mu.RLock()
	ok := c.connected
	c.mu.RUnlock()
	if !ok {
		return fmt.Errorf("xtb client not connected")
	}
	return c.stream.Send(obj)
}

// --- API helpers ---

type apiCommand struct {
	Command   string         `json:"command"`
	Arguments map[string]any `json:"arguments,omitempty"`
}

func (c *Client) executeAPI(cmd string, args map[string]any) (map[string]any, error) {
	req := apiCommand{Command: cmd, Arguments: args}
	if err := c.api.Send(req); err != nil {
		return nil, err
	}
	return c.api.ReadOne()
}

func (c *Client) login(ctx context.Context) (string, error) {
	if c.creds.UserID == "" || c.creds.Password == "" {
		return "", fmt.Errorf("missing XTB credentials (UserID/Password)")
	}
	resp, err := c.executeAPI("login", map[string]any{
		"userId":  c.creds.UserID,
		"password": c.creds.Password,
		"appName":  c.creds.AppName,
	})
	if err != nil {
		return "", err
	}
	// expected: {"status": true, "streamSessionId": "...", ...}
	status, _ := resp["status"].(bool)
	if !status {
		return "", fmt.Errorf("xtb login failed: %v", resp)
	}
	ssid, _ := resp["streamSessionId"].(string)
	if ssid == "" {
		return "", fmt.Errorf("xtb login missing streamSessionId: %v", resp)
	}
	return ssid, nil
}

func (c *Client) readStreamLoop() {
	for {
		select {
		case <-c.doneCh:
			return
		default:
			msg, err := c.stream.ReadOne()
			if err != nil {
				// On decode/connection error, stop emitting.
				return
			}
			b, _ := json.Marshal(msg)
			select {
			case c.recvCh <- string(b):
			case <-c.doneCh:
				return
			}
		}
	}
}

// --- Subscriptions (stream socket) ---

func (c *Client) ssid() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.streamSessionId
}

func (c *Client) SubscribePrice(symbol string) error {
	return c.Send(map[string]any{"command": StreamGetTickPrices, "symbol": symbol, "streamSessionId": c.ssid()})
}

func (c *Client) SubscribeBalance() error {
	return c.Send(map[string]any{"command": StreamGetBalance, "streamSessionId": c.ssid()})
}

func (c *Client) SubscribeTradeStatus() error {
	return c.Send(map[string]any{"command": StreamGetTradeStatus, "streamSessionId": c.ssid()})
}

func (c *Client) SubscribeNews() error {
	return c.Send(map[string]any{"command": StreamGetNews, "streamSessionId": c.ssid()})
}

func (c *Client) SubscribeProfits() error {
	return c.Send(map[string]any{"command": StreamGetProfits, "streamSessionId": c.ssid()})
}

func (c *Client) UnsubscribePrice(symbol string) error {
	return c.Send(map[string]any{"command": StreamStopTickPrices, "symbol": symbol, "streamSessionId": c.ssid()})
}

func (c *Client) UnsubscribeBalance() error {
	return c.Send(map[string]any{"command": StreamStopBalance, "streamSessionId": c.ssid()})
}

func (c *Client) UnsubscribeTradeStatus() error {
	return c.Send(map[string]any{"command": StreamStopTradeStatus, "streamSessionId": c.ssid()})
}

func (c *Client) UnsubscribeNews() error {
	return c.Send(map[string]any{"command": StreamStopNews, "streamSessionId": c.ssid()})
}

func (c *Client) UnsubscribeProfits() error {
	return c.Send(map[string]any{"command": StreamStopProfits, "streamSessionId": c.ssid()})
}
