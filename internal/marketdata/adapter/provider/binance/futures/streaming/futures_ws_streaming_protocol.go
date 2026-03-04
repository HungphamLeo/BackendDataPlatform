// futures_ws_streaming_protocol.go
package binance

import "encoding/json"

type futWSMethod string

const (
	futWSMethodSubscribe         futWSMethod = "SUBSCRIBE"
	futWSMethodUnsubscribe       futWSMethod = "UNSUBSCRIBE"
	futWSMethodListSubscriptions futWSMethod = "LIST_SUBSCRIPTIONS"
	futWSMethodSetProperty       futWSMethod = "SET_PROPERTY"
	futWSMethodGetProperty       futWSMethod = "GET_PROPERTY"
)

type FuturesWSRequest struct {
	Method futWSMethod `json:"method"`
	Params []any       `json:"params,omitempty"`
	ID     uint64      `json:"id"` // futures spec: unsigned int
}

type FuturesWSResponse struct {
	Result json.RawMessage `json:"result"`
	ID     uint64          `json:"id"`
	Error  *FuturesWSError `json:"error,omitempty"`
}

type FuturesWSError struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// Combined stream wrapper: {"stream":"<streamName>","data":<rawPayload>}
type FuturesCombinedEvent struct {
	Stream string          `json:"stream"`
	Data   json.RawMessage `json:"data"`
}
