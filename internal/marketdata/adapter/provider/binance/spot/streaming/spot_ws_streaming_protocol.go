// spot_ws_streaming_protocol.go
package binance

import "encoding/json"

type wsMethod string

const (
	wsMethodSubscribe         wsMethod = "SUBSCRIBE"
	wsMethodUnsubscribe       wsMethod = "UNSUBSCRIBE"
	wsMethodListSubscriptions wsMethod = "LIST_SUBSCRIPTIONS"
	wsMethodSetProperty       wsMethod = "SET_PROPERTY"
	wsMethodGetProperty       wsMethod = "GET_PROPERTY"
)

type WSRequest struct {
	Method wsMethod       `json:"method"`
	Params []any          `json:"params,omitempty"`
	ID     any            `json:"id"` // int64 / string / nil
}

type WSResponse struct {
	Result json.RawMessage `json:"result"`
	ID     any             `json:"id"`
	Error  *WSError        `json:"error,omitempty"`
}

type WSError struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	ID   any    `json:"id,omitempty"`
}

// Combined stream wrapper: {"stream":"<streamName>","data":<rawPayload>}
type CombinedEvent struct {
	Stream string          `json:"stream"`
	Data   json.RawMessage `json:"data"`
}
