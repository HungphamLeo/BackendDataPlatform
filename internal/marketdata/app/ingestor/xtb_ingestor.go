package ingestor

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/HungphamLeo/BackendDataPlatform/internal/infra/kafka"
	"github.com/HungphamLeo/BackendDataPlatform/internal/infra/redisclient"
	"github.com/HungphamLeo/BackendDataPlatform/internal/infra/xtb"
	"github.com/HungphamLeo/BackendDataPlatform/pkg/recorders"
)

type IngestorConfig struct {
	KafkaProducer kafka.Producer
	RedisClient   *redisclient.RedisClient
	XTBClient     xtb.XTBClient
	TopicPrefix   string
}

type Ingestor struct {
	conf     IngestorConfig
	channels []string

	receiveCh <-chan string
	wg        sync.WaitGroup
}

func NewIngestor(cfg IngestorConfig) *Ingestor {
	return &Ingestor{
		conf: cfg,
	}
}

func (i *Ingestor) AddChannel(ch string) {
	// channel names like: "ticker", "balance", "candles:1m"
	for _, c := range i.channels {
		if c == ch {
			return
		}
	}
	i.channels = append(i.channels, ch)
}

func (i *Ingestor) Run(ctx context.Context) error {
	// connect XTB
	if err := i.conf.XTBClient.Connect(ctx); err != nil {
		return err
	}

	msgCh, err := i.conf.XTBClient.Receive()
	if err != nil {
		return err
	}
	i.receiveCh = msgCh

	// subscribe on channels based on configuration
	if err := i.subscribeAll(); err != nil {
		log.Printf("subscribe warning: %v", err)
	}

	i.wg.Add(1)
	go i.readLoop(ctx)
	return nil
}

func (i *Ingestor) subscribeAll() error {
	// Build and send subscribe commands to XTB client. Implementation detail depends on protocol.
	// Here we send simple JSON "subscribe" commands. Adapt to actual XTB socket protocol.
	for _, ch := range i.channels {
		parts := strings.Split(ch, ":")
		cmd := map[string]interface{}{
			"command": "subscribe",
			"channel": parts[0],
		}
		if len(parts) > 1 {
			cmd["sub"] = parts[1]
		}
		_ = i.conf.XTBClient.Send(cmd)
	}
	return nil
}

func (i *Ingestor) readLoop(ctx context.Context) {
	defer i.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case raw, ok := <-i.receiveCh:
			if !ok {
				return
			}
			// process message in separate goroutine to maintain throughput
			i.wg.Add(1)
			go func(msg string) {
				defer i.wg.Done()
				if err := i.processMessage(ctx, msg); err != nil {
					log.Printf("processMessage error: %v", err)
				}
			}(raw)
		}
	}
}

func (i *Ingestor) processMessage(ctx context.Context, raw string) error {
	// raw is JSON string; parse into a generic envelope:
	var envelope map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &envelope); err != nil {
		return err
	}

	// Determine the type/command. This mapping should match XTB messages
	// e.g., envelope["command"] or envelope["type"] etc.
	cmd, _ := envelope["command"].(string)
	data := envelope["data"]

	switch cmd {
	case "tickPrices", "ticker":
		rec := recorders.TickerRecordFromRaw(data)
		// cache in redis
		b, _ := json.Marshal(rec)
		_ = i.conf.RedisClient.SetJSON(ctx, "xtb:ticker:"+rec.Symbol, string(b), 10*time.Second)
		_ = i.conf.RedisClient.SetJSON(ctx, "xtb:candles:"+timeframe, string(b), 2*time.Minute)

		// publish to kafka
		_ = i.conf.KafkaProducer.Publish(ctx, i.conf.TopicPrefix+"-ticker", []byte(rec.Symbol), b)
	case "candles", "candle":
		// map to internal candle structure(s)
		cands := recorders.CandlesRecordFromRaw(data)
		for timeframe, list := range cands {
			for _, c := range list {
				b, _ := json.Marshal(c)
				_ = i.conf.KafkaProducer.Publish(ctx, i.conf.TopicPrefix+"-candles-"+timeframe, []byte(c.StartTime.String()), b)
				// Optionally cache latest candle per timeframe
				_ = i.conf.RedisClient.SetJSON(ctx, "xtb:ticker:"+rec.Symbol, string(b), 10*time.Second)
				_ = i.conf.RedisClient.SetJSON(ctx, "xtb:candles:"+timeframe, string(b), 2*time.Minute)

		}
	default:
		// fallback: publish raw to an "events" topic
		b, _ := json.Marshal(envelope)
		_ = i.conf.KafkaProducer.Publish(ctx, i.conf.TopicPrefix+"-events", nil, b)
	}
	return nil
}