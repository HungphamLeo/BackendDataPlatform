# Tổng Thiết Kế Kiến Trúc Tổng Thể — Microservices + DDD cho Trading Platform (Go)

> Phiên bản: 1.0 — Ngày: 2025-12-30  
> Phạm vi: Kiến trúc tổng thể, thư mục mã nguồn, lựa chọn công nghệ, nguyên tắc SOLID/Design Patterns, vận hành & quan sát.  
> Giả định: Bounded contexts gồm Market Data, Orders, Accounts, Realtime Gateway, API Gateway. Môi trường: Cloud + Kubernetes.

---

Mục lục
1. Mục tiêu kiến trúc  
2. Nguyên tắc thiết kế (DDD + SOLID)  
3. Giao tiếp & dữ liệu  
4. Lựa chọn công nghệ  
5. Cây thư mục (Monorepo, multi-service)  
6. Dòng dữ liệu & giao tiếp (sync & async)  
7. Mô hình dịch vụ & boundaries  
8. SOLID & Design Patterns theo layer  
9. Observability & SLOs  
10. Bảo mật & tuân thủ  
11. Triển khai & mở rộng  
12. Kiểm thử & chất lượng  
13. Migrations  
14. Roadmap triển khai  
15. Ví dụ mã (rút gọn)  
16. Quy tắc “rào chắn” (Architecture Guardrails)  
17. Phụ lục: Dashboards & Alerts

---

## 1) Mục tiêu kiến trúc
- Tách domain theo bounded context, mỗi service có database riêng.  
- Hỗ trợ strong consistency cho giao dịch nội bộ và eventual consistency qua events.  
- Evolve an toàn: contract-first (protobuf), ADR, event versioning.  
- Production-grade observability: logs, metrics, traces, SLOs, alerts.  
- Thiết kế theo SOLID & Clean Architecture (ports/adapters, domain-first).

---

## 2) Nguyên tắc thiết kế (DDD + SOLID + Clean boundaries)
- Layering:
  - Domain: entities, value objects, domain services — không phụ thuộc framework.  
  - Application: use cases, workflows/sagas nhẹ, orchestration qua ports.  
  - Adapters (Infrastructure): DB, Kafka, Redis, clients.  
  - Transport/Interface: gRPC, HTTP, WebSocket handlers; mapper DTO ↔ Domain.
- Shared Kernel: chỉ value objects chung (Money, Symbol, IDs, TimeRange), lỗi chung — không chia sẻ entity.
- Contracts-first: protobuf (api/proto) và codegen; mapping ở adapter/transport.
- SOLID:
  - SRP: mỗi use case/handler một trách nhiệm.
  - OCP: mở rộng bằng interfaces, strategy, composition.
  - LSP/ISP/DIP: interfaces nhỏ, semantics rõ; application phụ thuộc vào ports.

---

## 3) Kiến trúc giao tiếp & dữ liệu
- Synchronous: gRPC cho intra-service calls (business-critical checks, queries).
- Asynchronous: Kafka cho domain events (orders.created, marketdata.tick.received, ...).
- Transactional Outbox pattern: ghi entity + outbox trong cùng transaction; relay xuất lên Kafka; DLQ cho thất bại.
- Idempotency: middleware + store (Redis/Postgres) để đảm bảo idempotency cho HTTP/gRPC/consumers.

---

## 4) Lựa chọn công nghệ (tổng quan)
- Ngôn ngữ chính: Go (gRPC, protobuf, concurrency primitives).  
- Contracts: Buf / Protobuf; codegen Go (v1) và có thể Python cho tooling.  
- Messaging: Kafka (producer/consumer, headers metadata).  
- DB: Postgres (dịch vụ chính) + Redis (cache, presence, rate-limit).  
- API: HTTP (chi), gRPC, WebSocket (realtime).  
- Infra: Docker, Kubernetes (kustomize), HPA, ConfigMap/Secrets.  
- Observability: Prometheus, Grafana, OpenTelemetry (OTLP), Jaeger.  
- CI/CD: Makefile + scripts, kustomize overlays; smoke/E2E/load tests.  
- Testing tools: Testcontainers, Pact, k6, Chaos Mesh.

---

## 5) Cây thư mục (Monorepo, multi-service — ví dụ)
```
trading-platform-go/
├── README.md
├── go.mod
├── Makefile
├── docs/
│   ├── architecture/
│   └── runbooks/
├── api/
│   └── proto/
│       ├── common/v1/*.proto
│       ├── marketdata/v1/*.proto
│       ├── orders/v1/*.proto
│       └── events/v1/*.proto
├── cmd/
│   ├── market-ingestor/
│   ├── stream-processor/
│   ├── history-writer/
│   ├── realtime-gateway/
│   └── api-gateway/
├── internal/
│   ├── sharedkernel/
│   ├── platform/
│   │   ├── config/
│   │   ├── observability/
│   │   ├── runtime/
│   │   ├── transport/
│   │   ├── messaging/kafka/
│   │   ├── outbox/
│   │   └── idempotency/
│   ├── marketdata/
│   ├── orders/
│   ├── accounts/
│   └── gateway/
├── deployments/
├── tests/
└── migrations/
```
Ghi chú: Contracts/proto không import trực tiếp vào domain/app; mapping ở adapter/transport.

---

## 6) Dòng dữ liệu & giao tiếp

### 6.1 Sync flow (REST → gRPC)
Client → API Gateway (auth, rate-limit) → OrdersService.PlaceOrder (gRPC)  
Handler → PlaceOrderUseCase → PlaceOrderWorkflow → AccountsService (gRPC) → Repository (create order + write outbox) → trả kết quả

### 6.2 Async flow (Kafka)
Market ingestor ghi tick + outbox → Relay → Kafka topic `marketdata.tick.received`  
Stream Processor consumes → aggregate → publish events → History Writer archives to S3/MinIO  
Realtime Gateway consumes events → broadcast to WebSocket clients

### 6.3 Outbox chuẩn
Outbox schema (ví dụ): outbox(id, topic, payload_bytes, event_type, created_at, published_at, failed_at, error, retries)  
- payload_bytes = protobuf bytes  
- metadata chứa version/type để versioning events

---

## 7) Mô hình dịch vụ & boundaries

| Service | Trách nhiệm | Database | Events |
|---|---:|---|---|
| Market Data | Ingest & aggregate ticks, produce market events | Postgres | marketdata.* |
| Orders | Quản lý lệnh, lifecycle, workflows & compensation | Postgres | orders.* |
| Accounts | Tài khoản, số dư, thanh toán | Postgres | accounts.* |
| API Gateway | BFF, auth, rate-limit, routing | Redis | — |
| Realtime Gateway | WebSocket broadcast & presence | Redis | — |
| Stream Processor | Aggregate / stream transforms | — | — |
| History Writer | Archival → S3/MinIO | S3 | — |

---

## 8) Đảm bảo SOLID & Design Patterns theo layer
- Domain: Factory, Aggregate Root, Value Objects, Domain Services, Optimistic Locking (version).  
- Application: Command objects, Workflow/Steps with Compensate (light saga), Strategy/Template Method.  
- Adapters: Repository pattern, Outbox, Idempotency middleware, Gateway clients, Mapper (DTO↔VO), Anti-Corruption Layer.  
- Transport: Middleware (auth, rate-limit, tracing), gRPC interceptors, error mapping.

---

## 9) Observability & SLOs
- Metrics: request duration, gRPC latency, Kafka lag, outbox backlog, DB connections, business counters (orders.placed).  
- Tracing: OpenTelemetry + Jaeger for distributed traces.  
- Dashboards: Grafana service-overview, kafka-metrics, api-gateway.  
- Alerts (ví dụ): HighErrorRate, HighLatency, OutboxBacklog.  
- SLO examples: API Gateway 99.9% availability / P99 < 200ms; Orders 99.95% / P99 < 100ms.

---

## 10) Bảo mật & tuân thủ
- AuthN: JWT/OIDC; AuthZ theo vai trò.  
- TLS everywhere; HMAC signing cho broker integrations when needed.  
- Secrets: K8s Secrets / External Secrets manager.  
- Logging: avoid PII, structured logs, audit trail.  
- Compliance: retention policies, encryption at rest/in transit.

---

## 11) Triển khai & mở rộng
- Kubernetes: base + overlays (dev/prod), HPA (CPU/QPS/Kafka lag), PodDisruptionBudgets, rolling/blue-green.  
- Scale: phân tách workloads — scale ingestor/processor theo throughput; realtime theo số kết nối; cân nhịp Kafka partitions.  
- Cloud infra as code: Terraform modules (EKS/VPC/RDS/MSK/ElastiCache).  
- Multi-region: read replicas, failover, backups & PITR.

---

## 12) Kiểm thử & chất lượng
- Unit tests: domain/app, pure logic, in-memory implementations.  
- Integration tests: Testcontainers (Postgres + Kafka) để verify outbox & relay.  
- Contract tests: Pact / proto validation.  
- Load tests: k6 (spike/stress).  
- Chaos: Chaos Mesh for resilience validation.  
- CI: lint → unit → integration → smoke/E2E → deploy pipelines.

---

## 13) Migrations
- Migrations versioned per service; avoid cross-service schema coupling.  
- Standard tables across services: outbox, idempotency.  
- Use migration tooling with environment-aware configs (dev/staging/prod).

---

## 14) Roadmap triển khai (8 tuần)
- W1–W2: khởi tạo monorepo, api/proto/events, platform/outbox/idempotency, sharedkernel.  
- W3–W4: Market Data + Outbox; Orders + Workflow + Accounts cơ bản.  
- W5–W6: API Gateway, Realtime Gateway, Observability (Prom/Graf/Tracing).  
- W7–W8: Load & Chaos testing, Security audit, SLOs, Go-live review.

---

## 15) Ví dụ mã Go (rút gọn)

```go
// internal/orders/domain/model.go
package domain

type OrderStatus string

const (
    Pending OrderStatus = "pending"
    Filled  OrderStatus = "filled"
)

type Order struct {
    ID       string
    UserID   string
    Symbol   string
    Side     string
    Price    int64 // Money VO recommended
    Quantity float64
    Status   OrderStatus
    Version  int
}

func (o *Order) Fill() error {
    if o.Status != Pending {
        return fmt.Errorf("can only fill pending order")
    }
    o.Status = Filled
    o.Version++
    return nil
}
```

```go
// internal/platform/outbox/outbox.go
package outbox

type Message struct {
    ID        string
    Topic     string
    Payload   []byte // protobuf bytes
    EventType string
    CreatedAt time.Time
}
```

```go
// internal/platform/workflow/step.go
package workflow

type Step interface {
    Name() string
    Execute(ctx context.Context) error
    Compensate(ctx context.Context) error
}
```

---

## 16) Quy tắc “rào chắn” (Architecture Guardrails)
1. Không import protobuf vào domain/app — mapper ở adapter/transport.  
2. Outbox payload phải là protobuf bytes (tránh giữ đồng thời JSON & protobuf).  
3. Workflow theo use case — không dùng saga framework generic trừ khi tái sử dụng >5 patterns.  
4. Shared kernel tối thiểu — chỉ phép VO & lỗi chung.  
5. Mọi quyết định P0/P1 phải có ADR & review định kỳ.

---

## 17) Phụ lục: Dashboards & Alerts (ví dụ)
Prometheus alert ví dụ:
```yaml
groups:
- name: critical
  rules:
  - alert: OutboxBacklog
    expr: count(outbox_pending) > 1000
    for: 5m
    labels:
      severity: critical
    annotations:
      summary: "Outbox backlog exceeds threshold"
```

Grafana: dashboard templates cho API Gateway, Orders service, Kafka consumer lag, Outbox backlog.

---

Kết luận  
Kiến trúc này ưu tiên clean boundaries, pragmatic contracts-first development, và production readiness (observability, SLOs, resilience). Nó cho phép mở rộng theo domain, phát triển an toàn qua contract & ADR, và vận hành ở quy mô cloud-native.
