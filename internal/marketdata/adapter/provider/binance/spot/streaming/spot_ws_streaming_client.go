package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type SpotWSOptions struct {
	BaseURL        string
	UseCombined    bool
	InitialStreams []string
	DialTimeout    time.Duration
	Header         http.Header

	// Spot limit: 5 incoming messages/s (including ping/pong)
	// default safe: 4
	ControlMsgsPerSec int
	PongDeadline      time.Duration
}

type SpotWSClient struct {
	core *WSCore
}

func NewSpotWSClient(opts SpotWSOptions) *SpotWSClient {
	if opts.BaseURL == "" {
		opts.BaseURL = "wss://stream.binance.com:9443"
	}
	if opts.ControlMsgsPerSec <= 0 {
		opts.ControlMsgsPerSec = 4
	}
	core := NewWSCore(WSCoreOptions{
		BaseURL:           opts.BaseURL,
		UseCombined:       opts.UseCombined,
		InitialStreams:    opts.InitialStreams,
		DialTimeout:       opts.DialTimeout,
		Header:            opts.Header,
		ControlMsgsPerSec: opts.ControlMsgsPerSec,
		PongDeadline:      opts.PongDeadline,
		ValidateID:        nil, // spot allows int/string/null
	})
	return &SpotWSClient{core: core}
}

func (c *SpotWSClient) Connect(ctx context.Context) error { return c.core.Connect(ctx) }
func (c *SpotWSClient) Close() error                      { return c.core.Close() }

func (c *SpotWSClient) Subscribe(ctx context.Context, id any, streams ...string) (<-chan json.RawMessage, error) {
	return c.core.Subscribe(ctx, id, "SUBSCRIBE", streams...)
}

func (c *SpotWSClient) Unsubscribe(ctx context.Context, id any, streams ...string) error {
	return c.core.Unsubscribe(ctx, id, "UNSUBSCRIBE", streams...)
}

func (c *SpotWSClient) ListSubscriptions(ctx context.Context, id any) ([]string, error) {
	resp, err := c.core.Call(ctx, id, map[string]any{"method": "LIST_SUBSCRIPTIONS"})
	if err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, fmt.Errorf("ws error %d: %s", resp.Error.Code, resp.Error.Msg)
	}

	var arr []string
	if len(resp.Result) == 0 || string(resp.Result) == "null" {
		return arr, nil
	}
	if err := json.Unmarshal(resp.Result, &arr); err != nil {
		return nil, err
	}
	return arr, nil
}

func (c *SpotWSClient) SetCombined(ctx context.Context, id any, combined bool) error {
	_, err := c.core.Call(ctx, id, map[string]any{
		"method": "SET_PROPERTY",
		"params": []any{"combined", combined},
	})
	return err
}

func (c *SpotWSClient) GetCombined(ctx context.Context, id any) (bool, error) {
	resp, err := c.core.Call(ctx, id, map[string]any{
		"method": "GET_PROPERTY",
		"params": []any{"combined"},
	})
	if err != nil {
		return false, err
	}
	if resp.Error != nil {
		return false, fmt.Errorf("ws error %d: %s", resp.Error.Code, resp.Error.Msg)
	}
	var v bool
	if len(resp.Result) == 0 {
		return false, nil
	}
	if err := json.Unmarshal(resp.Result, &v); err != nil {
		return false, err
	}
	return v, nil
}
