# Data Platform Plan — Financial Trading

## Top-Level Overview

**Goal:** Design and build a production-grade **Data Platform** layered on top of the existing Financial Trading ecosystem, targeting a data engineer/analyst audience. The platform will:

- Ingest **streaming market data** (real-time) using **Go + Kafka**
- Run **batch data processing** pipelines using **Python**
- Expose a **gRPC-first internal service mesh** so each platform component communicates via typed contracts
- Expose a **REST/gRPC public API** (Go) so external consumers can query processed data directly

The architecture reuses proven patterns already present in this codebase:
- Redis-backed real-time key-value store (from `market-maker-data-provider`)
- Multi-exchange WebSocket adapters (Python, 9 exchanges)
- Kafka topic-driven command/event bus (from `trading-bot-finance-service`)
- gRPC service definitions (from `trading-bot-protobuf`)
- Go service scaffolding (from `trading-bot-api`, `trading-bot-background-tradebot`)

**Scope:** New platform components only — no modification of existing trading-bot or market-maker services.

**Non-goals:**
- Replacing existing order-execution or bot-orchestration services
- Building a front-end UI
- Live trading logic

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│             EXTERNAL CONSUMERS (other teams, dashboards)            │
│              REST API  +  gRPC Public API (Go)                      │
└────────────────────────────┬────────────────────────────────────────┘
                             │
         ┌───────────────────┴──────────────────┐
         │         data-platform-api (Go)        │
         │  Exposes processed datasets, signals  │
         │  REST (Gin) + gRPC server             │
         └───────────────────┬──────────────────┘
                             │ gRPC
         ┌───────────────────┴──────────────────┐
         │      data-platform-query (Go)         │
         │  Reads from MySQL/Redis/S3-like store │
         │  Serves as internal data access layer │
         └───────────────────┬──────────────────┘
                             │ gRPC
    ┌────────────────────────┴─────────────────────────────┐
    │                                                      │
    ▼                                                      ▼
┌──────────────────────────┐          ┌──────────────────────────────┐
│  data-platform-streaming │          │  data-platform-batch (Python)│
│  (Go)                    │          │                              │
│  Kafka consumer(s)        │          │  Scheduled Python pipelines  │
│  Enriches + persists      │          │  Historical OHLCV, signals   │
│  OHLCV, orderbook, trades │          │  Indicator computation       │
│  → MySQL / TimescaleDB   │          │  → MySQL / Parquet files     │
└──────────────┬───────────┘          └──────────────────────────────┘
               │ Kafka topics
┌──────────────┴─────────────────────────────────────┐
│          data-platform-ingestor (Go)               │
│  Reads from Redis (populated by Python WebSocket)  │
│  Publishes to Kafka topics:                        │
│    market.ticker, market.orderbook, market.kline   │
└──────────────┬─────────────────────────────────────┘
               │ Redis read
┌──────────────┴─────────────────────────────────────┐
│       market-maker-data-provider (Python)          │  ← EXISTING
│  9-exchange WebSocket → Redis                      │
│  Keys: {SYMBOL}_{EXCHANGE}_{DATA_TYPE}             │
└────────────────────────────────────────────────────┘
```

---

## Sub-Tasks

---

### Sub-Task 1 — Define gRPC Service Contracts (Proto Definitions)

**Status:** `[ ] pending`

**Intent:**
All internal communication between data platform services will use gRPC. Before writing any service code, define the `.proto` files that act as the typed API contract between components. Following the pattern of `trading-bot-protobuf`.

**Expected Outcomes:**
- A new `data-platform-protobuf/` directory (or module within an existing proto repo)
- Proto files for: MarketData service, BatchData service, DataQuery service
- Generated Go stubs (`*.pb.go`, `*_grpc.pb.go`) ready for import
- `go.mod` pointing to the proto module

**Todo List:**
1. Create `data-platform-protobuf/` directory at workspace root
2. Define `market_data.proto`:
   - Service `MarketDataStream` with RPCs: `GetTicker`, `GetOrderBook`, `GetKline`, `StreamTicker` (server-streaming)
   - Messages: `TickerRequest`, `TickerResponse`, `OrderBookRequest`, `OrderBookResponse`, `KlineRequest`, `KlineResponse`
3. Define `batch_data.proto`:
   - Service `BatchData` with RPCs: `GetOHLCV`, `GetIndicator`, `ListAvailableSymbols`
   - Messages: `OHLCVRequest`, `OHLCVResponse`, `IndicatorRequest`, `IndicatorResponse`
4. Define `data_query.proto`:
   - Service `DataQuery` with RPCs: `QueryMarketSummary`, `QueryHistoricalRange`, `QuerySignals`
   - Messages: `MarketSummaryRequest`, `MarketSummaryResponse`, `HistoricalRangeRequest`, `SignalResponse`
5. Write `Makefile` with `protoc` commands (matching pattern in `trading-bot-protobuf/Makefile`)
6. Run `make` to generate Go stubs
7. Initialize `go.mod` for the new proto module

**Relevant Context:**
- Reference: `trading-bot-protobuf/Makefile` — protoc invocation pattern
- Reference: `trading-bot-protobuf/assistant/assistant.proto` — service + message patterns
- Reference: `trading-bot-protobuf/trading-gateway/trading-gateway.proto` — CRUD-style RPCs
- Reference: `trading-bot-protobuf/go.mod` — module naming convention

---

### Sub-Task 2 — data-platform-ingestor (Go): Redis → Kafka Bridge

**Status:** `[ ] pending`

**Intent:**
The existing `market-maker-data-provider` (Python) already writes real-time ticker, orderbook, and kline data to Redis. The ingestor service bridges that data into Kafka so downstream Go streaming consumers can subscribe to it. This separates the Python socket layer from the Go processing layer cleanly.

**Expected Outcomes:**
- New Go service `data-platform-ingestor/`
- Polls or subscribes to Redis keys matching `{SYMBOL}_{EXCHANGE}_{DATA_TYPE}`
- Publishes normalised JSON events to Kafka topics: `market.ticker`, `market.orderbook`, `market.kline.1m`, `market.kline.5m`, `market.kline.1h`
- Configurable symbol list, exchanges, and poll interval via `.env`
- Structured logging

**Todo List:**
1. Scaffold `data-platform-ingestor/` Go module (`go.mod`, `main.go`, `config/`, `redis/`, `kafka/`, `service/`)
2. Implement Redis client — reuse pattern from `redis/redis.go` in `trading-bot-background-tradebot`
3. Implement Kafka writer — reuse pattern from `kafka/writer.go` in `trading-bot-api`
4. Build ingestor loop:
   - On configurable ticker (default 500ms), scan Redis keys for configured symbols+exchanges
   - Deserialize JSON values into typed Go structs (Ticker, OrderBook, Kline)
   - Publish to appropriate Kafka topic with symbol+exchange as partition key
5. Add deduplication — track last-published value per key, skip if unchanged
6. Add `.env` file with: `REDIS_ADDR`, `KAFKA_BROKERS`, `SYMBOLS`, `EXCHANGES`, `POLL_INTERVAL_MS`
7. Add `Dockerfile` and `deploy.sh` following existing service patterns

**Relevant Context:**
- Reference Redis key pattern: `market-maker-data-provider/config.json` — `{SYMBOL}_{QUOTE}_{EXCHANGE}_{DATA_TYPE}`
- Reference: `trading-bot-background-tradebot/redis/redis.go` — Redis client init
- Reference: `trading-bot-api/kafka/writer.go` — Kafka writer pattern
- Reference: `trading-bot-finance-service/pkg/consumers/router.go` — Kafka consumer router

---

### Sub-Task 3 — data-platform-streaming (Go): Kafka Consumer + Persistence

**Status:** `[ ] pending`

**Intent:**
Consumes Kafka topics published by the ingestor. Enriches, normalises, and persists market data to a time-series-friendly MySQL (or TimescaleDB) schema. This is the real-time write path that keeps the queryable store up to date.

**Expected Outcomes:**
- New Go service `data-platform-streaming/`
- Kafka consumer group for `market.ticker`, `market.orderbook`, `market.kline.*`
- MySQL persistence layer with schema: `market_ticker`, `market_orderbook`, `market_kline`
- gRPC server implementing `MarketDataStream` service (from Sub-Task 1 proto)
- Configurable via `.env`

**Todo List:**
1. Scaffold `data-platform-streaming/` Go module (`go.mod`, `main.go`, `config/`, `kafka/`, `repository/`, `models/`, `server/`)
2. Define MySQL schema models:
   - `MarketTicker` (id, symbol, exchange, price, volume, timestamp)
   - `MarketOrderBook` (id, symbol, exchange, bids_json, asks_json, timestamp)
   - `MarketKline` (id, symbol, exchange, interval, open, high, low, close, volume, timestamp)
3. Implement Kafka consumers — one consumer group per topic, routing pattern from `trading-bot-finance-service/pkg/consumers/router.go`
4. Implement repository layer (GORM) — batch inserts with configurable flush interval
5. Implement gRPC server for `MarketDataStream`:
   - `GetTicker` / `GetOrderBook` / `GetKline` — query from MySQL
   - `StreamTicker` — server-side streaming via Kafka subscription
6. Register gRPC server in `main.go` on configurable port
7. Add `.env`, `Dockerfile`, `deploy.sh`

**Relevant Context:**
- Reference: `trading-bot-finance-service/pkg/consumers/router.go` — Kafka router pattern
- Reference: `trading-bot-background-tradebot/pkg/background-tradebot/repository/` — GORM repository pattern
- Reference: `trading-bot-trading-gateway` — gRPC server implementation pattern
- Reference: Sub-Task 1 proto: `market_data.proto` (MarketDataStream service)

---

### Sub-Task 4 — data-platform-batch (Python): Historical Data + Indicator Pipelines

**Status:** `[ ] pending`

**Intent:**
Python batch jobs that run on schedule to: (a) fetch historical OHLCV from exchanges, (b) compute technical indicators (MA, ATR, etc.), (c) persist results to MySQL for querying. Reuses indicator logic already present in `MarketMaker/indicators/`.

**Expected Outcomes:**
- New Python package `data-platform-batch/`
- CLI-runnable pipeline scripts: `fetch_ohlcv.py`, `compute_indicators.py`, `run_all.py`
- Scheduled via cron or simple loop with configurable interval
- Persists to MySQL tables: `historical_ohlcv`, `indicator_result`
- Configurable via `config.json` or `.env`

**Todo List:**
1. Scaffold `data-platform-batch/` with: `requirements.txt`, `config.py`, `database.py`, `fetch_ohlcv.py`, `compute_indicators.py`, `run_all.py`, `pipelines/`
2. `database.py`: MySQL connection using `mysql-connector-python` or `SQLAlchemy` — reuse pattern from `MarketMaker/database_mm.py`
3. `fetch_ohlcv.py`:
   - Pull historical candlestick data from exchange REST APIs (start with Binance `/api/v3/klines`)
   - Support configurable symbol list, timeframe, and lookback period
   - Persist to `historical_ohlcv` table (symbol, exchange, interval, open, high, low, close, volume, timestamp)
4. `compute_indicators.py`:
   - Load OHLCV from MySQL
   - Compute: SMA, EMA, ATR, RSI, Bollinger Bands — reuse logic from `MarketMaker/indicators/ma.py`, `atr.py`, `indi.py`
   - Persist results to `indicator_result` table (symbol, exchange, interval, indicator_name, value, timestamp)
5. `run_all.py`: orchestrates fetch → compute pipeline with configurable schedule (`schedule` library)
6. Add `deploy.sh` and `README.md` with setup instructions
7. Expose a lightweight gRPC server (via `grpcio`) implementing `BatchData` service from Sub-Task 1 proto — so the query layer can call it

**Relevant Context:**
- Reference: `MarketMaker/indicators/ma.py`, `atr.py`, `indi.py`, `smooth.py` — indicator implementations to reuse
- Reference: `MarketMaker/database_mm.py` — MySQL client pattern
- Reference: `MarketMaker/config.py` — config loading pattern
- Reference: Sub-Task 1 proto: `batch_data.proto` (BatchData service)

---

### Sub-Task 5 — data-platform-query (Go): Internal Data Access Layer (gRPC)

**Status:** `[ ] pending`

**Intent:**
A dedicated Go service that acts as the internal data access layer. It aggregates data from the streaming service (real-time) and the batch service (historical/indicators) via gRPC, providing a single internal query interface for the API layer. This keeps the API layer stateless and thin.

**Expected Outcomes:**
- New Go service `data-platform-query/`
- gRPC server implementing `DataQuery` service (from Sub-Task 1 proto)
- gRPC clients to call `MarketDataStream` (Sub-Task 3) and `BatchData` (Sub-Task 4)
- Caching layer: Redis (DB 10 reserved) for hot query results (TTL-based)
- Configurable via `.env`

**Todo List:**
1. Scaffold `data-platform-query/` Go module (`go.mod`, `main.go`, `config/`, `clients/`, `cache/`, `server/`)
2. Implement gRPC client wrappers:
   - `clients/streaming.go` → connects to `data-platform-streaming` MarketDataStream gRPC
   - `clients/batch.go` → connects to `data-platform-batch` BatchData gRPC (Python grpc server)
3. Implement Redis cache layer (DB 10):
   - Cache key pattern: `query:{request_hash}`
   - TTL: 5s for real-time queries, 60s for batch/historical queries
4. Implement gRPC server for `DataQuery`:
   - `QueryMarketSummary` — cache-first, then streaming client
   - `QueryHistoricalRange` — cache-first, then batch client
   - `QuerySignals` — aggregates indicator results from batch client
5. Register gRPC server in `main.go` on configurable port
6. Add `.env` (streaming address, batch address, Redis addr, ports), `Dockerfile`, `deploy.sh`

**Relevant Context:**
- Reference: `trading-bot-api/` — gRPC client setup pattern (LoadBalancerJobClient, AccountingTransactionClient)
- Reference: `trading-bot-background-tradebot/redis/redis.go` — Redis client
- Reference: Sub-Task 1 proto: `data_query.proto` (DataQuery service)
- Reference: Sub-Task 3: `data-platform-streaming` gRPC address
- Reference: Sub-Task 4: `data-platform-batch` gRPC address

---

### Sub-Task 6 — data-platform-api (Go): Public REST + gRPC API

**Status:** `[ ] pending`

**Intent:**
The external-facing service. Provides a REST API (Gin) and a public gRPC endpoint so external teams or dashboards can query data. Routes requests to the internal `data-platform-query` service. Handles authentication (JWT middleware) following the existing `trading-bot-api` auth pattern.

**Expected Outcomes:**
- New Go service `data-platform-api/`
- REST endpoints (Gin):
  - `GET /v1/market/ticker?symbol=BTC_USDT&exchange=binance`
  - `GET /v1/market/orderbook?symbol=BTC_USDT&exchange=binance`
  - `GET /v1/market/kline?symbol=BTC_USDT&exchange=binance&interval=1m&from=&to=`
  - `GET /v1/batch/ohlcv?symbol=BTC_USDT&exchange=binance&interval=1d&limit=100`
  - `GET /v1/batch/indicators?symbol=BTC_USDT&exchange=binance&indicator=ATR`
  - `GET /v1/signals?symbol=BTC_USDT`
- Public gRPC endpoint (same data, gRPC transport) for high-throughput consumers
- JWT auth middleware
- Swagger/OpenAPI docs generated from route annotations
- Configurable via `.env`

**Todo List:**
1. Scaffold `data-platform-api/` Go module (`go.mod`, `main.go`, `config/`, `middleware/`, `api/rest/`, `api/grpc/`, `clients/`)
2. Implement gRPC client to `data-platform-query` — `clients/query.go`
3. Implement JWT auth middleware — reuse pattern from `trading-bot-api/middlewares/auth.go`
4. Implement REST handlers in `api/rest/`:
   - `market.go` — ticker, orderbook, kline endpoints
   - `batch.go` — OHLCV, indicators endpoints
   - `signals.go` — aggregated signal endpoint
5. Implement public gRPC server in `api/grpc/` — re-exposes `DataQuery` service with auth interceptor
6. Register Gin router and gRPC server in `main.go`; run both on separate ports (REST: 8090, gRPC: 9090)
7. Add Swagger annotations and `swag init` to generate docs
8. Add `.env`, `Dockerfile`, `deploy.sh`, `README.md`

**Relevant Context:**
- Reference: `trading-bot-api/middlewares/auth.go` — JWT middleware
- Reference: `trading-bot-api/api/rest/` — Gin handler patterns, router registration
- Reference: `market-maker-api/` — simpler Gin service as alternative reference
- Reference: Sub-Task 5: `data-platform-query` gRPC address

---

### Sub-Task 7 — Infrastructure & Deployment Configuration

**Status:** `[ ] pending`

**Intent:**
Define the shared infrastructure config — Kafka topics, MySQL schema migrations, Redis DB allocations, Docker Compose for local development — so all data platform services can be started together consistently.

**Expected Outcomes:**
- `data-platform-infra/` directory with:
  - `docker-compose.yml` bringing up: Kafka, Zookeeper, MySQL, Redis, all 4 data platform Go services, Python batch service
  - `migrations/` SQL files for all new tables
  - `kafka-topics.sh` — script to create required Kafka topics
  - `README.md` — setup guide

**Todo List:**
1. Create `data-platform-infra/` directory
2. Write `migrations/001_market_ticker.sql`, `002_market_orderbook.sql`, `003_market_kline.sql`, `004_historical_ohlcv.sql`, `005_indicator_result.sql`
3. Write `kafka-topics.sh` — creates topics: `market.ticker`, `market.orderbook`, `market.kline.1m`, `market.kline.5m`, `market.kline.1h` with configurable partition count and replication factor
4. Write `docker-compose.yml`:
   - Services: `zookeeper`, `kafka`, `mysql`, `redis`
   - Platform services: `ingestor`, `streaming`, `batch`, `query`, `api`
   - Environment variable injection from `.env` file
5. Write `README.md` with: prerequisites, local setup steps, environment variable reference, service port map

**Relevant Context:**
- Reference: `trading-bot-api/.env.sample` — env var naming conventions
- Reference: `trading-bot-finance-service/deploy.sh` — deployment script pattern
- Kafka topic naming: follow existing pattern (dot-separated lowercase, e.g. `market.ticker`)
- Redis DB allocation map (existing: 0, 2, 7, 9) → new platform uses DB 10, 11

---

## Service Port Map

| Service | REST Port | gRPC Port |
|---|---|---|
| data-platform-api | 8090 | 9090 |
| data-platform-query | — | 9091 |
| data-platform-streaming | — | 9092 |
| data-platform-batch (Python) | — | 9093 |
| data-platform-ingestor | — | — (no gRPC, only Kafka out) |

## Redis DB Allocation

| DB | Owner | Purpose |
|---|---|---|
| 0 | trading-bot-api | Session/balance |
| 2 | background-tradebot | Bot job tracking |
| 7 | market-maker-execute | Strategy parameters |
| 9 | Multiple (shared) | Platform state |
| 10 | data-platform-query | Hot query cache |
| 11 | data-platform-ingestor | Dedup tracking |

## Kafka Topic Map

| Topic | Producer | Consumer |
|---|---|---|
| market.ticker | ingestor | streaming |
| market.orderbook | ingestor | streaming |
| market.kline.1m | ingestor | streaming |
| market.kline.5m | ingestor | streaming |
| market.kline.1h | ingestor | streaming |
