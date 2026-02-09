# 🎯 Hỏi Đáp & Lộ Trình Học Tập

**Câu hỏi thường gặp và cách giải quyết.**

---

## ❓ Tôi Là Fresher - Nên Bắt Đầu Từ Đâu?

### 📚 Thứ Tự Đọc Tài Liệu:

1. **HUONG_DAN_CODE_VIET.md** (file này - đọc A-Z, dễ nhất)
   - Giải thích architecture overview
   - Giải thích 4 layer
   - Ví dụ code chi tiết
   - Giải thích từng component

2. **DIAGRAMS_VA_LUONG_CHAY.md** (nếu bạn thích hình ảnh)
   - Visual architecture
   - Flow diagrams
   - Data flow through layers
   - Request/response lifecycle

3. **CODE_EXAMPLES_CHI_TIET.md** (file code để copy-paste)
   - Code mẫu cho từng file
   - SQL migrations
   - Protobuf examples

4. **HUONG_DAN_BAT_DAU_CODE.md** (khi sẵn sàng code)
   - Step-by-step hướng dẫn
   - Docker setup
   - Viết code từng bước
   - Test code

5. **QUICK_REFERENCE.md** (khi bạn đã hiểu + muốn tra cứu)
   - Cheat sheet
   - Common tasks
   - Shortcuts

---

## 🤔 Câu Hỏi Thường Gặp

### Q: Tôi không hiểu "4 Layer". Cái gì là "Domain"?

**A:** Tưởng tượng như một ngôi nhà:

```
┌──────────────────────────────┐
│ Transport Layer (Cửa ra/vào) │ ← Người nhân viên tiếp khách ở cầu
├──────────────────────────────┤
│ Application Layer (Nội bộ)   │ ← Manager điều phối các bộ phận
├──────────────────────────────┤
│ Domain Layer (Quy tắc)       │ ← Luật pháp công ty (không ai biết)
├──────────────────────────────┤
│ Adapter Layer (Công nghệ)    │ ← Xe hơi, máy móc để làm việc
└──────────────────────────────┘
```

- **Domain**: Luật kinh doanh (ví dụ: "không thể rút tiền hơn số dư")
- **Application**: Plan/làm việc (ví dụ: "bước 1 check tiền, 2 rút, 3 lưu")
- **Adapter**: Dùng công nghệ (ví dụ: "dùng database để lưu")
- **Transport**: Giao tiếp với bên ngoài (ví dụ: "nhận HTTP request, trả JSON")

---

### Q: Tại sao phải chia layer? Không viết chung một file được không?

**A:** Được, nhưng:

```
1 File (spaghetti code):
├─ HTTP parsing
├─ Business logic
├─ Database code
├─ Error handling
└─ Validation
→ Việc này khó test, khó hiểu, khó bảo dưỡng

4 Layers:
├─ Transport: Chỉ xử lý HTTP
├─ Application: Chỉ điều phối
├─ Domain: Chỉ business logic
└─ Adapter: Chỉ database
→ Dễ test, dễ hiểu, dễ thay đổi
```

Ví dụ:
- Muốn đổi database PostgreSQL → MongoDB? Chỉ đổi layer Adapter
- Muốn kiểm tra business logic? Test layer Domain mà không cần database
- Muốn add HTTP + gRPC cùng lúc? Thêm layer Transport khác, logic giữ nguyên

---

### Q: Tôi hiểu "Order" là entity gì đó, nhưng "Money" là sao?

**A:** 

**Order** = Cái chính (Entity)
```go
Order {
  ID,            // ORDER-123
  UserID,        // USER-456
  Symbol,        // "AAPL"
  Price,         // 18000 (số)
  Quantity,      // 10 (số)
  Status,        // "PENDING"
}
// Có ID duy nhất, có thể thay đổi (PENDING → FILLED)
```

**Money** = Giá trị (Value Object)
```go
Money {
  Amount:    1000       // 10.00 (vì 1000 cent = 10 dollar)
  Currency:  "USD"
}
// Không có ID duy nhất, không thay đổi được
// 10.00 USD = 10.00 USD (bất kỳ Money nào)
```

**Tại sao tách riêng Money?**
- Không bao giờ có "Money" âm (validate trong constructor)
- Không bao giờ cộng "10 USD" với "20 VND" (check currency)
- Có thể tái sử dụng ở nhiều chỗ (Order, Account, Payment)

---

### Q: Repository là cái gì? Database driver ấy chứ?

**A:** Không!

```go
// ❌ Database Driver (tập trung vào CÔNG NGHỆ)
type Database interface {
    Query(sql string, args ...interface{}) (Rows, error)
    Exec(sql string, args ...interface{}) (Result, error)
}

// ✅ Repository (tập trung vào BUSINESS)
type OrderRepository interface {
    Save(ctx context.Context, order *Order) error
    GetByID(ctx context.Context, id string) (*Order, error)
    GetByUserID(ctx context.Context, userID string) ([]*Order, error)
}
```

**Repository** là abstraction (interface) của "nơi lưu order".
- Có thể implement bằng PostgreSQL, MongoDB, File, Memory, v.v.
- Application không biết database là gì (chỉ gọi interface)
- Nếu đổi database, chỉ đổi implementation, không đổi business logic

---

### Q: Tôi thấy `.Scan()`, `.Exec()`, `.QueryRow()` cái đó là sao?

**A:** Cách giao tiếp với PostgreSQL driver Go:

```go
// Query một row
var name string
err := db.QueryRowContext(
    ctx,
    "SELECT name FROM users WHERE id = $1",
    userID,
).Scan(&name)

// Execute command (INSERT, UPDATE, DELETE)
result, err := db.ExecContext(
    ctx,
    "INSERT INTO users (id, name) VALUES ($1, $2)",
    userID, "John",
)

// Query nhiều rows
rows, err := db.QueryContext(
    ctx,
    "SELECT id, name FROM users WHERE age > $1",
    18,
)
defer rows.Close()

for rows.Next() {
    var id, name string
    err := rows.Scan(&id, &name)
    // use id, name
}
```

---

### Q: Event / Kafka / Outbox là cái gì? Tôi chỉ muốn save database thôi

**A:** Được, nhưng:

```
Scenario 1 (đơn giản):
User click "Đặt lệnh"
  ↓ Save order to database
  ✅ Done!

Scenario 2 (phức tạp - thực tế):
User click "Đặt lệnh"
  ↓ Save order to database
  ↓ Update user balance
  ↓ Notify realtime clients (WebSocket)
  ↓ Send email confirmation
  ↓ Archive to S3
  ↓ Update analytics

Nếu save success nhưng email fail, sao?
→ Email service không biết!
→ Update analytics không biết!

Giải pháp: Event + Kafka
1. Save order + "OrderPlaced" event → Kafka
2. Các service lắng nghe:
   a. Email service: gửi email
   b. Analytics: cập nhật stats
   c. Realtime Gateway: notify clients
   d. Archive: backup data
   
Nếu email fail, order vẫn được save ✅
```

**Outbox Pattern:** Tránh mất event nếu app crash

```
❌ Cách sai:
1. Save order
2. [APP CRASH!]
3. Send event (không bao giờ chạy)
→ Event mất!

✅ Cách đúng:
1. Save order + save event in outbox (1 transaction)
2. [APP CRASH!]
3. Restart
4. Relay: Lấy event từ outbox, gửi Kafka
→ Event không mất!
```

---

### Q: Tôi cần hiểu Protobuf không?

**A:** Không bắt buộc ngay, nhưng nên biết:

```protobuf
// Định nghĩa "hợp đồng" giữa services
message Order {
    string id = 1;
    string user_id = 2;
    int64 price = 3;
}

service OrderService {
    rpc PlaceOrder(PlaceOrderRequest) returns (PlaceOrderResponse);
}
```

**Tại sao?**
- Kích thước nhỏ (tốt cho mạng)
- Type-safe (không bị lỗi kiểu)
- Có thể sinh code Go/Python/Java tự động

**Bạn cần làm gì?**
- Học cú pháp (dễ, giống JSON)
- Tạo `.proto` files
- Chạy `buf generate` để sinh code Go

Không cần đi sâu, làm theo example là được!

---

## 🎓 Lộ Trình Học (Nếu Bạn Có 2 Tuần)

### Tuần 1

**Ngày 1-2:**
- [ ] Đọc HUONG_DAN_CODE_VIET.md (phần 1-3)
- [ ] Xem DIAGRAMS_VA_LUONG_CHAY.md
- [ ] Hiểu: 4 layer, Entity, Value Object, Repository

**Ngày 3-4:**
- [ ] Setup Docker (docker-compose.yml)
- [ ] Tạo database tables (SQL)
- [ ] Test connection

**Ngày 5-7:**
- [ ] Viết Money value object
- [ ] Viết Order entity
- [ ] Test domain logic
- [ ] Viết unit tests

### Tuần 2

**Ngày 1-2:**
- [ ] Viết OrderRepository interface
- [ ] Viết PostgresOrderRepository implementation
- [ ] Test với database

**Ngày 3-4:**
- [ ] Viết PlaceOrderUseCase
- [ ] Test với mock repositories
- [ ] Viết HTTP handler

**Ngày 5-7:**
- [ ] Implement Wire (dependency injection)
- [ ] Setup main.go
- [ ] Test API với curl
- [ ] Write integration tests

---

## 🐛 Debug Tips

### "Không hiểu code chạy cái gì"

**Giải pháp**: Dùng print statements

```go
func (uc *PlaceOrderUseCase) Execute(ctx context.Context, cmd *PlaceOrderCommand) {
    fmt.Printf("🔍 START: cmd=%+v\n", cmd)
    
    account, err := uc.accountRepo.GetByID(ctx, cmd.UserID)
    fmt.Printf("🔍 Account: balance=%d\n", account.AvailableBalance)
    
    totalPrice := cmd.Price * int64(cmd.Quantity)
    fmt.Printf("🔍 Total: %d, Has: %d\n", totalPrice, account.AvailableBalance)
    
    order, _ := domain.NewOrder(...)
    fmt.Printf("🔍 Order created: %+v\n", order)
}
```

Run:

```bash
go run ./cmd/api_gateway 2>&1 | grep 🔍
```

### "Query không return dữ liệu"

**Giải pháp**: EXPLAIN trong SQL

```sql
EXPLAIN ANALYZE SELECT * FROM orders WHERE user_id = 'USER-1';
```

Kiểm tra:
- Có index không?
- Query hợp lệ không?
- Có dữ liệu không?

### "Panic! interface{} value is nil"

**Giải pháp**: Check nil trước

```go
// ❌
var user *User
user.GetName()  // PANIC!

// ✅
var user *User
if user == nil {
    return fmt.Errorf("user is nil")
}
user.GetName()
```

---

## 📞 Resources

### Tài Liệu Chính Thức
- [Go Documentation](https://golang.org/doc/)
- [PostgreSQL Docs](https://www.postgresql.org/docs/)
- [Protobuf Guide](https://developers.google.com/protocol-buffers)

### Hữu Ích Cho Dự Án
- [GORM ORM](https://gorm.io/) - Alternative to raw SQL
- [Kafka Go Client](https://github.com/segmentio/kafka-go)
- [Redis Go Client](https://github.com/redis/go-redis)

### Community
- Go Discord: https://gophers.slack.com/
- Stack Overflow: tag `go`

---

## 💪 Để Trở Thành Một Lập Trình Viên Tốt

```
Viết code → Test code → Đọc code khác → Viết code tốt hơn

Tháng 1: Hiểu syntax
Tháng 3: Biết cách tổ chức code
Tháng 6: Hiểu design patterns
Tháng 12: Viết code "production-ready"
```

**Hàng ngày:**
- [ ] Viết code (2 giờ)
- [ ] Test code (30 phút)
- [ ] Đọc code khác / docs (30 phút)
- [ ] Refactor / improve (30 phút)

**Hàng tuần:**
- [ ] Code review (học từ team)
- [ ] Investigate bug (học debugging)
- [ ] Đọc blog/paper (học best practices)

---

## ✨ Phần Thưởng

Khi bạn hoàn thành project này, bạn sẽ hiểu:

- ✅ Clean Architecture + Domain-Driven Design
- ✅ Dependency Injection + Wire
- ✅ Testing strategies (unit, integration, E2E)
- ✅ Building microservices
- ✅ Event-driven architecture
- ✅ Database design + migrations
- ✅ Protocol Buffers + gRPC
- ✅ distributed systems concepts

**Đó là những skill mà một BE developer cấp senior phải biết!**

---

## 🎉 Bắt Đầu Hôm Nay!

```bash
# Step 1: Clone project
cd /mnt/c/Users/Admin/Downloads/Project/Github/BackendDataPlatform

# Step 2: Setup
docker-compose up -d

# Step 3: Read (in order)
cat HUONG_DAN_CODE_VIET.md
cat DIAGRAMS_VA_LUONG_CHAY.md

# Step 4: Code
cat CODE_EXAMPLES_CHI_TIET.md
# Copy-paste examples → tạo file mới

# Step 5: Test
go test -v ./...

# Step 6: Run
go run ./cmd/api_gateway/main.go

# Step 7: Celebrate 🎉
curl http://localhost:8080/health
```

---

**Good luck, fresher!** 🚀

Nhớ: Mỗi senior developer đã từng là fresher!
