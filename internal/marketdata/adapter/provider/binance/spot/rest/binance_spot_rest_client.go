package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/HungphamLeo/BackendDataPlatform/internal/marketdata/application/port/out"
	"github.com/HungphamLeo/BackendDataPlatform/internal/marketdata/domain/entity"
	"github.com/HungphamLeo/BackendDataPlatform/internal/marketdata/domain/value_object"
	"github.com/HungphamLeo/BackendDataPlatform/internal/platform/config"
	"github.com/HungphamLeo/BackendDataPlatform/internal/platform/logging"
)

// BinanceSpotRestClient triển khai MarketDataProviderPort cho Binance Spot REST API
type BinanceSpotRestClient struct {
	baseURL    string
	httpClient *http.Client
	logger     logging.Logger
	config     *config.BinanceConfig
}

// NewBinanceSpotRestClient tạo mới BinanceSpotRestClient
func NewBinanceSpotRestClient(
	config *config.BinanceConfig,
	logger logging.Logger,
) *BinanceSpotRestClient {
	return &BinanceSpotRestClient{
		baseURL: config.SpotBaseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger,
		config: config,
	}
}

// Đảm bảo BinanceSpotRestClient triển khai interface MarketDataProviderPort
var _ out.MarketDataProviderPort = (*BinanceSpotRestClient)(nil)

// Name trả về tên của provider
func (c *BinanceSpotRestClient) Name() string {
	return "BINANCE_SPOT"
}

// Connect kết nối đến provider
func (c *BinanceSpotRestClient) Connect(ctx context.Context) error {
	// Kiểm tra kết nối bằng cách gọi API ping
	url := fmt.Sprintf("%s/api/v3/ping", c.baseURL)
	
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("Failed to connect to Binance Spot API", logging.Fields{
			"error": err.Error(),
		})
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("binance spot api error: status %d", resp.StatusCode)
	}
	
	c.logger.Info("Connected to Binance Spot API", nil)
	return nil
}

// Disconnect ngắt kết nối khỏi provider
func (c *BinanceSpotRestClient) Disconnect(ctx context.Context) error {
	// Không cần làm gì vì REST API không duy trì kết nối
	return nil
}

// SubscribeTicker đăng ký nhận dữ liệu ticker
func (c *BinanceSpotRestClient) SubscribeTicker(ctx context.Context, symbols []value_object.Symbol) error {
	// REST API không hỗ trợ subscribe, chỉ có thể poll
	return nil
}

// SubscribeTrades đăng ký nhận dữ liệu giao dịch
func (c *BinanceSpotRestClient) SubscribeTrades(ctx context.Context, symbols []value_object.Symbol) error {
	// REST API không hỗ trợ subscribe, chỉ có thể poll
	return nil
}

// SubscribeOrderBook đăng ký nhận dữ liệu sổ lệnh
func (c *BinanceSpotRestClient) SubscribeOrderBook(ctx context.Context, symbols []value_object.Symbol, depth int) error {
	// REST API không hỗ trợ subscribe, chỉ có thể poll
	return nil
}

// SubscribeCandles đăng ký nhận dữ liệu nến
func (c *BinanceSpotRestClient) SubscribeCandles(ctx context.Context, symbols []value_object.Symbol, timeframes []value_object.Timeframe) error {
	// REST API không hỗ trợ subscribe, chỉ có thể poll
	return nil
}

// GetTicker lấy dữ liệu ticker
func (c *BinanceSpotRestClient) GetTicker(ctx context.Context, symbol value_object.Symbol) (*value_object.Quote, error) {
	// Chuyển đổi symbol sang định dạng của Binance
	binanceSymbol := symbol.Base + symbol.Quote
	
	url := fmt.Sprintf("%s/api/v3/ticker/bookTicker?symbol=%s", c.baseURL, binanceSymbol)
	
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("Failed to get ticker from Binance Spot API", logging.Fields{
			"error":  err.Error(),
			"symbol": binanceSymbol,
		})
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("binance spot api error: status %d", resp.StatusCode)
	}
	
	var result struct {
		Symbol   string `json:"symbol"`
		BidPrice string `json:"bidPrice"`
		BidQty   string `json:"bidQty"`
		AskPrice string `json:"askPrice"`
		AskQty   string `json:"askQty"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	// Chuyển đổi từ string sang float64
	bidPrice, _ := strconv.ParseFloat(result.BidPrice, 64)
	bidQty, _ := strconv.ParseFloat(result.BidQty, 64)
	askPrice, _ := strconv.ParseFloat(result.AskPrice, 64)
	askQty, _ := strconv.ParseFloat(result.AskQty, 64)
	
	// Lấy giá trung bình của bid và ask làm giá cuối cùng
	lastPrice := (bidPrice + askPrice) / 2
	
	return &value_object.Quote{
		Bid:     bidPrice,
		Ask:     askPrice,
		Last:    lastPrice,
		BidSize: bidQty,
		AskSize: askQty,
	}, nil
}

// GetTickers lấy dữ liệu ticker cho nhiều symbol
func (c *BinanceSpotRestClient) GetTickers(ctx context.Context, symbols []value_object.Symbol) (map[string]*value_object.Quote, error) {
	url := fmt.Sprintf("%s/api/v3/ticker/bookTicker", c.baseURL)
	
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("Failed to get tickers from Binance Spot API", logging.Fields{
			"error": err.Error(),
		})
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("binance spot api error: status %d", resp.StatusCode)
	}
	
	var results []struct {
		Symbol   string `json:"symbol"`
		BidPrice string `json:"bidPrice"`
		BidQty   string `json:"bidQty"`
		AskPrice string `json:"askPrice"`
		AskQty   string `json:"askQty"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, err
	}
	
	// Tạo map để lưu kết quả
	tickers := make(map[string]*value_object.Quote)
	
	// Tạo map để kiểm tra symbol cần lấy
	symbolMap := make(map[string]bool)
	for _, symbol := range symbols {
		binanceSymbol := symbol.Base + symbol.Quote
		symbolMap[binanceSymbol] = true
	}
	
	// Lọc kết quả theo symbol
	for _, result := range results {
		if _, ok := symbolMap[result.Symbol]; !ok {
			continue
		}
		
		// Chuyển đổi từ string sang float64
		bidPrice, _ := strconv.ParseFloat(result.BidPrice, 64)
		bidQty, _ := strconv.ParseFloat(result.BidQty, 64)
		askPrice, _ := strconv.ParseFloat(result.AskPrice, 64)
		askQty, _ := strconv.ParseFloat(result.AskQty, 64)
		
		// Lấy giá trung bình của bid và ask làm giá cuối cùng
		lastPrice := (bidPrice + askPrice) / 2
		
		tickers[result.Symbol] = &value_object.Quote{
			Bid:     bidPrice,
			Ask:     askPrice,
			Last:    lastPrice,
			BidSize: bidQty,
			AskSize: askQty,
		}
	}
	
	return tickers, nil
}

// GetTrades lấy dữ liệu giao dịch
func (c *BinanceSpotRestClient) GetTrades(ctx context.Context, symbol value_object.Symbol, limit int) ([]*entity.Trade, error) {
	// Chuyển đổi symbol sang định dạng của Binance
	binanceSymbol := symbol.Base + symbol.Quote
	
	url := fmt.Sprintf("%s/api/v3/trades?symbol=%s&limit=%d", c.baseURL, binanceSymbol, limit)
	
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("Failed to get trades from Binance Spot API", logging.Fields{
			"error":  err.Error(),
			"symbol": binanceSymbol,
		})
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("binance spot api error: status %d", resp.StatusCode)
	}
	
	var results []struct {
		ID           int64  `json:"id"`
		Price        string `json:"price"`
		Qty          string `json:"qty"`
		QuoteQty     string `json:"quoteQty"`
		Time         int64  `json:"time"`
		IsBuyerMaker bool   `json:"isBuyerMaker"`
		IsBestMatch  bool   `json:"isBestMatch"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, err
	}
	
	trades := make([]*entity.Trade, len(results))
	for i, result := range results {
		price, _ := strconv.ParseFloat(result.Price, 64)
		qty, _ := strconv.ParseFloat(result.Qty, 64)
		quoteQty, _ := strconv.ParseFloat(result.QuoteQty, 64)
		
		side := "BUY"
		if result.IsBuyerMaker {
			side = "SELL"
		}
		
		trades[i] = &entity.Trade{
			ID:           strconv.FormatInt(result.ID, 10),
			Symbol:       symbol,
			Exchange:     c.Name(),
			Price:        price,
			Quantity:     qty,
			QuoteQuantity: quoteQty,
			Timestamp:    time.UnixMilli(result.Time),
			Side:         side,
			IsBuyerMaker: result.IsBuyerMaker,
		}
	}
	
	return trades, nil
}

// GetHistoricalTrades lấy dữ liệu giao dịch lịch sử
func (c *BinanceSpotRestClient) GetHistoricalTrades(ctx context.Context, symbol value_object.Symbol, fromID string, limit int) ([]*entity.Trade, error) {
	// Chuyển đổi symbol sang định dạng của Binance
	binanceSymbol := symbol.Base + symbol.Quote
	
	url := fmt.Sprintf("%s/api/v3/historicalTrades?symbol=%s&limit=%d", c.baseURL, binanceSymbol, limit)
	if fromID != "" {
		url += fmt.Sprintf("&fromId=%s", fromID)
	}
	
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	
	// Thêm API key vào header
	req.Header.Set("X-MBX-APIKEY", c.config.APIKey)
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("Failed to get historical trades from Binance Spot API", logging.Fields{
			"error":  err.Error(),
			"symbol": binanceSymbol,
		})
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("binance spot api error: status %d", resp.StatusCode)
	}
	
	var results []struct {
		ID           int64  `json:"id"`
		Price        string `json:"price"`
		Qty          string `json:"qty"`
		QuoteQty     string `json:"quoteQty"`
		Time         int64  `json:"time"`
		IsBuyerMaker bool   `json:"isBuyerMaker"`
		IsBestMatch  bool   `json:"isBestMatch"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, err
	}
	
	trades := make([]*entity.Trade, len(results))
	for i, result := range results {
		price, _ := strconv.ParseFloat(result.Price, 64)
		qty, _ := strconv.ParseFloat(result.Qty, 64)
		quoteQty, _ := strconv.ParseFloat(result.QuoteQty, 64)
		
		side := "BUY"
		if result.IsBuyerMaker {
			side = "SELL"
		}
		
		trades[i] = &entity.Trade{
			ID:           strconv.FormatInt(result.ID, 10),
			Symbol:       symbol,
			Exchange:     c.Name(),
			Price:        price,
			Quantity:     qty,
			QuoteQuantity: quoteQty,
			Timestamp:    time.UnixMilli(result.Time),
			Side:         side,
			IsBuyerMaker: result.IsBuyerMaker,
		}
	}
	
	return trades, nil
}

// GetOrderBook lấy dữ liệu sổ lệnh
func (c *BinanceSpotRestClient) GetOrderBook(ctx context.Context, symbol value_object.Symbol, depth int) (*entity.OrderBook, error) {
	// Chuyển đổi symbol sang định dạng của Binance
	binanceSymbol := symbol.Base + symbol.Quote
	
	url := fmt.Sprintf("%s/api/v3/depth?symbol=%s&limit=%d", c.baseURL, binanceSymbol, depth)
	
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("Failed to get order book from Binance Spot API", logging.Fields{
			"error":  err.Error(),
			"symbol": binanceSymbol,
		})
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("binance spot api error: status %d", resp.StatusCode)
	}
	
	var result struct {
		LastUpdateID int64      `json:"lastUpdateId"`
		Bids         [][]string `json:"bids"`
		Asks         [][]string `json:"asks"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	// Chuyển đổi bids
	bids := make([]value_object.Quote, len(result.Bids))
	for i, bid := range result.Bids {
		price, _ := strconv.ParseFloat(bid[0], 64)
		qty, _ := strconv.ParseFloat(bid[1], 64)
		
		bids[i] = value_object.Quote{
			Bid:     price,
			BidSize: qty,
		}
	}
	
	// Chuyển đổi asks
	asks := make([]value_object.Quote, len(result.Asks))
	for i, ask := range result.Asks {
		price, _ := strconv.ParseFloat(ask[0], 64)
		qty, _ := strconv.ParseFloat(ask[1], 64)
		
		asks[i] = value_object.Quote{
			Ask:     price,
			AskSize: qty,
		}
	}
	
	return &entity.OrderBook{
		Symbol:       symbol,
		Exchange:     c.Name(),
		Bids:         bids,
		Asks:         asks,
		LastUpdateID: result.LastUpdateID,
		Timestamp:    time.Now(),
	}, nil
}

// GetCandles lấy dữ liệu nến
func (c *BinanceSpotRestClient) GetCandles(
	ctx context.Context,
	symbol value_object.Symbol,
	timeframe value_object.Timeframe,
	startTime, endTime time.Time,
	limit int,
) ([]*entity.Candle, error) {
	// Chuyển đổi symbol sang định dạng của Binance
	binanceSymbol := symbol.Base + symbol.Quote
	
	// Chuyển đổi timeframe sang định dạng của Binance
	binanceInterval := timeframe.String()
	
	url := fmt.Sprintf("%s/api/v3/klines?symbol=%s&interval=%s&limit=%d", c.baseURL, binanceSymbol, binanceInterval, limit)
	
	if !startTime.IsZero() {
		url += fmt.Sprintf("&startTime=%d", startTime.UnixMilli())
	}
	
	if !endTime.IsZero() {
		url += fmt.Sprintf("&endTime=%d", endTime.UnixMilli())
	}
	
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("Failed to get candles from Binance Spot API", logging.Fields{
			"error":    err.Error(),
			"symbol":   binanceSymbol,
			"interval": binanceInterval,
		})
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("binance spot api error: status %d", resp.StatusCode)
	}
	
	var results [][]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, err
	}
	
	candles := make([]*entity.Candle, len(results))
	for i, result := range results {
		openTime := time.UnixMilli(int64(result[0].(float64)))
		closeTime := time.UnixMilli(int64(result[6].(float64)))
		open, _ := strconv.ParseFloat(result[1].(string), 64)
		high, _ := strconv.ParseFloat(result[2].(string), 64)
		low, _ := strconv.ParseFloat(result[3].(string), 64)
		close, _ := strconv.ParseFloat(result[4].(string), 64)
		volume, _ := strconv.ParseFloat(result[5].(string), 64)
		quoteVolume, _ := strconv.ParseFloat(result[7].(string), 64)
		tradeCount := int64(result[8].(float64))
		
		candles[i] = &entity.Candle{
			Symbol:      symbol,
			Exchange:    c.Name(),
			Timeframe:   timeframe,
			OpenTime:    openTime,
			CloseTime:   closeTime,
			Open:        open,
			High:        high,
			Low:         low,
			Close:       close,
			Volume:      volume,
			QuoteVolume: quoteVolume,
			TradeCount:  tradeCount,
			Closed:      true,
		}
	}
	
	return candles, nil
}

// GetExchangeInfo lấy thông tin sàn giao dịch
func (c *BinanceSpotRestClient) GetExchangeInfo(ctx context.Context) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/api/v3/exchangeInfo", c.baseURL)
	
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("Failed to get exchange info from Binance Spot API", logging.Fields{
			"error": err.Error(),
		})
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("binance spot api error: status %d", resp.StatusCode)
	}
	
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	return result, nil
}

// IsConnected kiểm tra xem đã kết nối chưa
func (c *BinanceSpotRestClient) IsConnected() bool {
	// REST API không duy trì kết nối, luôn trả về true
	return true
}
