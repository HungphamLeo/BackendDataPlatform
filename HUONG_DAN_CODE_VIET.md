# 📚 Hướng Dẫn Code - Trading Platform Go - Cho Fresher

**Tác Giả**: Giải thích dành cho fresher  
**Ngôn Ngữ**: Tiếng Việt  
**Mục Đích**: Hiểu luồng code từ yêu cầu người dùng đến response từ server

---

## 📖 Mục Lục
1. [Kiến Trúc Tổng Thể](#1-kiến-trúc-tổng-thể)
2. [4 Layer Chính](#2-4-layer-chính-clean-architecture)
3. [Luồng Request HTTP](#3-luồng-request-http-chi-tiết)
4. [Ví Dụ: Feature "Lệnh Mua-Bán"](#4-ví-dụ-feature-lệnh-mua-bán)
5. [Cách Code Tương Tác](#5-cách-các-phần-code-tương-tác)
6. [Giải Thích Từng Thành Phần](#6-giải-thích-từng-thành-phần)

---

## 1. Kiến Trúc Tổng Thể

### 🏗️ Hình Dung Project

```
USER (Client - Điện thoại/Web)
    ↓ (gửi yêu cầu HTTP/gRPC)
┌─────────────────────────────────────────┐
│      API GATEWAY (Cổng chính)           │
│  - Kiểm tra quyền (auth)                │
│  - Giới hạn số request (rate-limit)     │
│  - Chuyển hướng yêu cầu                 │
└────────┬────────────────────────────────┘
         ↓
┌─────────────────────────────────────────┐
│   BUSINESS SERVICES (Dịch vụ chính)     │
│  - Orders Service (Quản lý lệnh)        │
│  - Accounts Service (Quản lý tài khoản) │
│  - Market Data Service (Giá thị trường)  │
└────┬─────────────────┬─────────────┬────┘
     ↓                 ↓             ↓
┌──────────┐   ┌──────────────┐  ┌────────┐
│PostgreSQL│   │Kafka (tin tức)│  │ Redis  │
│(dữ liệu) │   │(sự kiện)      │  │(cache) │
└──────────┘   └──────────────┘  └────────┘
```

### 💡 Các Dịch Vụ Chức Năng

| Dịch Vụ | Tên Trong Code | Chức Năng |
|---------|----------------|----------|
| API Gateway | `cmd/api_gateway` | Cổng vào, kiểm tra quyền, định tuyến |
| Orders | `internal/orders` | Quản lý lệnh mua-bán |
| Accounts | `internal/accounts` | Quản lý tài khoản, số dư tiền |
| Market Data | `internal/marketdata` | Xử lý giá thị trường |
| Realtime Gateway | `cmd/realtime_gateway` | Gửi update theo thời gian thực qua WebSocket |
| Stream Processor | `cmd/stream_processor` | Xử lý luồng dữ liệu từ Kafka |
| History Writer | `cmd/history_writer` | Lưu trữ dữ liệu cũ |

---

## 2. 4 Layer Chính (Clean Architecture)

Project sử dụng mô hình **4 Layer** - đây là mô hình **rất quan trọng** để hiểu code:

```
┌────────────────────────────────────┐
│ 4. TRANSPORT/INTERFACE             │
│    (HTTP Controller, gRPC Handler) │  ← Input từ người dùng
├────────────────────────────────────┤
│ 3. APPLICATION                     │
│    (Use Case, Business Logic)      │  ← Xử lý logic chính
├────────────────────────────────────┤
│ 2. DOMAIN                          │
│    (Entity, Value Object)          │  ← Dữ liệu và quy tắc
├────────────────────────────────────┤
│ 1. ADAPTER/INFRASTRUCTURE          │
│    (Database, Kafka, Redis)        │  ← Giao tiếp với công nghệ
└────────────────────────────────────┘
```

### 🔍 Giải Thích Từng Layer

#### **Layer 1: ADAPTER (Cơ Sở Hạ Tầng)**

```
📁 adapter/
├── 📁 store/postgres/    ← Giao tiếp với Database
│   ├── repo.go           (lưu & lấy dữ liệu từ DB)
│   └── mapper.go         (chuyển database row → object)
├── 📁 broker/kafka/      ← Giao tiếp với Kafka
│   └── publisher.go      (gửi message lên Kafka)
└── 📁 cache/redis/       ← Giao tiếp với Redis
    └── cache.go          (cache để tăng tốc độ)
```

**Chức năng**: Giao tiếp với công nghệ bên ngoài (Database, Kafka, Redis)
**Lưu ý**: Đây là nơi code "bẩn", phụ thuộc vào framework/library

**Ví dụ code adapter:**
```go
// Điều này là ADAPTER - nó biết về database
type PostgresRepository struct {
    db *sql.DB
}

func (r *PostgresRepository) SaveOrder(ctx context.Context, order *domain.Order) error {
    query := "INSERT INTO orders (id, user_id, symbol) VALUES ($1, $2, $3)"
    _, err := r.db.ExecContext(ctx, query, order.ID, order.UserID, order.Symbol)
    return err
}
```

---

#### **Layer 2: DOMAIN (Lõi Kinh Doanh)**

```
📁 domain/
├── account.go        ← Tài khoản (entity - đối tượng chính)
├── order.go          ← Lệnh mua-bán (entity)
├── money.go          ← Tiền (value object)
└── symbol.go         ← Mã cổ phiếu (value object)
```

**Chức năng**: Chứa quy tắc kinh doanh, không biết về database hay HTTP
**Tính Chất**: 
- Không phụ thuộc framework
- Tập trung vào "nó có ý nghĩa gì trong đời sống thực?"

**Ví dụ code domain:**
```go
// Đây là DOMAIN - nó là luật kinh doanh
package domain

type Account struct {
    ID       string
    UserID   string
    Balance  *Money        // Tiền (Value Object)
    Status   string
}

// Quy tắc: Không thể rút tiền hơn số dư
func (a *Account) Debit(amount *Money) error {
    if a.Balance.IsLess(amount) {
        return fmt.Errorf("insufficient balance")
    }
    a.Balance = a.Balance.Subtract(amount)
    return nil
}

// Value Object - giá trị không thay đổi
type Money struct {
    Amount   int64   // 1000 = 10.00 USD
    Currency string  // USD, VND, ...
}

func (m *Money) IsLess(other *Money) bool {
    return m.Amount < other.Amount
}
```

---

#### **Layer 3: APPLICATION (Use Case - Công Việc Cụ Thể)**

```
📁 app/
├── place_order_usecase.go      ← "Đặt lệnh mua" (công việc cụ thể)
├── cancel_order_usecase.go     ← "Hủy lệnh"
├── credit_usecase.go           ← "Nạp tiền vào tài khoản"
└── debit_usecase.go            ← "Rút tiền từ tài khoản"
```

**Chức năng**: 
- Điều phối các bước thực hiện công việc
- Kết hợp domain rules với adapter
- Là nơi "happy path" được định nghĩa

**Ví dụ code use case:**
```go
// Đây là USE CASE - nó điều phối "đặt lệnh mua"
package app

type PlaceOrderUseCase struct {
    orderRepo     OrderRepository      // adapter - lưu order
    accountRepo   AccountRepository    // adapter - lưu account
    eventPublisher EventPublisher      // adapter - gửi event
}

// Đây là "công việc" - đặt lệnh mua
func (uc *PlaceOrderUseCase) Execute(ctx context.Context, cmd *PlaceOrderCommand) (*OrderDTO, error) {
    // Bước 1: Kiểm tra tài khoản có đủ tiền?
    account, err := uc.accountRepo.GetByID(ctx, cmd.UserID)
    if err != nil {
        return nil, fmt.Errorf("account not found: %w", err)
    }
    
    // Bước 2: Tạo order mới (dùng domain logic)
    orderAmount := money.New(cmd.Price * int64(cmd.Quantity), "USD")
    order := domain.NewOrder(cmd.UserID, cmd.Symbol, cmd.Side, orderAmount, cmd.Quantity)
    
    // Bước 3: Giữ số tiền (reserve balance)
    if err := account.Reserve(orderAmount); err != nil {
        return nil, fmt.Errorf("cannot reserve balance: %w", err)
    }
    
    // Bước 4: Lưu vào database (adapter)
    if err := uc.orderRepo.Save(ctx, order); err != nil {
        return nil, fmt.Errorf("failed to save order: %w", err)
    }
    
    if err := uc.accountRepo.Save(ctx, account); err != nil {
        return nil, fmt.Errorf("failed to save account: %w", err)
    }
    
    // Bước 5: Thông báo sự kiện (adapter - Kafka)
    event := domain.OrderPlaced{...}
    if err := uc.eventPublisher.Publish(ctx, event); err != nil {
        return nil, fmt.Errorf("failed to publish event: %w", err)
    }
    
    // Bước 6: Trả kết quả
    return &OrderDTO{...}, nil
}
```

---

#### **Layer 4: TRANSPORT (HTTP/gRPC - Điểm Vào)**

```
📁 transport/
├── http/
│   ├── handler.go           ← "Làm gì khi nhân yêu cầu HTTP?"
│   └── middleware/
│       ├── auth.go          ← Kiểm tra đăng nhập
│       └── logging.go       ← Ghi lại log
└── grpc/
    ├── handler.go           ← Xử lý yêu cầu gRPC
    └── interceptors.go      ← Kiểm tra, ghi log, ...
```

**Chức năng**: 
- Nhận yêu cầu từ HTTP/gRPC
- Chuyển đổi dữ liệu (DTO ↔ Domain Object)
- Gọi use case
- Trả lại response

**Ví dụ code transport:**
```go
// Đây là TRANSPORT - nó biết về HTTP
package transport

type PlaceOrderHandler struct {
    usecase *app.PlaceOrderUseCase
}

// Hàm này được gọi khi API nhận POST /orders
func (h *PlaceOrderHandler) PlaceOrder(w http.ResponseWriter, r *http.Request) {
    // Bước 1: Lấy dữ liệu từ HTTP request
    var req PlaceOrderRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }
    
    // Bước 2: Chuyển đổi HTTP request → Use case command
    cmd := &app.PlaceOrderCommand{
        UserID:   req.UserID,
        Symbol:   req.Symbol,
        Side:     req.Side,
        Price:    req.Price,
        Quantity: req.Quantity,
    }
    
    // Bước 3: Gọi use case
    result, err := h.usecase.Execute(r.Context(), cmd)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    // Bước 4: Chuyển đổi kết quả → HTTP response
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(result)
}
```

---

## 3. Luồng Request HTTP - Chi Tiết

Hãy tưởng tượng bạn đang dùng app mua cổ phiếu, bạn nhấn "Mua Apple":

### 📱 Bước 1: Client gửi HTTP Request

```
POST https://api.tradingplatform.com/orders
Content-Type: application/json

{
  "symbol": "AAPL",
  "side": "BUY",
  "price": 18000,
  "quantity": 10
}
```

### 🚀 Bước 2: API Gateway nhận request

**File**: `cmd/api_gateway/main.go`
```go
// API Gateway nhận request
// - Kiểm tra: User có đăng nhập không?
// - Kiểm tra: User có quyền đặt lệnh không?
// - Kiểm tra: Có đang bị tấn công không (rate-limit)?
// Nếu OK → chuyển tới Orders Service
```

**Code middleware:**
```go
type AuthMiddleware struct {
    jwtValidator JWTValidator  // Kiểm tra token
}

func (m *AuthMiddleware) Handle(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Lấy token từ header
        token := r.Header.Get("Authorization")
        
        // Kiểm tra token có hợp lệ?
        user, err := m.jwtValidator.Validate(token)
        if err != nil {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        
        // Nếu OK, cho qua
        next.ServeHTTP(w, r)
    })
}
```

### 📨 Bước 3: Orders Transport Layer (handler HTTP)

**File**: `internal/orders/transport/http_handler.go` (chưa làm)
```go
// Handler này như "người tiếp nhân" tại quầy
func (h *PlaceOrderHandler) PlaceOrder(w http.ResponseWriter, r *http.Request) {
    // 1. Lấy dữ liệu từ request
    var req PlaceOrderRequest
    json.NewDecoder(r.Body).Decode(&req)
    
    // 2. Chuyển dữ liệu → command format
    cmd := &app.PlaceOrderCommand{...}
    
    // 3. Gọi use case
    result, err := h.usecase.Execute(r.Context(), cmd)
    
    // 4. Trả response
    json.NewEncoder(w).Encode(result)
}
```

### 🏢 Bước 4: Orders Application Layer (use case)

**File**: `internal/orders/app/place_order_usecase.go` (chưa làm)
```go
func (uc *PlaceOrderUseCase) Execute(ctx context.Context, cmd *PlaceOrderCommand) (*OrderDTO, error) {
    // 1️⃣ Kiểm tra: Tài khoản có đủ tiền?
    account, _ := uc.accountRepo.GetByID(ctx, cmd.UserID)
    if account.Balance < cmd.Price * cmd.Quantity {
        return nil, errors.New("Insufficient balance")
    }
    
    // 2️⃣ Tạo order mới (domain logic)
    order := domain.NewOrder(cmd.UserID, cmd.Symbol, cmd.Side, cmd.Price, cmd.Quantity)
    
    // 3️⃣ Lưu vào database (adapter)
    _ = uc.orderRepo.Save(ctx, order)
    
    // 4️⃣ Thông báo sự kiện (adapter - Kafka)
    _ = uc.eventPublisher.Publish(ctx, "orders.placed", order)
    
    return &OrderDTO{...}, nil
}
```

### 💾 Bước 5: Adapter Layer (database)

**File**: `internal/orders/adapter/store/postgres/repo.go` (chưa làm)
```go
func (r *PostgresRepository) Save(ctx context.Context, order *domain.Order) error {
    query := """
        INSERT INTO orders (id, user_id, symbol, side, price, quantity, status)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    """
    
    // Gói dữ liệu vào SQL query
    err := r.db.ExecContext(ctx, 
        query,
        order.ID, 
        order.UserID, 
        order.Symbol,
        order.Side,
        order.Price,
        order.Quantity,
        "PENDING",  // status mặc định
    )
    
    return err
}
```

### 📊 Bước 6: Database lưu dữ liệu

```
PostgreSQL
├── 📋 orders table
│   ├── id: "ORDER-001"
│   ├── user_id: "USER-123"
│   ├── symbol: "AAPL"
│   ├── side: "BUY"
│   ├── price: 18000
│   ├── quantity: 10
│   └── status: "PENDING"
└── 📋 outbox table (sự kiện chưa gửi)
    └── {topic: "orders.placed", data: {...}}
```

### 📢 Bước 7: Kafka Relay (gửi event)

**File**: `internal/platform/outbox/relay.go` (chưa làm)

Có một service chạy lặp lại, mỗi giây kiểm tra:
```go
func (r *OutboxRelay) Run(ctx context.Context) {
    ticker := time.NewTicker(time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        // Lấy tất cả message chưa gửi
        messages, _ := r.db.GetUnpublished(ctx)
        
        for _, msg := range messages {
            // Gửi lên Kafka
            _ = r.kafkaProducer.SendMessage(ctx, msg.Topic, msg.Payload)
            
            // Đánh dấu là đã gửi
            _ = r.db.MarkPublished(ctx, msg.ID)
        }
    }
}
```

### 📡 Bước 8: Kafka (Message Broker)

```
Kafka Topic: "orders.placed"
│
├─ Message 1: {OrderID: "ORDER-001", UserID: "USER-123", ...}
├─ Message 2: {...}
└─ Message 3: {...}
```

### 🔄 Bước 9: Stream Processor xử lý event

**File**: `cmd/stream_processor/main.go`
```go
// Stream Processor nghe events từ Kafka
func (p *Processor) Consumers(ctx context.Context) {
    // Lắng nghe topic "orders.placed"
    messages, _ := p.kafkaConsumer.Consume(ctx, "orders.placed")
    
    for msg := range messages {
        // Ví dụ: nếu order được placed, tính toán thống kê
        _ = p.aggregate(ctx, msg)
        
        // Commit offset (đánh dấu đã xử lý)
        _ = p.kafkaConsumer.CommitMessage(msg)
    }
}
```

### 📻 Bước 10: Realtime Gateway gửi áp dụng cho user

**File**: `cmd/realtime_gateway/ws_routes.go`
```go
// Realtime Gateway gửi update real-time cho các user đang xem
func (g *RealtimeGateway) broadcast(ctx context.Context, event Event) {
    // Tìm tất cả WebSocket connection
    for _, conn := range g.connections {
        // Gửi event cho từng user
        conn.SendJSON(event)
    }
}
```

### 📱 Bước 11: Response trả về client

```json
{
  "order_id": "ORDER-001",
  "symbol": "AAPL",
  "side": "BUY",
  "price": 18000,
  "quantity": 10,
  "status": "PENDING",
  "created_at": "2026-02-07T10:30:00Z"
}
```

---

## 4. Ví Dụ: Feature "Lệnh Mua-Bán"

Hãy cùng nhau xây dựng feature "Đặt lệnh mua" từ đầu, từng layer một.

### 📋 Bước 1: Viết Protobuf (Contract)

**File**: `api/proto/orders/v1/orders.proto`(chưa làm)

```protobuf
syntax = "proto3";

package orders.v1;

// Request: Client gửi thông tin lệnh
message PlaceOrderRequest {
    string user_id = 1;
    string symbol = 2;
    string side = 3;      // "BUY" hoặc "SELL"
    int64 price = 4;      // Giá (tính bằng cent, ví dụ 18000 = 180.00)
    double quantity = 5;  // Số lượng
}

// Response: Server trả về thông tin lệnh vừa tạo
message PlaceOrderResponse {
    string order_id = 1;
    string status = 2;
    int64 created_at = 3;  // Thời gian Unix timestamp
}

// Event: Sự kiện được gửi qua Kafka
message OrderPlacedEvent {
    string order_id = 1;
    string user_id = 2;
    string symbol = 3;
}
```

**Tại sao dùng Protobuf?**
- Nó định nghĩa "hợp đồng" giữa services
- Kích thước nhỏ (tốt cho mạng)
- Type-safe (không bị lỗi kiểu dữ liệu)
- Có thể sinh code Go tự động

---

### 🏛️ Bước 2: Viết Domain Model

**File**: `internal/orders/domain/order.go`

```go
package domain

import (
    "fmt"
    "time"
    "github.com/google/uuid"
)

// Order là entity chính - đối tượng quan trọng trong hệ thống
type Order struct {
    ID        string      // Unique ID (ORDER-xxx)
    UserID    string      // Ai đặt lệnh?
    Symbol    string      // Cái gì (AAPL, GOOGL)?
    Side      string      // Mua hay bán? (BUY/SELL)
    Price     int64       // Giá bao nhiêu? (tính theo cent)
    Quantity  float64     // Bao nhiêu cái?
    Status    string      // Trạng thái hiện tại (PENDING, FILLED, CANCELLED)
    CreatedAt time.Time   // Lúc nào tạo?
    Version   int         // Để kiểm tra conflict (optimistic locking)
}

// Factory function - hàm tạo order mới (phải tuân theo quy tắc)
func NewOrder(
    userID string,
    symbol string,
    side string,
    price int64,
    quantity float64,
) *Order {
    // Kiểm tra dữ liệu hợp lệ
    if price <= 0 {
        panic("Price must be positive")
    }
    if quantity <= 0 {
        panic("Quantity must be positive")
    }
    if side != "BUY" && side != "SELL" {
        panic("Side must be BUY or SELL")
    }
    
    return &Order{
        ID:        "ORDER-" + uuid.New().String()[:8],
        UserID:    userID,
        Symbol:    symbol,
        Side:      side,
        Price:     price,
        Quantity:  quantity,
        Status:    "PENDING",  // Trạng thái mặc định
        CreatedAt: time.Now(),
        Version:   1,
    }
}

// Domain logic: Check xem order có hợp lệ không?
func (o *Order) Validate() error {
    if o.UserID == "" {
        return fmt.Errorf("user_id is required")
    }
    if o.Symbol == "" {
        return fmt.Errorf("symbol is required")
    }
    if o.Price <= 0 {
        return fmt.Errorf("price must be positive")
    }
    if o.Quantity <= 0 {
        return fmt.Errorf("quantity must be positive")
    }
    return nil
}

// Domain logic: Bạn có thể cancel lệnh chứ?
func (o *Order) CanCancel() bool {
    return o.Status == "PENDING"  // Chỉ lệnh chưa fill mới cancel được
}

// Domain logic: Cancel một lệnh
func (o *Order) Cancel() error {
    if !o.CanCancel() {
        return fmt.Errorf("cannot cancel order in %s status", o.Status)
    }
    o.Status = "CANCELLED"
    o.Version++  // Tăng version khi change
    return nil
}

// Domain logic: Fill một lệnh (thực thi)
func (o *Order) Fill() error {
    if o.Status != "PENDING" {
        return fmt.Errorf("cannot fill order that is not pending")
    }
    o.Status = "FILLED"
    o.Version++
    return nil
}
```

**Lưu ý quan trọng:**
- `Order` là **Entity** (có identity duy nhất)
- Chứa **business logic** (quy tắc)
- Không biết về database hay HTTP
- Không import bất cứ framework nào

---

### 🧪 Bước 3: Tạo Use Case

**File**: `internal/orders/app/place_order_usecase.go`

```go
package app

import (
    "context"
    "fmt"
    "github.com/BackendDataPlatform/internal/orders/domain"
    "github.com/BackendDataPlatform/internal/platform/idempotency"
    "github.com/BackendDataPlatform/internal/platform/outbox"
)

// Command: Input từ client (tới từ HTTP handler)
type PlaceOrderCommand struct {
    UserID   string
    Symbol   string
    Side     string
    Price    int64
    Quantity float64
}

// DTO (Data Transfer Object): Output được trả về client
type PlaceOrderResult struct {
    OrderID   string
    Symbol    string
    Status    string
    CreatedAt string
}

// Use Case là người "điều phối" - nó biết cách làm việc, nhưng không làm trực tiếp
type PlaceOrderUseCase struct {
    orderRepo      domain.OrderRepository        // Adapter: lưu order
    accountRepo    domain.AccountRepository      // Adapter: lấy account
    eventPublisher domain.EventPublisher         // Adapter: gửi event
    idempotency    idempotency.IdempotencyStore // Adapter: chống duplicate
}

// Constructor
func NewPlaceOrderUseCase(
    orderRepo domain.OrderRepository,
    accountRepo domain.AccountRepository,
    eventPublisher domain.EventPublisher,
    idempotency idempotency.IdempotencyStore,
) *PlaceOrderUseCase {
    return &PlaceOrderUseCase{
        orderRepo:      orderRepo,
        accountRepo:    accountRepo,
        eventPublisher: eventPublisher,
        idempotency:    idempotency,
    }
}

// Execute: Phương thức chính - đặt lệnh
func (uc *PlaceOrderUseCase) Execute(ctx context.Context, cmd *PlaceOrderCommand, idempotencyKey string) (*PlaceOrderResult, error) {
    // 🔒 BƯỚC 1: Kiểm tra idempotent (tránh đặt 2 lần cùng lệnh)
    // Nếu client gửi 2 request giống nhau, chỉ tạo 1 lệnh thôi
    cachedResult, err := uc.idempotency.Get(ctx, idempotencyKey)
    if err == nil && cachedResult != nil {
        // Đã tạo rồi, trả kết quả cũ
        return cachedResult.(*PlaceOrderResult), nil
    }
    
    // ✅ BƯỚC 2: Validate input
    if cmd.UserID == "" || cmd.Symbol == "" || cmd.Price <= 0 {
        return nil, fmt.Errorf("invalid input")
    }
    
    // 🏦 BƯỚC 3: Lấy tài khoản từ database (adapter)
    account, err := uc.accountRepo.GetByID(ctx, cmd.UserID)
    if err != nil {
        return nil, fmt.Errorf("account not found: %w", err)
    }
    
    // 💰 BƯỚC 4: Kiểm tra: Có đủ tiền không?
    requiredBalance := cmd.Price * int64(cmd.Quantity)
    if account.AvailableBalance < requiredBalance {
        return nil, fmt.Errorf("insufficient balance: need %d, have %d", requiredBalance, account.AvailableBalance)
    }
    
    // 📝 BƯỚC 5: Tạo order mới (domain logic)
    order := domain.NewOrder(cmd.UserID, cmd.Symbol, cmd.Side, cmd.Price, cmd.Quantity)
    
    // 🔐 BƯỚC 6: Validate order (có hợp lệ không?)
    if err := order.Validate(); err != nil {
        return nil, fmt.Errorf("invalid order: %w", err)
    }
    
    // 💾 BƯỚC 7: Lưu order vào database (adapter)
    if err := uc.orderRepo.Save(ctx, order); err != nil {
        return nil, fmt.Errorf("failed to save order: %w", err)
    }
    
    // 💼 BƯỚC 8: Reserve balance (giữ tiền, chưa trừ hết)
    account.AvailableBalance -= requiredBalance
    if err := uc.accountRepo.Update(ctx, account); err != nil {
        return nil, fmt.Errorf("failed to reserve balance: %w", err)
    }
    
    // 📢 BƯỚC 9: Tạo event (sự kiện) để gửi tới dịch vụ khác
    // Sự kiện này sẽ được ghi vào Kafka
    event := domain.OrderPlacedEvent{
        OrderID:   order.ID,
        UserID:    order.UserID,
        Symbol:    order.Symbol,
        CreatedAt: order.CreatedAt,
    }
    
    if err := uc.eventPublisher.Publish(ctx, "orders.placed", event); err != nil {
        // Log error, nhưng vẫn trả lại success (event sẽ được retry)
        fmt.Printf("Warning: failed to publish event: %v\n", err)
    }
    
    // 📤 BƯỚC 10: Tạo result để trả về client
    result := &PlaceOrderResult{
        OrderID:   order.ID,
        Symbol:    order.Symbol,
        Status:    order.Status,
        CreatedAt: order.CreatedAt.Format(time.RFC3339),
    }
    
    // 🔒 BƯỚC 11: Lưu result vào cache (idempotency) - nếu duplicate lần tới, dùng cái này
    _ = uc.idempotency.Set(ctx, idempotencyKey, result, 24*time.Hour)
    
    return result, nil
}
```

**Chú ý:**
- Use Case **điều phối** (orchestrates)
- Nó gọi domain logic, adapter, rồi trả kết quả
- Nó không biết về HTTP hay database trực tiếp
- Nó biết về **interface** (OrderRepository, EventPublisher) chứ không biết implement

---

### 🌐 Bước 4: Viết Transport (HTTP Handler)

**File**: `internal/orders/transport/http_handler.go`

```go
package transport

import (
    "encoding/json"
    "net/http"
    "github.com/BackendDataPlatform/internal/orders/app"
)

// HTTP Request từ client
type PlaceOrderRequest struct {
    Symbol   string  `json:"symbol"`
    Side     string  `json:"side"`
    Price    int64   `json:"price"`
    Quantity float64 `json:"quantity"`
}

// HTTP Handler: Người tiếp nhân
type PlaceOrderHandler struct {
    usecase *app.PlaceOrderUseCase
}

// Constructor
func NewPlaceOrderHandler(usecase *app.PlaceOrderUseCase) *PlaceOrderHandler {
    return &PlaceOrderHandler{usecase: usecase}
}

// ServeHTTP: Được gọi khi nhận POST /orders
func (h *PlaceOrderHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // Bước 1: Lấy user ID từ JWT token (middleware đã add vào context)
    userID := r.Context().Value("user_id").(string)
    
    // Bước 2: Lấy idempotency key từ header
    idempotencyKey := r.Header.Get("Idempotency-Key")
    if idempotencyKey == "" {
        http.Error(w, "Idempotency-Key header required", http.StatusBadRequest)
        return
    }
    
    // Bước 3: Parse JSON body
    var req PlaceOrderRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }
    
    // Bước 4: Chuyển đổi HTTP request → Use case command
    cmd := &app.PlaceOrderCommand{
        UserID:   userID,
        Symbol:   req.Symbol,
        Side:     req.Side,
        Price:    req.Price,
        Quantity: req.Quantity,
    }
    
    // Bước 5: Gọi use case
    result, err := h.usecase.Execute(r.Context(), cmd, idempotencyKey)
    if err != nil {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
        return
    }
    
    // Bước 6: Trả response thành công
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(result)
}
```

---

### 💾 Bước 5: Viết Adapter (Database)

**File**: `internal/orders/adapter/store/postgres/repo.go`

```go
package postgres

import (
    "context"
    "database/sql"
    "fmt"
    "github.com/BackendDataPlatform/internal/orders/domain"
)

// PostgreSQL Repository: Giao tiếp với database
type OrderRepository struct {
    db *sql.DB
}

// Constructor
func NewOrderRepository(db *sql.DB) *OrderRepository {
    return &OrderRepository{db: db}
}

// Save: Lưu order vào database
func (r *OrderRepository) Save(ctx context.Context, order *domain.Order) error {
    query := `
        INSERT INTO orders (id, user_id, symbol, side, price, quantity, status, created_at, version)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
        ON CONFLICT (id) DO NOTHING
    `
    
    _, err := r.db.ExecContext(
        ctx,
        query,
        order.ID,
        order.UserID,
        order.Symbol,
        order.Side,
        order.Price,
        order.Quantity,
        order.Status,
        order.CreatedAt,
        order.Version,
    )
    
    if err != nil {
        return fmt.Errorf("failed to save order: %w", err)
    }
    
    return nil
}

// GetByID: Lấy order từ database bằng ID
func (r *OrderRepository) GetByID(ctx context.Context, orderID string) (*domain.Order, error) {
    query := `
        SELECT id, user_id, symbol, side, price, quantity, status, created_at, version
        FROM orders
        WHERE id = $1
    `
    
    var order domain.Order
    err := r.db.QueryRowContext(ctx, query, orderID).Scan(
        &order.ID,
        &order.UserID,
        &order.Symbol,
        &order.Side,
        &order.Price,
        &order.Quantity,
        &order.Status,
        &order.CreatedAt,
        &order.Version,
    )
    
    if err == sql.ErrNoRows {
        return nil, fmt.Errorf("order not found")
    }
    if err != nil {
        return nil, fmt.Errorf("failed to get order: %w", err)
    }
    
    return &order, nil
}
```

---

## 5. Cách Các Phần Code Tương Tác

### 🔗 Dependency Injection với Wire

Project dùng **Wire** - framework để "nối" các thành phần lại.

**File**: `cmd/api_gateway/wire.go`

```go
package main

import (
    "github.com/google/wire"
    "github.com/BackendDataPlatform/internal/orders/app"
    "github.com/BackendDataPlatform/internal/orders/transport"
    "github.com/BackendDataPlatform/internal/orders/adapter/store/postgres"
    "github.com/BackendDataPlatform/internal/platform/storage/postgres"
)

// ProviderSet: Nói Wire "làm cách nào để tạo các thành phần"
var ProviderSet = wire.NewSet(
    // 1️⃣ Tạo database connection
    postgres.NewDB,
    
    // 2️⃣ Tạo adapter (repository)
    postgres.NewOrderRepository,
    
    // 3️⃣ Tạo use case (gọi adapter)
    app.NewPlaceOrderUseCase,
    
    // 4️⃣ Tạo HTTP handler (gọi use case)
    transport.NewPlaceOrderHandler,
)

// InitializeApp: Wire tự động tạo toàn bộ object theo dependency
func InitializeApp() *App {
    wire.Build(ProviderSet)
    return &App{}
}
```

**Cách hoạt động:**
```
wire.Build(ProviderSet)
    ↓
Wire thấy: "Cần PlaceOrderHandler"
    ↓ (đọc: NewPlaceOrderHandler cần PlaceOrderUseCase)
Wire thấy: "Cần PlaceOrderUseCase"
    ↓ (đọc: NewPlaceOrderUseCase cần OrderRepository)
Wire thấy: "Cần OrderRepository"
    ↓ (đọc: NewOrderRepository cần *sql.DB)
Wire thấy: "Cần *sql.DB"
    ↓ (đọc: NewDB)
Wire tạo DB connection
Wire tạo OrderRepository
Wire tạo PlaceOrderUseCase
Wire tạo PlaceOrderHandler
```

---

## 6. Giải Thích Từng Thành Phần

### 🔐 1. Outbox Pattern (Chống Mất Data)

**Vấn đề:** Nếu app crash sau khi lưu order nhưng trước khi gửi Kafka, sự kiện sẽ mất!

**Giải pháp:**
```
┌──────────────────┐
│ Save Order       │ ─┐
├──────────────────┤  │ Một transaction (cùng lúc, tất hoặc không)
│ Save Outbox      │ ─┤ (nếu một cái fail, cả 2 coi như fail)
│ (sự kiện)        │ ─┘
└──────────────────┘
         ↓
   (Relay picks up)
         ↓
┌──────────────────┐
│ Send to Kafka    │
└──────────────────┘
         ↓
┌──────────────────┐
│ Mark Sent        │
└──────────────────┘
```

**Code:**
```go
// Lưu order + outbox trong 1 transaction
tx := db.BeginTx(ctx)

// Lưu order
tx.Exec("INSERT INTO orders...")

// Lưu outbox message
tx.Exec("INSERT INTO outbox (topic, payload) VALUES ('orders.placed', ...)")

// Commit cả 2
tx.Commit()
```

---

### 🔄 2. Idempotency (Chống Trùng)

**Vấn đề:** Nếu client gửi request 2 lần (vì timeout), server sẽ tạo 2 orders!

**Giải pháp:**
```
Request 1: POST /orders (Idempotency-Key: "KEY-123")
  ↓ Lưu vào cache: KEY-123 → OrderID-001
  ↓ Tạo order ORDER-001
  ↓ Trả response

Request 2: POST /orders (Idempotency-Key: "KEY-123") ← cùng key
  ↓ Tìm trong cache: KEY-123 → OrderID-001 (đã có!)
  ↓ Trả lại kết quả cũ (không tạo mới)
```

**Code:**
```go
// Check cache trước
if result, found := cache.Get(idempotencyKey); found {
    return result  // Trả kết quả cũ
}

// Không có, tạo mới
result := createOrder(...)

// Lưu vào cache
cache.Set(idempotencyKey, result, 24*time.Hour)
```

---

### 📢 3. Workflow (Saga) - Giao Dịch Phức Tạp

**Vấn đề:** Đặt lệnh cần:
1. Check account có tiền
2. Save order
3. Reserve balance
4. Gửi event

Nếu bước 3 fail, cần rollback bước 2.

**Giải pháp:**
```go
type PlaceOrderWorkflow struct {
    steps []Step
}

type Step struct {
    Execute   func() error
    Compensate func() error  // "Undo" nếu fail
}

// Chạy
for _, step := range workflow.steps {
    if err := step.Execute(); err != nil {
        // Undo các step trước đó
        for i := len(workflow.steps) - 2; i >= 0; i-- {
            workflow.steps[i].Compensate()  // Rollback
        }
        return err
    }
}
```

---

### 🔍 4. Value Object (Money, Symbol)

Value Objects là "giá trị" không có identity, không thay đổi.

**Ví dụ:**
```go
// Money là value object - bạn không care nó là object nào, chỉ care giá trị
type Money struct {
    Amount   int64   // 1000 = 10.00
    Currency string  // USD
}

// Hai Money có giá trị giống nhau là bằng nhau
money1 := Money{1000, "USD"}
money2 := Money{1000, "USD"}
money1 == money2  // true

// Money không thay đổi được (immutable)
money1.Amount = 2000  // Đừng làm việc này!
// Thay vào đó:
money3 := money1.Add(Money{1000, "USD"})  // Tạo object mới
```

---

### 🗂️ 5. Shared Kernel (Code chung)

Project có thư mục `internal/sharedkernel/` - nơi chứa code dùng chung.

```
📁 sharedkernel/
├── money.go       ← Money value object
├── symbol.go      ← Symbol (AAPL, GOOGL, ...)
├── ids.go         ← Hàm tạo ID unique
├── errors.go      ← Error định nghĩa chung
└── timerange.go   ← Khoảng thời gian
```

**Quy tắc:** Chỉ chia sẻ:
- Value Objects (Money, Symbol)
- Error definitions
- Utility functions

**KHÔNG chia sẻ:**
- Entity (Order, Account, ...)
- Repository
- Use case

---

## 📊 Tóm Tắt Từ Vựng

| Từ | Nghĩa | Ví Dụ |
|---|---|---|
| **Entity** | Đối tượng có identity duy nhất | Order, Account |
| **Value Object** | Giá trị (không care object) | Money, Symbol |
| **Repository** | Quản lý lưu/lấy data | OrderRepository |
| **Use Case** | Công việc cụ thể | PlaceOrderUseCase |
| **DTO** | Dữ liệu truyền giữa layer | PlaceOrderRequest |
| **Adapter** | Giao tiếp công nghệ | PostgresRepository |
| **Handler** | Xử lý HTTP/gRPC | PlaceOrderHandler |
| **Middleware** | Kiểm tra trước xử lý | Auth, Logging |
| **Event** | Sự kiện xảy ra | OrderPlacedEvent |
| **Outbox** | Lưu event chưa gửi | outbox table |
| **Idempotency** | Chống trùng | Hãy lưu request_id |

---

## 🚀 Các file sẽ code tiếp

Theo thứ tự ưu tiên:

1. **`internal/sharedkernel/`** - Value objects (Money, Symbol)
2. **`internal/orders/domain/`** - Entity (Order)
3. **`internal/accounts/domain/`** - Entity (Account)
4. **`internal/orders/app/`** - Use cases
5. **`internal/orders/adapter/`** - Repository
6. **`internal/orders/transport/`** - HTTP Handler
7. **`cmd/api_gateway/main.go`** - Entry point
8. **Protobuf files** - API contracts

---

## ❓ Câu Hỏi Thường Gặp

**Q: Domain model không connect database được, thế thì lấy data kiểu gì?**
A: Domain model không lấy data trực tiếp. Use Case gọi Repository (adapter) để lấy. Repository trả Domain object, Use Case xử lý logic.

**Q: Tại sao phải 4 layer? Không thể viết một file lớn được không?**
A: Có thể, nhưng:
- 4 layer giúp test dễ hơn
- Thay đổi database mà không ảnh hưởng logic
- Team lớn, mỗi người chịu trách nhiệm layer khác nhau

**Q: DTO là gì, sao không dùng Domain object trực tiếp?**
A: DTO chỉ chứa dữ liệu cần thiết cho HTTP response. Domain object chứa logic không cần cho client.

**Q: Repository interface nằm ở domain hay adapter?**
A: Interface nằm ở domain, implement nằm ở adapter. Vì domain phụ thuộc vào abstraction, không phụ thuộc vào implementation.

---

## 📞 Cần Giúp?

Để hiểu rõ hơn:
1. Tìm một feature yêu thích
2. Trace code từ main.go → handler → usecase → domain → adapter
3. Vẽ diagram riêng mình
4. Viết code cho feature đó

**Happy Coding!** 🚀
