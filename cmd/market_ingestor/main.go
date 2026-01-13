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
	xtb "github.com/HungphamLeo/BackendDataPlatform/internal/platform/integrations/xtb"
	kafka "github.com/HungphamLeo/BackendDataPlatform/internal/platform/messaging/kafka"
	redisclient "github.com/HungphamLeo/BackendDataPlatform/internal/platform/storage/redis"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	// Minimal config via env vars (safe defaults for local dev)
	kafkaBrokers := splitEnv("KAFKA_BROKERS", "localhost:9092")
	redisAddr := getenv("REDIS_ADDR", "localhost:6379")

	xtbAddress := getenv("XTB_ADDRESS", "xapi.xtb.com")
	apiPort := getenvInt("XTB_API_PORT", 5124)
	streamPort := getenvInt("XTB_STREAM_PORT", 5125)
	tlsEnabled := getenvBool("XTB_TLS", true)

	userID := os.Getenv("XTB_USER_ID")
	password := os.Getenv("XTB_PASSWORD")
	appName := getenv("XTB_APP_NAME", "go")

	symbols := splitEnv("XTB_SYMBOLS", "EURUSD")

	// infra: kafka producer
	kp, err := kafka.NewProducer(kafkaBrokers)
	if err != nil {
		log.Fatalf("kafka producer init failed: %v", err)
	}
	defer kp.Close()

	// infra: redis client
	rc := redisclient.New(redisAddr)
	defer rc.Close()

	// integration: xtb client
	cfg := xtb.DefaultConfig()
	cfg.Address = xtbAddress
	cfg.APIPort = apiPort
	cfg.StreamingPort = streamPort
	cfg.TLSEnabled = tlsEnabled

	xClient := xtb.NewClient(cfg, xtb.Credentials{UserID: userID, Password: password, AppName: appName})
	defer xClient.Close()

	ing := ingestor.NewIngestor(ingestor.IngestorConfig{
		KafkaProducer: kp,
		RedisClient:   rc,
		XTBClient:     xClient,
		TopicPrefix:   "xtb",
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
