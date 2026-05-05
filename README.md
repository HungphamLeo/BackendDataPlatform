Chào bạn, với tư cách là một Solution Architect, tôi rất hiểu những khó khăn của các bạn fresher khi tiếp cận một hệ thống Microservices kết hợp Event-Driven Architecture (EDA) thực tế. 

Dưới đây là bản refactor file `README.md` được thiết kế theo tư duy "Top-Down" (từ bức tranh tổng thể đến chi tiết triển khai code). Tài liệu này được tinh chỉnh riêng để các bạn fresher có thể đọc hiểu luồng nghiệp vụ (Binance/XTB, Orders), nắm bắt kiến trúc, và biết chính xác phải gõ lệnh gì để chạy và test.

---

# 🚀 Tấm Bản Đồ Khởi Hành: Trading Platform Backend (Golang)

Chào mừng bạn đến với dự án Hệ thống Giao dịch Tần suất cao (Trading Platform). Nếu bạn là Fresher và cảm thấy ngợp trước Microservices, đừng lo! File README này là kim chỉ nam chi tiết nhất để bạn nắm bắt dự án, hiểu kiến trúc và tự tay chạy được những dòng code đầu tiên.

## 🎯 1. Mục Tiêu Kinh Doanh (Business Objectives)
Hệ thống của chúng ta được xây dựng để giải quyết 3 bài toán lõi:
1. **Market Data & News Ingestion:** Trích xuất dữ liệu thị trường (giá cả, order book) và tin tức từ các sàn giao dịch lớn như **Binance** và **XTB**. Việc lấy dữ liệu diễn ra qua cả REST API (lấy snapshot) và Realtime WebSocket (stream dữ liệu) thông qua **Kafka**.
2. **Core Trading & Account Management:** Xây dựng các function/use-case xử lý logic nghiệp vụ khắt khe: đặt lệnh giao dịch (Orders), khớp lệnh, và quản lý số dư tài khoản (Accounts).
3. **API Gateway:** Cung cấp các Backend API Gateway chuẩn mực để Client (Web/App) gọi lấy dữ liệu và thực hiện giao dịch một cách bảo mật, ổn định.

## 🏗️ 2. Kiến Trúc & Công Nghệ

Dự án áp dụng mô hình **Microservices + Event-Driven Architecture (EDA)**. Thay vì gọi nhau trực tiếp và chờ đợi (gây nghẽn), các services giao tiếp bất đồng bộ thông qua các "Sự kiện" (Events) trên Kafka.



### 2.1. Tech Stack Cốt Lõi
*   **Ngôn ngữ chính:** Golang (Tối ưu cho xử lý đồng thời - concurrency).
*   **Message Broker:** Kafka (Xử lý hàng triệu event realtime từ Binance/XTB).
*   **Database & Cache:** PostgreSQL (Lưu trữ bền vững) & Redis (Cache tốc độ cao, rate-limit).
*   **Core Modules:**
    *   `viper`: Quản lý cấu hình (config) động từ file `.yaml` hoặc biến môi trường.
    *   `zap`: Logging tốc độ cực cao, định dạng JSON chuẩn cho môi trường Production.
    *   `gorm`: Thao tác với Database (PostgreSQL) thân thiện, dễ bảo trì.

### 2.2. Mô Hình 4 Layer (Clean Architecture)
Để code không trở thành "đống mì ý", chúng ta áp dụng mô hình 4 Layer. Hãy nhớ quy tắc: **Layer ngoài có thể gọi layer trong, layer trong không được biết về layer ngoài**[cite: 1].
*   **Domain (Lõi):** Chứa các quy tắc nghiệp vụ (VD: Cách tính tiền, logic của một Order)[cite: 1]. Không được chứa `gorm`, `gin` hay bất kỳ framework nào tại đây[cite: 1].
*   **Application (Use Cases):** Điều phối công việc (VD: Hàm `PlaceOrder` sẽ gọi hàm kiểm tra số dư, sau đó gọi hàm lưu DB)[cite: 1].
*   **Adapter (Hạ tầng):** Nơi giao tiếp với thế giới bên ngoài. Đây là nơi bạn dùng `gorm` để query PostgreSQL, hoặc dùng module Kafka để publish message[cite: 1].
*   **Transport (Giao tiếp):** Nơi tiếp nhận Request từ User (REST API, gRPC, WebSocket) và chuyển thành Command cho Application layer[cite: 1].

---

## 📂 3. Cấu Trúc Microservices & Thư Mục

Hệ thống được chia thành các service độc lập để dễ scale[cite: 1]:

```text
trading-platform/
├── cmd/
│   ├── api_gateway/        # Cổng giao tiếp cho Client gọi API
│   ├── market_ingestor/    # Service kết nối Binance/XTB qua WebSocket/API -> đẩy vào Kafka
│   ├── orders_service/     # Quản lý đặt lệnh, hủy lệnh
│   └── accounts_service/   # Quản lý số dư, nạp/rút
├── internal/
│   ├── pkg/                # Các thư viện dùng chung
│   │   ├── config/         # Setup Viper
│   │   ├── logger/         # Setup Zap
│   │   └── database/       # Setup Gorm connection
│   ├── orders/             # Source code của Order Service
│   │   ├── domain/         # Entity & Interfaces
│   │   ├── app/            # Use cases
│   │   ├── adapter/        # Gorm Repositories, Kafka Publishers
│   │   └── transport/      # Gin/Fiber/HTTP Handlers
├── deploy/                 # Docker-compose, K8s manifests
└── README.md
```

---

## 💻 4. Hướng Dẫn Chi Tiết Cho Fresher (Step-by-Step)

Là Fresher, bạn hãy làm tuần tự theo các bước sau. Đừng bỏ cóc!

### Bước 1: Chuẩn bị môi trường (Prerequisites)
1. Cài đặt **Golang** (Phiên bản $\ge$ 1.21).
2. Cài đặt **Docker & Docker Compose** (Bắt buộc để chạy Kafka, Postgres).
3. Clone project:
```bash
git clone https://github.com/your-repo/trading-platform.git
cd trading-platform
go mod tidy
```

### Bước 2: Khởi động Hạ Tầng (Infrastructure)
Chúng ta sẽ dùng Docker Compose để dựng toàn bộ Postgres, Redis, Zookeeper và Kafka lên máy local[cite: 1].
```bash
# Di chuyển vào thư mục chứa cấu hình docker
cd deploy

# Khởi động ở chế độ background (-d)
docker-compose up -d

# Kiểm tra xem các container đã chạy thành công chưa
docker-compose ps
```

### Bước 3: Cấu hình hệ thống (Viper)
Hệ thống sử dụng `viper` để đọc file `config.yaml`. Đảm bảo bạn đã copy file mẫu:
```bash
cp config.example.yaml config.yaml
```
*Mở file `config.yaml` và kiểm tra thông tin chuỗi kết nối Database, Kafka Broker (thường là `localhost:9092`), và API Keys của Binance/XTB.*

### Bước 4: Chạy Database Migrations (GORM)
Dự án sử dụng tính năng AutoMigrate của `gorm` (hoặc các tool migration tương đương) để tạo bảng[cite: 1].
Khi các service khởi động, GORM sẽ tự động đồng bộ struct model trong code thành table trong PostgreSQL.

### Bước 5: Chạy thử các Microservices
Mở các terminal khác nhau để chạy từng service:

**Terminal 1: Khởi động API Gateway**
```bash
go run cmd/api_gateway/main.go
```

**Terminal 2: Khởi động Market Ingestor (Kết nối Binance/XTB)**
```bash
go run cmd/market_ingestor/main.go
# Bạn sẽ thấy Log (Zap) in ra: "Connected to Binance WebSocket..."
```

**Terminal 3: Khởi động Orders Service**
```bash
go run cmd/orders_service/main.go
```

---

## 🧪 5. Các Lệnh Code & Test Thường Dùng

Là developer, bạn sẽ cần test code liên tục. Đây là bộ bí kíp lệnh dành cho bạn:

### 5.1. Chạy Unit Test & Xem Độ Phủ (Coverage)
Chạy toàn bộ test trong dự án và kiểm tra xem có phát hiện lỗi nào không:
```bash
# Chạy tất cả các bài test
go test -v ./...

# Chạy test và xuất báo cáo độ phủ code (Coverage)
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out  # Mở báo cáo trên trình duyệt
```

### 5.2. Test API Bằng cURL (Hoặc Postman)
Sau khi `api_gateway` chạy ở port `8080`, hãy thử đặt một lệnh mua[cite: 1]:
```bash
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: req-12345" \
  -d '{
    "symbol": "BTCUSDT",
    "side": "BUY",
    "price": 65000,
    "quantity": 0.1
  }'
```
*Lưu ý: Header `Idempotency-Key` dùng để chống trùng lặp request. Nếu bạn gửi 2 lần cùng một Key, hệ thống (thông qua Redis) sẽ chỉ xử lý 1 lần và trả lại kết quả cũ[cite: 1].*

### 5.3. Xem Log Của Hệ Thống (Zap)
Vì chúng ta dùng `zap`, log in ra sẽ có cấu trúc JSON cực kỳ chi tiết, giúp bạn dễ dàng debug.
Ví dụ một dòng log chuẩn:
```json
{"level":"info","ts":"2026-05-04T21:00:00Z","caller":"market_ingestor/binance.go:45","msg":"Received tick","symbol":"BTCUSDT","price":65050.5}
```

---

## 🧠 6. Bí Kíp Dành Cho Fresher Của Dự Án Này

Để không bị bỡ ngỡ, hãy ghi nhớ 3 pattern xương sống sau:

1. **Outbox Pattern (Chống mất data):** Khi user đặt lệnh, hệ thống lưu Order vào Database (bằng Gorm), ĐỒNG THỜI lưu một "Sự kiện" (Event) vào bảng Outbox trong cùng 1 Transaction[cite: 1]. Sẽ có một worker riêng đọc bảng Outbox và đẩy lên Kafka. Điều này đảm bảo dù ứng dụng có crash thì sự kiện vẫn không bị mất[cite: 1].
2. **Idempotency (Chống click đúp):** Mọi API thay đổi trạng thái (như POST, PUT) đều cần middleware kiểm tra Redis xem `Idempotency-Key` đã được xử lý chưa trước khi gọi xuống Application layer[cite: 1].
3. **Làm việc với GORM ở Adapter Layer:** Bạn tuyệt đối **không** được import `gorm.DB` vào các file ở thư mục `domain` hay `app`. Hãy định nghĩa một Interface (ví dụ: `OrderRepository`) ở Domain, và dùng Gorm để implement interface đó tại thư mục `adapter`[cite: 1].

---
*Chúc bạn code vui vẻ! Đừng ngại đọc mã nguồn của các file `_test.go` để hiểu cách các functions hoạt động nhé. Thời gian + kiên trì + practice = thành công[cite: 1]. 🚀*