# 🚀 Master Handoff & Development Plan: PingGopher

> **Context Transition Document**: Use this document to initialize the new context window for building **PingGopher** — a modern, production-grade Uptime & Synthetic Monitoring SaaS powered by **`gopher-queue`**.

---

## 📌 1. Foundational Background: `gopher-queue` Status

The underlying background queue engine **`gopher-queue`** has been fully engineered, tested, and published to GitHub.

* **GitHub Repository**: [`https://github.com/SirZeck/gopher-queue.git`](https://github.com/SirZeck/gopher-queue.git)
* **Release Version**: `v1.0.0`
* **License**: MIT (`⚖️ MIT`)
* **CI/CD**: GitHub Actions Build & Test (`.github/workflows/ci.yml`)

### Key Architecture of `gopher-queue`:
* **Atomic Leasing**: Server-side Redis Lua scripts (`$O(1)$` atomic lease claims, preventing double execution).
* **Multi-Level Priority Queues**: `ready:1` (High), `ready:2` (Medium), `ready:3` (Low).
* **Delayed & Scheduled Jobs**: Redis Sorted Sets (`ZSET`) with background promoter loops.
* **Watchdog Failover**: Automatic supervisor recovery for crashed worker leases.
* **Dead-Letter Queue (DLQ)**: Exponential backoff retries and manual re-driving.
* **Observability & UI**: Built-in REST API (`gopher-server`), administrative CLI (`gopherctl`), Prometheus `/metrics`, and an embedded real-time Web Dashboard (`/dashboard`) with Server-Sent Events (SSE) streaming (`/v1/stream`).

---

## 🏛️ 2. New Product Vision: `ping-gopher`

**`ping-gopher`** is a standalone, full-stack SaaS application that uses `gopher-queue` as its core task execution engine to provide uptime, synthetic user flow, and SSL certificate monitoring.

### 🌟 Key Differentiators & Features:
1. **External Synthetic Uptime Monitoring**: Measures HTTP status, response latency (ms), and SSL certificate expiration from an external perspective.
2. **Synthetic User Flow Execution**: Simulates critical user transactions (e.g. `Login -> Add to Cart -> Checkout`) using headless workers.
3. **Smart Alert Escalation (Anti-Alert Fatigue)**: Uses `gopher-queue` retry windows to confirm outages before triggering emergency email/webhook notifications.
4. **Hosted Public Status Pages**: Customizable public status pages (e.g., `/status/[slug]`) for transparent incident reporting.

---

## 🏗️ 3. Software Architecture & Strategy

* **Pattern**: **Modular Monolith ("Deploy Any Role" Strategy)**
  - Single codebase organized into decoupled domain packages (`internal/auth`, `internal/monitor`, `internal/worker`, `internal/notifier`).
  - **Single Binary / Multi-Role Deployment**:
    - API Role: `ping-gopher --role=api` (Handles user traffic, REST API, dashboard).
    - Worker Role: `ping-gopher --role=worker` (Runs `gopher-queue` ping & SSL background workers).
* **Technology Stack**:
  - **Language**: Go 1.22+
  - **Queue & Broker**: `gopher-queue` + Redis 7.0+
  - **Database**: SQLite (Local Dev) / PostgreSQL (Production)
  - **Frontend**: Single-Page App / Server-Rendered HTML Dashboard

---

## 📁 4. Project File & Directory Structure

```text
ping-gopher/
├── cmd/
│   └── pinggopher/        # Main entrypoint (API & Worker role launcher)
├── internal/
│   ├── api/               # HTTP REST API server, router & CORS
│   ├── auth/              # User registration, authentication, JWT tokens
│   ├── monitor/           # Target URLs, SSL inspection, latency statistics
│   ├── worker/            # gopher-queue task handlers (ping:check, ssl:check, alert:send)
│   ├── notifier/          # Email (SMTP/Resend), Webhooks, Slack alerts
│   └── db/                # Database models, connection, and migrations
├── web/                   # Frontend Web Application & Dashboard assets
├── go.mod
├── Makefile
└── README.md
```

---

## 🗄️ 5. Database Schema & Data Models

### **`users`**
- `id` (UUID, Primary Key)
- `email` (String, Unique)
- `password_hash` (String)
- `created_at`, `updated_at` (Timestamp)

### **`monitors`**
- `id` (UUID, Primary Key)
- `user_id` (UUID, Foreign Key)
- `name` (String, e.g. "Production API")
- `url` (String, e.g. "https://api.example.com")
- `check_interval_seconds` (Integer, default 60)
- `status` (Enum: `UP`, `DOWN`, `DEGRADED`, `PAUSED`)
- `ssl_expiration_date` (Timestamp)
- `created_at`, `updated_at` (Timestamp)

### **`ping_logs`**
- `id` (UUID, Primary Key)
- `monitor_id` (UUID, Foreign Key)
- `status_code` (Integer, e.g. 200)
- `response_time_ms` (Integer)
- `error_message` (Text, Nullable)
- `created_at` (Timestamp)

### **`incidents`**
- `id` (UUID, Primary Key)
- `monitor_id` (UUID, Foreign Key)
- `started_at` (Timestamp)
- `resolved_at` (Timestamp, Nullable)
- `cause` (Text)
- `status` (Enum: `OPEN`, `INVESTIGATING`, `RESOLVED`)

---

## 📅 6. Master Execution Roadmap

* **Phase 1: Project Setup & Database Foundations**
  - Initialize `ping-gopher` repository & `go.mod`.
  - Connect `github.com/SirZeck/gopher-queue` as a dependency.
  - Implement SQLite/Postgres database layer and migrations for `users`, `monitors`, `ping_logs`, `incidents`.

* **Phase 2: Core Ping & Worker Engine**
  - Implement `check:http_uptime` and `check:ssl_cert` worker handlers in `internal/worker`.
  - Connect `gopher-queue` client & background scheduler.

* **Phase 3: REST API & Authentication**
  - Implement JWT authentication endpoints (`POST /v1/auth/signup`, `POST /v1/auth/login`).
  - Implement CRUD endpoints for monitors (`POST /v1/monitors`, `GET /v1/monitors`, `GET /v1/monitors/:id/logs`).

* **Phase 4: Web Application Dashboard**
  - Build real-time user dashboard displaying live SLA charts (99.9%), response latency graphs, and monitor controls.

* **Phase 5: Alerting & Public Status Pages**
  - Integrate email/webhook notifications on incident creation.
  - Build hosted public status page renderer (`GET /status/:slug`).
