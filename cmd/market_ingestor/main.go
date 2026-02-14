package market_ingestor

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	ingestor "github.com/HungphamLeo/BackendDataPlatform/internal/marketdata/app/ingestor"
	binance "github.com/HungphamLeo/BackendDataPlatform/internal/platform/integrations/binance"
	kafka "github.com/HungphamLeo/BackendDataPlatform/internal/platform/messaging/kafka"
	redisclient "github.com/HungphamLeo/BackendDataPlatform/internal/platform/storage/redis"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	// Minimal config via env vars (safe defaults for local dev)
	kafkaBrokers := splitEnv("KAFKA_BROKERS", "localhost:9092")
	redisAddr := getenv("REDIS_ADDR", "localhost:6379")

	binanceAddress := getenv("binance_ADDRESS", "xapi.binance.com")
	apiPort := getenvInt("binance_API_PORT", 5124)
	streamPort := getenvInt("binance_STREAM_PORT", 5125)
	tlsEnabled := getenvBool("binance_TLS", true)

	userID := os.Getenv("binance_USER_ID")
	password := os.Getenv("binance_PASSWORD")
	appName := getenv("binance_APP_NAME", "go")

	symbols := splitEnv("binance_SYMBOLS", "EURUSD")

	// infra: kafka producer
	kp, err := kafka.NewProducer(kafkaBrokers)
	if err != nil {
		log.Fatalf("kafka producer init failed: %v", err)
	}
	defer kp.Close()

	// infra: redis client
	rc := redisclient.New(redisAddr)
	defer rc.Close()

	// integration: binance client
	cfg := binance.DefaultConfig()
	cfg.Address = binanceAddress
	cfg.APIPort = apiPort
	cfg.StreamingPort = streamPort
	cfg.TLSEnabled = tlsEnabled

	xClient := binance.NewClient(cfg, binance.Credentials{UserID: userID, Password: password, AppName: appName})
	defer xClient.Close()

	ing := ingestor.NewIngestor(ingestor.IngestorConfig{
		KafkaProducer: kp,
		RedisClient:   rc,
		binanceClient: xClient,
		TopicPrefix:   "binance",
		Symbols:       symbols,
	})

	// Channels you want to subscribe (mimics Python logic)
	ing.AddChannel("ticker")
	ing.AddChannel("balance")
	ing.AddChannel("order_status")

	go func() {
		if err := ing.Run(ctx); err != nil {
			log.Printf("ingestor run finished with error: %v", err)
			cancel()
		}
	}()

	<-ctx.Done()
	time.Sleep(500 * time.Millisecond)
	log.Println("shutdown complete")
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func splitEnv(key, def string) []string {
	v := getenv(key, def)
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func getenvInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	var n int
	_, err := fmt.Sscanf(v, "%d", &n)
	if err != nil {
		return def
	}
	return n
}

func getenvBool(key string, def bool) bool {
	v := strings.ToLower(os.Getenv(key))
	if v == "" {
		return def
	}
	switch v {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		return def
	}
}
