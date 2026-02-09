# 🚀 Hướng Dẫn Bắt Đầu Code - Từng Bước Từ A-Z

**Cho fresher muốn bắt đầu viết code ngay cho project này.**

---

## 📋 Mục Lục
1. [Bước Chuẩn Bị](#bước-1-chuẩn-bị)
2. [Viết Shared Kernel (Value Objects)](#bước-2-viết-shared-kernel)
3. [Viết Domain Models](#bước-3-viết-domain-models)
4. [Viết Use Cases](#bước-4-viết-use-cases)
5. [Viết Adapters (Database)](#bước-5-viết-adapters-database)
6. [Viết Handlers (Transport)](#bước-6-viết-handlers-transport)
7. [Chạy & Test](#bước-7-chạy--test)

---

## Bước 1: Chuẩn Bị

### 1.1 Setup Project Structure

```bash
cd /mnt/c/Users/Admin/Downloads/Project/Github/BackendDataPlatform

# Kiểm tra go version
go version
# Nên ≥ 1.21

# Init modules (nếu chưa có)
go mod init github.com/yourusername/BackendDataPlatform
go mod tidy

# Cài tools cần thiết
go install github.com/bufbuild/buf/cmd/buf@latest
go install github.com/google/wire/cmd/wire@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### 1.2 Tạo File Go Đơn Giản Để Test

Tạo file `main.go` tạm thời:

```bash
cat > main.go << 'EOF'
package main

import "fmt"

func main() {
    fmt.Println("Hello, Trading Platform!")
}
EOF

go run main.go
# Output: Hello, Trading Platform!
```

### 1.3 Tạo Docker Setup (Database & Kafka)

Tạo file `docker-compose.yml`:

```bash
cat > docker-compose.yml << 'EOF'
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_USER: trader
      POSTGRES_PASSWORD: password123
      POSTGRES_DB: trading_db
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

  kafka:
    image: confluentinc/cp-kafka:7.5.0
    environment:
      KAFKA_BROKER_ID: 1
      KAFKA_ZOOKEEPER_CONNECT: zookeeper:2181
      KAFKA_ADVERTISED_LISTENERS: PLAINTEXT://kafka:29092,PLAINTEXT_HOST://localhost:9092
      KAFKA_LISTENER_SECURITY_PROTOCOL_MAP: PLAINTEXT:PLAINTEXT,PLAINTEXT_HOST:PLAINTEXT
      KAFKA_INTER_BROKER_LISTENER_NAME: PLAINTEXT
    ports:
      - "9092:9092"
    depends_on:
      - zookeeper

  zookeeper:
    image: confluentinc/cp-zookeeper:7.5.0
    environment:
      ZOOKEEPER_CLIENT_PORT: 2181

volumes:
  postgres_data:
EOF

# Chạy
docker-compose up -d

# Kiểm tra
docker-compose ps
```

### 1.4 Cài Migrate Tool Để Chạy SQL

```bash
# Cài migrate
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.15.2/migrate.linux-amd64.tar.gz | tar xvz

# Hoặc dùng Go
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Kiểm tra
migrate -version
```

---

## Bước 2: Viết Shared Kernel

### 2.1 Tạo File Money Value Object

**File**: `internal/sharedkernel/money.go`

```go
package sharedkernel

import (
    "fmt"
)

// Money: Giá trị tiền (Value Object)
type Money struct {
    Amount   int64   // 1000 = 10.00
    Currency string  // USD, VND
}

func NewMoney(amount int64, currency string) (*Money, error) {
    if amount < 0 {
        return nil, fmt.Errorf("amount cannot be negative")
    }
    if currency == "" {
        return nil, fmt.Errorf("currency is required")
    }
    return &Money{Amount: amount, Currency: currency}, nil
}

func (m *Money) Add(other *Money) (*Money, error) {
    if m.Currency != other.Currency {
        return nil, fmt.Errorf("cannot add different currencies")
    }
    return &Money{
        Amount:   m.Amount + other.Amount,
        Currency: m.Currency,
    }, nil
}

func (m *Money) IsGreaterThan(other *Money) bool {
    return m.Amount > other.Amount
}
```

### 2.2 Tạo File Symbol Value Object

**File**: `internal/sharedkernel/symbol.go`

```go
package sharedkernel

import (
    "fmt"
    "regexp"
)

type Symbol struct {
    Code string
}

func NewSymbol(code string) (*Symbol, error) {
    if len(code) == 0 || len(code) > 10 {
        return nil, fmt.Errorf("symbol length must be 1-10")
    }
    matched, _ := regexp.MatchString(`^[A-Z0-9]+$`, code)
    if !matched {
        return nil, fmt.Errorf("symbol must be uppercase letters only")
    }
    return &Symbol{Code: code}, nil
}

func (s *Symbol) String() string {
    return s.Code
}
```

### 2.3 Test Shared Kernel

Tạo file `internal/sharedkernel/money_test.go`:

```go
package sharedkernel

import "testing"

func TestNewMoney(t *testing.T) {
    // Test hợp lệ
    m, err := NewMoney(1000, "USD")
    if err != nil {
        t.Fatalf("expected no error, got %v", err)
    }
    if m.Amount != 1000 {
        t.Errorf("expected 1000, got %d", m.Amount)
    }

    // Test không hợp lệ
    m, err = NewMoney(-100, "USD")
    if err == nil {
        t.Error("expected error for negative amount")
    }
}

func TestMoneyAdd(t *testing.T) {
    m1, _ := NewMoney(1000, "USD")
    m2, _ := NewMoney(500, "USD")
    
    result, _ := m1.Add(m2)
    if result.Amount != 1500 {
        t.Errorf("expected 1500, got %d", result.Amount)
    }
}
```

**Chạy test:**

```bash
go test -v ./internal/sharedkernel/...
```

---

## Bước 3: Viết Domain Models

### 3.1 Tạo Order Entity

**File**: `internal/orders/domain/order.go`

```go
package domain

import (
    "fmt"
    "time"
    "github.com/google/uuid"
)

type OrderStatus string

const (
    OrderPending   OrderStatus = "PENDING"
    OrderFilled    OrderStatus = "FILLED"
    OrderCancelled OrderStatus = "CANCELLED"
)

type Order struct {
    ID        string
    UserID    string
    Symbol    string  // Tạm dùng string, sau convert sang Symbol
    Side      string  // BUY, SELL
    Price     int64
    Quantity  float64
    Status    OrderStatus
    CreatedAt time.Time
    Version   int
}

func NewOrder(userID, symbol, side string, price int64, quantity float64) (*Order, error) {
    if userID == "" || symbol == "" {
        return nil, fmt.Errorf("invalid input")
    }
    if side != "BUY" && side != "SELL" {
        return nil, fmt.Errorf("invalid side")
    }
    if price <= 0 || quantity <= 0 {
        return nil, fmt.Errorf("price and quantity must be positive")
    }

    return &Order{
        ID:        "ORDER-" + uuid.New().String()[:8],
        UserID:    userID,
        Symbol:    symbol,
        Side:      side,
        Price:     price,
        Quantity:  quantity,
        Status:    OrderPending,
        CreatedAt: time.Now(),
        Version:   1,
    }, nil
}

func (o *Order) Validate() error {
    if o.ID == "" || o.UserID == "" {
        return fmt.Errorf("invalid order")
    }
    return nil
}

func (o *Order) Cancel() error {
    if o.Status != OrderPending {
        return fmt.Errorf("cannot cancel %s order", o.Status)
    }
    o.Status = OrderCancelled
    o.Version++
    return nil
}
```

### 3.2 Tạo Repository Interface

**File**: `internal/orders/domain/repository.go`

```go
package domain

import "context"

type OrderRepository interface {
    Save(ctx context.Context, order *Order) error
    GetByID(ctx context.Context, id string) (*Order, error)
    Update(ctx context.Context, order *Order) error
}

type EventPublisher interface {
    Publish(ctx context.Context, topic string, event interface{}) error
}
```

### 3.3 Test Domain

**File**: `internal/orders/domain/order_test.go`

```go
package domain

import "testing"

func TestNewOrder(t *testing.T) {
    order, err := NewOrder("USER-1", "AAPL", "BUY", 18000, 10)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if order.Status != OrderPending {
        t.Errorf("expected PENDING, got %s", order.Status)
    }
}

func TestOrderCancel(t *testing.T) {
    order, _ := NewOrder("USER-1", "AAPL", "BUY", 18000, 10)
    
    err := order.Cancel()
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    
    if order.Status != OrderCancelled {
        t.Errorf("expected CANCELLED, got %s", order.Status)
    }
}
```

**Chạy:**

```bash
go test -v ./internal/orders/domain/...
```

---

## Bước 4: Viết Use Cases

### 4.1 Tạo PlaceOrderUseCase

**File**: `internal/orders/app/place_order_usecase.go`

```go
package app

import (
    "context"
    "fmt"
    "github.com/BackendDataPlatform/internal/orders/domain"
)

type PlaceOrderCommand struct {
    UserID   string
    Symbol   string
    Side     string
    Price    int64
    Quantity float64
}

type PlaceOrderResult struct {
    OrderID string
    Status  string
}

type PlaceOrderUseCase struct {
    orderRepo domain.OrderRepository
    publisher domain.EventPublisher
}

func NewPlaceOrderUseCase(
    orderRepo domain.OrderRepository,
    publisher domain.EventPublisher,
) *PlaceOrderUseCase {
    return &PlaceOrderUseCase{
        orderRepo: orderRepo,
        publisher: publisher,
    }
}

func (uc *PlaceOrderUseCase) Execute(
    ctx context.Context,
    cmd *PlaceOrderCommand,
) (*PlaceOrderResult, error) {
    // Validate
    if cmd.UserID == "" || cmd.Symbol == "" {
        return nil, fmt.Errorf("invalid input")
    }

    // Create order (domain logic)
    order, err := domain.NewOrder(cmd.UserID, cmd.Symbol, cmd.Side, cmd.Price, cmd.Quantity)
    if err != nil {
        return nil, fmt.Errorf("failed to create order: %w", err)
    }

    // Validate
    if err := order.Validate(); err != nil {
        return nil, fmt.Errorf("invalid order: %w", err)
    }

    // Save (adapter)
    if err := uc.orderRepo.Save(ctx, order); err != nil {
        return nil, fmt.Errorf("failed to save order: %w", err)
    }

    // Publish event
    _ = uc.publisher.Publish(ctx, "orders.placed", order)

    return &PlaceOrderResult{
        OrderID: order.ID,
        Status:  string(order.Status),
    }, nil
}
```

### 4.2 Test Use Case Với Mock

**File**: `internal/orders/app/place_order_usecase_test.go`

```go
package app

import (
    "context"
    "testing"
    "github.com/BackendDataPlatform/internal/orders/domain"
)

// Mock repository
type MockOrderRepository struct {
    SaveCalled bool
}

func (m *MockOrderRepository) Save(ctx context.Context, order *domain.Order) error {
    m.SaveCalled = true
    return nil
}

func (m *MockOrderRepository) GetByID(ctx context.Context, id string) (*domain.Order, error) {
    return nil, nil
}

func (m *MockOrderRepository) Update(ctx context.Context, order *domain.Order) error {
    return nil
}

// Mock publisher
type MockEventPublisher struct{}

func (m *MockEventPublisher) Publish(ctx context.Context, topic string, event interface{}) error {
    return nil
}

// Test
func TestPlaceOrderUseCase(t *testing.T) {
    mockRepo := &MockOrderRepository{}
    mockPub := &MockEventPublisher{}
    
    uc := NewPlaceOrderUseCase(mockRepo, mockPub)
    
    cmd := &PlaceOrderCommand{
        UserID:   "USER-1",
        Symbol:   "AAPL",
        Side:     "BUY",
        Price:    18000,
        Quantity: 10,
    }
    
    result, err := uc.Execute(context.Background(), cmd)
    
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    
    if result.OrderID == "" {
        t.Error("expected order ID")
    }
    
    if !mockRepo.SaveCalled {
        t.Error("repository Save not called")
    }
}
```

**Chạy:**

```bash
go test -v ./internal/orders/app/...
```

---

## Bước 5: Viết Adapters (Database)

### 5.1 Tạo SQL Migrations

**File**: `migrations/orders/000001_orders_init.up.sql`

```sql
CREATE TABLE IF NOT EXISTS orders (
    id              VARCHAR(36) PRIMARY KEY,
    user_id         VARCHAR(36) NOT NULL,
    symbol          VARCHAR(10) NOT NULL,
    side            VARCHAR(10) NOT NULL,
    price           BIGINT NOT NULL,
    quantity        DOUBLE PRECISION NOT NULL,
    status          VARCHAR(20) NOT NULL,
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    version         INTEGER NOT NULL DEFAULT 1
);

CREATE INDEX idx_orders_user_id ON orders(user_id);
```

**File**: `migrations/orders/000001_orders_init.down.sql`

```sql
DROP TABLE IF EXISTS orders;
```

### 5.2 Chạy Migrations

```bash
# Cài migrate
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Chạy
migrate -path migrations/orders -database "postgresql://trader:password123@localhost:5432/trading_db?sslmode=disable" up

# Check
psql -U trader -d trading_db -c "\\dt"
```

### 5.3 Implement OrderRepository

**File**: `internal/orders/adapter/store/postgres/repo.go`

```go
package postgres

import (
    "context"
    "database/sql"
    "fmt"
    "github.com/BackendDataPlatform/internal/orders/domain"
)

type OrderRepository struct {
    db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
    return &OrderRepository{db: db}
}

func (r *OrderRepository) Save(ctx context.Context, order *domain.Order) error {
    query := `
        INSERT INTO orders (id, user_id, symbol, side, price, quantity, status, created_at, version)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
    `
    
    _, err := r.db.ExecContext(
        ctx, query,
        order.ID, order.UserID, order.Symbol, order.Side,
        order.Price, order.Quantity, order.Status, order.CreatedAt, order.Version,
    )
    
    if err != nil {
        return fmt.Errorf("failed to save order: %w", err)
    }
    
    return nil
}

func (r *OrderRepository) GetByID(ctx context.Context, id string) (*domain.Order, error) {
    query := `
        SELECT id, user_id, symbol, side, price, quantity, status, created_at, version
        FROM orders WHERE id = $1
    `
    
    var o domain.Order
    err := r.db.QueryRowContext(ctx, query, id).Scan(
        &o.ID, &o.UserID, &o.Symbol, &o.Side, &o.Price,
        &o.Quantity, (*string)(&o.Status), &o.CreatedAt, &o.Version,
    )
    
    if err == sql.ErrNoRows {
        return nil, fmt.Errorf("order not found")
    }
    if err != nil {
        return nil, fmt.Errorf("failed to get order: %w", err)
    }
    
    return &o, nil
}

func (r *OrderRepository) Update(ctx context.Context, order *domain.Order) error {
    query := `UPDATE orders SET status = $1, version = $2 WHERE id = $3`
    
    result, err := r.db.ExecContext(ctx, query, order.Status, order.Version, order.ID)
    if err != nil {
        return fmt.Errorf("failed to update order: %w", err)
    }
    
    affected, _ := result.RowsAffected()
    if affected == 0 {
        return fmt.Errorf("order not found")
    }
    
    return nil
}
```

---

## Bước 6: Viết Handlers (Transport)

### 6.1 Setup HTTP Server

**File**: `cmd/orders_service/main.go`

```go
package main

import (
    "database/sql"
    "fmt"
    "log"
    "net/http"
    
    _ "github.com/lib/pq"
    "github.com/BackendDataPlatform/internal/orders/adapter/store/postgres"
    "github.com/BackendDataPlatform/internal/orders/app"
    "github.com/BackendDataPlatform/internal/orders/transport"
)

func main() {
    // 1. Connect to database
    db, err := sql.Open("postgres", "postgres://trader:password123@localhost:5432/trading_db?sslmode=disable")
    if err != nil {
        log.Fatalf("failed to connect to database: %v", err)
    }
    defer db.Close()
    
    // 2. Create repository
    orderRepo := postgres.NewOrderRepository(db)
    
    // 3. Create publisher (mock for now)
    publisher := &MockPublisher{}
    
    // 4. Create use case
    uc := app.NewPlaceOrderUseCase(orderRepo, publisher)
    
    // 5. Create HTTP handler
    handler := transport.NewPlaceOrderHandler(uc)
    
    // 6. Setup routes
    http.HandleFunc("/orders", handler.ServeHTTP)
    http.HandleFunc("/health", healthHandler)
    
    // 7. Start server
    fmt.Println("Server starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(`{"status": "ok"}`))
}

// Mock publisher
type MockPublisher struct{}

func (m *MockPublisher) Publish(ctx context.Context, topic string, event interface{}) error {
    return nil
}
```

### 6.2 Implement HTTP Handler

**File**: `internal/orders/transport/http_handler.go`

```go
package transport

import (
    "encoding/json"
    "fmt"
    "net/http"
    "github.com/BackendDataPlatform/internal/orders/app"
)

type PlaceOrderRequest struct {
    Symbol   string  `json:"symbol"`
    Side     string  `json:"side"`
    Price    int64   `json:"price"`
    Quantity float64 `json:"quantity"`
}

type PlaceOrderHandler struct {
    usecase *app.PlaceOrderUseCase
}

func NewPlaceOrderHandler(usecase *app.PlaceOrderUseCase) *PlaceOrderHandler {
    return &PlaceOrderHandler{usecase: usecase}
}

func (h *PlaceOrderHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    
    // Tạm dùng hardcoded user ID (thực tế từ JWT)
    userID := "USER-1"
    
    var req PlaceOrderRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }
    
    cmd := &app.PlaceOrderCommand{
        UserID:   userID,
        Symbol:   req.Symbol,
        Side:     req.Side,
        Price:    req.Price,
        Quantity: req.Quantity,
    }
    
    result, err := h.usecase.Execute(r.Context(), cmd)
    if err != nil {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(result)
}
```

---

## Bước 7: Chạy & Test

### 7.1 Docker Up

```bash
docker-compose up -d
docker-compose logs postgres  # Check postgres started

# Chạy migrations
migrate -path migrations/orders -database "postgresql://trader:password123@localhost:5432/trading_db?sslmode=disable" up
```

### 7.2 Chạy Server

```bash
# Từ project root
cd cmd/orders_service
go run main.go

# Output: Server starting on :8080
```

### 7.3 Test HTTP

```bash
# Mở terminal khác

# Health check
curl http://localhost:8080/health

# Place order (test hợp lệ)
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{
    "symbol": "AAPL",
    "side": "BUY",
    "price": 18000,
    "quantity": 10
  }'

# Response:
# {"OrderID":"ORDER-a1b2c3d4","Status":"PENDING"}

# Kiểm tra database
psql -U trader -d trading_db -c "SELECT * FROM orders;"
```

### 7.4 Chạy Unit Tests

```bash
# Test all
go test -v ./...

# Test specific package
go test -v ./internal/orders/domain/...
go test -v ./internal/orders/app/...

# With coverage
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

---

## 📊 Checklist - Công Việc Đã Làm

- ✅ Bước 1: Setup & Docker
- ✅ Bước 2: Shared Kernel (Money, Symbol)
- ✅ Bước 3: Domain Models (Order entity)
- ✅ Bước 4: Use Cases (PlaceOrderUseCase)
- ✅ Bước 5: Adapters (PostgreSQL Repository)
- ✅ Bước 6: Transport (HTTP Handler)
- ✅ Bước 7: Run & Test

---

## 🎯 Tiếp Theo Là Gì?

```
Sau khi hoàn thành 7 bước trên, bạn sẽ:

1️⃣ Implement thêm Services:
   - Accounts Service (tương tự Orders)
   - Market Data Service
   
2️⃣ Thêm Features:
   - Cancel Order use case
   - Get Order by ID
   - List Orders by User
   
3️⃣ Kafka Integration:
   - Setup Kafka Publisher
   - Setup Outbox Pattern
   - Setup Consumers
   
4️⃣ gRPC Communication:
   - Define Protobuf messages
   - Implement gRPC services
   
5️⃣ Observability:
   - Add logging
   - Add metrics
   - Add tracing
   
6️⃣ Deploy:
   - Docker Compose cho local
   - Kubernetes manifests
   - CI/CD pipeline
```

---

## ❓ Troubleshooting

### Database connection error

```bash
# Check postgres is running
docker-compose ps

# If not, restart
docker-compose down
docker-compose up -d

# Check connection
psql -U trader -d trading_db -c "SELECT 1"
```

### Migration failed

```bash
# Check migration status
migrate -path migrations/orders -database "postgresql://..." version

# Force version
migrate -path migrations/orders -database "postgresql://..." force 1

# Try again
migrate -path migrations/orders -database "postgresql://..." up
```

### Port already in use

```bash
# Find process using port 8080
lsof -i :8080

# Kill it
kill -9 <PID>

# Or use different port
go run main.go -port 8081
```

### Import errors

```bash
# Download dependencies
go mod download
go mod tidy

# Verify imports
go vet ./...
```

---

**Chúc bạn viết code vui vẻ!** 🚀

Nếu gặp vấn đề, hãy check lại từng bước và đảm bảo:
- ✅ Docker services running
- ✅ Database migrations done
- ✅ Go modules tidy
- ✅ Code compiles without errors
- ✅ Tests pass
