package rest

import (
	"net/http"
	"strconv"

	"github.com/HungphamLeo/BackendDataPlatform/api/clients"
)

// BatchHandler serves /v1/batch/* endpoints.
type BatchHandler struct {
	queryClient *clients.QueryClient
}

// NewBatchHandler creates a BatchHandler.
func NewBatchHandler(qc *clients.QueryClient) *BatchHandler {
	return &BatchHandler{queryClient: qc}
}

// GetOHLCV handles GET /v1/batch/ohlcv?symbol=BTC_USDT&exchange=binance&interval=1d&limit=100
func (h *BatchHandler) GetOHLCV(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	symbol := q.Get("symbol")
	exchange := q.Get("exchange")
	interval := q.Get("interval")
	limit, _ := strconv.Atoi(q.Get("limit"))
	if symbol == "" || exchange == "" {
		writeError(w, http.StatusBadRequest, "symbol and exchange are required")
		return
	}
	if interval == "" {
		interval = "1d"
	}
	if limit == 0 {
		limit = 100
	}

	// TODO: call queryClient.QueryHistoricalRange once stubs generated
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"symbol": symbol, "exchange": exchange, "interval": interval,
		"limit": limit, "bars": []interface{}{},
	})
}

// GetIndicators handles GET /v1/batch/indicators?symbol=BTC_USDT&exchange=binance&indicator=ATR
func (h *BatchHandler) GetIndicators(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	symbol := q.Get("symbol")
	exchange := q.Get("exchange")
	indicator := q.Get("indicator")
	interval := q.Get("interval")
	if symbol == "" || exchange == "" {
		writeError(w, http.StatusBadRequest, "symbol and exchange are required")
		return
	}
	if indicator == "" {
		indicator = "ATR"
	}
	if interval == "" {
		interval = "1d"
	}

	// TODO: call queryClient.QuerySignals once stubs generated
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"symbol": symbol, "exchange": exchange, "indicator": indicator, "interval": interval,
		"points": []interface{}{},
	})
}
