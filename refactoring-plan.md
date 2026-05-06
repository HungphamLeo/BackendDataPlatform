# Kế hoạch Refactoring Marketdata Service

## 1. Cấu trúc thư mục theo Clean Architecture

```
internal/
  marketdata/
    domain/                     # Domain Layer - Business rules & entities
      entity/                   # Business entities
      value_object/             # Value objects
      aggregate/                # Aggregate roots
      repository/               # Repository interfaces
      service/                  # Domain service interfaces
      event/                    # Domain events
    
    application/                # Application Layer - Use cases & orchestration
      dto/                      # Data Transfer Objects
      port/                     # Input & Output ports
        in/                     # Input ports (use cases)
        out/                    # Output ports (repositories, messaging)
      usecase/                  # Use case implementations
      mapper/                   # Mappers between domain & DTOs
    
    adapter/                    # Adapter Layer - External services & persistence
      provider/                 # External service providers
        binance/                # Binance integration
          spot/                 # Binance Spot
            rest/               # REST API client
            streaming/          # WebSocket client
          futures/              # Binance Futures
            rest/               # REST API client
            streaming/          # WebSocket client
          common/               # Common code for Binance
            mapper/             # Mappers for Binance data
            model/              # Binance-specific models
      persistence/              # Database implementations
        postgres/               # PostgreSQL repositories
        redis/                  # Redis repositories
      messaging/                # Messaging implementations
        kafka/                  # Kafka producers/consumers
    
    infrastructure/             # Infrastructure Layer - Technical concerns
      config/                   # Configuration
        yaml/                   # YAML config files
      logging/                  # Logging setup
      metrics/                  # Metrics & monitoring
      cache/                    # Caching mechanisms
      database/                 # Database connections
      messaging/                # Messaging connections
    
    transport/                  # Transport Layer - API endpoints
      http/                     # HTTP handlers
      grpc/                     # gRPC handlers
      websocket/                # WebSocket handlers
```

## 2. Domain Layer

### 2.1. Entities & Value Objects
- `Symbol` (Value Object): Chuẩn hóa symbol giữa các sàn
- `Timeframe` (Value Object): Chuẩn hóa timeframe giữa các sàn
- `Quote` (Value Object): Giá bid/ask
- `Candle` (Entity): Nến OHLCV
- `Trade` (Entity): Giao dịch

### 2.2. Aggregates
- `Ticker` (Aggregate Root): Quản lý thông tin giá của một symbol
- `OrderBook` (Aggregate Root): Quản lý order book của một symbol

### 2.3. Repositories
- `TickerRepository`: Lưu trữ và truy xuất ticker
- `CandleRepository`: Lưu trữ và truy xuất candle
- `TradeRepository`: Lưu trữ và truy xuất trade
- `OrderBookRepository`: Lưu trữ và truy xuất order book

### 2.4. Domain Events
- `PriceUpdated`: Khi giá thay đổi
- `CandleClosed`: Khi nến đóng
- `NewTrade`: Khi có giao dịch mới
- `OrderBookUpdated`: Khi order book cập nhật

## 3. Application Layer

### 3.1. DTOs
- `TickerDTO`: Data transfer object cho ticker
- `CandleDTO`: Data transfer object cho candle
- `TradeDTO`: Data transfer object cho trade
- `OrderBookDTO`: Data transfer object cho order book

### 3.2. Use Cases
- `GetLatestPriceUseCase`: Lấy giá mới nhất
- `GetCandlesUseCase`: Lấy dữ liệu nến
- `GetTradesUseCase`: Lấy dữ liệu giao dịch
- `GetOrderBookUseCase`: Lấy dữ liệu order book
- `StreamPriceUseCase`: Stream giá realtime
- `StreamCandlesUseCase`: Stream nến realtime
- `StreamTradesUseCase`: Stream giao dịch realtime
- `StreamOrderBookUseCase`: Stream order book realtime

### 3.3. Ports
- Input Ports: Interface cho use cases
- Output Ports: Interface cho repositories và messaging

## 4. Adapter Layer

### 4.1. Binance Integration
- REST API Clients:
  - `BinanceSpotRestClient`: Client cho Binance Spot REST API
  - `BinanceFuturesRestClient`: Client cho Binance Futures REST API
- WebSocket Clients:
  - `BinanceSpotWebSocketClient`: Client cho Binance Spot WebSocket
  - `BinanceFuturesWebSocketClient`: Client cho Binance Futures WebSocket
- Mappers:
  - `BinanceSpotMapper`: Chuyển đổi dữ liệu từ Binance Spot sang domain model
  - `BinanceFuturesMapper`: Chuyển đổi dữ liệu từ Binance Futures sang domain model

### 4.2. Persistence
- PostgreSQL Repositories:
  - `PostgresTickerRepository`: Triển khai TickerRepository với PostgreSQL
  - `PostgresCandleRepository`: Triển khai CandleRepository với PostgreSQL
  - `PostgresTradeRepository`: Triển khai TradeRepository với PostgreSQL
- Redis Repositories:
  - `RedisTickerRepository`: Cache cho ticker
  - `RedisOrderBookRepository`: Cache cho order book

### 4.3. Messaging
- Kafka Producers:
  - `KafkaPriceProducer`: Đẩy giá lên Kafka
  - `KafkaCandleProducer`: Đẩy nến lên Kafka
  - `KafkaTradeProducer`: Đẩy giao dịch lên Kafka
- Kafka Consumers:
  - `KafkaPriceConsumer`: Nhận giá từ Kafka
  - `KafkaCandleConsumer`: Nhận nến từ Kafka
  - `KafkaTradeConsumer`: Nhận giao dịch từ Kafka

## 5. Infrastructure Layer

### 5.1. Configuration
- YAML Config Files:
  - `binance.yaml`: Cấu hình cho Binance (endpoints, API keys)
  - `database.yaml`: Cấu hình cho PostgreSQL
  - `redis.yaml`: Cấu hình cho Redis
  - `kafka.yaml`: Cấu hình cho Kafka
  - `logging.yaml`: Cấu hình cho logging

### 5.2. Logging
- `ZapLoggerManager`: Quản lý logger với Zap
- `LoggerFactory`: Factory để tạo logger

### 5.3. Database
- `PostgresConnection`: Kết nối PostgreSQL
- `RedisConnection`: Kết nối Redis

### 5.4. Messaging
- `KafkaConnection`: Kết nối Kafka

## 6. Transport Layer

### 6.1. HTTP Handlers
- `TickerHandler`: Handler cho ticker API
- `CandleHandler`: Handler cho candle API
- `TradeHandler`: Handler cho trade API
- `OrderBookHandler`: Handler cho order book API

### 6.2. gRPC Handlers
- `MarketDataServiceServer`: gRPC server cho market data

### 6.3. WebSocket Handlers
- `MarketDataWebSocketHandler`: WebSocket handler cho market data

## 7. Kế hoạch triển khai

1. **Bước 1**: Tạo cấu trúc thư mục
2. **Bước 2**: Triển khai Domain Layer
3. **Bước 3**: Triển khai Application Layer
4. **Bước 4**: Triển khai Infrastructure Layer (Config, Logging)
5. **Bước 5**: Triển khai Adapter Layer (Binance Integration)
6. **Bước 6**: Triển khai Adapter Layer (Persistence, Messaging)
7. **Bước 7**: Triển khai Transport Layer
8. **Bước 8**: Tích hợp và kiểm thử
