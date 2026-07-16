package rest

import (
	"net/http"

	"github.com/HungphamLeo/BackendDataPlatform/api/clients"
)

// SignalsHandler serves /v1/signals endpoints.
type SignalsHandler struct {
	queryClient *clients.QueryClient
}

// NewSignalsHandler creates a SignalsHandler.
func NewSignalsHandler(qc *clients.QueryClient) *SignalsHandler {
	return &SignalsHandler{queryClient: qc}
}

// GetSignals handles GET /v1/signals?symbol=BTC_USDT&exchange=binance&interval=1d
func (h *SignalsHandler) GetSignals(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	symbol := q.Get("symbol")
	exchange := q.Get("exchange")
	interval := q.Get("interval")
	if symbol == "" {
		writeError(w, http.StatusBadRequest, "symbol is required")
		return
	}
	if exchange == "" {
		exchange = "binance"
	}
	if interval == "" {
		interval = "1d"
	}

	// TODO: aggregate signals from queryClient.QuerySignals once stubs generated
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"symbol": symbol, "exchange": exchange, "interval": interval,
		"signals": []interface{}{},
	})
}
