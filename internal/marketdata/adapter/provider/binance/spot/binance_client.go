package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type SpotREST struct {
	baseURL string
	client  *http.Client
}

func NewSpotREST(baseURL string) *SpotREST {
	return &SpotREST{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (s *SpotREST) GetExchangeInfo(ctx context.Context) ([]byte, error) {

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		s.baseURL+"/api/v3/exchangeInfo",
		nil,
	)

	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).DecodeBytes()
}

func (s *SpotREST) GetDepth(ctx context.Context, symbol string, limit int) ([]byte, error) {

	params := url.Values{}
	params.Add("symbol", symbol)
	params.Add("limit", fmt.Sprintf("%d", limit))

	endpoint := fmt.Sprintf(
		"%s/api/v3/depth?%s",
		s.baseURL,
		params.Encode(),
	)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
	)

	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var raw json.RawMessage
	err = json.NewDecoder(resp.Body).Decode(&raw)

	return raw, err
}
