package rest

import (
	"net/http"
	"strconv"
	"time"

	"github.com/HungphamLeo/BackendDataPlatform/api/clients"
)

// MarketHandler serves /v1/market/* endpoints.
type MarketHandler struct {
	queryClient *clients.QueryClient
}

// NewMarketHandler creates a MarketHandler.
func NewMarketHandler(qc *clients.QueryClient) *MarketHandler {
	return &MarketHandler{queryClient: qc}
}

// GetTicker handles GET /v1/market/ticker?symbol=BTC_USDT&exchange=binance
func (h *MarketHandler) GetTicker(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")
	exchange := r.URL.Query().Get("exchange")
	if symbol == "" || exchange == "" {
		writeError(w, http.StatusBadRequest, "symbol and exchange are required")
		return
	}

	// TODO: call queryClient.QueryMarketSummary once stubs are generated.
	// Returning stub response for scaffold.
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"symbol":   symbol,
		"exchange": exchange,
		"note":     "implement after proto stubs generated",
	})
}

// GetOrderBook handles GET /v1/market/orderbook?symbol=BTC_USDT&exchange=binance
func (h *MarketHandler) GetOrderBook(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")
	exchange := r.URL.Query().Get("exchange")
	if symbol == "" || exchange == "" {
		writeError(w, http.StatusBadRequest, "symbol and exchange are required")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"symbol": symbol, "exchange": exchange, "bids": []interface{}{}, "asks": []interface{}{},
	})
}

// GetKline handles GET /v1/market/kline?symbol=BTC_USDT&exchange=binance&interval=1m&from=&to=
func (h *MarketHandler) GetKline(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	symbol := q.Get("symbol")
	exchange := q.Get("exchange")
	interval := q.Get("interval")
	if symbol == "" || exchange == "" {
		writeError(w, http.StatusBadRequest, "symbol and exchange are required")
		return
	}
	if interval == "" {
		interval = "1m"
	}
	from, _ := parseTimeParam(q.Get("from"))
	to, _ := parseTimeParam(q.Get("to"))
	if to.IsZero() {
		to = time.Now()
	}
	if from.IsZero() {
		from = to.Add(-24 * time.Hour)
	}
	limit, _ := strconv.Atoi(q.Get("limit"))

	// TODO: call queryClient.QueryHistoricalRange
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"symbol": symbol, "exchange": exchange, "interval": interval,
		"from": from, "to": to, "limit": limit, "klines": []interface{}{},
	})
}

func parseTimeParam(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	// try unix timestamp first
	if ts, err := strconv.ParseInt(s, 10, 64); err == nil {
		return time.Unix(ts, 0), nil
	}
	return time.Parse(time.RFC3339, s)
}
