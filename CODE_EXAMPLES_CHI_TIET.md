
# 💻 Code Examples - Hướng Dẫn Chi Tiết Từng File

Đây là các đoạn code mẫu bạn sẽ viết cho từng tệp trong project.

## 1️⃣ Shared Kernel (Code Chung)

### File: `internal/sharedkernel/money.go`

---
```go
package sharedkernel

import (
    "fmt"
)

// Money: Đối tượng giá trị (Value Object) - không thay đổi được
type Money struct {
    Amount   int64   // Số tiền (tính theo cent, ví dụ: 1000 = 10.00 USD)
    Currency string  // Loại tiền (USD, VND, EUR)
}

// NewMoney: Tạo Money mới
func NewMoney(amount int64, currency string) (*Money, error) {
    if amount < 0 {
        return nil, fmt.Errorf("amount cannot be negative: %d", amount)
    }
    if currency == "" {
        return nil, fmt.Errorf("currency is required")
    }
    return &Money{
        Amount:   amount,
        Currency: currency,
    }, nil
}

// Add: Cộng hai Money (tạo object mới, không thay đổi hiện tại)
func (m *Money) Add(other *Money) (*Money, error) {
    if m.Currency != other.Currency {
        return nil, fmt.Errorf("cannot add different currencies: %s + %s", m.Currency, other.Currency)
    }
    return &Money{
        Amount:   m.Amount + other.Amount,
        Currency: m.Currency,
    }, nil
}

// Subtract: Trừ hai Money
func (m *Money) Subtract(other *Money) (*Money, error) {
    if m.Currency != other.Currency {
        return nil, fmt.Errorf("cannot subtract different currencies")
    }
    if m.Amount < other.Amount {
        return nil, fmt.Errorf("insufficient balance: %d < %d", m.Amount, other.Amount)
    }
    return &Money{
        Amount:   m.Amount - other.Amount,
        Currency: m.Currency,
    }, nil
}

// IsGreaterThan: So sánh xem có nhiều hơn không
func (m *Money) IsGreaterThan(other *Money) bool {
    if m.Currency != other.Currency {
        return false
    }
    return m.Amount > other.Amount
}

// IsEqual: Hai Money có bằng nhau không
func (m *Money) IsEqual(other *Money) bool {
    return m.Amount == other.Amount && m.Currency == other.Currency
}

// String: Để in ra (ví dụ: "10.00 USD")
func (m *Money) String() string {
    return fmt.Sprintf("%.2f %s", float64(m.Amount)/100, m.Currency)
}
```
---

**Cách dùng:**
---
```go
// Tạo 10 USD
money1, _ := sharedkernel.NewMoney(1000, "USD")

// Tạo 5 USD
money2, _ := sharedkernel.NewMoney(500, "USD")

// Cộng lại
total, _ := money1.Add(money2)
fmt.Println(total)  // "15.00 USD"

// Trừ
remaining, _ := money1.Subtract(money2)
fmt.Println(remaining)  // "5.00 USD"

// So sánh
if money1.IsGreaterThan(money2) {
    fmt.Println("money1 nhiều hơn")
}
```
---

### File: `internal/sharedkernel/symbol.go`

---
```go
package sharedkernel

import (
    "fmt"
    "regexp"
)

// Symbol: Mã chứng chỉ (AAPL, GOOGL, ...)
type Symbol struct {
    Code string
}

// NewSymbol: Tạo Symbol mới (phải đúng format)
func NewSymbol(code string) (*Symbol, error) {
    if len(code) < 1 || len(code) > 10 {
        return nil, fmt.Errorf("symbol must be 1-10 characters")
    }
    
    // Kiểm tra chỉ chứa chữ cái
    if matched, _ := regexp.MatchString(`^[A-Z0-9]+$`, code); !matched {
        return nil, fmt.Errorf("symbol must be uppercase letters/numbers only")
    }
    
    return &Symbol{Code: code}, nil
}

// String: Để in ra
func (s *Symbol) String() string {
    return s.Code
}

// IsEqual: Hai Symbol có bằng nhau không
func (s *Symbol) IsEqual(other *Symbol) bool {
    return s.Code == other.Code
}
```
---

### File: `internal/sharedkernel/errors.go`

---
```go
package sharedkernel

import "fmt"

// DomainError: Lỗi domain (sự kiện gì không hợp lệ trong logic kinh doanh)
type DomainError struct {
    Code    string
    Message string
}

func (e *DomainError) Error() string {
    return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Error definitions
var (
    ErrInsufficientBalance = &DomainError{
        Code:    "INSUFFICIENT_BALANCE",
        Message: "Account balance is insufficient for this operation",
    }
    
    ErrInvalidOrder = &DomainError{
        Code:    "INVALID_ORDER",
        Message: "Order is invalid",
    }
    
    ErrOrderNotFound = &DomainError{
        Code:    "ORDER_NOT_FOUND",
        Message: "Order not found",
    }
)
```
---

## 2️⃣ Domain Layer

### File: `internal/orders/domain/order.go`

---
```go
package domain

import (
    "fmt"
    "time"
    "github.com/google/uuid"
    "github.com/BackendDataPlatform/internal/sharedkernel"
)

// OrderStatus: Trạng thái đơn hàng
type OrderStatus string

const (
    OrderPending   OrderStatus = "PENDING"    // Chờ thực hiện
    OrderFilled    OrderStatus = "FILLED"     // Đã thực hiện
    OrderPartial   OrderStatus = "PARTIAL"    // Thực hiện một phần
    OrderCancelled OrderStatus = "CANCELLED"  // Đã hủy
)

// Order: Entity chính - Đơn hàng
type Order struct {
    ID        string
    UserID    string
    Symbol    *sharedkernel.Symbol
    Side      string  // "BUY" hoặc "SELL"
    Price     *sharedkernel.Money
    Quantity  float64
    Status    OrderStatus
    CreatedAt time.Time
    UpdatedAt time.Time
    Version   int  // Optimistic locking
}

// NewOrder: Factory để tạo order mới
func NewOrder(
    userID string,
    symbol *sharedkernel.Symbol,
    side string,
    price *sharedkernel.Money,
    quantity float64,
) (*Order, error) {
    // Validate input
    if userID == "" {
        return nil, fmt.Errorf("userID is required")
    }
    if side != "BUY" && side != "SELL" {
        return nil, fmt.Errorf("side must be BUY or SELL")
    }
    if quantity <= 0 {
        return nil, fmt.Errorf("quantity must be positive")
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
        UpdatedAt: time.Now(),
        Version:   1,
    }, nil
}

// Validate: Kiểm tra order có hợp lệ không
func (o *Order) Validate() error {
    if o.ID == "" {
        return fmt.Errorf("order ID is required")
    }
    if o.UserID == "" {
        return fmt.Errorf("user ID is required")
    }
    if o.Symbol == nil {
        return fmt.Errorf("symbol is required")
    }
    if o.Price == nil {
        return fmt.Errorf("price is required")
    }
    if o.Quantity <= 0 {
        return fmt.Errorf("quantity must be positive")
    }
    return nil
}

// GetTotalValue: Tính tổng giá trị order
func (o *Order) GetTotalValue() *sharedkernel.Money {
    // totalPrice = price * quantity
    singlePrice := o.Price.Amount
    totalAmount := singlePrice * int64(o.Quantity)
    total, _ := sharedkernel.NewMoney(totalAmount, o.Price.Currency)
    return total
}

// CanCancel: Có thể hủy order không?
func (o *Order) CanCancel() bool {
    return o.Status == OrderPending || o.Status == OrderPartial
}

// Cancel: Hủy order
func (o *Order) Cancel() error {
    if !o.CanCancel() {
        return fmt.Errorf("cannot cancel order in %s status", o.Status)
    }
    o.Status = OrderCancelled
    o.UpdatedAt = time.Now()
    o.Version++
    return nil
}

// Fill: Thực hiện order
func (o *Order) Fill() error {
    if o.Status == OrderPending || o.Status == OrderPartial {
        o.Status = OrderFilled
        o.UpdatedAt = time.Now()
        o.Version++
        return nil
    }
    return fmt.Errorf("cannot fill order in %s status", o.Status)
}
```
---

### File: `internal/orders/domain/repository.go`

---
```go
package domain

import (
    "context"
)

// OrderRepository: Interface để lưu/lấy Order từ database
// Lưu ý: Interface định nghĩa ở domain (rules), implement ở adapter
type OrderRepository interface {
    Save(ctx context.Context, order *Order) error
    GetByID(ctx context.Context, id string) (*Order, error)
    GetByUserID(ctx context.Context, userID string) ([]*Order, error)
    Update(ctx context.Context, order *Order) error
    Delete(ctx context.Context, id string) error
}

// EventPublisher: Interface để gửi event (sự kiện)
type EventPublisher interface {
    Publish(ctx context.Context, topic string, event interface{}) error
}
```
---

### File: `internal/accounts/domain/account.go`

---
```go
package domain

import (
    "fmt"
    "time"
    "github.com/BackendDataPlatform/internal/sharedkernel"
)

// Account: Entity - Tài khoản
type Account struct {
    ID               string
    UserID           string
    Balance          *sharedkernel.Money  // Tổng số dư
    ReservedBalance  *sharedkernel.Money  // Số dư bị giữ (đặt order)
    AvailableBalance *sharedkernel.Money  // Số dư có thể dùng
    Status           string               // ACTIVE, FROZEN, CLOSED
    CreatedAt        time.Time
    UpdatedAt        time.Time
    Version          int
}

// NewAccount: Tạo account mới
func NewAccount(userID string, initialBalance *sharedkernel.Money) (*Account, error) {
    if userID == "" {
        return nil, fmt.Errorf("userID is required")
    }
    if initialBalance == nil {
        return nil, fmt.Errorf("initialBalance is required")
    }
    
    return &Account{
        ID:               fmt.Sprintf("ACC-%s", userID)[:20],
        UserID:           userID,
        Balance:          initialBalance,
        ReservedBalance:  &sharedkernel.Money{0, initialBalance.Currency},
        AvailableBalance: initialBalance,
        Status:           "ACTIVE",
        CreatedAt:        time.Now(),
        UpdatedAt:        time.Now(),
        Version:          1,
    }, nil
}

// Deposit: Nạp tiền vào tài khoản
func (a *Account) Deposit(amount *sharedkernel.Money) error {
    if a.Status != "ACTIVE" {
        return fmt.Errorf("account is not active: %s", a.Status)
    }
    
    newBalance, err := a.Balance.Add(amount)
    if err != nil {
        return fmt.Errorf("failed to add balance: %w", err)
    }
    
    newAvailable, err := a.AvailableBalance.Add(amount)
    if err != nil {
        return fmt.Errorf("failed to update available balance: %w", err)
    }
    
    a.Balance = newBalance
    a.AvailableBalance = newAvailable
    a.UpdatedAt = time.Now()
    a.Version++
    
    return nil
}

// Withdraw: Rút tiền từ tài khoản
func (a *Account) Withdraw(amount *sharedkernel.Money) error {
    if a.Status != "ACTIVE" {
        return fmt.Errorf("account is not active")
    }
    
    // Kiểm tra có đủ tiền không
    if !a.AvailableBalance.IsGreaterThan(amount) && !a.AvailableBalance.IsEqual(amount) {
        return fmt.Errorf("insufficient balance")
    }
    
    newBalance, _ := a.Balance.Subtract(amount)
    newAvailable, _ := a.AvailableBalance.Subtract(amount)
    
    a.Balance = newBalance
    a.AvailableBalance = newAvailable
    a.UpdatedAt = time.Now()
    a.Version++
    
    return nil
}

// Reserve: Giữ tiền (để đặt order)
// Ví dụ: balance = 1000, reserve = 500 → available = 500
func (a *Account) Reserve(amount *sharedkernel.Money) error {
    if !a.AvailableBalance.IsGreaterThan(amount) && !a.AvailableBalance.IsEqual(amount) {
        return fmt.Errorf("insufficient available balance to reserve")
    }
    
    newReserved, _ := a.ReservedBalance.Add(amount)
    newAvailable, _ := a.AvailableBalance.Subtract(amount)
    
    a.ReservedBalance = newReserved
    a.AvailableBalance = newAvailable
    a.UpdatedAt = time.Now()
    a.Version++
    
    return nil
}

// Release: Thả tiền (hủy order)
func (a *Account) Release(amount *sharedkernel.Money) error {
    if !a.ReservedBalance.IsGreaterThan(amount) && !a.ReservedBalance.IsEqual(amount) {
        return fmt.Errorf("insufficient reserved balance to release")
    }
    
    newReserved, _ := a.ReservedBalance.Subtract(amount)
    newAvailable, _ := a.AvailableBalance.Add(amount)
    
    a.ReservedBalance = newReserved
    a.AvailableBalance = newAvailable
    a.UpdatedAt = time.Now()
    a.Version++
    
    return nil
}
```
---

## 3️⃣ Application Layer (Use Cases)

### File: `internal/orders/app/place_order_usecase.go`

---
```go
package app

import (
    "context"
    "fmt"
    "time"
    "github.com/BackendDataPlatform/internal/orders/domain"
    "github.com/BackendDataPlatform/internal/accounts/domain"
    "github.com/BackendDataPlatform/internal/sharedkernel"
    "github.com/BackendDataPlatform/internal/platform/idempotency"
)

// PlaceOrderCommand: Input từ client (HTTP request)
type PlaceOrderCommand struct {
    UserID   string
    Symbol   string  // "AAPL"
    Side     string  // "BUY" hoặc "SELL"
    Price    int64   // 18000 = 180.00
    Quantity float64 // 10
}

// PlaceOrderResult: Output trả về client (HTTP response)
type PlaceOrderResult struct {
    OrderID   string `json:"order_id"`
    Symbol    string `json:"symbol"`
    Side      string `json:"side"`
    Price     int64  `json:"price"`
    Quantity  float64 `json:"quantity"`
    Status    string `json:"status"`
    CreatedAt string `json:"created_at"`
}

// PlaceOrderUseCase: Use case chính - Đặt lệnh mua/bán
type PlaceOrderUseCase struct {
    orderRepo      domain.OrderRepository
    accountRepo    domain.AccountRepository  // từ accounts service
    eventPublisher domain.EventPublisher
    idempotency    idempotency.IdempotencyStore
}

// NewPlaceOrderUseCase: Constructor
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

// Execute: Thực hiện case - Đặt lệnh
func (uc *PlaceOrderUseCase) Execute(
    ctx context.Context,
    cmd *PlaceOrderCommand,
    idempotencyKey string,
) (*PlaceOrderResult, error) {
    // ✅ BƯỚC 1: Kiểm tra idempotency
    // Nếu client gửi 2 request giống nhau, chỉ xử lý 1 lần
    cachedResult, err := uc.idempotency.Get(ctx, idempotencyKey)
    if err == nil && cachedResult != nil {
        return cachedResult.(*PlaceOrderResult), nil
    }
    
    // ✅ BƯỚC 2: Validate input
    if cmd.UserID == "" || cmd.Symbol == "" || cmd.Price <= 0 || cmd.Quantity <= 0 {
        return nil, fmt.Errorf("invalid input: missing required fields")
    }
    
    // ✅ BƯỚC 3: Lấy tài khoản
    account, err := uc.accountRepo.GetByID(ctx, cmd.UserID)
    if err != nil {
        return nil, fmt.Errorf("account not found: %w", err)
    }
    
    // ✅ BƯỚC 4: Tạo Money object
    orderPrice, err := sharedkernel.NewMoney(cmd.Price, "USD")
    if err != nil {
        return nil, fmt.Errorf("invalid price: %w", err)
    }
    
    // ✅ BƯỚC 5: Tạo Symbol object
    symbol, err := sharedkernel.NewSymbol(cmd.Symbol)
    if err != nil {
        return nil, fmt.Errorf("invalid symbol: %w", err)
    }
    
    // ✅ BƯỚC 6: Tính tổng giá (price * quantity)
    totalPrice := orderPrice.Amount * int64(cmd.Quantity)
    totalMoney, _ := sharedkernel.NewMoney(totalPrice, "USD")
    
    // ✅ BƯỚC 7: Kiểm tra có đủ tiền không
    if !account.AvailableBalance.IsGreaterThan(totalMoney) && !account.AvailableBalance.IsEqual(totalMoney) {
        return nil, fmt.Errorf("insufficient balance: need %v, have %v",
            totalMoney, account.AvailableBalance)
    }
    
    // ✅ BƯỚC 8: Tạo order mới (domain logic)
    order, err := domain.NewOrder(cmd.UserID, symbol, cmd.Side, orderPrice, cmd.Quantity)
    if err != nil {
        return nil, fmt.Errorf("failed to create order: %w", err)
    }
    
    // ✅ BƯỚC 9: Validate order
    if err := order.Validate(); err != nil {
        return nil, fmt.Errorf("invalid order: %w", err)
    }
    
    // ✅ BƯỚC 10: Lưu order vào database
    if err := uc.orderRepo.Save(ctx, order); err != nil {
        return nil, fmt.Errorf("failed to save order: %w", err)
    }
    
    // ✅ BƯỚC 11: Reserve balance (giữ tiền)
    if err := account.Reserve(totalMoney); err != nil {
        return nil, fmt.Errorf("failed to reserve balance: %w", err)
    }
    
    // ✅ BƯỚC 12: Lưu account (với balance đã bị giữ)
    if err := uc.accountRepo.Update(ctx, account); err != nil {
        return nil, fmt.Errorf("failed to update account: %w", err)
    }
    
    // ✅ BƯỚC 13: Tạo event (sự kiện) để gửi
    event := map[string]interface{}{
        "order_id": order.ID,
        "user_id": order.UserID,
        "symbol": order.Symbol.String(),
        "side": order.Side,
        "price": order.Price.Amount,
        "quantity": order.Quantity,
        "created_at": order.CreatedAt.Unix(),
    }
    
    // ✅ BƯỚC 14: Gửi event (Kafka)
    if err := uc.eventPublisher.Publish(ctx, "orders.placed", event); err != nil {
        fmt.Printf("Warning: failed to publish event: %v\n", err)
        // Không return error, event sẽ retry
    }
    
    // ✅ BƯỚC 15: Tạo result để trả về
    result := &PlaceOrderResult{
        OrderID:   order.ID,
        Symbol:    order.Symbol.String(),
        Side:      order.Side,
        Price:     order.Price.Amount,
        Quantity:  order.Quantity,
        Status:    string(order.Status),
        CreatedAt: order.CreatedAt.Format(time.RFC3339),
    }
    
    // ✅ BƯỚC 16: Lưu vào cache (idempotency)
    _ = uc.idempotency.Set(ctx, idempotencyKey, result, 24*time.Hour)
    
    return result, nil
}
```
---

## 4️⃣ Adapter Layer (Database)

### File: `internal/orders/adapter/store/postgres/repo.go`

---
```go
package postgres

import (
    "context"
    "database/sql"
    "fmt"
    "sync"
    "github.com/BackendDataPlatform/internal/orders/domain"
)

// PostgresOrderRepository: Implement OrderRepository với PostgreSQL
type PostgresOrderRepository struct {
    db *sql.DB
    mu sync.RWMutex  // Để thread-safe
}

// NewPostgresOrderRepository: Constructor
func NewPostgresOrderRepository(db *sql.DB) *PostgresOrderRepository {
    return &PostgresOrderRepository{
        db: db,
    }
}

// Save: Lưu order vào database
func (r *PostgresOrderRepository) Save(ctx context.Context, order *domain.Order) error {
    query := `
        INSERT INTO orders (id, user_id, symbol, side, price, quantity, status, created_at, updated_at, version)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
        ON CONFLICT (id) DO NOTHING
        RETURNING id
    `
    
    var returnedID string
    err := r.db.QueryRowContext(
        ctx,
        query,
        order.ID,
        order.UserID,
        order.Symbol.String(),
        order.Side,
        order.Price.Amount,
        order.Quantity,
        order.Status,
        order.CreatedAt,
        order.UpdatedAt,
        order.Version,
    ).Scan(&returnedID)
    
    if err != nil {
        return fmt.Errorf("failed to save order: %w", err)
    }
    
    return nil
}

// GetByID: Lấy order từ database
func (r *PostgresOrderRepository) GetByID(ctx context.Context, id string) (*domain.Order, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    query := `
        SELECT id, user_id, symbol, side, price, quantity, status, created_at, updated_at, version
        FROM orders
        WHERE id = $1
    `
    
    var (
        o       domain.Order
        symbolStr string
        statusStr string
    )
    
    err := r.db.QueryRowContext(ctx, query, id).Scan(
        &o.ID,
        &o.UserID,
        &symbolStr,
        &o.Side,
        &o.Price,
        &o.Quantity,
        &statusStr,
        &o.CreatedAt,
        &o.UpdatedAt,
        &o.Version,
    )
    
    if err == sql.ErrNoRows {
        return nil, fmt.Errorf("order not found: %s", id)
    }
    if err != nil {
        return nil, fmt.Errorf("failed to get order: %w", err)
    }
    
    // Chuyển string → object
    symbol, _ := sharedkernel.NewSymbol(symbolStr)
    o.Symbol = symbol
    o.Status = domain.OrderStatus(statusStr)
    
    return &o, nil
}

// GetByUserID: Lấy tất cả order của user
func (r *PostgresOrderRepository) GetByUserID(ctx context.Context, userID string) ([]*domain.Order, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    query := `
        SELECT id, user_id, symbol, side, price, quantity, status, created_at, updated_at, version
        FROM orders
        WHERE user_id = $1
        ORDER BY created_at DESC
    `
    
    rows, err := r.db.QueryContext(ctx, query, userID)
    if err != nil {
        return nil, fmt.Errorf("failed to query orders: %w", err)
    }
    defer rows.Close()
    
    var orders []*domain.Order
    for rows.Next() {
        var (
            o       domain.Order
            symbolStr string
            statusStr string
        )
        
        err := rows.Scan(
            &o.ID,
            &o.UserID,
            &symbolStr,
            &o.Side,
            &o.Price,
            &o.Quantity,
            &statusStr,
            &o.CreatedAt,
            &o.UpdatedAt,
            &o.Version,
        )
        
        if err != nil {
            return nil, fmt.Errorf("failed to scan order: %w", err)
        }
        
        symbol, _ := sharedkernel.NewSymbol(symbolStr)
        o.Symbol = symbol
        o.Status = domain.OrderStatus(statusStr)
        
        orders = append(orders, &o)
    }
    
    if err = rows.Err(); err != nil {
        return nil, fmt.Errorf("row error: %w", err)
    }
    
    return orders, nil
}

// Update: Cập nhật order (với optimistic locking)
func (r *PostgresOrderRepository) Update(ctx context.Context, order *domain.Order) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    query := `
        UPDATE orders
        SET status = $1, updated_at = $2, version = $3
        WHERE id = $4 AND version = $5
        RETURNING id
    `
    
    var returnedID string
    err := r.db.QueryRowContext(
        ctx,
        query,
        order.Status,
        order.UpdatedAt,
        order.Version,
        order.ID,
        order.Version - 1,  // Kiểm tra version cũ
    ).Scan(&returnedID)
    
    if err == sql.ErrNoRows {
        return fmt.Errorf("order not found or version conflict")
    }
    if err != nil {
        return fmt.Errorf("failed to update order: %w", err)
    }
    
    return nil
}

// Delete: Xóa order
func (r *PostgresOrderRepository) Delete(ctx context.Context, id string) error {
    query := "DELETE FROM orders WHERE id = $1"
    
    result, err := r.db.ExecContext(ctx, query, id)
    if err != nil {
        return fmt.Errorf("failed to delete order: %w", err)
    }
    
    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("failed to get rows affected: %w", err)
    }
    
    if rowsAffected == 0 {
        return fmt.Errorf("order not found")
    }
    
    return nil
}
```
---

## 5️⃣ Transport Layer (HTTP)

### File: `internal/orders/transport/http_handler.go`

---
```go
package transport

import (
    "encoding/json"
    "fmt"
    "net/http"
    "github.com/BackendDataPlatform/internal/orders/app"
)

// PlaceOrderRequest: HTTP request từ client
type PlaceOrderRequest struct {
    Symbol   string  `json:"symbol"`    // "AAPL"
    Side     string  `json:"side"`      // "BUY"
    Price    int64   `json:"price"`     // 18000
    Quantity float64 `json:"quantity"`  // 10
}

// PlaceOrderResponse: HTTP response trả về client
type PlaceOrderResponse struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
    Error   string      `json:"error,omitempty"`
}

// PlaceOrderHandler: HTTP handler
type PlaceOrderHandler struct {
    usecase *app.PlaceOrderUseCase
}

// NewPlaceOrderHandler: Constructor
func NewPlaceOrderHandler(usecase *app.PlaceOrderUseCase) *PlaceOrderHandler {
    return &PlaceOrderHandler{usecase: usecase}
}

// ServeHTTP: Được gọi khi POST /orders
// Signature: func(http.ResponseWriter, *http.Request)
func (h *PlaceOrderHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // Chỉ accept POST
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    
    // ✅ BƯỚC 1: Lấy user ID từ JWT middleware
    // Middleware đã put vào context
    userID, ok := r.Context().Value("user_id").(string)
    if !ok || userID == "" {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(PlaceOrderResponse{
            Success: false,
            Error:   "User not authenticated",
        })
        return
    }
    
    // ✅ BƯỚC 2: Lấy Idempotency-Key từ header
    idempotencyKey := r.Header.Get("Idempotency-Key")
    if idempotencyKey == "" {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(PlaceOrderResponse{
            Success: false,
            Error:   "Idempotency-Key header is required",
        })
        return
    }
    
    // ✅ BƯỚC 3: Parse JSON body
    var req PlaceOrderRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(PlaceOrderResponse{
            Success: false,
            Error:   fmt.Sprintf("Invalid JSON: %v", err),
        })
        return
    }
    
    // ✅ BƯỚC 4: Tạo command
    cmd := &app.PlaceOrderCommand{
        UserID:   userID,
        Symbol:   req.Symbol,
        Side:     req.Side,
        Price:    req.Price,
        Quantity: req.Quantity,
    }
    
    // ✅ BƯỚC 5: Gọi use case
    result, err := h.usecase.Execute(r.Context(), cmd, idempotencyKey)
    if err != nil {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(PlaceOrderResponse{
            Success: false,
            Error:   err.Error(),
        })
        return
    }
    
    // ✅ BƯỚC 6: Trả response thành công
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)  // 201
    json.NewEncoder(w).Encode(PlaceOrderResponse{
        Success: true,
        Data:    result,
    })
}
```
---

## 6️⃣ SQL Migrations

### File: `migrations/orders/000001_orders_init.up.sql`

---
```sql
-- Tạo table orders
CREATE TABLE IF NOT EXISTS orders (
    id              VARCHAR(36) PRIMARY KEY,
    user_id         VARCHAR(36) NOT NULL,
    symbol          VARCHAR(10) NOT NULL,
    side            VARCHAR(10) NOT NULL,  -- BUY, SELL
    price           BIGINT NOT NULL,       -- tính theo cent
    quantity        DOUBLE PRECISION NOT NULL,
    status          VARCHAR(20) NOT NULL,  -- PENDING, FILLED, ...
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    version         INTEGER NOT NULL DEFAULT 1,  -- optimistic locking
    
    -- Indexes để tăng tốc độ query
    CONSTRAINT orders_user_symbol_check CHECK (symbol != '')
);

CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_created_at ON orders(created_at DESC);

-- Tạo table outbox (cho event publishing)
CREATE TABLE IF NOT EXISTS outbox (
    id              VARCHAR(36) PRIMARY KEY,
    topic           VARCHAR(100) NOT NULL,
    payload         BYTEA NOT NULL,         -- protobuf bytes
    event_type      VARCHAR(100) NOT NULL,
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    published_at    TIMESTAMP,
    failed_at       TIMESTAMP,
    error           TEXT,
    retries         INTEGER NOT NULL DEFAULT 0,
    
    CONSTRAINT outbox_topic_check CHECK (topic != '')
);

CREATE INDEX idx_outbox_published_at ON outbox(published_at) WHERE published_at IS NULL;
CREATE INDEX idx_outbox_created_at ON outbox(created_at);
```
---

### File: `migrations/accounts/000001_accounts_init.up.sql`

---
```sql
-- Tạo table accounts
CREATE TABLE IF NOT EXISTS accounts (
    id                  VARCHAR(36) PRIMARY KEY,
    user_id             VARCHAR(36) UNIQUE NOT NULL,
    balance             BIGINT NOT NULL DEFAULT 0,            -- tính theo cent
    reserved_balance    BIGINT NOT NULL DEFAULT 0,            -- balance bị giữ
    available_balance   BIGINT NOT NULL DEFAULT 0,            -- balance có thể dùng
    status              VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
    created_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    version             INTEGER NOT NULL DEFAULT 1
);

CREATE INDEX idx_accounts_user_id ON accounts(user_id);
CREATE INDEX idx_accounts_status ON accounts(status);

-- Constraint: available = balance - reserved
-- (MySQL: CHECK (available_balance = balance - reserved_balance))
```
---

## 7️⃣ Protobuf Definitions

### File: `api/proto/orders/v1/orders.proto`

---
```protobuf
syntax = "proto3";

package orders.v1;

import "google/protobuf/timestamp.proto";

// PlaceOrderRequest: Yêu cầu đặt lệnh
message PlaceOrderRequest {
    string symbol = 1;      // AAPL
    string side = 2;        // BUY, SELL
    
    int64 price = 3;        // 18000 = 180.00
    double quantity = 4;    // 10
}

// PlaceOrderResponse: Phản hồi thành công
message PlaceOrderResponse {
    string order_id = 1;
    string symbol = 2;
    string side = 3;
    int64 price = 4;
    double quantity = 5;
    string status = 6;
    google.protobuf.Timestamp created_at = 7;
}

// OrderPlacedEvent: Sự kiện được gửi qua Kafka
message OrderPlacedEvent {
    string order_id = 1;
    string user_id = 2;
    string symbol = 3;
    string side = 4;
    int64 price = 5;
    double quantity = 6;
    google.protobuf.Timestamp created_at = 7;
}

// OrderService: gRPC service
service OrderService {
    rpc PlaceOrder(PlaceOrderRequest) returns (PlaceOrderResponse) {}
}
```
---

## 📝 Cách Compile Protobuf

---
```bash
# Cài đặt tools
go install github.com/bufbuild/buf/cmd/buf@latest

# Tạo buf.gen.yml
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

# Generate Go code
buf generate api/proto
```
---

**Các bước tiếp theo để bạn code:**

1. ✅ Copy các file ở trên vào project
2. ✅ Tạo database migrations
3. ✅ Implement các Repository interface khác
4. ✅ Viết unit tests cho domain logic
5. ✅ Implement gRPC handlers
6. ✅ Setup Wire dependency injection

Happy coding! 🚀
