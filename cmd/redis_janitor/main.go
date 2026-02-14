package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	redisclient "github.com/HungphamLeo/BackendDataPlatform/internal/platform/storage/redis"
	"github.com/HungphamLeo/BackendDataPlatform/internal/platform/storage/redis/janitor"
)

func main() {
	redisAddr := getenv("REDIS_ADDR", "localhost:6379")

	// Examples:
	// REDIS_CLEAN_INTERVAL=10m
	// REDIS_CLEAN_PATTERNS=binance:ticker:*,binance:candles:*,binance:balance:*
	interval := mustDuration(getenv("REDIS_CLEAN_INTERVAL", "10m"))
	patterns := splitCSV(getenv("REDIS_CLEAN_PATTERNS", "binance:ticker:*,binance:candles:*"))

	threshold := getenvFloat("REDIS_MEM_THRESHOLD", 0.80)

	rdb := redisclient.New(redisAddr)
	defer rdb.Close()

	conf := janitor.Config{
		Interval:         interval,
		MemoryThreshold:  threshold,
		Patterns:         patterns,
		ScanCount:        1000,
		DeleteBatchSize:  500,
		MaxDeletesPerRun: 200000,
	}

	j := janitor.New(rdb, conf)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := j.Run(ctx); err != nil && err != context.Canceled {
			log.Printf("redis janitor stopped: %v", err)
		}
	}()

	log.Printf("redis janitor running: addr=%s interval=%s patterns=%v threshold=%.2f",
		redisAddr, interval, patterns, threshold)

	// graceful shutdown
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	cancel()
	time.Sleep(300 * time.Millisecond)
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func mustDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 10 * time.Minute
	}
	return d
}

func getenvFloat(k string, d float64) float64 {
	if v := os.Getenv(k); v != "" {
		// cheap parse
		var f float64
		_, err := fmt.Sscanf(v, "%f", &f)
		if err == nil {
			return f
		}
	}
	return d
}
