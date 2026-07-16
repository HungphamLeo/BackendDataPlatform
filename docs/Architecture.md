# Architecture Guide — BackendDataPlatform

> Tài liệu này dành cho **fresher data engineer / backend engineer** muốn nhanh chóng nắm bắt kiến trúc hệ thống, lý thuyết đằng sau từng quyết định thiết kế, và cách code trong dự án này áp dụng chúng trong thực tế.
>
> Đọc từ trên xuống theo thứ tự. Mỗi phần sẽ giải thích **"tại sao"** trước rồi mới đến **"như thế nào trong code"**.

---

## Mục Lục

1. [Bức tranh toàn cảnh: Luồng ETL](#1-bức-tranh-toàn-cảnh-luồng-etl)
2. [Tech Stack & Lý do chọn](#2-tech-stack--lý-do-chọn)
3. [Clean Architecture — 4 Layer](#3-clean-architecture--4-layer)
4. [Domain Layer — Trái tim của hệ thống](#4-domain-layer--trái-tim-của-hệ-thống)
5. [Application Layer — Điều phối Use Case](#5-application-layer--điều-phối-use-case)
6. [Adapter Layer — Cầu nối hạ tầng](#6-adapter-layer--cầu-nối-hạ-tầng)
7. [Transport Layer — Giao tiếp với thế giới ngoài](#7-transport-layer--giao-tiếp-với-thế-giới-ngoài)
8. [Platform — Shared Infrastructure](#8-platform--shared-infrastructure)
9. [Kafka — Event-Driven Backbone](#9-kafka--event-driven-backbone)
10. [Outbox Pattern — Đảm bảo không mất event](#10-outbox-pattern--đảm-bảo-không-mất-event)
11. [Idempotency — Chống xử lý trùng lặp](#11-idempotency--chống-xử-lý-trùng-lặp)
12. [Dead Letter Queue — Xử lý lỗi có kiểm soát](#12-dead-letter-queue--xử-lý-lỗi-có-kiểm-soát)
13. [Observability — Tracing, Metrics, Logging](#13-observability--tracing-metrics-logging)
14. [Luồng ETL End-to-End](#14-luồng-etl-end-to-end)
15. [Cấu hình hệ thống (Viper)](#15-cấu-hình-hệ-thống-viper)
16. [Bản đồ file quan trọng](#16-bản-đồ-file-quan-trọng)

---

## 1. Bức tranh toàn cảnh: Luồng ETL

Hệ thống này là một **data platform tài chính thời gian thực** với nhiệm vụ chính là thu thập (Extract), chuẩn hóa (Transform), và phục vụ (Load/Serve) dữ liệu thị trường từ các sàn giao dịch.

```
╔══════════════════════════════════════════════════════════════════════╗
║                         EXTRACT                                      ║
║                                                                      ║
║   Binance Spot/Futures          XTB                                  ║
║   WebSocket + REST API          WebSocket JSON-Socket                ║
║         │                            │                               ║
║         └──────────┬─────────────────┘                              ║
║                    ▼                                                 ║
║             ingestor service                                         ║
║         (internal/platform/integrations/)                            ║
╚════════════════════════╤═════════════════════════════════════════════╝
                         │  publish JSON/Protobuf
                         ▼
╔══════════════════════════════════════════════════════════════════════╗
║                      TRANSFORM                                       ║
║                                                                      ║
║              Apache Kafka (broker :9092)                             ║
║   Topics: market-data.price-updated / new-trade / alert-triggered   ║
║                         │                                            ║
║                         ▼                                            ║
║              stream_processor service                                ║
║    (chuẩn hóa dữ liệu → Domain Entities → lưu DB)                   ║
║                         │                                            ║
║          ┌──────────────┴──────────────┐                             ║
║          ▼                             ▼                             ║
║       MySQL 8.0                  batch service (Python)              ║
║  (market_ticker, kline,          fetch OHLCV lịch sử                ║
║   orderbook, ohlcv)              tính technical indicators           ║
╚══════════╤══════════════════════════╤══════════════════════════════╝
           │                          │
╔══════════╪══════════════════════════╪══════════════════════════════╗
║          │           LOAD / SERVE   │                              ║
║          ▼                          ▼                              ║
║   streaming service            batch service                       ║
║   gRPC :9092                   gRPC :9093                          ║
║          │                          │                              ║
║          └──────────┬───────────────┘                             ║
║                     ▼                                              ║
║              query service (Redis cache)                           ║
║              gRPC :9091                                            ║
║                     │                                              ║
║          ┌──────────┴──────────┐                                   ║
║          ▼                     ▼                                   ║
║    api service           realtime_gateway                          ║
║    REST :8090             WebSocket                                ║
║    gRPC :9090                                                      ║
╚══════════════════════════════════════════════════════════════════╝
```

**Tóm gọn:** Dữ liệu đi qua ba giai đoạn rõ ràng — **thu thập từ sàn → xử lý/lưu trữ qua Kafka → phục vụ client qua API/WebSocket**.

---

## 2. Tech Stack & Lý do chọn

### Bảng tổng quan

| Thành phần | Công nghệ | Lý do chọn |
|---|---|---|
| **Ngôn ngữ chính** | Go 1.23 | Concurrency native (goroutines), latency thấp, compile-time safety |
| **Batch / Indicators** | Python 3 | Thư viện tính toán tài chính phong phú (pandas, ta-lib) |
| **Message Broker** | Apache Kafka (Confluent 7.5) | Throughput cao, persistent, replay được, phù hợp stream processing |
| **Database (streaming)** | MySQL 8.0 | Dễ dùng, phổ biến, phù hợp time-series đơn giản |
| **Database (trading)** | PostgreSQL 15 | ACID đầy đủ, transactions phức tạp, JSON support tốt |
| **Cache** | Redis 7 | Sub-millisecond read, support pub/sub, TTL tự động |
| **Internal transport** | gRPC + Protobuf | Type-safe, binary encoding nhanh hơn JSON ~5x, streaming bidirectional |
| **Public transport** | REST + WebSocket | REST cho truy vấn, WebSocket cho realtime push |
| **Config** | Viper | Hỗ trợ YAML + env override, hot reload |
| **Logging** | Uber Zap | Zero-allocation logging, cấu trúc JSON, production-ready |
| **ORM** | GORM | Type-safe query builder, migration support |
| **Observability** | OpenTelemetry | Vendor-neutral, hỗ trợ trace + metric + log đồng thời |
| **Container** | Docker + Docker Compose | Reproducible environment, dễ local dev |
| **Orchestration** | Kubernetes | Auto-scaling, self-healing, declarative deployments |

### Tại sao Kafka thay vì REST call trực tiếp?

Khi Ingestor nhận được tick từ Binance, nó có hai lựa chọn:
1. Gọi trực tiếp vào database → **vấn đề:** Ingestor phụ thuộc vào DB, nếu DB chậm thì Ingestor bị chặn, mất tick.
2. Publish vào Kafka → **ưu điểm:** Ingestor và Stream Processor **decoupled** hoàn toàn, Kafka lưu message bền vững, có thể replay lại nếu consumer crash.

```
Ingestor ─publish→ Kafka ─consume→ StreamProcessor ─write→ MySQL
   ↑                 ↑                    ↑
  (nhanh)         (buffer)           (chậm cũng OK,
                                     Kafka chờ)
```

---

## 3. Clean Architecture — 4 Layer

### Lý thuyết

Clean Architecture (Robert C. Martin) đặt ra nguyên tắc: **business rules không được phụ thuộc vào infrastructure**. Nếu muốn đổi từ PostgreSQL sang MongoDB, chỉ cần thay Adapter Layer, không cần đụng vào Domain hay Application.

```
┌─────────────────────────────────────────────────────────────┐
│  Transport  (HTTP handlers / gRPC / WebSocket)              │
│  Nhận request từ bên ngoài, trả response                    │
├─────────────────────────────────────────────────────────────┤
│  Adapter  (Repository impl / Kafka publisher)               │
│  Implement các interface được định nghĩa ở Application      │
├─────────────────────────────────────────────────────────────┤
│  Application  (Use Cases / DTOs / Ports)                    │
│  Điều phối luồng xử lý, không chứa business rules           │
├─────────────────────────────────────────────────────────────┤
│  Domain  (Entities / Value Objects / Aggregates / Events)   │
│  Chứa business rules thuần túy, KHÔNG import bất kỳ         │
│  framework nào (không có gorm, gin, kafka ở đây)            │
└─────────────────────────────────────────────────────────────┘

Quy tắc phụ thuộc: mũi tên chỉ từ ngoài vào trong
Transport → Application → Domain  ✅
Domain → Transport                 ❌ (KHÔNG được phép)
```

### Trong dự án này

```
internal/
├── marketdata/
│   ├── domain/          ← Layer 1: Business rules
│   ├── application/     ← Layer 2: Use cases
│   ├── adapter/         ← Layer 3: Infra implementations
│   └── transport/       ← Layer 4: HTTP/gRPC handlers
└── gateway/
    ├── app/
    ├── adapter/
    ├── contracts/
    └── transport/
```

---

## 4. Domain Layer — Trái tim của hệ thống

> **Nguyên tắc:** Domain layer không được `import` bất kỳ thư viện nào ngoài thư viện chuẩn của Go. Không có `gorm`, `kafka-go`, `gin` tại đây.

### 4.1 Value Objects — Dữ liệu bất biến, có nghĩa nghiệp vụ

Value Object là loại dữ liệu **không có identity riêng** — hai Symbol có cùng `Base` và `Quote` là hoàn toàn giống nhau. Chúng là bất biến (immutable) sau khi tạo.

**`Symbol`** — chuẩn hóa cặp giao dịch giữa các sàn:

```go
// internal/marketdata/domain/value_object/symbol.go

type Symbol struct {
    Base  string // Đồng cơ sở: BTC, ETH, EUR
    Quote string // Đồng báo giá: USDT, USD
}

// ParseSymbol hiểu nhiều định dạng khác nhau từ các sàn:
// "BTCUSDT" (Binance), "BTC/USDT" (FTX), "BTC-USDT" (generic)
func ParseSymbol(s string) (Symbol, error) { ... }

// ToExchangeFormat chuyển sang định dạng của sàn cụ thể
func (s Symbol) ToExchangeFormat(exchange string) string {
    switch strings.ToUpper(exchange) {
    case "BINANCE":
        return s.Base + s.Quote  // → "BTCUSDT"
    case "FTX":
        return s.Base + "/" + s.Quote  // → "BTC/USDT"
    }
}
```

> **Tại sao cần Symbol thay vì dùng `string`?** Vì `"BTCUSDT"` và `"BTC/USDT"` là cùng một symbol nhưng khác string. Nếu dùng string thẳng, code sẽ bị bug khi so sánh. Value Object đảm bảo tất cả symbol đi qua cùng một chuẩn hóa.

**`Timeframe`** — chuẩn hóa khung thời gian:

```go
// internal/marketdata/domain/value_object/timeframe.go

type Timeframe struct {
    Value int           // Ví dụ: 5
    Unit  TimeframeUnit // Ví dụ: Minute
}

// Binance dùng "1m", FTX dùng "60" (seconds), Kraken dùng "1" (minutes)
func (t Timeframe) ToExchangeFormat(exchange string) string { ... }

// Dùng trong Candle.NewCandle để tính CloseTime
func (t Timeframe) ToDuration() time.Duration {
    return time.Duration(t.ToSeconds()) * time.Second
}
```

**`Quote`** — báo giá bid/ask tại một thời điểm:

```go
// internal/marketdata/domain/value_object/quote.go

type Quote struct {
    Symbol    Symbol
    Bid       float64   // Giá người mua sẵn sàng trả
    Ask       float64   // Giá người bán sẵn sàng bán
    BidSize   float64   // Khối lượng tại giá Bid
    AskSize   float64   // Khối lượng tại giá Ask
    Timestamp time.Time
    Source    string
}

// Business methods tính toán từ Bid/Ask
func (q Quote) MidPrice() float64       { return (q.Bid + q.Ask) / 2 }
func (q Quote) Spread() float64         { return q.Ask - q.Bid }
func (q Quote) SpreadPercentage() float64 { ... }
func (q Quote) IsStale() bool           { return time.Since(q.Timestamp) > 5*time.Minute }
```

### 4.2 Entities — Đối tượng có identity, có thể thay đổi trạng thái

Entity có **ID riêng** và thay đổi trạng thái theo thời gian. Hai Candle có cùng symbol/timeframe nhưng `ID` khác nhau là hai entity khác nhau.

**`Candle`** — nến OHLCV (Open, High, Low, Close, Volume):

```go
// internal/marketdata/domain/entity/candle.go

type Candle struct {
    Symbol      value_object.Symbol
    Timeframe   value_object.Timeframe
    OpenTime    time.Time
    CloseTime   time.Time  // = OpenTime + Timeframe.ToDuration()
    Open        float64
    High        float64
    Low         float64
    Close       float64
    Volume      float64
    QuoteVolume float64
    TradeCount  int64
    Closed      bool    // Trạng thái: đang hình thành hay đã đóng
    Source      string
}

// Update cập nhật nến khi có tick mới đến
func (c *Candle) Update(price, volume, quoteVolume float64) {
    if c.Closed { return } // Bảo vệ: nến đã đóng không cập nhật được
    if price > c.High { c.High = price }
    if price < c.Low  { c.Low  = price }
    c.Close   = price
    c.Volume += volume
    c.TradeCount++
}

// Business methods — logic hiểu nến nằm trong Entity, không rải khắp nơi
func (c *Candle) IsGreen() bool { return c.Close > c.Open }
func (c *Candle) IsDoji()  bool { return abs(c.Close-c.Open) < c.Open*0.001 }
func (c *Candle) Validate() error { ... } // Kiểm tra tính hợp lệ
```

**`Trade`** — một giao dịch khớp lệnh:

```go
// internal/marketdata/domain/entity/trade.go

type Trade struct {
    ID            string
    Symbol        value_object.Symbol
    Price         float64
    Quantity      float64
    QuoteQuantity float64  // = Price * Quantity
    Timestamp     time.Time
    IsBuyerMaker  bool  // true = người mua là maker (passive), false = taker (aggressive)
    Source        string
}

// Trong Binance: nếu IsBuyerMaker=true → giao dịch được tính là "SELL" 
// vì người bán là taker (người chủ động)
func (t *Trade) IsBuy() bool  { return !t.IsBuyerMaker }
func (t *Trade) Side() string {
    if t.IsBuy() { return "BUY" }
    return "SELL"
}
```

### 4.3 Domain Events — Thông báo "điều gì đã xảy ra"

Domain Events là cách domain **thông báo** cho các phần khác biết rằng một điều quan trọng đã xảy ra, mà không cần biết ai đang lắng nghe.

```go
// internal/marketdata/domain/event/domain_event.go

// Interface chung — mọi event đều phải implement
type DomainEvent interface {
    EventType()   string    // "PriceUpdated", "NewTrade", ...
    AggregateID() string    // ID của entity tạo ra event
    OccurredAt()  time.Time // Khi nào xảy ra
    Source()      string    // "BINANCE_SPOT", "XTB", ...
}

// BaseDomainEvent — struct nhúng (embed) để tái sử dụng
type BaseDomainEvent struct {
    EventName  string
    EntityID   string
    Timestamp  time.Time
    SourceName string
}

// Các event cụ thể embed BaseDomainEvent và thêm data riêng

type PriceUpdated struct {
    BaseDomainEvent
    TickerID string
    Symbol   value_object.Symbol
    OldPrice float64
    NewPrice float64
    Quantity float64
}

type NewTrade struct {
    BaseDomainEvent
    Symbol   value_object.Symbol
    TradeID  string
    Price    float64
    Quantity float64
    Side     string  // "BUY" hoặc "SELL"
}

type CandleClosed struct {
    BaseDomainEvent
    Symbol    value_object.Symbol
    Timeframe value_object.Timeframe
    Open, High, Low, Close float64
    Volume     float64
    TradeCount int64
}

type AlertTriggered struct {
    BaseDomainEvent
    Symbol         value_object.Symbol
    AlertType      string   // "PRICE_ABOVE", "PRICE_BELOW", ...
    Condition      string
    CurrentValue   float64
    ThresholdValue float64
}
```

> **Tại sao dùng Domain Events thay vì gọi hàm trực tiếp?**
> Nếu khi giá thay đổi ta gọi thẳng `kafkaProducer.Publish(...)` từ trong Domain, thì Domain sẽ phụ thuộc vào Kafka — vi phạm Clean Architecture. Thay vào đó, Domain chỉ tạo ra `PriceUpdated` event. Application Layer nhận event đó và quyết định publish lên Kafka.

---

## 5. Application Layer — Điều phối Use Case

Application Layer không chứa business rules. Nó chỉ **điều phối**: gọi Repository để lấy data, gọi Domain để xử lý, rồi gọi EventPublisher để publish event.

### Ports — Interface Dependency Inversion

```
internal/marketdata/application/port/
├── in/     ← Input Ports: Use case interfaces (GetTickerUseCase, ...)
└── out/    ← Output Ports: Repository & messaging interfaces
```

**Input Ports** (Use Case interfaces) — Transport Layer sẽ gọi vào đây:

```go
// Ví dụ minh họa cấu trúc (từ refactoring-plan.md)

// Input Ports — Transport gọi Application thông qua interface này
type GetTickerUseCase interface {
    Execute(ctx context.Context, exchange, symbol string) (*dto.TickerDTO, error)
}

type GetCandlesUseCase interface {
    Execute(ctx context.Context, req dto.GetCandlesRequest) ([]*dto.CandleDTO, error)
}

type StreamPriceUseCase interface {
    Execute(ctx context.Context, symbol string, out chan<- *dto.TickerDTO) error
}
```

**Output Ports** — Application gọi Adapter thông qua interface này:

```go
// Output Ports — Application định nghĩa, Adapter implement

type TickerRepositoryPort interface {
    Save(ctx context.Context, ticker *domain.Ticker) error
    FindBySymbol(ctx context.Context, symbol string) (*domain.Ticker, error)
}

type EventPublisherPort interface {
    Publish(ctx context.Context, event domain.DomainEvent) error
}

type MarketDataProviderPort interface {
    GetCurrentPrice(ctx context.Context, symbol string) (float64, error)
    Subscribe(ctx context.Context, symbol string, handler func(tick domain.Tick)) error
}
```

### DTOs — Dữ liệu trao đổi qua layer boundary

DTOs (Data Transfer Objects) là các struct đơn giản dùng để **truyền dữ liệu qua ranh giới layer**, tránh để Domain Entity bị lộ trực tiếp ra ngoài.

```go
// Ví dụ minh họa
type TickerDTO struct {
    Symbol    string    `json:"symbol"`
    Price     float64   `json:"price"`
    Bid       float64   `json:"bid"`
    Ask       float64   `json:"ask"`
    Volume    float64   `json:"volume"`
    Timestamp time.Time `json:"timestamp"`
    Exchange  string    `json:"exchange"`
}

type CandleDTO struct {
    Symbol    string    `json:"symbol"`
    Timeframe string    `json:"timeframe"`
    Open      float64   `json:"open"`
    High      float64   `json:"high"`
    Low       float64   `json:"low"`
    Close     float64   `json:"close"`
    Volume    float64   `json:"volume"`
    OpenTime  time.Time `json:"openTime"`
    CloseTime time.Time `json:"closeTime"`
    Closed    bool      `json:"closed"`
}
```

---

## 6. Adapter Layer — Cầu nối hạ tầng

Adapter Layer **implement** các Output Port interfaces của Application Layer. Đây là nơi duy nhất trong `internal/marketdata/` được phép import `gorm`, `kafka-go`, hay bất kỳ thư viện hạ tầng nào.

```
internal/marketdata/adapter/
├── provider/    ← Gọi ra Binance/XTB WebSocket/REST
├── persistence/ ← Implement Repository bằng GORM (PostgreSQL)
├── messaging/   ← Implement EventPublisher bằng Kafka
├── broker/      ← Kafka broker-specific adapters
└── store/       ← Redis store adapters
```

### Binance WebSocket Client

```go
// internal/platform/integrations/binance/client.go

// Client kết nối WebSocket thuần — nhận message thô từ Binance
type Client struct {
    conn *websocket.Conn
}

func NewClient(url string) (*Client, error) {
    conn, _, err := websocket.DefaultDialer.Dial(url, nil)
    if err != nil { return nil, err }
    return &Client{conn}, nil
}

// Read nhận message liên tục và gọi handler callback
func (c *Client) Read(handler func([]byte)) {
    for {
        _, message, err := c.conn.ReadMessage()
        if err != nil { return }
        handler(message) // Adapter Layer sẽ parse JSON và convert sang Domain Entity
    }
}
```

### XTB Client — Dual Socket (API + Streaming)

XTB sử dụng kiến trúc hai kết nối: một socket cho lệnh API (login, query), một socket riêng cho streaming realtime.

```go
// internal/platform/integrations/xtb/client.go

// Client quản lý cả API socket và Streaming socket
type Client struct {
    cfg    Config
    creds  Credentials
    api    *jsonSocket     // Socket gọi lệnh (login, get data)
    stream *jsonSocket     // Socket nhận stream realtime
    streamSessionId string // Token nhận được sau khi login
    recvCh chan string      // Channel trả về message cho caller
}

// Luồng kết nối XTB:
// 1. Connect API socket
// 2. Login → nhận streamSessionId
// 3. Connect Streaming socket
// 4. Subscribe các symbol

func (c *Client) Connect(ctx context.Context) error {
    // 1) Kết nối socket API
    c.api.Connect(ctx)
    // 2) Login để lấy streamSessionId
    ssid, _ := c.login(ctx)
    c.streamSessionId = ssid
    // 3) Kết nối socket stream
    c.stream.Connect(ctx)
    // 4) Bắt đầu goroutine đọc stream
    go c.readStreamLoop()
    return nil
}

func (c *Client) SubscribePrice(symbol string) error {
    return c.Send(map[string]any{
        "command":         StreamGetTickPrices,
        "symbol":          symbol,
        "streamSessionId": c.ssid(),
    })
}
```

### Kafka Producer (Output Port implementation)

```go
// internal/platform/messaging/kafka/producer.go

// Producer interface — được định nghĩa ở đây để Adapter dùng
type Producer interface {
    Publish(ctx context.Context, topic string, key, value []byte) error
    Close() error
}

type producer struct {
    writer *kafka.Writer
}

func NewProducer(brokers []string) (Producer, error) {
    w := &kafka.Writer{
        Addr:         kafka.TCP(brokers...),
        Balancer:     &kafka.LeastBytes{}, // Cân bằng tải theo partition ít data nhất
        RequiredAcks: kafka.RequireOne,    // Chờ leader ACK (cân bằng durability vs latency)
        Async:        false,               // Sync: chờ xác nhận trước khi return
        BatchTimeout: 10 * time.Millisecond,
    }
    return &producer{writer: w}, nil
}

func (p *producer) Publish(ctx context.Context, topic string, key, value []byte) error {
    return p.writer.WriteMessages(ctx, kafka.Message{
        Topic: topic,
        Key:   key,   // Key dùng để routing vào partition (cùng key → cùng partition → đúng thứ tự)
        Value: value,
    })
}
```

### Outbox Record — Cấu trúc dữ liệu trung chuyển

```go
// internal/platform/outbox/record.go

// TickerRecord là message format publish ra Kafka
type TickerRecord struct {
    Symbol  string  `json:"symbol"`
    AskPr   float64 `json:"askPr"`
    BidPr   float64 `json:"bidPr"`
    BestBid float64 `json:"bestBid"`
    BestAsk float64 `json:"bestAsk"`
    LastPr  float64 `json:"lastPr"`
    Spread  float64 `json:"spread"`
    TS      int64   `json:"ts"` // Unix milliseconds
}

// TickerRecordFromRaw chuyển đổi raw map[string]interface{} sang typed struct
// Xử lý defensive vì data từ WebSocket không luôn có cùng format
func TickerRecordFromRaw(raw interface{}) TickerRecord {
    switch v := raw.(type) {
    case map[string]interface{}:
        out.Symbol  = toString(v["symbol"])
        out.BestBid = toFloat(v["bestBid"])
        out.BestAsk = toFloat(v["bestAsk"])
        out.Spread  = toFloat(v["spreadRaw"])
        // Mid price = (bid + ask) / 2
        out.LastPr  = (out.BestBid + out.BestAsk) / 2
    }
}
```

---

## 7. Transport Layer — Giao tiếp với thế giới ngoài

Transport Layer nhận request từ bên ngoài (HTTP, gRPC, WebSocket) và chuyển thành lệnh cho Application Layer. Nó **không chứa business logic** — chỉ validate input, gọi Use Case, và format response.

```
internal/marketdata/transport/
├── http/   ← REST handlers (gorilla/mux)
└── grpc/   ← gRPC service implementations
```

### HTTP Handler pattern

```go
// Ví dụ minh họa pattern HTTP handler trong dự án

type TickerHandler struct {
    useCase application.GetTickerUseCase // Phụ thuộc vào interface, không phải impl
}

func (h *TickerHandler) GetTicker(w http.ResponseWriter, r *http.Request) {
    // 1. Parse input
    exchange := mux.Vars(r)["exchange"]
    symbol   := mux.Vars(r)["symbol"]

    // 2. Gọi Use Case (không biết gì về DB hay Kafka)
    ticker, err := h.useCase.Execute(r.Context(), exchange, symbol)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // 3. Format response
    json.NewEncoder(w).Encode(ticker)
}
```

### gRPC — Communication giữa services

Services nội bộ giao tiếp qua gRPC thay vì REST vì:
- Binary encoding (Protobuf) nhanh hơn JSON ~5x
- Strongly-typed contracts qua `.proto` files
- Hỗ trợ streaming bidirectional tự nhiên

```protobuf
// protobuf/market-data/market_data.proto (ví dụ cấu trúc)
service MarketDataService {
    rpc GetTicker(GetTickerRequest) returns (TickerResponse);
    rpc StreamTicker(StreamTickerRequest) returns (stream TickerResponse);
}
```

---

## 8. Platform — Shared Infrastructure

`internal/platform/` chứa các thành phần **dùng chung** cho tất cả services. Đây là "stdlib mở rộng" của dự án.

```
internal/platform/
├── config/        ← Viper wrapper
├── logging/       ← Zap logger
├── observability/ ← OpenTelemetry (trace + metrics)
├── messaging/     ← Kafka producer/consumer
│   └── kafka/
├── storage/       ← PostgreSQL + Redis connection pools
├── idempotency/   ← Redis-backed idempotency middleware
├── outbox/        ← Outbox pattern implementation
├── transport/     ← HTTP utils, gRPC helpers
├── workflow/      ← Step runner với observability
├── runtime/       ← App lifecycle (graceful shutdown, worker pool, retry)
├── errors/        ← Error types chuẩn hóa
├── integrations/
│   ├── binance/   ← Binance WebSocket client
│   └── xtb/       ← XTB dual-socket client
└── testing/       ← Test helpers
```

### Config với Viper

```go
// internal/platform/config/config.go

// Config là cấu trúc dữ liệu mapping 1:1 với config/marketdata.yaml
type Config struct {
    App       AppConfig       `mapstructure:"app"`
    HTTP      HTTPConfig      `mapstructure:"http"`
    GRPC      GRPCConfig      `mapstructure:"grpc"`
    Logging   LoggingConfig   `mapstructure:"logging"`
    Database  DatabaseConfig  `mapstructure:"database"`
    Redis     RedisConfig     `mapstructure:"redis"`
    Kafka     KafkaConfig     `mapstructure:"kafka"`
    Binance   BinanceConfig   `mapstructure:"binance"`
    XTB       XTBConfig       `mapstructure:"xtb"`
    Endpoints EndpointsConfig `mapstructure:"endpoints"`
}

func LoadConfig(configPath string) (*Config, error) {
    v := viper.New()
    v.SetConfigFile(configPath)
    v.ReadInConfig()

    // Biến môi trường override config file
    // Ví dụ: MARKETDATA_KAFKA_BOOTSTRAP_SERVERS=kafka:9092
    v.SetEnvPrefix("MARKETDATA")
    v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
    v.AutomaticEnv()

    // Expand ${ENV_VAR} syntax trong YAML
    // Ví dụ: api_key: "${BINANCE_SPOT_API_KEY}"
    for _, key := range v.AllKeys() {
        val := v.GetString(key)
        if strings.HasPrefix(val, "${") {
            envKey := val[2 : len(val)-1]
            if envVal := os.Getenv(envKey); envVal != "" {
                v.Set(key, envVal)
            }
        }
    }

    var config Config
    v.Unmarshal(&config)
    return &config, nil
}
```

### Runtime — Graceful Shutdown

Một service production cần shutdown an toàn: hoàn thành request đang xử lý, flush Kafka messages, đóng DB connections.

```go
// internal/platform/runtime/ (minh họa pattern)

// App lắng nghe OS signal (Ctrl+C, SIGTERM từ Kubernetes)
// và gọi shutdown hooks theo thứ tự ngược lại
func Run(ctx context.Context, services ...Service) error {
    ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
    defer stop()

    // Khởi động tất cả services
    for _, svc := range services {
        go svc.Start(ctx)
    }

    // Chờ tín hiệu shutdown
    <-ctx.Done()

    // Shutdown gracefully với timeout
    shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    for i := len(services) - 1; i >= 0; i-- {
        services[i].Stop(shutdownCtx)
    }
    return nil
}
```

---

## 9. Kafka — Event-Driven Backbone

### Tại sao Kafka là xương sống của hệ thống?

Kafka giải quyết bài toán cốt lõi của data platform: **dữ liệu thị trường đến rất nhanh** (hàng nghìn tick/giây từ Binance), nhưng các consumer (lưu DB, tính indicator, cập nhật cache) có tốc độ xử lý khác nhau.

```
Binance WebSocket → 1000 ticks/giây
                         ↓
              Kafka (buffer bất đồng bộ)
              ┌────────────────────────┐
              │   Partition 0: BTC     │
              │   Partition 1: ETH     │
              │   Partition 2: BNB     │
              └────────────────────────┘
                   ↙        ↓        ↘
          DB Writer    Cache Updater   Indicator Engine
          (100/s)        (500/s)         (50/s)
```

Mỗi consumer hoạt động **độc lập** theo tốc độ của mình. Nếu DB Writer chậm, message vẫn nằm trong Kafka, không mất.

### Kafka Topics trong dự án

```
market-data.price-updated   ← Tick mới từ Binance/XTB
market-data.new-trade        ← Giao dịch khớp lệnh
market-data.alert-triggered  ← Cảnh báo giá vượt ngưỡng
market-data.events           ← Default topic (events không phân loại)
```

### Kafka Message Key — Đảm bảo thứ tự

```go
// Khi publish, dùng Symbol làm key
// → Tất cả message của BTCUSDT đi vào cùng partition
// → Đảm bảo thứ tự xử lý trong cùng symbol

producer.Publish(ctx,
    "market-data.price-updated",
    []byte("BTCUSDT"),       // ← KEY: routing vào đúng partition
    messageBytes,
)
```

### Codec — Serialization

```
internal/platform/messaging/kafka/
├── codec.go           ← Interface Codec
├── codec_protobuf.go  ← Protobuf encoding (faster, binary)
└── headers.go         ← Kafka message headers (metadata)
```

---

## 10. Outbox Pattern — Đảm bảo không mất event

### Vấn đề

Khi nhận được một tick mới từ Binance, service cần làm hai việc:
1. Lưu vào database
2. Publish event lên Kafka

Nếu làm tuần tự: lưu DB thành công → crash → chưa kịp publish → **event bị mất vĩnh viễn**.

### Giải pháp: Outbox Pattern

```
┌─────────────────────────────────────────────────────────────┐
│                     Service Process                         │
│                                                             │
│   BEGIN TRANSACTION                                         │
│     INSERT INTO market_ticker (symbol, price, ...)          │
│     INSERT INTO outbox (event_type, payload, status='NEW')  │
│   COMMIT                                                    │
│                                                             │
│   ↑ Đây là atomic operation — hoặc cả hai thành công,       │
│     hoặc cả hai fail. KHÔNG bao giờ mất dữ liệu.           │
└─────────────────────┬───────────────────────────────────────┘
                      │ Kafka vẫn chưa nhận được message
                      ↓
┌─────────────────────────────────────────────────────────────┐
│                   Outbox Relay Worker                       │
│   (chạy riêng, poll bảng outbox liên tục)                  │
│                                                             │
│   SELECT * FROM outbox WHERE status='NEW' LIMIT 100         │
│   FOR EACH record:                                          │
│     Publish to Kafka                                        │
│     UPDATE outbox SET status='PUBLISHED'                    │
└─────────────────────────────────────────────────────────────┘
```

### Files trong dự án

```
internal/platform/outbox/
├── record.go        ← TickerRecord, CandleRecord — message format
├── claimer.go       ← Claim messages để xử lý (tránh race condition)
├── relay.go         ← Poll outbox → publish to Kafka
├── publisher.go     ← Kafka publishing logic
├── repository.go    ← Interface để đọc/ghi outbox table
└── state_machine.go ← Quản lý trạng thái: NEW → CLAIMED → PUBLISHED → FAILED
```

### State Machine của Outbox Record

```
NEW ──────────────────────────────────────────────────────────────────►
 │                                                                     │
 │ Relay claimer picks up                                              │
 ▼                                                                     │
CLAIMED                                                                │
 │                                                                     │
 │ Publish to Kafka OK         Publish failed (timeout, etc.)          │
 ▼                                   ▼                                 │
PUBLISHED                          FAILED ──── retry count > N ───────►
                                     │
                                     └─── retry count ≤ N → NEW (retry)
```

---

## 11. Idempotency — Chống xử lý trùng lặp

### Vấn đề

Client gửi request đặt lệnh → timeout → client không biết request thành công hay chưa → client gửi lại → **lệnh bị đặt hai lần**.

### Giải pháp

```
Client
  │
  │ POST /api/v1/orders
  │ Idempotency-Key: req-unique-12345
  │
  ▼
Idempotency Middleware
  │
  ├─ Kiểm tra Redis: key "req-unique-12345" tồn tại?
  │   ├─ CÓ → trả về cached response (không xử lý lại)
  │   └─ KHÔNG → tiếp tục xử lý
  │
  ▼
Use Case (PlaceOrder, ...)
  │
  ▼
Response → lưu vào Redis với key "req-unique-12345", TTL 24h
```

### Files trong dự án

```
internal/platform/idempotency/
├── middleware.go  ← HTTP middleware: check & store trong Redis
├── store.go       ← Redis store implementation
├── keys.go        ← Chuẩn hóa format của idempotency key
└── idempotency.go ← Core logic
```

---

## 12. Dead Letter Queue — Xử lý lỗi có kiểm soát

### Vấn đề

Một Kafka consumer nhận message nhưng xử lý thất bại (parse error, DB timeout, logic bug). Nếu không có cơ chế DLQ:
- Retry mãi mãi → block toàn bộ partition
- Bỏ qua → mất dữ liệu

### Giải pháp: Dead Letter Queue

```
Kafka Topic: market-data.price-updated
     │
     ▼
Consumer Group: stream-processor
     │
     ├─ Process OK → Commit offset
     │
     └─ Process FAIL (retry 1, 2, 3 lần)
              │
              └─ Vẫn fail → Publish to DLQ topic
                            market-data.price-updated.DLQ
                                    │
                                    ▼
                            DLQ Processor (manual review)
                            hoặc alert team engineering
```

```
internal/platform/messaging/kafka/
├── dlq.go          ← DLQ publisher logic
├── retry_policy.go ← Exponential backoff retry strategy
└── handler.go      ← Message handler interface với retry wrapper
```

---

## 13. Observability — Tracing, Metrics, Logging

Production systems cần **trả lời được 3 câu hỏi**:
1. Request này mất bao lâu từ đầu đến cuối? **(Tracing)**
2. Hệ thống đang hoạt động như thế nào về hiệu năng? **(Metrics)**
3. Chuyện gì đã xảy ra tại thời điểm cụ thể? **(Logging)**

### Structured Logging với Zap

```go
// internal/platform/logging/ và platform/observability/logger.go

// Zap log theo format JSON, structured fields thay vì string interpolation
logger.Info("tick received",
    zap.String("symbol", "BTCUSDT"),
    zap.Float64("price", 65050.5),
    zap.String("exchange", "BINANCE_SPOT"),
    zap.Duration("latency", time.Since(receivedAt)),
)

// Output JSON (dễ parse bởi Elasticsearch, Loki, CloudWatch):
// {"level":"info","ts":"2025-06-15T10:00:00Z","msg":"tick received",
//  "symbol":"BTCUSDT","price":65050.5,"exchange":"BINANCE_SPOT","latency":"2ms"}
```

> **Tại sao không dùng `fmt.Sprintf("received %s price %f"...)`?**
> String formatting tạo heap allocation, chậm hơn ~10x. Với hàng nghìn tick/giây, điều này ảnh hưởng đáng kể.

### OpenTelemetry

```
internal/platform/observability/
├── obs.go        ← Khởi tạo OTel provider
├── tracing.go    ← Distributed tracing (OpenTelemetry Traces)
├── metrics.go    ← Metrics (latency histogram, counter, gauge)
├── logger.go     ← Logging integration
└── attributes.go ← Common attribute keys (span tags)
```

### Workflow Runner — Observable Pipeline

```go
// internal/platform/workflow/

// Thay vì viết:
// step1(); step2(); step3()

// Dùng workflow runner để tự động trace + measure mỗi step:
workflow.NewRunner(tracer, meter).
    Step("validate_input",    validateInput).
    Step("enrich_data",       enrichWithMarketData).
    Step("save_to_db",        persistToDatabase).
    Step("publish_event",     publishToKafka).
    Run(ctx, input)

// → Mỗi step tự động có:
//   - OpenTelemetry span
//   - Latency metric
//   - Error counter
```

---

## 14. Luồng ETL End-to-End

Dưới đây là luồng hoàn chỉnh từ khi Binance gửi tick đến khi client nhận được dữ liệu.

### E — Extract: Thu thập từ Binance

```
[Binance WebSocket]
  │ JSON message: {"e":"trade","s":"BTCUSDT","p":"65050.50","q":"0.001",...}
  │
  ▼
[internal/platform/integrations/binance/Client.Read(handler)]
  │ handler([]byte) được gọi với raw message
  │
  ▼
[Ingestor Service (cmd/ingestor)]
  │ Parse JSON
  │ Convert sang TickerRecord (internal/platform/outbox/record.go)
  │
  ▼
[Kafka Producer]
  │ Publish to "market-data.price-updated"
  │ Key: "BTCUSDT" (để đảm bảo thứ tự)
```

### T — Transform: Xử lý trong Stream Processor

```
[Kafka Consumer - stream_processor]
  │ Nhận message từ "market-data.price-updated"
  │
  ▼
[Domain Layer - Parse & Validate]
  │ ParseSymbol("BTCUSDT") → Symbol{Base:"BTC", Quote:"USDT"}
  │ NewTrade(id, symbol, price, qty, ts, isBuyerMaker, source)
  │ trade.Validate() → kiểm tra price > 0, qty > 0, ...
  │
  ▼
[Application Layer - Use Case]
  │ GetOrCreateCandle(symbol, timeframe, openTime)
  │ candle.Update(price, volume, quoteVolume)
  │ if candle.CloseTime ≤ now: candle.Close() → emit CandleClosed event
  │
  ▼
[Adapter Layer - Persist]
  │ BEGIN TRANSACTION
  │   INSERT INTO market_ticker (symbol, price, volume, ts)
  │   INSERT INTO market_kline  (symbol, tf, open, high, low, close, ...)
  │ COMMIT
```

### L — Load/Serve: Phục vụ client

```
[Query Service]
  │ Nhận gRPC request từ API Service
  │
  ├─ Check Redis cache (TTL: 5s cho ticker, 30s cho candles)
  │   ├─ Cache HIT → trả về ngay (sub-millisecond)
  │   └─ Cache MISS → query Streaming Service (gRPC)
  │
  ▼
[Streaming Service (gRPC :9092)]
  │ SELECT * FROM market_ticker WHERE symbol=? ORDER BY ts DESC LIMIT 1
  │ Trả về qua gRPC
  │
  ▼
[API Service (REST :8090)]
  │ JSON response cho HTTP client
  │
  ▼
[Realtime Gateway (WebSocket)]
  │ Push realtime khi có update mới (sub từ Kafka)
  │ Thay vì client phải polling REST mỗi giây
```

---

## 15. Cấu hình hệ thống (Viper)

File cấu hình chính: [`config/marketdata.yaml`](../config/marketdata.yaml)

### Nguyên tắc cấu hình

1. **Giá trị mặc định** nằm trong YAML file (commit vào git, không chứa secret)
2. **Secret** (API keys, passwords) nằm trong biến môi trường (KHÔNG commit)
3. **Override theo môi trường** qua `MARKETDATA_` prefix

```yaml
# config/marketdata.yaml
binance:
  spot:
    api_key: "${BINANCE_SPOT_API_KEY}"     # ← expand từ env var
    secret_key: "${BINANCE_SPOT_SECRET_KEY}"
    symbols: ["BTCUSDT", "ETHUSDT", "BNBUSDT"]
    rate_limit:
      requests_per_minute: 1200

kafka:
  bootstrap_servers: "localhost:9092"       # ← override bằng env:
  topics:                                   #   MARKETDATA_KAFKA_BOOTSTRAP_SERVERS=kafka:9092
    price_updated: "market-data.price-updated"
```

### Biến môi trường bắt buộc

```bash
BINANCE_SPOT_API_KEY=xxx
BINANCE_SPOT_SECRET_KEY=xxx
BINANCE_FUTURES_API_KEY=xxx
BINANCE_FUTURES_SECRET_KEY=xxx
XTB_API_KEY=xxx
XTB_PASSWORD=xxx
```

---

## 16. Bản đồ file quan trọng

Khi cần đọc code để hiểu một khái niệm, đây là danh sách file nên đọc theo thứ tự:

### Hiểu Domain Model

| Muốn hiểu | Đọc file |
|---|---|
| Symbol/Timeframe/Quote là gì | `internal/marketdata/domain/value_object/` |
| Candle, Trade, OrderBook, Tick | `internal/marketdata/domain/entity/` |
| Domain Events (PriceUpdated, NewTrade...) | `internal/marketdata/domain/event/domain_event.go` |
| Aggregate Ticker | `internal/marketdata/domain/aggregate/` |

### Hiểu Infrastructure

| Muốn hiểu | Đọc file |
|---|---|
| Kết nối Binance WebSocket | `internal/platform/integrations/binance/client.go` |
| Kết nối XTB (dual socket) | `internal/platform/integrations/xtb/client.go` |
| Kafka Producer | `internal/platform/messaging/kafka/producer.go` |
| Config loading (Viper) | `internal/platform/config/config.go` |
| Message formats (outbox) | `internal/platform/outbox/record.go` |

### Hiểu Patterns

| Pattern | Files liên quan |
|---|---|
| Outbox Pattern | `internal/platform/outbox/*.go` |
| Idempotency | `internal/platform/idempotency/*.go` |
| DLQ / Retry | `internal/platform/messaging/kafka/dlq.go`, `retry_policy.go` |
| Workflow Runner | `internal/platform/workflow/*.go` |
| Graceful Shutdown | `internal/platform/runtime/shutdown.go`, `app.go` |

### Entry Points (khởi động service)

| Service | Entry Point | Mô tả |
|---|---|---|
| Ingestor | `cmd/ingestor/main.go` | Thu thập từ Binance/XTB |
| Stream Processor | `cmd/stream_processor/main.go` | Kafka → MySQL |
| API | `cmd/api/main.go` | REST :8090 |
| Realtime Gateway | `cmd/realtime_gateway/main.go` | WebSocket |
| Query | `cmd/query/main.go` | gRPC aggregator |

---

## Checklist cho Fresher

Đọc xong tài liệu này, bạn nên tự kiểm tra:

- [ ] Giải thích được tại sao Domain Layer không được import `gorm`
- [ ] Vẽ được sơ đồ luồng từ Binance tick → MySQL row
- [ ] Giải thích được Outbox Pattern giải quyết vấn đề gì
- [ ] Biết `Symbol{Base:"BTC", Quote:"USDT"}` khác `string("BTCUSDT")` ở điểm nào
- [ ] Hiểu tại sao dùng Kafka Key = symbol để đảm bảo thứ tự
- [ ] Giải thích được Idempotency Key dùng để làm gì
- [ ] Biết khi nào dùng `TickerRecord` vs `domain.Ticker` vs `TickerDTO`
- [ ] Đọc được log JSON và biết field nào có nghĩa gì
- [ ] Biết service nào chạy ở port nào và giao tiếp với nhau bằng giao thức gì

---

*Module: `github.com/HungphamLeo/BackendDataPlatform` — Go 1.23 + Python 3 + Kafka + MySQL + PostgreSQL + Redis*
