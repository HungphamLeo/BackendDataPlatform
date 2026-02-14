package main

import (
	"log"

	"binance-streaming/internal/kafka"
	"binance-streaming/internal/websocket"
)

func main() {

	wsURL :=
		"wss://stream.binance.com:9443/ws/btcusdt@trade"

	kafkaBrokers := []string{
		"localhost:9092",
	}
	topic := "binance-trades"

	// Kafka producer
	producer := kafka.NewProducer(
		kafkaBrokers,
		topic,
	)

	defer producer.Close()

	// WebSocket client
	client, err :=
		websocket.NewClient(wsURL)

	if err != nil {

		log.Fatal(err)
	}

	defer client.Close()

	log.Println("Connected to Binance")

	client.Read(func(message []byte) {

		log.Println(string(message))

		err :=
			producer.Send(message)

		if err != nil {
			log.Println(err)
		}

	})
}
