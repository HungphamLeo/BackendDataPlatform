package binance

import (
	"context"
	"encoding/json"
)

type FuturesWSFacade struct {
	Client *FuturesWSClient
}

func NewFuturesWSFacade(client *FuturesWSClient) *FuturesWSFacade {
	return &FuturesWSFacade{Client: client}
}

func (f *FuturesWSFacade) SubscribeMarkPrice(ctx context.Context, id uint64, symbol string, oneSecond bool) (<-chan FuturesMarkPriceEvent, string, error) {
	stream := FuturesStreamMarkPrice(symbol, oneSecond)
	raw, err := f.Client.Subscribe(ctx, id, stream)
	if err != nil {
		return nil, "", err
	}
	out := make(chan FuturesMarkPriceEvent, 256)
	go futDecodeLoop(raw, out, func(b json.RawMessage) (FuturesMarkPriceEvent, error) {
		var e FuturesMarkPriceEvent
		return e, json.Unmarshal(b, &e)
	})
	return out, stream, nil
}

func futDecodeLoop[T any](raw <-chan json.RawMessage, out chan<- T, decode func(json.RawMessage) (T, error)) {
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
