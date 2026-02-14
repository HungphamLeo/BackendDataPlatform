package usecase

import (
	"context"
	"fmt"
	"time"

	"internal/marketdata/application/dto"
	"internal/marketdata/application/ports"
	"internal/marketdata/domain/value_object"
)

// ✅ USE CASE: Get Latest Price
type GetLatestPriceUseCase struct {
	// Ports (interfaces defined in ports.go)
	tickerRepo    ports.TickerRepository // ��� Infrastructure port
	cacheStore    ports.CacheStore       // 👈 Platform port
	binanceClient ports.binanceClient    // 👈 External API client
	logger        ports.Logger           // 👈 Platform logging
	metrics       ports.Metrics          // 👈 Platform metrics
}

func NewGetLatestPriceUseCase(
	tickerRepo ports.TickerRepository,
	cacheStore ports.CacheStore,
	binanceClient ports.binanceClient,
	logger ports.Logger,
	metrics ports.Metrics,
) *GetLatestPriceUseCase {
	return &GetLatestPriceUseCase{
		tickerRepo:    tickerRepo,
		cacheStore:    cacheStore,
		binanceClient: binanceClient,
		logger:        logger,
		metrics:       metrics,
	}
}

// 🎯 Execute use case
func (uc *GetLatestPriceUseCase) Execute(
	ctx context.Context,
	req *dto.GetLatestPriceRequest,
) (*dto.GetLatestPriceResponse, error) {
	// Step 1: Validate input
	if err := req.Validate(); err != nil {
		uc.logger.Warn(ctx, fmt.Sprintf("invalid request: %v", err))
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	symbol := value_object.Symbol(req.Symbol)

	// Step 2: Try cache first (platform/cache)
	cachedPrice, err := uc.cacheStore.Get(ctx, "price:"+req.Symbol)
	if err == nil && cachedPrice != "" {
		uc.logger.Debug(ctx, "price found in cache", "symbol", req.Symbol)
		uc.metrics.RecordCacheHit("price")

		return &dto.GetLatestPriceResponse{
			Symbol: req.Symbol,
			Bid:    cachedPrice.Bid,
			Ask:    cachedPrice.Ask,
			Source: "CACHE",
		}, nil
	}

	// Step 3: Call binance REST API (infrastructure/binance_client)
	uc.logger.Info(ctx, "fetching price from binance", "symbol", req.Symbol)
	quote, err := uc.binanceClient.GetQuote(ctx, symbol)
	if err != nil {
		uc.logger.Error(ctx, "failed to get quote from binance", err)
		uc.metrics.RecordError("get_price", "binance_api")
		return nil, fmt.Errorf("failed to get quote: %w", err)
	}

	// Step 4: Load into cache for next time (platform/cache)
	_ = uc.cacheStore.Set(ctx, "price:"+req.Symbol, quote, 5*time.Minute)

	// Step 5: Save ticker to database (persistence)
	ticker, err := uc.tickerRepo.GetBySymbol(ctx, symbol)
	if err != nil {
		ticker = domain.NewTicker(symbol, "FOREX")
	}

	ticker.UpdatePrice(quote)
	if err := uc.tickerRepo.Save(ctx, ticker); err != nil {
		uc.logger.Error(ctx, "failed to save ticker", err)
		// Don't fail - still return price to user
	}

	// Step 6: Record metrics (platform/metrics)
	uc.metrics.RecordPrice(symbol.String(), quote.MidPrice())

	// Step 7: Return response
	return &dto.GetLatestPriceResponse{
		Symbol: req.Symbol,
		Bid:    quote.Bid,
		Ask:    quote.Ask,
		Source: "binance_API",
	}, nil
}
