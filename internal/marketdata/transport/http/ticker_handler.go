package http

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/HungphamLeo/BackendDataPlatform/internal/marketdata/application/dto"
	"github.com/HungphamLeo/BackendDataPlatform/internal/marketdata/application/port/in"
	"github.com/HungphamLeo/BackendDataPlatform/internal/platform/logging"
)

// TickerHandler xử lý các request HTTP liên quan đến ticker
type TickerHandler struct {
	getTickerUseCase  in.GetTickerUseCase
	getTickersUseCase in.GetTickersUseCase
	logger            logging.Logger
}

// NewTickerHandler tạo mới TickerHandler
func NewTickerHandler(
	getTickerUseCase in.GetTickerUseCase,
	getTickersUseCase in.GetTickersUseCase,
	logger logging.Logger,
) *TickerHandler {
	return &TickerHandler{
		getTickerUseCase:  getTickerUseCase,
		getTickersUseCase: getTickersUseCase,
		logger:            logger,
	}
}

// RegisterRoutes đăng ký các route cho TickerHandler
func (h *TickerHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/api/v1/tickers/{exchange}/{symbol}", h.GetTicker).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/tickers/{exchange}", h.GetTickers).Methods(http.MethodGet)
}

// GetTicker xử lý request lấy dữ liệu ticker
func (h *TickerHandler) GetTicker(w http.ResponseWriter, r *http.Request) {
	// Lấy tham số từ URL
	vars := mux.Vars(r)
	exchange := vars["exchange"]
	symbol := vars["symbol"]
	
	// Tạo request DTO
	request := &dto.GetTickerRequest{
		Symbol:   symbol,
		Exchange: exchange,
	}
	
	// Gọi use case
	response, err := h.getTickerUseCase.Execute(r.Context(), request)
	if err != nil {
		h.logger.Error("Failed to get ticker", logging.Fields{
			"error":    err.Error(),
			"symbol":   symbol,
			"exchange": exchange,
		})
		
		// Trả về lỗi
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})
		return
	}
	
	// Trả về kết quả
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetTickers xử lý request lấy dữ liệu ticker cho nhiều symbol
func (h *TickerHandler) GetTickers(w http.ResponseWriter, r *http.Request) {
	// Lấy tham số từ URL
	vars := mux.Vars(r)
	exchange := vars["exchange"]
	
	// Lấy danh sách symbol từ query parameter
	symbols := r.URL.Query()["symbol"]
	if len(symbols) == 0 {
		// Trả về lỗi
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "at least one symbol is required",
		})
		return
	}
	
	// Tạo request DTO
	request := &dto.GetTickersRequest{
		Symbols:  symbols,
		Exchange: exchange,
	}
	
	// Gọi use case
	response, err := h.getTickersUseCase.Execute(r.Context(), request)
	if err != nil {
		h.logger.Error("Failed to get tickers", logging.Fields{
			"error":    err.Error(),
			"symbols":  symbols,
			"exchange": exchange,
		})
		
		// Trả về lỗi
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})
		return
	}
	
	// Trả về kết quả
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
