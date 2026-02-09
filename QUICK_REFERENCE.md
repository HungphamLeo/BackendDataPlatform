# 📚 Quick Reference - Tra Cứu Nhanh

**Dành cho khi bạn cần tra cứu nhanh cách làm một việc gì đó.**

---

## 🏗️ Project Structure

```
internal/
├── sharedkernel/      ← Value Objects, lỗi chung
│   ├── money.go       # Giá trị tiền
│   ├── symbol.go      # Mã SP
│   └── errors.go      # Lỗi shared
│
├── orders/
│   ├── domain/        ← Business Logic (không dependency)
│   │   ├── order.go   # Order entity
│   │   └── repo.go    # Repository interface
│   ├── app/           ← Use cases (điều phối)
│   │   └── place_order_usecase.go
│   ├── adapter/       ← Công nghệ (DB, Kafka, Redis)
│   │   └── store/postgres/repo.go
│   └── transport/     ← HTTP/gRPC handlers
│       ├── http_handler.go
│       └── rest.go
│
├── accounts/          ← Tương tự như orders/
│   ├── domain/
│   ├── app/
│   ├── adapter/
│   └── transport/
│
├── platform/          ← Shared infrastructure
│   ├── config/        # Configuration
│   ├── storage/       # Database, Redis
│   ├── messaging/     # Kafka
│   ├── transport/     # HTTP/gRPC shared
│   ├── observability/ # Logging, Metrics
│   ├── outbox/        # Outbox pattern
│   ├── idempotency/   # Idempotency
│   ├── runtime/       # App lifecycle
│   ├── workflow/      # Saga pattern
│   └── testing/       # Test utils
│
└── gateway/           ← API Gateway logic
    ├── domain/
    ├── app/
    └── transport/

cmd/
├── api_gateway/       ← Entry point 1
│   ├── main.go
│   ├── wire.go
│   └── routes.go
├── orders_service/    ← Entry point 2
│   ├── main.go
│   └── wire.go
├── accounts_service/
├── market_ingestor/
├── history_writer/
├── realtime_gateway/
├── stream_processor/
└── redis_janitor/

migrations/
├── orders/
│   ├── 000001_orders_init.up.sql
│   └── 000001_orders_init.down.sql
├── accounts/
└── shared/

api/proto/            ← Protobuf definitions
├── orders/v1/
├── accounts/v1/
├── marketdata/v1/
└── common/v1/

tests/
├── unit/              ← Domain + App tests
├── integration/       ← DB + Kafka tests
├── contract/          ← gRPC contract tests
├── load/k6/           ← Load tests
└── chaos/             ← Chaos engineering

deployments/
├── docker/            ← Dockerfiles
├── docker_compose/    ← Local setup
└── k8s/               ← Kubernetes manifests
```

---

## 🔥 Common Tasks

### Tạo Entity Mới

**Domain Model** (`internal/orders/domain/entity.go`):

```go
type Entity struct {
    ID    string
    /* fields */
}

func NewEntity(...) (*Entity, error) {
    // Validate input
    // Return Entity
}

// Domain method (no dependencies)
func (e *Entity) DoSomething() error {
    // Business logic
}
```

**Repository Interface** (`internal/orders/domain/repository.go`):

```go
type EntityRepository interface {
    Save(ctx context.Context, entity *Entity) error
    GetByID(ctx context.Context, id string) (*Entity, error)
    Update(ctx context.Context, entity *Entity) error
}
```

---

### Tạo Use Case

```go
// internal/orders/app/my_usecase.go

type MyCommand struct {
    /* input fields */
}

type MyResult struct {
    /* output fields */
}

type MyUseCase struct {
    repo1 domain.Repo1Repository
    repo2 domain.Repo2Repository
    publisher domain.EventPublisher
}

func NewMyUseCase(repo1, repo2, publisher) *MyUseCase {
    return &MyUseCase{...}
}

func (uc *MyUseCase) Execute(ctx context.Context, cmd *MyCommand) (*MyResult, error) {
    // 1. Validate input
    // 2. Get data from adapter
    // 3. Call domain logic
    // 4. Save data via adapter
    // 5. Publish event
    // 6. Return result or error
}
```

---

### Test Use Case

```go
// internal/orders/app/my_usecase_test.go

type MockRepository struct{
    SaveCalled bool
}

func (m *MockRepository) Save(ctx context.Context, entity *Entity) error {
    m.SaveCalled = true
    return nil
}

// ... other interface methods

func TestMyUseCase(t *testing.T) {
    mockRepo := &MockRepository{}
    uc := NewMyUseCase(mockRepo, ...)
    
    result, err := uc.Execute(context.Background(), &MyCommand{...})
    
    if err != nil {
        t.Fatalf("expected no error: %v", err)
    }
    if !mockRepo.SaveCalled {
        t.Error("Save not called")
    }
}
```

---

### Implement Repository (Database)

```go
// internal/orders/adapter/store/postgres/repo.go

type PostgresRepository struct {
    db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
    return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Save(ctx context.Context, entity *Entity) error {
    query := "INSERT INTO table_name (...) VALUES (...)"
    _, err := r.db.ExecContext(ctx, query, /* values */)
    if err != nil {
        return fmt.Errorf("failed to save: %w", err)
    }
    return nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id string) (*Entity, error) {
    query := "SELECT * FROM table_name WHERE id = $1"
    var entity Entity
    err := r.db.QueryRowContext(ctx, query, id).Scan(/* fields */)
    if err == sql.ErrNoRows {
        return nil, fmt.Errorf("not found")
    }
    if err != nil {
        return nil, fmt.Errorf("failed to get: %w", err)
    }
    return &entity, nil
}
```

---

### Viết HTTP Handler

```go
// internal/orders/transport/http_handler.go

type MyRequest struct {
    Field1 string `json:"field_1"`
    Field2 int64  `json:"field_2"`
}

type MyResultResponse struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
    Error   string      `json:"error,omitempty"`
}

type MyHandler struct {
    usecase *app.MyUseCase
}

func NewMyHandler(usecase *app.MyUseCase) *MyHandler {
    return &MyHandler{usecase: usecase}
}

func (h *MyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // 1. Check method
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    // 2. Parse request
    var req MyRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(map[string]string{
            "error": fmt.Sprintf("Invalid JSON: %v", err),
        })
        return
    }

    // 3. Create command
    cmd := &app.MyCommand{
        Field1: req.Field1,
        Field2: req.Field2,
    }

    // 4. Call use case
    result, err := h.usecase.Execute(r.Context(), cmd)
    if err != nil {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(MyResultResponse{
            Success: false,
            Error:   err.Error(),
        })
        return
    }

    // 5. Return success
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(MyResultResponse{
        Success: true,
        Data:    result,
    })
}
```

---

### Viết SQL Migration

```bash
# Tạo file migration
touch migrations/orders/000002_add_feature.up.sql
touch migrations/orders/000002_add_feature.down.sql
```

**File.up.sql** (khi migrate up):

```sql
-- Add columns
ALTER TABLE orders ADD COLUMN new_field VARCHAR(255);

-- Create table
CREATE TABLE new_table (
    id VARCHAR(36) PRIMARY KEY,
    field1 VARCHAR(100) NOT NULL,
    field2 BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create index
CREATE INDEX idx_new_table_field1 ON new_table(field1);
```

**File.down.sql** (khi migrate down):

```sql
DROP TABLE IF EXISTS new_table;
ALTER TABLE orders DROP COLUMN new_field;
```

**Chạy:**

```bash
# Up
migrate -path migrations/orders -database "postgresql://..." up

# Down
migrate -path migrations/orders \
  -database "postgresql://..." down

# Version cụ thể
migrate -path migrations/orders \
  -database "postgresql://..." goto 2
```

---

## 🧪 Testing Cheat Sheet

### Unit Test Domain

```bash
go test -v ./internal/orders/domain/...
```

### Unit Test App (với mock)

```bash
go test -v ./internal/orders/app/...
```

### Get Coverage

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out  # Mở browser
```

### Test One Function

```bash
go test -v -run TestFunctionName ./internal/orders/domain/...
```

### Benchmark

```bash
go test -bench=. -benchmem ./internal/orders/domain/...
```

### Check Race Conditions

```bash
go test -race ./...
```

---

## 🐳 Docker & Database

### Start Services

```bash
# Start all
docker-compose up -d

# Start specific
docker-compose up -d postgres
docker-compose up -d redis

# Check status
docker-compose ps

# View logs
docker-compose logs postgres
```

### Database Operations

```bash
# Connect to DB
psql -U trader -d trading_db

# List tables
\dt

# Describe table
\d orders

# Query
SELECT * FROM orders;

# Exit
\q
```

### Run Migrations

```bash
# Up all
migrate -path migrations/orders \
  -database "postgresql://trader:password123@localhost:5432/trading_db?sslmode=disable" up

# Down all
migrate -path migrations/orders \
  -database "postgresql://trader:password123@localhost:5432/trading_db?sslmode=disable" down

# Check current version
migrate -path migrations/orders \
  -database "postgresql://trader:password123@localhost:5432/trading_db?sslmode=disable" version
```

---

## 🌐 HTTP Testing

### With curl

```bash
# GET
curl http://localhost:8080/orders/ORDER-123

# POST
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{"symbol":"AAPL","side":"BUY"}'

# With header
curl -H "Authorization: Bearer TOKEN" \
  http://localhost:8080/orders

# With custom header
curl -H "Idempotency-Key: key-123" \
  http://localhost:8080/orders
```

### With httpie (better than curl)

```bash
# Install
go install github.com/httpie/cli@latest
# or: brew install httpie

# GET
http :8080/orders/ORDER-123

# POST
http POST :8080/orders symbol=AAPL side=BUY

# JSON body
http POST :8080/orders < body.json

# Pretty print
http --pretty=all POST :8080/orders
```

---

## 🔍 Debugging

### View Stack Trace

```bash
# Run and get debug info
go run ./cmd/api_gateway -v

# Check specific line
grep -n "function_name" internal/orders/app/file.go
```

### Use Debugger (Delve)

```bash
# Install
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug main
dlv debug ./cmd/api_gateway
# (dlv) break main.main
# (dlv) continue
# (dlv) next

# Debug test
dlv test ./internal/orders/app
```

### Structured Logging

```go
import "fmt"

// Simple
fmt.Printf("Debug: order_id=%s, status=%s\n", orderID, status)

// Better (for production)
type Logger interface {
    Info(msg string, fields map[string]interface{})
    Error(msg string, err error)
}
```

---

## 📊 Protobuf

### Define Message

**File**: `api/proto/orders/v1/orders.proto`

```protobuf
syntax = "proto3";

package orders.v1;

message Order {
    string id = 1;
    string user_id = 2;
    string symbol = 3;
    string side = 4;
    int64 price = 5;
    double quantity = 6;
}

service OrderService {
    rpc PlaceOrder(PlaceOrderRequest) returns (PlaceOrderResponse);
}
```

### Generate Go Code

```bash
# Install buf
go install github.com/bufbuild/buf/cmd/buf@latest

# Create buf.gen.yml
cat > buf.gen.yml << 'EOF'
version: v1
plugins:
  - plugin: go
    out: internal/pb
    opt: paths=source_relative
  - plugin: go-grpc
    out: internal/pb
    opt: paths=source_relative
EOF

# Generate
buf generate api/proto

# Check generated files
ls -la internal/pb/orders/v1/
```

---

## ⚙️ Wire (Dependency Injection)

### Define Providers

**File**: `cmd/api_gateway/wire.go`

```go
package main

import (
    "github.com/google/wire"
    "github.com/BackendDataPlatform/internal/orders/app"
    "github.com/BackendDataPlatform/internal/orders/adapter/store/postgres"
)

var ProviderSet = wire.NewSet(
    postgres.NewDB,
    postgres.NewOrderRepository,
    app.NewPlaceOrderUseCase,
)

func InitializeApp() *App {
    wire.Build(ProviderSet)
    return &App{}
}
```

### Generate Wire Code

```bash
# Install wire
go install github.com/google/wire/cmd/wire@latest

# Generate wire_gen.go
wire ./cmd/api_gateway

# Check generated
cat cmd/api_gateway/wire_gen.go
```

---

## 📈 Performance Tips

### Optimize Database

```sql
-- Add indexes trước khi query
CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_created_at ON orders(created_at DESC);

-- Check slow queries
EXPLAIN ANALYZE SELECT * FROM orders WHERE user_id = 'USER-1';
```

### Batch Operations

```go
// Instead of loop
for _, order := range orders {
    repo.Save(ctx, order)  // ❌ 1000 queries
}

// Do batch
repo.SaveBatch(ctx, orders)  // ✅ 1 query
```

### Use Connection Pooling

```go
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(5 * time.Minute)
```

### Cache Frequently Accessed Data

```go
// Cache in Redis
userCache.Set("user:123", userData, 1*time.Hour)
userCache.Get("user:123")
```

---

## 🔐 Common Errors & Fixes

### "connection refused"

```bash
# Check Docker
docker-compose ps

# Restart
docker-compose restart postgres

# Wait for ready
docker-compose up -d postgres && sleep 5
```

### "syntax error in SQL"

```bash
# Check syntax
SHOW QUERY BUFFER;

# Run step by step
\set ECHO all
\i migrations/orders/001_up.sql
```

### "import cycle detected"

```
// Domain (no imports from app/adapter) ✅
// App (imports domain + adapter interfaces) ✅
// Adapter (imports domain) ✅
// Transport (imports application) ✅

// ❌ Wrong: Domain imports App
// ❌ Wrong: Adapter imports Transport
```

### "interface{} does not implement"

```go
// Make sure all methods match
type MyInterface interface {
    Method1(ctx context.Context) error
    Method2(ctx context.Context, id string) (*Entity, error)
}

// Implement ALL methods
func (m *MyImplementation) Method1(ctx context.Context) error {
    // ...
}

func (m *MyImplementation) Method2(ctx context.Context, id string) (*Entity, error) {
    // ...
}
```

---

## 🎯 Shortcuts & Aliases

**Add to `.bashrc` or `.zshrc`:**

```bash
# Go commands
alias gt='go test -v ./...'
alias gtc='go test -v -coverprofile=cov.out ./... && go tool cover -html=cov.out'
alias gr='go run'
alias gb='go build'
alias gm='go mod tidy'

# Database
alias db-up='docker-compose up -d postgres && sleep 3'
alias db-down='docker-compose down'
alias db-migrate='migrate -path migrations/orders -database "postgresql://trader:password123@localhost:5432/trading_db?sslmode=disable" up'

# Git
alias gs='git status'
alias gc='git commit'
alias gp='git push'

# Project
alias proj-up='docker-compose up -d && sleep 5 && db-migrate'
alias proj-test='go test -v -race ./...'
```

Then use:
```bash
gt        # Run all tests
gtc       # Run tests with coverage
proj-up   # Start project
proj-test # Test project
```

---

**Bookmark this page!** 📌

Lần sau khi cần tra cứu, chỉ cần mở file này lên! 🚀
