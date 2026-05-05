package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/HungphamLeo/BackendDataPlatform/internal/marketdata/adapter/kafka"
	"github.com/HungphamLeo/BackendDataPlatform/internal/marketdata/adapter/provider/binance/futures/streaming"
	"github.com/HungphamLeo/BackendDataPlatform/internal/marketdata/app"
	"github.com/HungphamLeo/BackendDataPlatform/internal/platform/logging"
)

func main() {
	// Khởi tạo Logger và ghi vào thư mục logs (Hệ thống sẽ tự tạo thư mục nếu chưa có)
	logging.InitLogger("logs/market_ingestor/app.log")
	logger := logging.NewLogger()

	logger.Info("Starting Market Ingestor Service (Binance Futures)...")

	// Bật Prometheus endpoint (non-blocking) trên port 9090
	logging.StartPrometheusEndpoint()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Khởi tạo Adapter Layer: Tách biệt hoàn toàn, dễ dàng đổi nhà cung cấp
	binanceFuturesProvider := streaming.NewBinanceFuturesClient(logger)
	kafkaPublisher := kafka.NewKafkaPublisher()

	// 2. Khởi tạo Application Layer: Inject dependencies vào Use Case
	useCase := app.NewMarketIngestorUseCase(binanceFuturesProvider, kafkaPublisher, "market_data.ticks", logger)

	// 3. Khởi chạy quá trình Ingestion trong một Goroutine riêng (Non-blocking)
	go func() {
		// Các cặp coin cần stream giá realtime
		symbols := []string{"BTCUSDT", "ETHUSDT"}
		if err := useCase.Start(ctx, symbols); err != nil {
			logger.Fatal("Ingestor stopped with error")
		}
	}()

	// 4. Graceful Shutdown (lắng nghe tín hiệu từ OS/Docker để đóng tài nguyên)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	logger.Info("Market Ingestor Service gracefully shutting down...")
	logging.Sync() // Đảm bảo mọi log trong buffer đều được xả ra File trước khi service tắt hoàn toàn
}