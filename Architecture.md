
---
```go

# Tổng Thiết Kế Kiến Trúc Tổng Thể: Microservices + DDD cho Trading Platform (Go)

> **Phiên bản**: 1.0 — Ngày: 2025-12-30  
> **Phạm vi**: Kiến trúc tổng thể, thư mục mã nguồn, lựa chọn công nghệ, nguyên tắc SOLID/Design Patterns, vận hành & quan sát (Prometheus/Grafana), mở rộng domain & cloud, và chiến lược kiểm thử/triển khai.  
> **Giả định**: Bounded context gồm *Market Data*, *Orders*, *Accounts*, *Realtime Gateway*, *API Gateway*. Môi trường Cloud/Kubernetes.

---

## 1) Mục tiêu kiến trúc

- **Khả năng mở rộng theo domain**: tách theo bounded context, mỗi service có database riêng; cân bằng strong consistency (giao dịch nội bộ) và eventual consistency qua sự kiện.  
- **Evolve an toàn**: hợp đồng (protobuf) & ADR; event versioning; fitness functions để bảo toàn đặc tính kiến trúc trọng yếu.  
- **Vận hành/quan sát production-grade**: logs, metrics, traces; SLO/alerts; dashboards Grafana; readiness/liveness; idempotency/outbox; retry/backoff.  
- **Đáp ứng SOLID**: domain thuần; DIP qua ports/adapters; SRP trong use cases; LSP/ISP trong interfaces; OCP bằng composition & strategy.

---

## 2) Nguyên tắc thiết kế (DDD + SOLID + Clean boundaries)

- **Layering (DDD)**:  
  - **Domain**: entity/value objects, domain services, rules; *không phụ thuộc framework*.  
  - **Application**: use cases (orchestrate), workflow (saga nhẹ); gọi ports; xử lý transaction.  
  - **Adapters (Infrastructure)**: implementations cho ports (DB/Kafka/Redis/gRPC client).  
  - **Transport/Interface**: gRPC/HTTP/WebSocket handlers; mapping DTO ↔ domain.  
- **Shared Kernel (tối thiểu)**: chỉ *value objects* chung (Money, Symbol, IDs, TimeRange) và lỗi chung; *không chia sẻ entity*.  
- **Contracts-first**: Protobuf cho RPC & events; tách *api/proto/* khỏi *internal/*; mapper riêng.  
- **SOLID**:  
  - SRP: mỗi use case/mỗi handler đảm nhiệm 1 trách nhiệm.  
  - OCP: mở rộng chức năng qua interface/strategy/workflow step.  
  - LSP: interfaces có semantics rõ ràng; thay thế implementation không phá vỡ logic.  
  - ISP: ports tách nhỏ (Repository, EventPublisher…).  
  - DIP: Application phụ thuộc **ports** (interfaces) thay vì adapters cụ thể.

---

## 3) Kiến trúc giao tiếp & dữ liệu

- **Synchronous (gRPC)**: xác thực/balance check; queries.  
- **Asynchronous (Kafka)**: domain events (orders.created, marketdata.tick.received…).  
- **Transactional Outbox**: ghi **order** và **outbox** cùng transaction; relay → Kafka; at-least-once; DLQ.  
- **Idempotency**: middleware + store (Redis/Postgres) cho HTTP/gRPC; key theo request.

---

## 4) Lựa chọn công nghệ

- **Ngôn ngữ**: Go (gRPC, protobuf, concurrency primitives).  
- **Contract**: Buf/protobuf; codegen Go/Python.  
- **Messaging**: Kafka (producer/consumer, topics, headers).  
- **DB**: Postgres (dịch vụ chính) + Redis (cache/presence/rate-limit).  
- **API**: HTTP (chi), gRPC, WebSocket.  
- **Infra**: Docker + Kubernetes (kustomize); HPA; secrets/ConfigMap.  
- **Observability**: Prometheus + Grafana dashboards; OpenTelemetry (OTLP) + Jaeger.  
- **CI/CD**: Makefile + scripts; kustomize overlays; smoke/E2E/load tests.  
- **Testing**: Testcontainers, Pact (contract), k6 (load), Chaos Mesh.

---

## 5) Cây thư mục (Monorepo, multi-service)

```
---
trading-platform-go/
├── README.md
├── go.mod
├── Makefile
├── docs/
│   ├── architecture/
│   │   ├── overview.md
│   │   ├── data-flow.md
│   │   └── decision-records/
│   │       ├── 001-monorepo-multi-service.md
│   │       ├── 002-event-schema-protobuf.md
│   │       ├── 003-outbox-pattern.md
│   │       ├── 004-shared-kernel-only.md
│   │       └── 005-workflow-per-usecase.md
│   ├── runbooks/
│   └── api/
│       ├── rest-api.md
│       └── websocket-protocol.md
├── api/
│   └── proto/
│       ├── buf.yaml
│       ├── buf.gen.yaml
│       ├── common/v1/{money.proto,time.proto,ids.proto,error.proto}
│       ├── marketdata/v1/{models.proto,service.proto}
│       ├── orders/v1/{models.proto,service.proto}
│       ├── accounts/v1/{models.proto,service.proto}
│       └── events/v1/{marketdata_events.proto,orders_events.proto,accounts_events.proto}
├── cmd/
│   ├── market-ingestor/{main.go,wire.go,wire_gen.go}
│   ├── stream-processor/{main.go,wire.go,wire_gen.go}
│   ├── history-writer/{main.go,wire.go,wire_gen.go}
│   ├── realtime-gateway/{main.go,wire.go,wire_gen.go}
│   └── api-gateway/{main.go,wire.go,wire_gen.go}
├── internal/
│   ├── sharedkernel/{money.go,symbol.go,ids.go,time.go,errors.go}
│   ├── platform/
│   │   ├── config/{config.go,loader.go,validator.go}
│   │   ├── observability/{logger.go,metrics.go,tracing.go}
│   │   ├── runtime/{app.go,shutdown.go,workerpool.go,backoff.go,retry.go,health.go}
│   │   ├── transport/{grpc/{server.go,client.go,interceptors.go},http/{server.go,router.go,middleware.go}}
│   │   ├── messaging/kafka/{producer.go,consumer.go,headers.go,codec.go,health.go}
│   │   ├── outbox/{outbox.go,store.go,relay.go,publisher.go}
│   │   ├── idempotency/{idempotency.go,store.go,middleware.go}
│   │   └── testing/{fixtures.go,integration.go,testcontainers.go,mocks.go}
│   ├── marketdata/{domain,app,adapter,transport}
│   ├── orders/{domain,app/workflows,adapter/{store,broker,clients},transport/grpc}
│   ├── accounts/{domain,app,adapter/{store,broker,cache},transport/grpc}
│   └── gateway/{contracts,app,adapter/{clients,ratelimit},transport/{http,grpc}}
├── deployments/
│   ├── k8s/{base,infrastructure,services,overlays/{dev,prod},kustomization.yaml}
│   └── observability/{prometheus/{alerts,recording},grafana/{dashboards},jaeger,slo}
├── tests/{unit,integration,contract,load,chaos}
└── migrations/{marketdata,orders,accounts}
---
```

> **Ghi chú**: Contracts/proto **không import** vào domain/app; mapping nằm ở adapter/transport.

---

## 6) Dòng dữ liệu & giao tiếp

### 6.1 Sync flow (REST → gRPC)
- Client → **API Gateway** (auth, rate-limit) → gRPC **OrdersService.PlaceOrder**.  
- Handler → **Application** (`PlaceOrderUseCase`) → **Workflow** (`PlaceOrderWorkflow`) → gRPC **AccountsService** (check/deduct) → **Repository** (create order + write outbox) → trả kết quả.

### 6.2 Async flow (Kafka)
- Ingestor ghi *tick* + outbox; **Relay** → Kafka `marketdata.tick.received`.  
- **Stream Processor** aggregate candles; **History Writer** lưu trữ dài hạn.  
- **Realtime Gateway** consume sự kiện, broadcast WebSocket.

### 6.3 Outbox chuẩn
- Bảng `outbox(id, topic, payload_bytes, event_type, created_at, published_at, failed_at, error, retries)`; `payload_bytes` là **protobuf bytes**; metadata chứa version/type.

---

## 7) Mô hình dịch vụ & boundaries

| Service | Trách nhiệm | Database | Events |
|---|---|---|---|
| Market Data | Ingest/aggregate | Postgres | `marketdata.*` |
| Orders | Quản lý lệnh/compensation | Postgres | `orders.*` |
| Accounts | Tài khoản/số dư/auth | Postgres | `accounts.*` |
| API Gateway | BFF, auth, rate-limit | Redis | None |
| Realtime Gateway | Broadcast WSS/presence | Redis | None |
| Stream Processor | Aggregate/stateless | — | None |
| History Writer | Archival → S3/MinIO | S3 | None |

---

## 8) Đảm bảo SOLID & Design Patterns theo layer

- **Domain**: Factory (`NewOrder`), Aggregate root, Entities/VO; **Policy/Domain Service** cho quy tắc phức tạp; Optimistic locking `version`.  
- **Application**: **Command** object (PlaceOrderCommand), **Workflow** (chuỗi **Steps** có Compensate), **Template Method/Strategy** cho quyết định logic.  
- **Adapters**: **Repository pattern** (ports + impl), **Outbox**, **Idempotency**, **Gateway** clients; **Mapper** cho DTO↔VO; **Anti-corruption layer** giữa dữ liệu ngoài và domain.  
- **Transport**: **Middleware** (auth, rate-limit, tracing); **Interceptor** gRPC; **Error mapping**.

---

## 9) Observability & SLOs

- **Metrics**: HTTP request duration, gRPC latency, Kafka lag, outbox backlog, DB connections, business counters (orders.placed…).  
- **Dashboards Grafana**: *service-overview*, *kafka-metrics*, *api-gateway*; **Jaeger** cho traces xuyên dịch vụ; **Prometheus alerts** (HighErrorRate, HighLatency, OutboxBacklog).  
- **SLO** (ví dụ): API Gateway 99.9% / P99<200ms; Orders 99.95% / P99<100ms; Market Data 99.5% / P99<50ms.

---

## 10) Bảo mật & tuân thủ

- AuthN JWT/OIDC; AuthZ theo vai trò; ký HMAC khi tích hợp broker; TLS; secrets quản lý bởi K8s/External Secrets; logging không lộ PII; audit trail.

---

## 11) Triển khai & mở rộng

- **Kubernetes**: `deployments/k8s` với base + overlays; HPA theo CPU/QPS/Kafka lag; pod disruption budgets; multi-AZ; rolling/blue-green.  
- **Scale**: scale **ingestor/processor** theo throughput; **Orders/Accounts** scale theo QPS; **Realtime** scale theo số kết nối WSS; cân nhịp Kafka partitions.  
- **Cloud**: Terraform module EKS/VPC/RDS/MSK/ElastiCache; multi-region (read-models), failover; backup & PITR.

---

## 12) Kiểm thử & chất lượng

- **Unit**: domain/app thuần, in-memory.  
- **Integration**: Testcontainers (Postgres+Kafka) cho outbox/relay; verify event & DB.  
- **Contract**: Pact/OpenAPI-proto compliance.  
- **Load**: k6 (spike/stress); latency budget; bottleneck analysis.  
- **Chaos**: mạng delay/kill pods; validate resilience.  
- **CI**: lint/test/build; smoke/E2E trước prod.

---

## 13) Migrations (chuẩn hoá)

- Bảng business riêng cho từng service; **outbox/idempotency** table chuẩn; versioned migrations; không cross-schema coupling.

---

## 14) Roadmap triển khai (8 tuần)

- **W1–W2**: khởi tạo monorepo; `api/proto/events` + codegen; `internal/platform/{outbox,idempotency}`; `sharedkernel`.  
- **W3–W4**: Market Data + Outbox; Orders + Workflow + Accounts.  
- **W5–W6**: API/Realtime Gateway; Observability (Prom/Graf); Integration tests.  
- **W7–W8**: Load/Chaos; Security audit; SLO/alerts; Go-live review.

---

## 15) Ví dụ mã Go (rút gọn)

```
---
// internal/orders/domain/model.go
// Aggregate root + optimistic locking

type Order struct {
    id        sharedkernel.OrderID
    userID    sharedkernel.UserID
    symbol    sharedkernel.Symbol
    side      Side
    orderType OrderType
    price     sharedkernel.Money
    quantity  float64
    status    OrderStatus
    version   int
}

func (o *Order) Fill() error {
    if o.status != Pending { return fmt.Errorf("can only fill pending") }
    o.status = Filled; o.version++; return nil
}
---
```

```
---
// internal/platform/outbox/outbox.go
// Store protobuf bytes + metadata

type Message struct {
    ID        string
    Topic     string
    Payload   []byte // protobuf bytes
    EventType string
    CreatedAt time.Time
}
---
```

```
---
// internal/platform/workflow/step.go
// Minimal workflow toolkit để test/trace steps

type Step interface {
    Name() string
    Execute(ctx context.Context) error
    Compensate(ctx context.Context) error
}
---
```

---

## 16) Quy tắc “rào chắn” (Architecture Guardrails)

1. **Không import protobuf vào domain/app** — mapper ở adapter/transport.  
2. **Outbox payload = protobuf bytes** — tránh song song JSON/Protobuf.  
3. **Workflow theo use case** — không framework saga generic trừ khi >5 pattern giống nhau.  
4. **Shared kernel tối thiểu** — chỉ VO; không entity chung.  
5. **ADR & review định kỳ** — mọi quyết định P0/P1 phải có ADR.

---

## 17) Phụ lục: Dashboard & Alerts (ví dụ)

```
---
# prometheus/alerts/critical.yml
- alert: OutboxBacklog
  expr: count(outbox_pending) > 1000
  for: 5m
  labels: { severity: "critical" }
---
```

```
---
// grafana/dashboards/api-gateway.json (lược)
{
  "title": "API Gateway Overview",
  "panels": [ { "type": "graph", "targets": [ { "expr": "histogram_quantile(0.99, sum(rate(http_request_duration_seconds_bucket[5m])) by (le))" } ] } ]
}
---
```

---

**Kết luận**: Kiến trúc này tối ưu hoà *clean boundaries* + *pragmatism*; cho phép scale theo domain, evolve hợp đồng an toàn, và vận hành/giám sát theo chuẩn production. Tuân thủ nguyên tắc SOLID và DDD, triển khai nhanh với Go/k8s, và mở rộng thuận lợi khi thêm domain hoặc lên multi-cloud.

```
---
