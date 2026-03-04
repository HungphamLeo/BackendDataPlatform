package binance

import (
	"context"
	"encoding/json"
)

type SpotWSFacade struct {
	Client *SpotWSClient
}

func NewSpotWSFacade(client *SpotWSClient) *SpotWSFacade {
	return &SpotWSFacade{Client: client}
}

// Ví dụ 1 facade giữ nguyên style decode
func (f *SpotWSFacade) SubscribeBookTicker(ctx context.Context, id any, symbol string) (<-chan BookTickerEvent, error) {
	stream := StreamBookTicker(symbol)
	rawCh, err := f.Client.Subscribe(ctx, id, stream)
	if err != nil {
		return nil, err
	}

	out := make(chan BookTickerEvent, 256)
	go decodeLoop(rawCh, out, func(b json.RawMessage) (BookTickerEvent, error) {
		var e BookTickerEvent
		return e, json.Unmarshal(b, &e)
	})
	return out, nil
}

func decodeLoop[T any](raw <-chan json.RawMessage, out chan<- T, decode func(json.RawMessage) (T, error)) {
	defer close(out)
	for b := range raw {
		ev, err := decode(b)
		if err != nil {
			continue
		}
		select {
		case out <- ev:
		default:
		}
	}
}
