package server

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/HungphamLeo/BackendDataPlatform/query/cache"
	"github.com/HungphamLeo/BackendDataPlatform/query/clients"
)

// QueryServer implements the DataQuery gRPC service.
// After proto stubs are generated, embed the UnimplementedDataQueryServer here
// and register with grpc.Server.
type QueryServer struct {
	streamingClient *clients.StreamingClient
	batchClient     *clients.BatchClient
	queryCache      *cache.QueryCache
}

// New creates a QueryServer.
func New(
	streamingClient *clients.StreamingClient,
	batchClient *clients.BatchClient,
	queryCache *cache.QueryCache,
) *QueryServer {
	return &QueryServer{
		streamingClient: streamingClient,
		batchClient:     batchClient,
		queryCache:      queryCache,
	}
}

// --- cache helpers ---

func cacheKey(prefix string, params ...string) string {
	h := md5.New()
	for _, p := range params {
		fmt.Fprintln(h, p)
	}
	return fmt.Sprintf("query:%s:%x", prefix, h.Sum(nil))
}

// --- RPC implementations (cache-first pattern) ---

// MarketSummaryResult is the DTO returned by QueryMarketSummary.
type MarketSummaryResult struct {
	Symbol       string    `json:"symbol"`
	Exchange     string    `json:"exchange"`
	LastPrice    float64   `json:"last_price"`
	Bid          float64   `json:"bid"`
	Ask          float64   `json:"ask"`
	Volume24h    float64   `json:"volume_24h"`
	Timestamp    time.Time `json:"timestamp"`
}

// QueryMarketSummary fetches current ticker; cache-first with 5s TTL.
func (s *QueryServer) QueryMarketSummary(ctx context.Context, symbol, exchange string) (*MarketSummaryResult, error) {
	key := cacheKey("summary", symbol, exchange)

	// Cache hit?
	if cached, err := s.queryCache.Get(ctx, key); err == nil && cached != "" {
		var result MarketSummaryResult
		if err := json.Unmarshal([]byte(cached), &result); err == nil {
			return &result, nil
		}
	}

	// Fetch from streaming service.
	ticker, err := s.streamingClient.GetTicker(ctx, symbol, exchange)
	if err != nil {
		return nil, status.Errorf(codes.Unavailable, "streaming unavailable: %v", err)
	}

	result := &MarketSummaryResult{
		Symbol:    ticker.Symbol,
		Exchange:  ticker.Exchange,
		LastPrice: ticker.Price,
		Bid:       ticker.Bid,
		Ask:       ticker.Ask,
		Volume24h: ticker.Volume,
		Timestamp: ticker.Timestamp,
	}

	// Populate cache.
	if b, err := json.Marshal(result); err == nil {
		_ = s.queryCache.Set(ctx, key, string(b), cache.RealtimeTTL)
	}

	return result, nil
}

// HistoricalResult is the DTO returned by QueryHistoricalRange.
type HistoricalResult struct {
	Bars []clients.OHLCVBar `json:"bars"`
}

// QueryHistoricalRange fetches OHLCV from batch; cache-first with 60s TTL.
func (s *QueryServer) QueryHistoricalRange(
	ctx context.Context,
	symbol, exchange, interval string,
	from, to time.Time,
	limit int,
) (*HistoricalResult, error) {
	key := cacheKey("historical", symbol, exchange, interval,
		fmt.Sprintf("%d-%d-%d", from.Unix(), to.Unix(), limit))

	if cached, err := s.queryCache.Get(ctx, key); err == nil && cached != "" {
		var result HistoricalResult
		if err := json.Unmarshal([]byte(cached), &result); err == nil {
			return &result, nil
		}
	}

	bars, err := s.batchClient.GetOHLCV(ctx, symbol, exchange, interval, from, to, limit)
	if err != nil {
		return nil, status.Errorf(codes.Unavailable, "batch unavailable: %v", err)
	}

	result := &HistoricalResult{Bars: bars}
	if b, err := json.Marshal(result); err == nil {
		_ = s.queryCache.Set(ctx, key, string(b), cache.BatchTTL)
	}
	return result, nil
}

// SignalsResult is the DTO returned by QuerySignals.
type SignalsResult struct {
	Signals []clients.IndicatorPoint `json:"signals"`
}

// QuerySignals fetches indicators from batch; cache-first with 60s TTL.
func (s *QueryServer) QuerySignals(
	ctx context.Context,
	symbol, exchange, interval, indicatorName string,
) (*SignalsResult, error) {
	key := cacheKey("signals", symbol, exchange, interval, indicatorName)

	if cached, err := s.queryCache.Get(ctx, key); err == nil && cached != "" {
		var result SignalsResult
		if err := json.Unmarshal([]byte(cached), &result); err == nil {
			return &result, nil
		}
	}

	zeroTime := time.Time{}
	points, err := s.batchClient.GetIndicator(ctx, symbol, exchange, interval, indicatorName, zeroTime, zeroTime, 500)
	if err != nil {
		return nil, status.Errorf(codes.Unavailable, "batch unavailable: %v", err)
	}

	result := &SignalsResult{Signals: points}
	if b, err := json.Marshal(result); err == nil {
		_ = s.queryCache.Set(ctx, key, string(b), cache.BatchTTL)
	}
	return result, nil
}
