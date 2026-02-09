# 📊 Hình Ảnh & Biểu Đồ Luồng Chạy Code

Dành cho những bạn thích hình ảnh và biểu đồ hơn là văn bản!

---

## 1️⃣ Kiến Trúc Toàn Bộ Project

```
┌─────────────────────────────────────────────────────────────────┐
│                        🌐 CLIENTS                               │
│              (Web Browser, Mobile App, API Tools)               │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             │ HTTP/gRPC
                             ↓
┌─────────────────────────────────────────────────────────────────┐
│                    🏰 API GATEWAY (Cổng Vào)                    │
│           - Kiểm tra: User đã đăng nhập chưa? (Auth)           │
│           - Kiểm tra: Không bị spam? (Rate-limit)              │
│           - Định tuyến: Gửi tới service nào                    │
│           - Ghi log: Ai gửi request gì?                         │
└───┬──────────────┬──────────────┬──────────────┬─────────────────┘
    │              │              │              │
    │ gRPC         │ gRPC         │ gRPC         │ WebSocket
    ↓              ↓              ↓              ↓
┌────────────┐ ┌────────────┐ ┌────────────┐ ┌──────────────────┐
│  ORDERS    │ │ ACCOUNTS   │ │ MARKET     │ │ REALTIME         │
│  SERVICE   │ │ SERVICE    │ │ DATA       │ │ GATEWAY          │
│            │ │            │ │ SERVICE    │ │ (WebSocket)      │
│ (Đặt lệnh) │ │ (Quản lý   │ │ (Giá cổ    │ │ (Update real-    │
│            │ │  tài khoản)│ │  phiếu)    │ │  time nghe       │
└─────┬──────┘ └─────┬──────┘ └─────┬──────┘ │  thị trường)     │
      │              │              │        └────┬───────────────┘
      │   Postgresql │ Postgresql   │             │
      ▼              ▼              ▼             ▼
┌─────────────────────────────────────────────────────────────────┐
│                   🗄️  DATABASE (PostgreSQL)                     │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐                │
│  │ orders     │  │ accounts   │  │ marketdata │                │
│  │ table      │  │ table      │  │ table      │                │
│  └────────────┘  └────────────┘  └────────────┘                │
│  ┌────────────────────────────────────────────┐                │
│  │ outbox table (sự kiện chưa gửi)            │                │
│  │ ┌─────────────────────────────────────────┐│                │
│  │ │ (orders.placed, orders.cancelled, ...)  ││                │
│  │ └─────────────────────────────────────────┘│                │
│  └────────────────────────────────────────────┘                │
└──────┬──────────────────────────────────────────────────────────┘
       │
       │ Outbox Relay (chạy liên tục)
       │ - Lấy message chưa gửi từ outbox
       │ - Gửi lên Kafka
       │ - Đánh dấu là đã gửi
       ↓
┌──────────────────────────────────────────────────────────────────┐
│                      📢 KAFKA (Message Queue)                    │
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────┐             │
│  │orders.placed │ │accounts.     │ │marketdata.   │             │
│  │              │ │credited      │ │tick.received │             │
│  └──────┬───────┘ └──────┬───────┘ └──────┬───────┘             │
│         │                │                 │                     │
└─────────┼────────────────┼─────────────────┼─────────────────────┘
          │                │                 │
          │ Consumers      │ Consumers       │ Consumers
          ↓                ↓                 ↓
    ┌──────────────┐ ┌──────────────┐ ┌──────────────────┐
    │ Stream       │ │ History      │ │ Realtime Gateway │
    │ Processor    │ │ Writer       │ │ (lắng nghe       │
    │ (Tính toán   │ │ (Lưu trữ)    │ │  events →        │
    │  thống kê)   │ │              │ │  broadcast WS)   │
    └──────────────┘ └──────────────┘ └──────────────────┘
          │
          │
          ↓
┌──────────────────────────────────────────────────────────────────┐
│                    💾 REDIS (Cache)                              │
│  - Session user                                                  │
│  - Idempotency keys (chống trùng)                               │
│  - Rate-limit counters                                           │
│  - User presence (ai online)                                     │
└──────────────────────────────────────────────────────────────────┘
```

---

## 2️⃣ Luồng Chạy Chi Tiết: "Đặt Lệnh Mua"

### 🔄 Scenario: User đặt lệnh mua 10 cổ phiếu Apple (AAPL)

```
CLIENT (User nhấn "Mua")
│
├─ POST /api/orders
│  Content-Type: application/json
│  Authorization: Bearer <JWT_TOKEN>
│  Idempotency-Key: "KEY-12345"
│
│  {
│    "symbol": "AAPL",
│    "side": "BUY",
│    "price": 18000,      # 180.00 USD
│    "quantity": 10
│  }
│
└─────────────────────────────────────────────────────────────────

API GATEWAY
├─ 🔐 Middleware: Auth
│   - Kiểm tra: Token hợp lệ không?
│   - ❌ Không → Response 401 Unauthorized
│   - ✅ Có → Extract UserID, put vào context
│
├─ 🚫 Middleware: Rate-limit
│   - Check redis: User này đã gửi bao nhiêu request?
│   - ❌ Quá giới hạn → Response 429 Too Many Requests
│   - ✅ OK → Let through
│
├─ 📋 Middleware: Request ID / Logging
│   - Tạo unique request ID
│   - Log: từ khi request tới lúc response
│
├─ 🎯 Định Tuyến
│   - Route: /api/orders → PlaceOrderHandler
│
└─────────────────────────────────────────────────────────────────

ORDERS SERVICE - TRANSPORT LAYER (HTTP Handler)
├─ 1️⃣ Extract UserID từ context (middleware đã add)
├─ 2️⃣ Validate Idempotency-Key có trong header không
├─ 3️⃣ Parse JSON body
│   - symbol: "AAPL"
│   - side: "BUY"
│   - price: 18000
│   - quantity: 10
│
├─ 4️⃣ Tạo PlaceOrderCommand
│   {
│     UserID: "USER-123",
│     Symbol: "AAPL",
│     Side: "BUY",
│     Price: 18000,
│     Quantity: 10
│   }
│
├─ 5️⃣ Gọi UseCase: uc.Execute(ctx, cmd, "KEY-12345")
│
└─────────────────────────────────────────────────────────────────

ORDERS SERVICE - APPLICATION LAYER (Use Case)
├─ 1️⃣ Check Idempotency
│   - Query Redis: "KEY-12345" → ?
│   - ❌ Không tìm thấy → Tiếp tục
│   - ✅ Tìm thấy → Trả kết quả cũ (không làm gì cả)
│
├─ 2️⃣ Validate Input
│   - symbol != ""? ✅
│   - side = "BUY" hoặc "SELL"? ✅
│   - price > 0? ✅
│   - quantity > 0? ✅
│
├─ 3️⃣ Get Account from Database
│   - Query: "SELECT * FROM accounts WHERE user_id = 'USER-123'"
│   - Account {
│       ID: "ACC-USER-123",
│       Balance: 100000,        # 1000.00 USD
│       ReservedBalance: 30000,  # Đã giữ
│       AvailableBalance: 70000, # Còn lại
│     }
│
├─ 4️⃣ Calculate Total Price
│   - Total = price * quantity = 18000 * 10 = 180000 (1800.00 USD)
│
├─ 5️⃣ Check Balance
│   - AvailableBalance (70000) >= TotalPrice (180000)?
│   - ❌ Không đủ tiền
│   - Return Error: "insufficient balance"
│
│   [Nếu đủ tiền, tiếp tục...]
│
├─ 6️⃣ Create Order (Domain Logic)
│   - Gọi OrderFactory.NewOrder(...)
│   - Tạo object:
│       Order {
│         ID: "ORDER-a1b2c3d4",
│         UserID: "USER-123",
│         Symbol: Symbol("AAPL"),
│         Side: "BUY",
│         Price: Money(18000, "USD"),
│         Quantity: 10,
│         Status: "PENDING",
│         CreatedAt: now(),
│         Version: 1
│       }
│
├─ 7️⃣ Validate Order (Domain Rules)
│   - order.Validate()
│   - Check: ID không rỗng, userID không rỗng, symbol không null, ...
│
├─ 8️⃣ Save Order to Database
│   - INSERT INTO orders (...) VALUES (...)
│   - INSERT INTO outbox (topic='orders.placed', payload=...) VALUES (...)
│   - ⚠️ Cùng 1 transaction! Nếu fail một cái, cả 2 rollback
│
├─ 9️⃣ Reserve Balance
│   - Account.Reserve(Money(180000, "USD"))
│   - Update:
│       ReservedBalance: 30000 → 210000
│       AvailableBalance: 70000 → -110000 (❌ LỖI!)
│
│   [Nếu không đủ tiền, use case đã return error ở bước 5]
│
├─ 🔟 Update Account in Database
│   - UPDATE accounts SET reserved_balance = ..., available_balance = ...
│
├─ 1️⃣1️⃣ Publish Event (Kafka)
│   - eventPublisher.Publish(ctx, "orders.placed", event)
│   - Event này được lấy từ outbox (relay)
│
├─ 1️⃣2️⃣ Save to Idempotency Cache
│   - Redis.Set("KEY-12345", result, 24hours)
│   - Lần sau nếu request giống nhau, return kết quả này
│
├─ 1️⃣3️⃣ Return Result
│   {
│     order_id: "ORDER-a1b2c3d4",
│     symbol: "AAPL",
│     side: "BUY",
│     price: 18000,
│     quantity: 10,
│     status: "PENDING",
│     created_at: "2026-02-07T10:30:00Z"
│   }
│
└─────────────────────────────────────────────────────────────────

ORDERS SERVICE - ADAPTER LAYER (Database)
├─ Save Order (Repository)
│   - String SQL query
│   - Execute thông qua database connection
│   - Check: Order tồn tại chưa? (ON CONFLICT)
│
├─ Save Outbox Message (Platform Outbox)
│   - INSERT INTO outbox (topic, payload, event_type)
│   - Payload là protobuf binary (nhỏ hơn JSON)
│
└─────────────────────────────────────────────────────────────────

DATABASE - PERSISTENCE
├─ 💾 orders table
│   INSERT INTO orders (id, user_id, symbol, side, price, quantity, status, created_at, version)
│   VALUES ('ORDER-a1b2c3d4', 'USER-123', 'AAPL', 'BUY', 18000, 10.0, 'PENDING', now(), 1)
│
├─ 💾 outbox table
│   INSERT INTO outbox (id, topic, payload, event_type, created_at)
│   VALUES ('OUTBOX-xxx', 'orders.placed', [binary protobuf], 'OrderPlacedEvent', now())
│
├─ 💾 accounts table
│   UPDATE accounts 
│   SET reserved_balance = 210000, available_balance = -110000, updated_at = now(), version = 2
│   WHERE id = 'ACC-USER-123'
│   [❌ FAILED: negative available_balance! Use case đã check trước rồi]
│
└─────────────────────────────────────────────────────────────────

OUTBOX RELAY (Service chạy liên tục)
├─ Mỗi giây, query:
│   SELECT * FROM outbox WHERE published_at IS NULL LIMIT 100
│
├─ Lấy được message:
│   {
│     id: 'OUTBOX-xxx',
│     topic: 'orders.placed',
│     payload: [binary protobuf],
│     event_type: 'OrderPlacedEvent'
│   }
│
├─ Gửi lên Kafka:
│   kafkaProducer.SendMessage(topic='orders.placed', message=payload)
│   Message ID: OUTBOX-xxx
│
├─ Nếu thành công:
│   UPDATE outbox SET published_at = now() WHERE id = 'OUTBOX-xxx'
│
├─ Nếu thất bại (retry):
│   UPDATE outbox SET retries = retries + 1, failed_at = now()
│   (Retry lại lần tiếp theo)
│
└─────────────────────────────────────────────────────────────────

KAFKA (Message Broker)
├─ Topic: orders.placed
├─ Message 1: Payload = [binary protobuf of OrderPlacedEvent]
├─ Message 2: ...
├─ Message 3: ...
│
├─ Multiple consumers are listening:
│   - Stream Processor (tính toán thống kê)
│   - Terminal Writer (lưu trữ)
│   - Realtime Gateway (broadcast WebSocket)
│
└─────────────────────────────────────────────────────────────────

STREAM PROCESSOR (Consumer)
├─ Mỗi giây, consume messages từ 'orders.placed'
├─ Dữ liệu nhận được:
│   {order_id: 'ORDER-a1b2c3d4', user_id: 'USER-123', symbol: 'AAPL', ...}
│
├─ Xử lý:
│   - Ghi lại: User này vừa đặt lệnh mua AAPL
│   - Update cache: "AAPL_last_order_time" = now()
│   - Cập nhật metrics: "orders.placed.count" += 1
│
├─ Emit event mới (nếu cần):
│   - Publish: "orders.stats.updated" → Kafka
│
├─ Commit offset:
│   - Đánh dấu: "Đã xử lý message này rồi"
│   - Nếu crash, sẽ retry từ đây, không bị trùng
│
└─────────────────────────────────────────────────────────────────

HISTORY WRITER (Consumer)
├─ Consume: 'orders.placed'
├─ Lưu vào Archive (S3/MinIO):
│   s3://trading-platform/orders/2026-02-07/ORDER-a1b2c3d4.json
│
├─ Delete từ active database sau khi backup
│   DELETE FROM orders WHERE created_at < 30 days ago
│
└─────────────────────────────────────────────────────────────────

REALTIME GATEWAY (Consumer)
├─ Consume: 'orders.placed'
├─ Tìm tất cả WebSocket connections của user này
├─ Broadcast event tới client:
│   {
│     type: "order_placed",
│     data: {order_id, symbol, side, ...}
│   }
│
├─ Client (Browser) nhận được:
│   - Update UI: Hiển thị "Đặt lệnh thành công"
│   - Chart: Hiển thị lệnh mua mới
│   - History: Thêm order vào history list
│
└─────────────────────────────────────────────────────────────────

HTTP RESPONSE (Trở lại Client)
├─ Status Code: 201 Created
├─ Content-Type: application/json
├─ Body:
│   {
│     "success": true,
│     "data": {
│       "order_id": "ORDER-a1b2c3d4",
│       "symbol": "AAPL",
│       "side": "BUY",
│       "price": 18000,
│       "quantity": 10,
│       "status": "PENDING",
│       "created_at": "2026-02-07T10:30:00Z"
│     }
│   }
│
└─────────────────────────────────────────────────────────────────

CLIENT (Browser)
├─ Nhận Response
├─ Show notification: "✅ Lệnh mua thành công!"
├─ Update UI:
│   - Thêm order vào "My Orders" list
│   - Update balance: 1000.00 - 1800.00 = -800.00 (❌ NEGATIVE!)
│   - [Use case đã check này, vì không có tiền nên không đặt được]
│
└─────────────────────────────────────────────────────────────────

[Thực tế, order không tạo được vì không đủ tiền]
[Response sẽ là]
└─ Status Code: 400 Bad Request
   Body:
   {
     "success": false,
     "error": "insufficient balance: need 1800.00 USD, have 700.00 USD"
   }
```

---

## 3️⃣ Kiến Trúc Từng Layer

### Layer 1: Domain

```
┌────────────────────────────────────────────────────┐
│          DOMAIN LAYER (Lõi Kinh Doanh)             │
├────────────────────────────────────────────────────┤
│                                                    │
│  📊 Entity                                         │
│  ├─ Order {                                        │
│  │   ID, UserID, Symbol, Side, Price, Status      │
│  │   Methods: Validate(), Fill(), Cancel()        │
│  │ }                                               │
│  ├─ Account {                                      │
│  │   ID, UserID, Balance, ReservedBalance         │
│  │   Methods: Deposit(), Withdraw(), Reserve()    │
│  │ }                                               │
│  │                                                │
│  💰 Value Object (Không thay đổi, không ID)      │
│  ├─ Money { Amount, Currency }                    │
│  │  Methods: Add(), Subtract(), IsGreaterThan()  │
│  ├─ Symbol { Code }                               │
│  │  Methods: IsEqual()                            │
│  │                                                │
│  📚 Interface (Abstractions)                       │
│  ├─ OrderRepository                               │
│  │  Methods: Save(), GetByID(), Update()         │
│  ├─ AccountRepository                             │
│  │  Methods: GetByID(), Update()                 │
│  ├─ EventPublisher                                │
│  │  Methods: Publish()                           │
│                                                   │
│  ⚡ Domain Services (Quy tắc logic)               │
│  ├─ Không thể reserve balance > balance           │
│  ├─ Không thể đặt order với giá âm               │
│  ├─ Status chỉ có thể thay đổi trong một cách    │
│                                                   │
│  ❌ Không được import:                            │
│  ├─ Database driver (sql, grpc)                   │
│  ├─ HTTP/JSON                                     │
│  ├─ Cache client (redis)                          │
│  ├─ Framework bất kỳ                              │
│                                                   │
│  ✅ Chỉ có:                                        │
│  ├─ Pure logic                                    │
│  ├─ Interfaces (abstractions)                     │
│  ├─ Errors                                        │
│                                                   │
└────────────────────────────────────────────────────┘
```

### Layer 2: Application

```
┌────────────────────────────────────────────────────┐
│       APPLICATION LAYER (Điều Phối)                │
├────────────────────────────────────────────────────┤
│                                                    │
│  Use Cases (Công việc cụ thể)                     │
│  ├─ PlaceOrderUseCase {                           │
│  │   - Get Account from DB (adapter)              │
│  │   - Validate có tiền không (domain rule)       │
│  │   - Create Order (domain factory)              │
│  │   - Save Order (adapter)                       │
│  │   - Reserve Balance (domain method)            │
│  │   - Publish Event (adapter)                    │
│  │   - Return Result                              │
│  │ }                                               │
│  ├─ CancelOrderUseCase { ... }                    │
│  ├─ DepositUseCase { ... }                        │
│  └─ WithdrawUseCase { ... }                       │
│                                                   │
│  ⚙️ Commands (Input)                              │
│  ├─ PlaceOrderCommand { UserID, Symbol, ... }    │
│  ├─ CancelOrderCommand { UserID, OrderID }       │
│  └─ DepositCommand { UserID, Amount }            │
│                                                   │
│  📤 Results/DTOs (Output)                         │
│  ├─ PlaceOrderResult { OrderID, Status,... }     │
│  ├─ AccountBalance { Available, Reserved }       │
│  └─ OrderList { Orders[] }                       │
│                                                   │
│  📋 Workflow (Saga - giao dịch phức tạp)         │
│  ├─ Step 1: Check Balance                        │
│  ├─ Step 2: Reserve Balance                      │
│  ├─ Step 3: Create Order                         │
│  ├─ Compensate (nếu fail):                       │
│  │   - Release Balance                           │
│  │   - Cancel Order                              │
│                                                   │
│  ✅ Biết về:                                      │
│  ├─ Domain Entities & Rules                      │
│  ├─ Adapter Interfaces                           │
│                                                   │
│  ❌ Không biết về:                               │
│  ├─ HTTP/REST                                    │
│  ├─ Database (chỉ biết interface)               │
│  ├─ Kafka (chỉ biết interface)                  │
│                                                   │
└────────────────────────────────────────────────────┘
```

### Layer 3: Adapter

```
┌────────────────────────────────────────────────────┐
│    ADAPTER / INFRASTRUCTURE LAYER (Công Nghệ)     │
├────────────────────────────────────────────────────┤
│                                                    │
│  📊 Database Adapter                              │
│  ├─ PostgresOrderRepository implements            │
│  │  OrderRepository {                             │
│  │    - Save (INSERT SQL)                         │
│  │    - GetByID (SELECT SQL)                      │
│  │    - Update (UPDATE SQL + optimistic locking)  │
│  │  }                                              │
│  ├─ PostgresAccountRepository { ... }             │
│                                                   │
│  📢 Messaging Adapter                             │
│  ├─ KafkaEventPublisher implements               │
│  │  EventPublisher {                              │
│  │    - Publish (send to Kafka)                   │
│  │  }                                              │
│  ├─ KafkaConsumer {                              │
│  │    - Consume (read from Kafka)                 │
│  │  }                                              │
│                                                   │
│  💾 Cache Adapter                                 │
│  ├─ RedisIdempotencyStore {                      │
│  │    - Get (Redis GET)                          │
│  │    - Set (Redis SET)                          │
│  │  }                                              │
│  ├─ RedisCache { ... }                           │
│                                                   │
│  🔄 Outbox Pattern                               │
│  ├─ OutboxRepository {                           │
│  │    - Save message to outbox table             │
│  │  }                                              │
│  ├─ OutboxRelay {                                │
│  │    - Query unpublished messages               │
│  │    - Send to Kafka                            │
│  │    - Mark as published                        │
│  │  }                                              │
│                                                   │
│  🔗 External Service Clients                      │
│  ├─ AccountServiceClient (gRPC) {                │
│  │    - Call remote Accounts Service             │
│  │  }                                              │
│  ├─ MarketDataClient (gRPC) { ... }              │
│                                                   │
│  🗺️ Mappers (DTO ↔ Domain)                       │
│  ├─ OrderMapper {                                │
│  │    - FromDB(row) → Domain Order               │
│  │    - ToDB(order) → SQL values                 │
│  │    - ToDTO(order) → Response                  │
│  │  }                                              │
│                                                   │
│  ✅ Bẩn, nhưng cần thiết:                        │
│  ├─ SQL queries                                  │
│  ├─ Framework-specific code                      │
│  ├─ Configuration handling                       │
│                                                   │
└────────────────────────────────────────────────────┘
```

### Layer 4: Transport

```
┌────────────────────────────────────────────────────┐
│   TRANSPORT / INTERFACE LAYER (Điểm Vào/Ra)       │
├────────────────────────────────────────────────────┤
│                                                    │
│  🌐 HTTP Handlers                                 │
│  ├─ PlaceOrderHandler {                          │
│  │    - Extract user from JWT (context)          │
│  │    - Parse JSON body                          │
│  │    - Convert to Command                       │
│  │    - Call Use Case                            │
│  │    - Convert Result to JSON                   │
│  │    - Return HTTP Response                     │
│  │  }                                              │
│  ├─ GetOrderHandler { ... }                      │
│  ├─ CancelOrderHandler { ... }                   │
│                                                   │
│  🔌 gRPC Services                                 │
│  ├─ OrderServiceServer {                         │
│  │    - PlaceOrder RPC (gRPC instead HTTP)       │
│  │    - CancelOrder RPC                          │
│  │    - GetOrder RPC                             │
│  │  }                                              │
│                                                   │
│  🔐 Middleware                                    │
│  ├─ Auth Middleware {                            │
│  │    - Extract JWT token                        │
│  │    - Validate signature                       │
│  │    - Extract user info                        │
│  │    - Put in context                           │
│  │  }                                              │
│  ├─ RateLimit Middleware {                       │
│  │    - Check Redis counter                      │
│  │    - Increment if within limit                │
│  │    - Reject if exceeded                       │
│  │  }                                              │
│  ├─ Logging Middleware {                         │
│  │    - Log incoming request                     │
│  │    - Log outgoing response                    │
│  │    - Record execution time                    │
│  │  }                                              │
│  ├─ Recovery Middleware {                        │
│  │    - Catch panic                              │
│  │    - Return 500 error                         │
│  │  }                                              │
│                                                   │
│  📊 Error Handlers                                │
│  ├─ Map domain errors → HTTP status              │
│  │  InsufficientBalance → 400 Bad Request        │
│  │  OrderNotFound → 404 Not Found                │
│  │  Etc...                                       │
│                                                   │
│  ✅ Biết về:                                      │
│  ├─ HTTP Status codes                            │
│  ├─ JSON serialization                           │
│  ├─ gRPC protocol                                │
│  ├─ Request/Response formats                     │
│                                                   │
│  ❌ Không biết về:                               │
│  ├─ Database                                     │
│  ├─ Business logic (chỉ gọi use case)           │
│                                                   │
└────────────────────────────────────────────────────┘
```

---

## 4️⃣ Request Flow - Visual

```
HTTP Request đến
        │
        ↓
═════════════════════════════════════ TRANSPORT LAYER
│ PlaceOrderHandler.ServeHTTP()
│ ├─ Parse JSON
│ ├─ Extract user from context
│ └─ Create Command
│
├─ uc.Execute(ctx, cmd, idempotency)
│        │
│        ↓
═════════════════════════════════════ APPLICATION LAYER
│ PlaceOrderUseCase.Execute()
│ ├─ Check idempotency (Cache)
│ ├─ Validate input
│ ├─ Get account (Adapter call)
│ │   ├─ Check balance (Domain logic)
│ │
│ ├─ Create Order (Domain logic)
│ ├─ Validate Order (Domain logic)
│ ├─ Save Order (Adapter call)
│ ├─ Reserve Balance (Domain logic)
│ ├─ Update Account (Adapter call)
│ ├─ Publish Event (Adapter call)
│ │   └─ Save to Outbox (Database)
│ │
│ └─ Return Result
│        │
│        ↓
═════════════════════════════════════ DOMAIN LAYER
│ Account.Reserve()
│ Order.Validate()
│ Money.IsGreaterThan()
│
├─ Query/Command to Database
│        │
│        ↓
═════════════════════════════════════ ADAPTER LAYER
│ PostgresAccountRepository.GetByID()
│ ├─ Execute SQL: SELECT * FROM accounts WHERE id = ?
│ ├─ Map row to Account object
│ └─ Return Account
│
│ PostgresOrderRepository.Save()
│ ├─ Execute SQL: INSERT INTO orders (...)
│ └─ Save to Outbox
│
│ KafkaEventPublisher.Publish()
│ ├─ Insert into outbox table
│ └─ Relay will pick up later
│        │
│        ↓
═════════════════════════════════════ DATABASE
        │
    PostgreSQL
    ├─ orders table
    ├─ accounts table
    └─ outbox table
        │
        ↓
═════════════════════════════════════ MESSAGE QUEUE
        │
    Kafka (via Relay)
    ├─ orders.placed topic
        │
        ↓
════════════════════════════════════ CONSUMERS
        │
    ├─ Stream Processor
    ├─ History Writer
    └─ Realtime Gateway
        │
        ↓
Return HTTP Response ← PlaceOrderHandler
        │
Client receives 201 Created ✅
```

---

## 5️⃣ Dependencies Flow (Dependency Injection)

```
Using WIRE Framework:

wire.Build(
    ├─ postgres.NewDB (gọi)
    │   ↓ return *sql.DB
    │
    ├─ postgres.NewOrderRepository (cần *sql.DB)
    │   ↓ return OrderRepository
    │
    ├─ app.NewPlaceOrderUseCase (cần OrderRepository, ...)
    │   ↓ return PlaceOrderUseCase
    │
    └─ transport.NewPlaceOrderHandler (cần PlaceOrderUseCase)
        ↓ return PlaceOrderHandler

Sau khi Wire.Build():
PlaceOrderHandler
  └─ PlaceOrderUseCase
      ├─ OrderRepository (PostgresOrderRepository)
      │   └─ *sql.DB (database connection)
      ├─ AccountRepository
      │   └─ *sql.DB
      ├─ EventPublisher
      │   └─ Kafka client
      └─ IdempotencyStore
          └─ Redis client

Lúc chạy, mỗi layer có thể gọi layer dưới:
Transport → calls Application
Application → calls Domain + Adapter (Repository)
Domain → pure logic (không call gì)
Adapter → calls external systems (DB, Kafka, Redis)
```

---

## 6️⃣ Error Handling Flow

```
PlaceOrderUseCase.Execute()
    │
    ├─ idempotency.Get(...) → ERROR
    │   └─ Log & continue
    │
    ├─ Validate input → ERROR: "missing required fields"
    │   └─ Return error immediately to handler
    │
    ├─ accountRepo.GetByID(...) → ERROR: "account not found"
    │   └─ Wrap with fmt.Errorf() & return
    │
    ├─ account.Reserve(...) → ERROR: "insufficient balance"
    │   └─ Return domain error
    │
    ├─ orderRepo.Save(...) → ERROR: "database error"
    │   └─ Wrap & return
    │
    └─ eventPublisher.Publish(...) → ERROR
        └─ Log warning (don't fail, will retry)

Handler catches error:
    │
    ├─ if err != nil {
    │   ├─ HTTP Status 400 Bad Request
    │   ├─ JSON: {"success": false, "error": "..."}
    │   └─ Return to client
    │ }

Client receives error response ❌
```

---

## 7️⃣ Data Flow Through Layers

```
Client Input:
{
  "symbol": "AAPL",
  "side": "BUY",
  "price": 18000,
  "quantity": 10
}
        │ JSON
        ↓
TRANSPORT: Parse to PlaceOrderRequest
        │
        ↓
APPLICATION: Convert to PlaceOrderCommand
        │
        ↓
DOMAIN: Create objects
        ├─ Symbol("AAPL")
        ├─ Money(18000, "USD")
        ├─ Order(userID, symbol, side, price, quantity)
        └─ Account (from DB)
        │
        ↓
DOMAIN LOGIC: Validate & Calculate
        ├─ Check Money.IsGreaterThan()
        ├─ Check Order.Validate()
        └─ Calculate: totalPrice = price * quantity
        │
        ↓
ADAPTER: Save to Database
        ├─ String SQL: "INSERT INTO orders WHERE ..."
        ├─ ByteArray Protobuf: "INSERT INTO outbox ..."
        └─ Update: "UPDATE accounts WHERE ..."
        │
        ↓
DATABASE: Persist
        ├─ orders table: [id, user_id, symbol, ...]
        ├─ accounts table: [id, user_id, balance, ...]
        └─ outbox table: [id, topic, payload, ...]
        │
        ↓
ADAPTER: Prepare Response
        ├─ Convert Order to PlaceOrderResult
        └─ JSON: {"order_id": "ORDER-xxx", ...}
        │
        ↓
TRANSPORT: Send HTTP Response
        │
        ↓
Client Output:
{
  "success": true,
  "data": {
    "order_id": "ORDER-a1b2c3d4",
    ...
  }
}
```

---

## 8️⃣ Testing Layers Independently

```
Domain Testing (Pure Unit Tests)
├─ No database
├─ No HTTP
├─ No dependencies
├─ Test Order.Validate(), Money.Add(), etc.
└─ Fast & Independent ✅

Application Testing (With Mocks)
├─ Mock all adapters (Repository, EventPublisher)
├─ Test use case logic
├─ Test workflows & compensations
└─ No real database ✅

Integration Testing (With Testcontainers)
├─ Real PostgreSQL (in Docker)
├─ Real Kafka (in Docker)
├─ Test adapter + application together
├─ Test outbox relay
└─ Slow but comprehensive ✅

Transport Testing (HTTP/gRPC)
├─ Mock use cases
├─ Test request parsing
├─ Test response formatting
├─ Test middleware (auth, rate-limit)
└─ Fast ✅

E2E Testing (Full Stack)
├─ All services running
├─ Real dependencies
├─ Test complete workflows
└─ Very slow ⚠️
```

---

## ✅ Tóm Tắt

```
🎯 Main Concepts:

1️⃣ 4 LAYERS:
   Domain (Pure logic)
   ↓
   Application (Orchestration)
   ↓
   Adapter (External systems)
   ↓
   Transport (HTTP/gRPC entry points)

2️⃣ DEPENDENCY:
   Transport calls Application
   Application calls Domain + Adapter
   Domain never calls anything
   Adapter calls external systems

3️⃣ DATA FLOW:
   Client → Transport → Application → Domain + Adapter → Database → Kafka → Consumers

4️⃣ TESTING:
   Domain: Pure unit tests
   Application: With mocks
   Adapter: With Testcontainers
   Transport: With mocks

5️⃣ PATTERNS:
   Repository (adapter for DB)
   Outbox (prevent data loss)
   Idempotency (prevent duplicates)
   Saga (complex transactions)
   Event Publishing (async communication)
   Value Objects (Money, Symbol)
   Factory (NewOrder)
```
