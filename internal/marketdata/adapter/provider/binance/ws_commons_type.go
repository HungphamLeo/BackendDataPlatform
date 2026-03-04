package binance

import "encoding/json"

// Combined stream wrapper: {"stream":"<streamName>","data":<rawPayload>}
type WSCombinedEvent struct {
	Stream string          `json:"stream"`
	Data   json.RawMessage `json:"data"`
}

// Generic ws error format (both spot & futures use the same shape)
type WSError struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// Generic control response, id can be any (spot) or uint64 (futures wrapper will restrict)
type WSControlResponse struct {
	Result json.RawMessage `json:"result"`
	ID     any             `json:"id"`
	Error  *WSError        `json:"error,omitempty"`
}
