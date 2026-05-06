package binance

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type WSCoreOptions struct {
	BaseURL string

	// true: /stream?streams=a/b/c (combined)
	// false: /ws/<stream> (raw)
	UseCombined bool
	InitialStreams []string

	// Dial timeout
	DialTimeout time.Duration
	Header      http.Header

	// Control messages limiter (SUB/UNSUB/LIST/SET/GET)
	ControlMsgsPerSec int

	// Ping handler settings: when ping received, reply with pong + copy payload.
	// Default deadline for WriteControl
	PongDeadline time.Duration

	// Optional: validate id type (e.g. futures requires uint64)
	ValidateID func(id any) error
}

type WSCore struct {
	opts WSCoreOptions

	mu   sync.RWMutex
	conn *websocket.Conn

	// stream routing
	subsMu sync.RWMutex
	subs   map[string][]chan json.RawMessage // stream -> fanout channels

	// request/response correlation
	pendingMu sync.Mutex
	pending   map[string]chan WSControlResponse // idKey -> chan

	limiter <-chan time.Time
}

func NewWSCore(opts WSCoreOptions) *WSCore {
	if opts.DialTimeout <= 0 {
		opts.DialTimeout = 10 * time.Second
	}
	if opts.ControlMsgsPerSec <= 0 {
		opts.ControlMsgsPerSec = 4 // safe default; wrapper can set to 8 for futures
	}
	if opts.PongDeadline <= 0 {
		opts.PongDeadline = 10 * time.Second
	}

	tick := time.NewTicker(time.Second / time.Duration(opts.ControlMsgsPerSec))

	return &WSCore{
		opts:     opts,
		subs:    make(map[string][]chan json.RawMessage),
		pending: make(map[string]chan WSControlResponse),
		limiter: tick.C,
	}
}

func (c *WSCore) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		return nil
	}
	if c.opts.BaseURL == "" {
		return errors.New("BaseURL is required")
	}

	dialer := websocket.Dialer{HandshakeTimeout: c.opts.DialTimeout}
	url := c.buildURL()

	conn, _, err := dialer.DialContext(ctx, url, c.opts.Header)
	if err != nil {
		return err
	}

	// MUST pong back with copy of ping payload
	conn.SetPingHandler(func(appData string) error {
		return conn.WriteControl(websocket.PongMessage, []byte(appData), time.Now().Add(c.opts.PongDeadline))
	})

	c.conn = conn
	go c.readLoop()
	return nil
}

func (c *WSCore) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return nil
	}
	_ = c.conn.WriteControl(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, "bye"),
		time.Now().Add(2*time.Second),
	)
	err := c.conn.Close()
	c.conn = nil
	return err
}

func (c *WSCore) Subscribe(ctx context.Context, id any, method string, streams ...string) (<-chan json.RawMessage, error) {
	if len(streams) == 0 {
		return nil, errors.New("no streams")
	}
	if c.opts.ValidateID != nil {
		if err := c.opts.ValidateID(id); err != nil {
			return nil, err
		}
	}

	out := make(chan json.RawMessage, 256)

	// register local routing
	c.subsMu.Lock()
	for _, s := range streams {
		c.subs[s] = append(c.subs[s], out)
	}
	c.subsMu.Unlock()

	// send control request
	req := map[string]any{
		"method": method,
		"params": streams,
		"id":     id,
	}

	if _, err := c.sendRequestAwait(ctx, id, req); err != nil {
		// rollback
		c.subsMu.Lock()
		for _, s := range streams {
			c.subs[s] = removeRawChan(c.subs[s], out)
			if len(c.subs[s]) == 0 {
				delete(c.subs, s)
			}
		}
		c.subsMu.Unlock()
		close(out)
		return nil, err
	}

	return out, nil
}

func (c *WSCore) Unsubscribe(ctx context.Context, id any, method string, streams ...string) error {
	if len(streams) == 0 {
		return errors.New("no streams")
	}
	if c.opts.ValidateID != nil {
		if err := c.opts.ValidateID(id); err != nil {
			return err
		}
	}

	req := map[string]any{
		"method": method,
		"params": streams,
		"id":     id,
	}
	_, err := c.sendRequestAwait(ctx, id, req)
	return err
}

func (c *WSCore) Call(ctx context.Context, id any, req map[string]any) (WSControlResponse, error) {
	if c.opts.ValidateID != nil {
		if err := c.opts.ValidateID(id); err != nil {
			return WSControlResponse{}, err
		}
	}
	req["id"] = id
	return c.sendRequestAwait(ctx, id, req)
}

func (c *WSCore) buildURL() string {
	if c.opts.UseCombined {
		if len(c.opts.InitialStreams) > 0 {
			return fmt.Sprintf("%s/stream?streams=%s", c.opts.BaseURL, joinStreamsSlash(c.opts.InitialStreams))
		}
		return fmt.Sprintf("%s/stream", c.opts.BaseURL)
	}

	// raw
	if len(c.opts.InitialStreams) == 1 {
		return fmt.Sprintf("%s/ws/%s", c.opts.BaseURL, c.opts.InitialStreams[0])
	}
	return fmt.Sprintf("%s/ws", c.opts.BaseURL)
}

func joinStreamsSlash(streams []string) string {
	out := ""
	for i, s := range streams {
		if i > 0 {
			out += "/"
		}
		out += s
	}
	return out
}

func idKey(id any) string {
	return fmt.Sprintf("%T:%v", id, id)
}

func (c *WSCore) sendRequestAwait(ctx context.Context, id any, req any) (WSControlResponse, error) {
	select {
	case <-c.limiter:
	case <-ctx.Done():
		return WSControlResponse{}, ctx.Err()
	}

	key := idKey(id)

	ch := make(chan WSControlResponse, 1)
	c.pendingMu.Lock()
	c.pending[key] = ch
	c.pendingMu.Unlock()

	if err := c.writeJSON(req); err != nil {
		c.pendingMu.Lock()
		delete(c.pending, key)
		c.pendingMu.Unlock()
		return WSControlResponse{}, err
	}

	select {
	case resp := <-ch:
		return resp, nil
	case <-ctx.Done():
		c.pendingMu.Lock()
		delete(c.pending, key)
		c.pendingMu.Unlock()
		return WSControlResponse{}, ctx.Err()
	}
}

func (c *WSCore) writeJSON(v any) error {
	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()
	if conn == nil {
		return errors.New("ws not connected")
	}
	return conn.WriteJSON(v)
}

func removeRawChan(list []chan json.RawMessage, target chan json.RawMessage) []chan json.RawMessage {
	out := list[:0]
	for _, ch := range list {
		if ch != target {
			out = append(out, ch)
		}
	}
	return out
}

func (c *WSCore) readLoop() {
	for {
		c.mu.RLock()
		conn := c.conn
		c.mu.RUnlock()
		if conn == nil {
			return
		}

		_, msg, err := conn.ReadMessage()
		if err != nil {
			return
		}

		// 1) Try control response
		if handled := c.tryHandleControl(msg); handled {
			continue
		}

		// 2) Try combined wrapper
		if handled := c.tryHandleCombined(msg); handled {
			continue
		}

		// 3) Raw payload: broadcast to all subscribers
		c.broadcastToAll(json.RawMessage(msg))
	}
}

func (c *WSCore) tryHandleControl(msg []byte) bool {
	if !containsIDField(msg) {
		return false
	}

	var resp WSControlResponse
	if err := json.Unmarshal(msg, &resp); err != nil {
		return false
	}
	if resp.ID == nil {
		return false
	}

	key := idKey(resp.ID)

	c.pendingMu.Lock()
	ch, ok := c.pending[key]
	if ok {
		delete(c.pending, key)
	}
	c.pendingMu.Unlock()

	if ok {
		ch <- resp
		close(ch)
		return true
	}

	// Has id but not pending: ignore
	return true
}

func containsIDField(b []byte) bool {
	for i := 0; i+4 < len(b); i++ {
		if b[i] == '"' && b[i+1] == 'i' && b[i+2] == 'd' && b[i+3] == '"' {
			return true
		}
	}
	return false
}

func (c *WSCore) tryHandleCombined(msg []byte) bool {
	var ev WSCombinedEvent
	if err := json.Unmarshal(msg, &ev); err != nil {
		return false
	}
	if ev.Stream == "" || len(ev.Data) == 0 {
		return false
	}

	c.subsMu.RLock()
	list := append([]chan json.RawMessage(nil), c.subs[ev.Stream]...)
	c.subsMu.RUnlock()

	for _, ch := range list {
		select {
		case ch <- ev.Data:
		default:
			// drop on slow consumer
		}
	}
	return true
}

func (c *WSCore) broadcastToAll(raw json.RawMessage) {
	c.subsMu.RLock()
	defer c.subsMu.RUnlock()

	for _, list := range c.subs {
		for _, ch := range list {
			select {
			case ch <- raw:
			default:
			}
		}
	}
}
