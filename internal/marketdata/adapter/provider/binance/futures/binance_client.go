package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type FuturesREST struct {
	baseURL string
	client  *http.Client
	apiKey  string
}

func NewFuturesREST(baseURL, apiKey string) *FuturesREST {
	return &FuturesREST{
		baseURL: baseURL,
		apiKey:  apiKey,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (f *FuturesREST) GetExchangeInfo(ctx context.Context) ([]byte, error) {

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		f.baseURL+"/fapi/v1/exchangeInfo",
		nil,
	)

	if err != nil {
		return nil, err
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var raw json.RawMessage
	err = json.NewDecoder(resp.Body).Decode(&raw)

	return raw, err
}

func (f *FuturesREST) GetFundingRate(
	ctx context.Context,
	symbol string,
	limit int,
) ([]byte, error) {

	params := url.Values{}
	params.Add("symbol", symbol)
	params.Add("limit", fmt.Sprintf("%d", limit))

	endpoint := fmt.Sprintf(
		"%s/fapi/v1/fundingRate?%s",
		f.baseURL,
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

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var raw json.RawMessage
	err = json.NewDecoder(resp.Body).Decode(&raw)

	return raw, err
}
