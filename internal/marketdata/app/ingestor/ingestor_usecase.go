package app

import (
	"context"

	"github.com/HungphamLeo/BackendDataPlatform/internal/marketdata/domain"
	"github.com/HungphamLeo/BackendDataPlatform/internal/platform/logging"
)

// MarketIngestorUseCase điều phối luồng dữ liệu (ETL / Ingestion)
type MarketIngestorUseCase struct {
	provider  domain.MarketDataProvider
	publisher domain.MarketDataPublisher
	topic     string
	logger    logging.Logger
}

func NewMarketIngestorUseCase(provider domain.MarketDataProvider, publisher domain.MarketDataPublisher, topic string, logger logging.Logger) *MarketIngestorUseCase {
	return &MarketIngestorUseCase{
		provider:  provider,
		publisher: publisher,
		topic:     topic,
		logger:    logger,
	}
}

func (u *MarketIngestorUseCase) Start(ctx context.Context, symbols []string) error {
	if err := u.provider.Connect(ctx); err != nil {
		return err
	}
	if err := u.provider.Subscribe(symbols); err != nil {
		return err
	}

	tickCh := make(chan *domain.Tick, 1000)
	errCh := make(chan error, 10)

	go u.provider.Stream(ctx, tickCh, errCh)

	for {
		select {
		case <-ctx.Done():
			return u.provider.Close()
		case tick := <-tickCh:
			_ = u.publisher.PublishTick(ctx, u.topic, tick)
		case err := <-errCh:
			logging.LogEvent(u.logger, "ERROR", logging.EventLogSchema{
				Topic:   logging.TopicMarketData,
				Service: "market_ingestor",
				Action:  "STREAM_TICK",
				Status:  "FAILED",
				Payload: map[string]string{"error": err.Error()},
			}, "Error from Market Data Provider")
		}
	}
}