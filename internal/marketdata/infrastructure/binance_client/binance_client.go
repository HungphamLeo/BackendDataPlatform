package binance_client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"internal/marketdata/domain/value_object"
	"internal/platform/http_client"
)

// ✅ binance CLIENT - REST API ADAPTER
type RestClient struct {
	client    http_client.HTTPClient
	baseURL   string
	apiKey    string
	sessionID string
}

func NewRestClient(baseURL string, apiKey string) *RestClient {
	return &RestClient{
		client:  http_client.NewHTTPClient(10*time.Second, 3),
		baseURL: baseURL,
		apiKey:  apiKey,
	}
}

// GetQuote - Fetch quote from binance REST API
func (c *RestClient) GetQuote(
	ctx context.Context,
	symbol value_object.Symbol,
) (*value_object.Quote, error) {
	// ✅ Call external API (REST)
	endpoint := fmt.Sprintf("%s/api/v1/getQuote", c.baseURL)

	body := map[string]interface{}{
		"symbol": symbol.String(),
	}

	req, _ := json.Marshal(body)
	resp, err := c.client.Post(endpoint, req, map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", c.apiKey),
		"Content-Type":  "application/json",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to call binance API: %w", err)
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)

	// ✅ Map API response to domain value object
	var apiResponse struct {
		Bid  float64 `json:"bid"`
		Ask  float64 `json:"ask"`
		Time int64   `json:"time"`
	}
	_ = json.Unmarshal(data, &apiResponse)

	quote := &value_object.Quote{
		Bid:       apiResponse.Bid,
		Ask:       apiResponse.Ask,
		Spread:    apiResponse.Ask - apiResponse.Bid,
		Timestamp: apiResponse.Time,
	}

	return quote, nil
}
