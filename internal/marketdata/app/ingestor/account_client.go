// TODO: File này vi phạm Clean Architecture. 
// Cần di chuyển file này sang: internal/marketdata/adapter/provider/binance/rest/

package rest

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/HungphamLeo/BackendDataPlatform/internal/platform/logging"
)

// BinanceFuturesAccountClient trích xuất Account Data (Yêu cầu API Key & Signature)
type BinanceFuturesAccountClient struct {
	baseURL    string
	apiKey     string
	secretKey  string
	httpClient *http.Client
	logger     logging.Logger
}

// BaseURL, APIKey, SecretKey phải được load từ config.yaml và inject vào từ bên ngoài
func NewBinanceFuturesAccountClient(baseURL, apiKey, secretKey string, logger logging.Logger) *BinanceFuturesAccountClient {
	return &BinanceFuturesAccountClient{
		baseURL:    baseURL,
		apiKey:     apiKey,
		secretKey:  secretKey,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		logger:     logger,
	}
}

// signRequest tạo chữ ký HMAC SHA256 cho Binance
func (c *BinanceFuturesAccountClient) signRequest(query string) string {
	mac := hmac.New(sha256.New, []byte(c.secretKey))
	mac.Write([]byte(query))
	return hex.EncodeToString(mac.Sum(nil))
}

// GetBalances lấy số dư tài khoản (USDT, BUSD, v.v...)
func (c *BinanceFuturesAccountClient) GetBalances(ctx context.Context) ([]map[string]interface{}, error) {
	timestamp := time.Now().UnixMilli()
	params := url.Values{}
	params.Add("timestamp", fmt.Sprintf("%d", timestamp))

	signature := c.signRequest(params.Encode())
	finalURL := fmt.Sprintf("%s/fapi/v2/balance?%s&signature=%s", c.baseURL, params.Encode(), signature)

	return c.sendSignedRequest(ctx, http.MethodGet, finalURL)
}

// GetPositions lấy danh sách vị thế (Positions) hiện tại
func (c *BinanceFuturesAccountClient) GetPositions(ctx context.Context) ([]map[string]interface{}, error) {
	timestamp := time.Now().UnixMilli()
	params := url.Values{}
	params.Add("timestamp", fmt.Sprintf("%d", timestamp))

	signature := c.signRequest(params.Encode())
	finalURL := fmt.Sprintf("%s/fapi/v2/positionRisk?%s&signature=%s", c.baseURL, params.Encode(), signature)

	return c.sendSignedRequest(ctx, http.MethodGet, finalURL)
}

// sendSignedRequest là hàm helper đính kèm Header API Key và decode JSON
func (c *BinanceFuturesAccountClient) sendSignedRequest(ctx context.Context, method, finalURL string) ([]map[string]interface{}, error) {
	req, err := http.NewRequestWithContext(ctx, method, finalURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-MBX-APIKEY", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("Failed to fetch signed data from Binance", logging.zap.String("url", finalURL))
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.logger.Error(fmt.Sprintf("Binance returned status: %d", resp.StatusCode))
		return nil, fmt.Errorf("binance api error: status %d", resp.StatusCode)
	}

	var result []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}