package streaming

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type FuturesWSOptions struct {
	BaseURL        string
	UseCombined    bool
	InitialStreams []string
	DialTimeout    time.Duration
	Header         http.Header

	// Futures limit: 10 incoming messages/s
	// default safe: 8
	ControlMsgsPerSec int
	PongDeadline      time.Duration
}

// Futures spec: id must be unsigned int. We'll accept uint64 only.
func validateFuturesID(id any) error {
	if id == nil {
		return errors.New("futures ws id must be uint64 (non-nil)")
	}
	switch id.(type) {
	case uint64:
		return nil
	default:
		return fmt.Errorf("futures ws id must be uint64, got %T", id)
	}
}

type FuturesWSClient struct {
	core *WSCore
}

func NewFuturesWSClient(opts FuturesWSOptions) *FuturesWSClient {
	if opts.BaseURL == "" {
		opts.BaseURL = "wss://fstream.binance.com"
	}
	if opts.ControlMsgsPerSec <= 0 {
		opts.ControlMsgsPerSec = 8
	}

	core := NewWSCore(WSCoreOptions{
		BaseURL:           opts.BaseURL,
		UseCombined:       opts.UseCombined,
		InitialStreams:    opts.InitialStreams,
		DialTimeout:       opts.DialTimeout,
		Header:            opts.Header,
		ControlMsgsPerSec: opts.ControlMsgsPerSec,
		PongDeadline:      opts.PongDeadline,
		ValidateID:        validateFuturesID,
	})

	return &FuturesWSClient{core: core}
}

func (c *FuturesWSClient) Connect(ctx context.Context) error { return c.core.Connect(ctx) }
func (c *FuturesWSClient) Close() error                      { return c.core.Close() }

// Futures wrapper methods accept uint64 directly (cleaner)
func (c *FuturesWSClient) Subscribe(ctx context.Context, id uint64, streams ...string) (<-chan json.RawMessage, error) {
	return c.core.Subscribe(ctx, id, "SUBSCRIBE", streams...)
}

func (c *FuturesWSClient) Unsubscribe(ctx context.Context, id uint64, streams ...string) error {
	return c.core.Unsubscribe(ctx, id, "UNSUBSCRIBE", streams...)
}

func (c *FuturesWSClient) ListSubscriptions(ctx context.Context, id uint64) ([]string, error) {
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

func (c *FuturesWSClient) SetCombined(ctx context.Context, id uint64, combined bool) error {
	_, err := c.core.Call(ctx, id, map[string]any{
		"method": "SET_PROPERTY",
		"params": []any{"combined", combined},
	})
	return err
}

func (c *FuturesWSClient) GetCombined(ctx context.Context, id uint64) (bool, error) {
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
