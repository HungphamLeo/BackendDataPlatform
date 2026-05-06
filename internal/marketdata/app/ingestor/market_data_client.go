// TODO: File này vi phạm Clean Architecture. 
// Cần di chuyển file này sang: internal/marketdata/adapter/provider/binance/rest/

package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/HungphamLeo/BackendDataPlatform/internal/platform/logging"
)

// BinanceFuturesMarketDataClient lấy dữ liệu Public API (Không cần API Key)
type BinanceFuturesMarketDataClient struct {
	baseURL    string
	httpClient *http.Client
	logger     logging.Logger
}

// Thay vì hardcode URL, baseURL sẽ được inject từ Config YAML thông qua DI
func NewBinanceFuturesMarketDataClient(baseURL string, logger logging.Logger) *BinanceFuturesMarketDataClient {
	return &BinanceFuturesMarketDataClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger,
	}
}

// GetPremiumIndex trả về giá Mark Price và Funding Rate hiện tại của một Symbol
func (c *BinanceFuturesMarketDataClient) GetPremiumIndex(ctx context.Context, symbol string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/fapi/v1/premiumIndex?symbol=%s", c.baseURL, symbol)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("Failed to fetch premium index", logging.zap.Error(err))
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetKlines lấy dữ liệu nến (OHLCV) để dựng biểu đồ hoặc backtest
func (c *BinanceFuturesMarketDataClient) GetKlines(ctx context.Context, symbol, interval string, limit int) ([]interface{}, error) {
	url := fmt.Sprintf("%s/fapi/v1/klines?symbol=%s&interval=%s&limit=%d", c.baseURL, symbol, interval, limit)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("Failed to fetch klines", logging.zap.Error(err))
		return nil, err
	}
	defer resp.Body.Close()

	var result []interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}