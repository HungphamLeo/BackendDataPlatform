package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config đại diện cho cấu hình của ứng dụng
type Config struct {
	App      AppConfig      `mapstructure:"app"`
	HTTP     HTTPConfig     `mapstructure:"http"`
	GRPC     GRPCConfig     `mapstructure:"grpc"`
	Logging  LoggingConfig  `mapstructure:"logging"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Kafka    KafkaConfig    `mapstructure:"kafka"`
	Binance  BinanceConfig  `mapstructure:"binance"`
	XTB      XTBConfig      `mapstructure:"xtb"`
	Endpoints EndpointsConfig `mapstructure:"endpoints"`
}

// AppConfig đại diện cho cấu hình của ứng dụng
type AppConfig struct {
	Name        string `mapstructure:"name"`
	Version     string `mapstructure:"version"`
	Environment string `mapstructure:"environment"`
}

// HTTPConfig đại diện cho cấu hình của HTTP server
type HTTPConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	IdleTimeout     time.Duration `mapstructure:"idle_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

// GRPCConfig đại diện cho cấu hình của gRPC server
type GRPCConfig struct {
	Host                 string        `mapstructure:"host"`
	Port                 int           `mapstructure:"port"`
	MaxConcurrentStreams uint32        `mapstructure:"max_concurrent_streams"`
	ConnectionTimeout    time.Duration `mapstructure:"connection_timeout"`
	MaxConnectionIdle    time.Duration `mapstructure:"max_connection_idle"`
	MaxConnectionAge     time.Duration `mapstructure:"max_connection_age"`
}

// LoggingConfig đại diện cho cấu hình của logging
type LoggingConfig struct {
	Level    string `mapstructure:"level"`
	Format   string `mapstructure:"format"`
	Output   string `mapstructure:"output"`
	FilePath string `mapstructure:"file_path"`
}

// DatabaseConfig đại diện cho cấu hình của database
type DatabaseConfig struct {
	Driver          string        `mapstructure:"driver"`
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	Username        string        `mapstructure:"username"`
	Password        string        `mapstructure:"password"`
	Database        string        `mapstructure:"database"`
	SSLMode         string        `mapstructure:"ssl_mode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

// RedisConfig đại diện cho cấu hình của Redis
type RedisConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	Password        string        `mapstructure:"password"`
	DB              int           `mapstructure:"db"`
	PoolSize        int           `mapstructure:"pool_size"`
	MinIdleConns    int           `mapstructure:"min_idle_conns"`
	DialTimeout     time.Duration `mapstructure:"dial_timeout"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	PoolTimeout     time.Duration `mapstructure:"pool_timeout"`
	IdleTimeout     time.Duration `mapstructure:"idle_timeout"`
	MaxRetries      int           `mapstructure:"max_retries"`
	MinRetryBackoff time.Duration `mapstructure:"min_retry_backoff"`
	MaxRetryBackoff time.Duration `mapstructure:"max_retry_backoff"`
}

// KafkaConfig đại diện cho cấu hình của Kafka
type KafkaConfig struct {
	BootstrapServers   string        `mapstructure:"bootstrap_servers"`
	ClientID           string        `mapstructure:"client_id"`
	ConsumerGroup      string        `mapstructure:"consumer_group"`
	AutoOffsetReset    string        `mapstructure:"auto_offset_reset"`
	EnableAutoCommit   bool          `mapstructure:"enable_auto_commit"`
	AutoCommitInterval time.Duration `mapstructure:"auto_commit_interval"`
	SessionTimeout     time.Duration `mapstructure:"session_timeout"`
	HeartbeatInterval  time.Duration `mapstructure:"heartbeat_interval"`
	MaxPollRecords     int           `mapstructure:"max_poll_records"`
	Topics             KafkaTopics   `mapstructure:"topics"`
}

// KafkaTopics đại diện cho cấu hình của các topic Kafka
type KafkaTopics struct {
	PriceUpdated   string `mapstructure:"price_updated"`
	NewTrade       string `mapstructure:"new_trade"`
	AlertTriggered string `mapstructure:"alert_triggered"`
	Default        string `mapstructure:"default"`
}

// BinanceConfig đại diện cho cấu hình của Binance
type BinanceConfig struct {
	Spot    BinanceSpotConfig    `mapstructure:"spot"`
	Futures BinanceFuturesConfig `mapstructure:"futures"`
}

// BinanceSpotConfig đại diện cho cấu hình của Binance Spot
type BinanceSpotConfig struct {
	BaseURL    string             `mapstructure:"base_url"`
	WSURL      string             `mapstructure:"ws_url"`
	APIKey     string             `mapstructure:"api_key"`
	SecretKey  string             `mapstructure:"secret_key"`
	Symbols    []string           `mapstructure:"symbols"`
	Timeframes []string           `mapstructure:"timeframes"`
	RateLimit  BinanceRateLimit   `mapstructure:"rate_limit"`
}

// BinanceFuturesConfig đại diện cho cấu hình của Binance Futures
type BinanceFuturesConfig struct {
	BaseURL    string             `mapstructure:"base_url"`
	WSURL      string             `mapstructure:"ws_url"`
	APIKey     string             `mapstructure:"api_key"`
	SecretKey  string             `mapstructure:"secret_key"`
	Symbols    []string           `mapstructure:"symbols"`
	Timeframes []string           `mapstructure:"timeframes"`
	RateLimit  BinanceRateLimit   `mapstructure:"rate_limit"`
}

// BinanceRateLimit đại diện cho cấu hình của rate limit Binance
type BinanceRateLimit struct {
	RequestsPerMinute int `mapstructure:"requests_per_minute"`
	OrdersPerSecond   int `mapstructure:"orders_per_second"`
	OrdersPerDay      int `mapstructure:"orders_per_day"`
}

// XTBConfig đại diện cho cấu hình của XTB
type XTBConfig struct {
	BaseURL    string       `mapstructure:"base_url"`
	APIKey     string       `mapstructure:"api_key"`
	Password   string       `mapstructure:"password"`
	Symbols    []string     `mapstructure:"symbols"`
	Timeframes []string     `mapstructure:"timeframes"`
	RateLimit  XTBRateLimit `mapstructure:"rate_limit"`
}

// XTBRateLimit đại diện cho cấu hình của rate limit XTB
type XTBRateLimit struct {
	RequestsPerMinute int `mapstructure:"requests_per_minute"`
}

// EndpointsConfig đại diện cho cấu hình của các endpoint
type EndpointsConfig struct {
	Ticker    EndpointConfig `mapstructure:"ticker"`
	Tickers   EndpointConfig `mapstructure:"tickers"`
	Candles   EndpointConfig `mapstructure:"candles"`
	Trades    EndpointConfig `mapstructure:"trades"`
	OrderBook EndpointConfig `mapstructure:"order_book"`
}

// EndpointConfig đại diện cho cấu hình của một endpoint
type EndpointConfig struct {
	Path      string        `mapstructure:"path"`
	Method    string        `mapstructure:"method"`
	RateLimit int           `mapstructure:"rate_limit"`
	CacheTTL  time.Duration `mapstructure:"cache_ttl"`
}

// LoadConfig loads configuration from file and environment variables
func LoadConfig(configPath string) (*Config, error) {
	v := viper.New()
	
	// Set default configuration file path
	if configPath == "" {
		configPath = "config/marketdata.yaml"
	}
	
	// Set configuration file path
	v.SetConfigFile(configPath)
	
	// Read configuration file
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	
	// Set environment variables prefix
	v.SetEnvPrefix("MARKETDATA")
	
	// Replace dots with underscores in environment variables
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	
	// Enable environment variables override
	v.AutomaticEnv()
	
	// Expand environment variables in configuration values
	for _, key := range v.AllKeys() {
		val := v.GetString(key)
		if strings.HasPrefix(val, "${") && strings.HasSuffix(val, "}") {
			envKey := val[2 : len(val)-1]
			envVal := os.Getenv(envKey)
			if envVal != "" {
				v.Set(key, envVal)
			}
		}
	}
	
	// Unmarshal configuration into struct
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	
	return &config, nil
}
