package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {

	Binance struct {

		Spot struct {

			WebSocket struct {

				BaseURL string `yaml:"base_url"`
				Streams []string `yaml:"streams"`

				ReconnectIntervalSec int `yaml:"reconnect_interval_sec"`
				PingIntervalSec int `yaml:"ping_interval_sec"`

			} `yaml:"websocket"`

		} `yaml:"spot"`

	} `yaml:"binance"`

	Kafka struct {

		Brokers []string `yaml:"brokers"`

		Topic string `yaml:"topic"`

	} `yaml:"kafka"`
}

func Load(path string) (*Config, error) {

	data, err :=
		os.ReadFile(path)

	if err != nil {
		return nil, err
	}

	var cfg Config

	err =
		yaml.Unmarshal(data, &cfg)

	return &cfg, err
}
