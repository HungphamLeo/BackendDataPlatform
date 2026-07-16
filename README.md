# BackendDataPlatform

Nền tảng dữ liệu thị trường tài chính thời gian thực, được xây dựng bằng **Golang** (backend chính), theo mô hình **Microservices + Event-Driven Architecture (EDA)**. Hệ thống thu thập, xử lý và cung cấp dữ liệu từ các sàn giao dịch **Binance** và **XTB** qua REST, gRPC và WebSocket.

---

## Mục Lục

1. [Tổng quan kiến trúc](#1-tổng-quan-kiến-trúc)
2. [Tech Stack](#2-tech-stack)
3. [Cấu trúc thư mục](#3-cấu-trúc-thư-mục)
4. [Các Microservices](#4-các-microservices)
5. [Clean Architecture (4 Layer)](#5-clean-architecture-4-layer)
6. [Các Pattern Quan Trọng](#6-các-pattern-quan-trọng)
7. [Hướng dẫn chạy Local](#7-hướng-dẫn-chạy-local)
8. [Lệnh Build & Test](#8-lệnh-build--test)
9. [Cấu hình](#9-cấu-hình)
10. [Database Migrations](#10-database-migrations)

---

## 1. Tổng quan kiến trúc

Hệ thống được thiết kế theo luồng **Ingest → Process → Serve**:

```
Binance / XTB
     │  WebSocket / REST API
     ▼
┌─────────────┐      Kafka Topics       ┌──────────────────┐
│   Ingestor  │ ──────────────────────► │ Stream Processor │
│  (Go)       │  market-data.price-     │  (Go)            │
└─────────────┘  updated / new-trade    └────────┬─────────┘
                                                  │ MySQL
                                        ┌─────────▼─────────┐
                                        │  Streaming Service │ gRPC :9092
                                        │  (Go)             │
                                        └─────────┬─────────┘
                                                  │
                                        ┌─────────▼─────────┐
┌──────────────┐   Batch gRPC :9093     │   Query Service   │ gRPC :9091
│ Batch Svc    │ ──────────────────────►│   (Go)            │
│ (Python)     │                        └─────────┬─────────┘
└──────────────┘                                  │ Redis Cache
                                        ┌─────────▼─────────┐
                                        │    API Service    │ REST :8090 / gRPC :9090
                                        │    (Go)           │
                                        └───────────────────┘
```

**Kafka Broker:** `localhost:9092`  
**Realtime Gateway (WebSocket):** `cmd/realtime_gateway`  
**API Gateway (HTTP + gRPC):** `cmd/api_gateway`

---

## 2. Tech Stack

| Thành phần | Công nghệ |
|---|---|
| Ngôn ngữ chính | Go 1.23 |
| Batch / Indicators | Python 3 |
| Message Broker | Apache Kafka (Confluent 7.5) |
| Relational DB (streaming) | MySQL 8.0 |
| Relational DB (trading) | PostgreSQL 15 |
| Cache | Redis 7 |
| Transport internal | gRPC + Protobuf |
| Transport public | REST (gorilla/mux) + WebSocket |
| Config | Viper (YAML + env override) |
| Logging | Uber Zap (JSON structured) |
| ORM | GORM (PostgreSQL / MySQL) |
| Observability | OpenTelemetry (tracing + metrics) |
| Container | Docker / Docker Compose |
| Orchestration | Kubernetes (manifests tại `infra/kubernetes/`) |

---

## 3. Cấu trúc thư mục

```text
BackendDataPlatform/
├── cmd/                        # Entry point của từng service
│   ├── api/                    # REST + gRPC API public
│   ├── api_gateway/            # API Gateway (HTTP + gRPC routes, wire)
│   ├── ingestor/               # Thu thập dữ liệu từ Binance/XTB → Kafka
│   ├── market_ingestor/        # Market ingestor (phiên bản cũ / song song)
│   ├── query/                  # Query service (đọc từ streaming + batch)
│   ├── realtime_gateway/       # WebSocket gateway cho client
│   ├── redis_janitor/          # Dọn dẹp các key hết hạn trong Redis
│   ├── stream_processor/       # Xử lý event từ Kafka → DB
│   ├── streaming/              # Streaming service (cung cấp gRPC)
│   └── history_writer/         # Ghi dữ liệu lịch sử (consumers + wire)
│
├── internal/
│   ├── gateway/                # Domain: Gateway service
│   │   ├── adapter/
│   │   ├── app/
│   │   ├── contracts/
│   │   └── transport/http/
│   ├── marketdata/             # Domain: Market Data service
│   │   ├── domain/             # Entities, Value Objects, Aggregates, Events
│   │   ├── application/        # Use Cases, DTOs, Input/Output Ports
│   │   ├── adapter/            # Repo implementations, Kafka publisher
│   │   ├── app/
│   │   └── transport/          # HTTP handlers, gRPC services
│   ├── platform/               # Shared infrastructure code
│   │   ├── config/             # Viper setup
│   │   ├── logging/            # Zap wrapper
│   │   ├── observability/      # OpenTelemetry (tracing, metrics, logger)
│   │   ├── messaging/kafka/    # Producer, Consumer, DLQ, Retry Policy, Codec
│   │   ├── storage/            # PostgreSQL + Redis setup
│   │   ├── idempotency/        # Middleware + store (Redis-backed)
│   │   ├── outbox/             # Outbox pattern (claimer, relay, publisher)
│   │   ├── transport/          # HTTP utils, gRPC helpers
│   │   ├── workflow/           # Step runner với observability
│   │   ├── runtime/            # App bootstrap, graceful shutdown, worker pool
│   │   ├── errors/             # Centralized error types
│   │   ├── integrations/
│   │   │   ├── binance/        # Binance WebSocket client
│   │   │   └── xtb/            # XTB JSON socket client
│   │   └── testing/            # Test helpers
│   └── sharedkernel/           # (Dự phòng)
│
├── api/                        # REST handlers (batch, market, signals)
├── batch/                      # Python: fetch OHLCV, compute indicators (gRPC server)
├── streaming/                  # Go: Kafka consumer → MySQL, gRPC server
├── query/                      # Go: Query aggregator (streaming + batch + Redis)
├── ingestor/                   # Go: Binance/XTB ingestor → Kafka
│
├── protobuf/                   # .proto definitions
│   ├── market-data/
│   ├── data-query/
│   └── batch-data/
│
├── config/
│   └── marketdata.yaml         # Config mẫu cho Market Data Service
│
├── migrations/                 # SQL migrations theo domain
│   ├── marketdata/
│   ├── orders/
│   ├── accounts/
│   └── shared/
│
├── infra/
│   ├── docker/                 # Dockerfiles + data-platform-compose.yml
│   ├── kubernetes/             # K8s deployments, services, ingress
│   ├── migrations/             # DDL cho MySQL (market_ticker, orderbook, kline…)
│   ├── messaging/              # Kafka, NATS configs
│   ├── security/               # RBAC, TLS, secrets
│   └── kafka-topics.sh         # Script tạo Kafka topics
│
├── tests/
│   ├── chaos/                  # Chaos Engineering (network_delay, pod_kill)
│   ├── load/k6/                # Load testing với k6
│   ├── unit/                   # Unit tests (đang phát triển)
│   └── intergration/           # Integration tests (đang phát triển)
│
├── docs/
│   ├── architecture/           # ADR, principles
│   └── runbooks/               # Incident runbooks (DB locks, Kafka lag, WebSocket)
│
├── Makefile                    # Các lệnh build, test, lint, docker
├── buf.yaml / buf.gen.yml      # Buf config để generate proto
├── go.mod / go.sum
└── docker-compose.yml          # Compose đơn giản (dev nhanh)
```

---

## 4. Các Microservices

| Service | Entry Point | Port | Mô tả |
|---|---|---|---|
| **ingestor** | `cmd/ingestor` | — | Kết nối Binance/XTB qua WebSocket, đẩy event vào Kafka |
| **stream_processor** | `cmd/stream_processor` | — | Consumer Kafka, xử lý event, ghi vào MySQL |
| **streaming** | `cmd/streaming` | gRPC :9092 | Cung cấp dữ liệu realtime qua gRPC |
| **batch** | `batch/` (Python) | gRPC :9093 | Fetch OHLCV lịch sử, tính toán indicators |
| **query** | `cmd/query` | gRPC :9091 | Tổng hợp từ streaming + batch + Redis cache |
| **api** | `cmd/api` | REST :8090 / gRPC :9090 | API public cho client |
| **realtime_gateway** | `cmd/realtime_gateway` | WS | WebSocket gateway cho client realtime |
| **api_gateway** | `cmd/api_gateway` | HTTP + gRPC | API Gateway (routing, auth, idempotency) |
| **history_writer** | `cmd/history_writer` | — | Ghi dữ liệu lịch sử từ Kafka vào DB |
| **redis_janitor** | `cmd/redis_janitor` | — | Dọn dẹp key Redis hết hạn |

---

## 5. Clean Architecture (4 Layer)

Các domain service (`internal/marketdata`, `internal/gateway`) được tổ chức theo 4 layer. Quy tắc: **layer ngoài phụ thuộc vào layer trong, không chiều ngược lại**.

```
┌─────────────────────────────────────────────┐
│  Transport  (HTTP handlers, gRPC services)  │  ← Nhận request từ bên ngoài
├─────────────────────────────────────────────┤
│  Adapter    (Repository impl, Kafka pub)    │  ← Giao tiếp DB, Kafka, ext API
├─────────────────────────────────────────────┤
│  Application (Use Cases, DTOs, Ports)       │  ← Điều phối business logic
├─────────────────────────────────────────────┤
│  Domain     (Entities, Value Objects,       │  ← Quy tắc nghiệp vụ thuần túy
│              Aggregates, Domain Events)     │     Không import framework nào
└─────────────────────────────────────────────┘
```

**Các thành phần Domain hiện tại (`internal/marketdata/domain`):**
- **Value Objects:** `Symbol`, `Timeframe`, `Quote`
- **Entities:** `Candle`, `Trade`, `OrderBook`
- **Aggregates:** `Ticker`
- **Domain Events:** `PriceUpdated`, `NewTrade`, `AlertTriggered`

---

## 6. Các Pattern Quan Trọng

### Outbox Pattern
Chống mất event khi service crash. Khi ghi dữ liệu vào DB, một record cũng được ghi đồng thời vào bảng `outbox` trong cùng transaction. Worker riêng (`outbox/relay.go`) đọc bảng outbox và publish lên Kafka.

```
[Write DB] ──┐ (1 transaction)
[Write Outbox] ──┘
              ↓
    [Outbox Relay Worker] → Kafka
```

> Triển khai tại: [`internal/platform/outbox/`](internal/platform/outbox/)

### Idempotency
Mọi API thay đổi trạng thái đều có middleware kiểm tra `Idempotency-Key` trong Redis. Request trùng key sẽ trả lại kết quả cached thay vì xử lý lại.

> Triển khai tại: [`internal/platform/idempotency/`](internal/platform/idempotency/)

### Dead Letter Queue (DLQ)
Các Kafka message không xử lý được sau N lần retry sẽ được đẩy vào topic DLQ riêng để phân tích sau.

> Triển khai tại: [`internal/platform/messaging/kafka/dlq.go`](internal/platform/messaging/kafka/dlq.go)

### Workflow Runner
Chuỗi các bước (steps) xử lý được bọc trong `workflow.Runner` với observability tích hợp (tracing + metrics).

> Triển khai tại: [`internal/platform/workflow/`](internal/platform/workflow/)

---

## 7. Hướng dẫn chạy Local

### Prerequisites
- **Go** ≥ 1.23
- **Docker** + **Docker Compose**
- **Python** ≥ 3.9 (cho batch service)
- **buf** (cho generate protobuf — tuỳ chọn)

### Bước 1: Clone và cài dependencies

```bash
git clone https://github.com/HungphamLeo/BackendDataPlatform.git
cd BackendDataPlatform
go mod tidy
```

### Bước 2: Khởi động infrastructure

```bash
# Dùng compose chính (đầy đủ stack: Kafka, MySQL, Redis, các services)
make docker-compose-up

# Hoặc dùng compose đơn giản (chỉ infra, không build services)
docker compose up -d

# Kiểm tra trạng thái
docker compose ps
```

### Bước 3: Tạo Kafka Topics

```bash
make topics
# Tương đương: bash infra/kafka-topics.sh localhost:9092
```

### Bước 4: Cấu hình

Copy file cấu hình mẫu và chỉnh sửa theo môi trường:

```bash
# Cho từng service (mỗi service có .env.sample riêng)
cp ingestor/.env.sample ingestor/.env
cp streaming/.env.sample streaming/.env
cp batch/.env.sample batch/.env
cp query/.env.sample query/.env
cp api/.env.sample api/.env
```

Hoặc dùng file YAML cho market data service:
```bash
# Chỉnh sửa config/marketdata.yaml
# Điền BINANCE_SPOT_API_KEY, XTB_API_KEY, ... vào biến môi trường
```

### Bước 5: Chạy các services

```bash
# Terminal 1: Ingestor (Binance/XTB → Kafka)
go run ./cmd/ingestor

# Terminal 2: Stream Processor (Kafka → MySQL)
go run ./cmd/stream_processor

# Terminal 3: Streaming gRPC service
go run ./cmd/streaming

# Terminal 4: Batch Python service
cd batch && pip install -r requirements.txt && python batch_grpc_server.py

# Terminal 5: Query service
go run ./cmd/query

# Terminal 6: API public
go run ./cmd/api

# Terminal 7: Realtime WebSocket Gateway
go run ./cmd/realtime_gateway
```

---

## 8. Lệnh Build & Test

```bash
# Build tất cả binary Go
make build-all

# Build từng service
make build-ingestor
make build-streaming
make build-query
make build-api

# Chạy toàn bộ tests
make test
# Tương đương: go test ./...

# Lint
make lint
# Tương đương: golangci-lint run ./...

# Generate protobuf (Go)
make proto-go

# Generate protobuf (Python cho batch)
make proto-py

# Dừng toàn bộ stack
make docker-compose-down
```

### Binary output
Sau `make build-*`, binary được đặt tại `bin/`:
```
bin/ingestor
bin/streaming
bin/query
bin/api
```

---

## 9. Cấu hình

Hệ thống dùng **Viper** để đọc config. Độ ưu tiên: **biến môi trường** > **file YAML**.

File cấu hình chính: [`config/marketdata.yaml`](config/marketdata.yaml)

Các section quan trọng:

| Section | Mô tả |
|---|---|
| `app` | Tên service, version, môi trường |
| `http` | Host, port, timeouts của HTTP server |
| `grpc` | Host, port, stream config của gRPC server |
| `logging` | Level, format (json/console), output |
| `database` | PostgreSQL DSN, pool settings |
| `redis` | Redis host, pool, retry settings |
| `kafka` | Brokers, consumer group, topic mapping |
| `binance` | Spot + Futures: base URL, WS URL, API keys, symbols, rate limits |
| `xtb` | WebSocket URL, credentials, symbols, rate limit |
| `endpoints` | Path, rate limit, cache TTL của từng REST endpoint |

**Biến môi trường bắt buộc khi chạy production:**
```
BINANCE_SPOT_API_KEY
BINANCE_SPOT_SECRET_KEY
BINANCE_FUTURES_API_KEY
BINANCE_FUTURES_SECRET_KEY
XTB_API_KEY
XTB_PASSWORD
```

**Kafka Broker mặc định (local):** `localhost:9092`  
**Kafka Topics:**
```
market-data.price-updated
market-data.new-trade
market-data.alert-triggered
market-data.events
```

---

## 10. Database Migrations

### MySQL (cho streaming/ingestor service)
Các file DDL tại `infra/migrations/` được mount vào MySQL container khi chạy Docker Compose:

| File | Bảng |
|---|---|
| `001_market_ticker.up.sql` | `market_ticker` |
| `002_market_orderbook.up.sql` | `market_orderbook` |
| `003_market_kline.up.sql` | `market_kline` |
| `004_historical_ohlcv.up.sql` | `historical_ohlcv` |
| `005_indicator_result.up.sql` | `indicator_result` |

### PostgreSQL (cho trading services)
Migrations theo domain tại `migrations/`:
```
migrations/
├── marketdata/
├── orders/
├── accounts/
└── shared/
```

---

## Tài liệu liên quan

- [`docs/architecture/`](docs/architecture/) — Quyết định kiến trúc (ADR), nguyên tắc thiết kế
- [`docs/runbooks/`](docs/runbooks/) — Xử lý sự cố: DB locks, Kafka lag, WebSocket
- [`refactoring-summary.md`](refactoring-summary.md) — Tổng kết quá trình refactor Market Data Service
- [`infra/kubernetes/`](infra/kubernetes/) — Kubernetes manifests (deployments, services, ingress)

---

*Module: `github.com/HungphamLeo/BackendDataPlatform` — Go 1.23*
