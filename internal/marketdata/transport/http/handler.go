package http

import (
	"encoding/json"
	"net/http"

	"internal/marketdata/application/dto"
	"internal/marketdata/application/usecase"
	"internal/platform/observability"
	"internal/platform/transport/http_utils"
)

// ✅ HTTP HANDLER - Expose use case via REST API
type PriceHandler struct {
	getLatestPriceUC *usecase.GetLatestPriceUseCase
	logger           observability.Logger
}

func NewPriceHandler(
	getLatestPriceUC *usecase.GetLatestPriceUseCase,
	logger observability.Logger,
) *PriceHandler {
	return &PriceHandler{
		getLatestPriceUC: getLatestPriceUC,
		logger:           logger,
	}
}

// GET /api/v1/prices/:symbol
func (h *PriceHandler) GetLatestPrice(w http.ResponseWriter, r *http.Request) {
	// Step 1: Parse request
	symbol := r.PathValue("symbol")

	req := &dto.GetLatestPriceRequest{
		Symbol: symbol,
	}

	// Step 2: Call use case
	resp, err := h.getLatestPriceUC.Execute(r.Context(), req)
	if err != nil {
		h.logger.Error(r.Context(), "failed to get price", err)
		http_utils.ErrorResponse(w, "Failed to get price", http.StatusInternalServerError)
		return
	}

	// Step 3: Format response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// WebSocket: GET /api/v1/prices/stream?symbols=EUR/USD,AAPL
func (h *PriceHandler) StreamPrices(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	symbols := r.URL.Query()["symbols"]

	// Upgrade HTTP to WebSocket
	// (implementation details...)
}
