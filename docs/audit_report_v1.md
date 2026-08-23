# 🛡️ PingGopher Production-Readiness Audit Report

**Target Repository:** `github.com/SirZeck/ping-gopher`  
**Auditor Role:** Senior Staff Engineer (Pre-Launch Review)  
**Date:** August 23, 2026  

---

## 1. Executive Summary

This production-readiness audit evaluated the `ping-gopher` codebase across Security, Software Engineering Best Practices, Operational Readiness, and Documentation & Maintainability. **The application is currently NOT production-ready** and contains critical vulnerabilities, architectural defects, and operational flaws that will lead to security breaches, data leaks, and service downtime under initial production load.

### Key Findings Summary:
1. **Critical SSRF Vulnerability**: Synthetic HTTP and SSL probes accept any arbitrary target URL without host or IP validation, enabling attackers to probe internal cloud metadata services (`http://169.254.169.254`), loopback, and internal VPC networks.
2. **Critical Multi-Tenant Data Leak**: The public status page endpoint (`GET /v1/status/public`) executes un-scoped queries across all database records, leaking internal monitor names, IDs, statuses, and outage logs of all tenants to unauthenticated users.
3. **Critical Architectural Defect (Non-Functional Worker Pool)**: Claims of being "Powered by `gopher-queue`" with distributed Redis task queuing are incorrect. The scheduler bypasses Redis and spawns in-process goroutines, while dedicated `--role=worker` instances print a log line and immediately exit.
4. **Critical SQLite Concurrent Write Locking (`SQLITE_BUSY`)**: SQLite initialization omits Write-Ahead Logging (`WAL` mode), busy timeouts, and connection pool limits (`SetMaxOpenConns`). Concurrent database writes from API requests and background probes will immediately trigger `database is locked` panics.
5. **Critical Process Lifecycle & Graceful Shutdown Defect**: In API role mode (`--role=api`), `http.Server.ListenAndServe()` runs synchronously on the main thread, bypassing OS signal handlers (`SIGINT`/`SIGTERM`). In `--role=all` mode, HTTP servers and background probe goroutines are terminated abruptly without flushing state.
6. **High Denial of Service (DoS) Vulnerability**: API endpoints process JSON payloads without `http.MaxBytesReader` body size limits or input length constraints, allowing attackers to trigger container Out-Of-Memory (OOM) crashes or CPU exhaustion via oversized password hashing.
7. **High Unbounded Telemetry Query & Database Expansion**: Telemetry logs (`PingLog`) lack partition/retention policies and composite indexes `(monitor_id, created_at DESC)`. `GetMonitorLogsHandler` accepts un-capped `?limit=` parameters, allowing callers to load millions of log rows into RAM.
8. **High Disconnected Alerting Pipeline**: Incident creation and resolution logic in `WorkerEngine` fails to invoke `NotificationEngine`. Webhooks defined in `internal/notifier` are dead code; zero outage alerts are dispatched to users.

---

## 2. Comprehensive Audit Findings Table

| Severity | Category | File:Line | Issue Type | Finding & Root Cause | Concrete Recommendation & Fix |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **CRITICAL** | Security | [monitor_handlers.go:39-42](file:///c:/Users/HP/Projects/ping-gopher/internal/api/monitor_handlers.go#L39-L42)<br>[probe.go:43](file:///c:/Users/HP/Projects/ping-gopher/internal/worker/probe.go#L43)<br>[probe.go:99](file:///c:/Users/HP/Projects/ping-gopher/internal/worker/probe.go#L99) | Confirmed Vulnerability | **Server-Side Request Forgery (SSRF)**: `CreateMonitorHandler` accepts any URL. Probe engines dial loopback (`127.0.0.1`), private RFC 1918 IPs (`10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16`), and AWS metadata (`169.254.169.254`). | Validate URLs using `net/url` and parse resolved IP addresses against a strict blocklist before dialing HTTP or TLS connections. Reject non-HTTP/HTTPS schemes. |
| **CRITICAL** | Security | [status_handlers.go:37-41](file:///c:/Users/HP/Projects/ping-gopher/internal/api/status_handlers.go#L37-L41)<br>[status_handlers.go:70](file:///c:/Users/HP/Projects/ping-gopher/internal/api/status_handlers.go#L70) | Confirmed Vulnerability | **Multi-Tenant Data & Privacy Leak**: Unauthenticated public status endpoint `GET /v1/status/public` queries all monitors and active incidents across *all* tenant accounts without filtering. | Associate public status pages with specific tenant IDs/slugs (`GET /v1/status/:tenant_slug`) and filter DB queries strictly by tenant ID. |
| **CRITICAL** | Architecture | [main.go:90](file:///c:/Users/HP/Projects/ping-gopher/cmd/pinggopher/main.go#L90)<br>[scheduler.go:94-98](file:///c:/Users/HP/Projects/ping-gopher/internal/scheduler/scheduler.go#L94-L98)<br>[worker.go:30](file:///c:/Users/HP/Projects/ping-gopher/internal/worker/worker.go#L30) | Confirmed Defect | **Broken Distributed Architecture & Fake Task Queue**: Scheduler spawns local in-process goroutines (`go ProcessHTTPCheck`) instead of pushing tasks to Redis. `--role=worker` is non-functional and exits on startup. | Implement genuine task producer/consumer logic (e.g. using `gopher-queue` or Redis client) in `scheduler.go` and `startWorkerRole`. |
| **CRITICAL** | Database | [db.go:14-20](file:///c:/Users/HP/Projects/ping-gopher/internal/db/db.go#L14-L20) | Confirmed Defect | **SQLite `SQLITE_BUSY` Write Locking**: SQLite DSN only sets `_pragma=foreign_keys(1)`. Omits `WAL` journal mode, `busy_timeout`, and connection pool limits (`SetMaxOpenConns(1)`), causing database locks during concurrent API writes and probe inserts. | Update DSN to `%s?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)`. Set `sqlDB.SetMaxOpenConns(1)` for SQLite in `InitDB`. |
| **CRITICAL** | Operations | [main.go:44-69](file:///c:/Users/HP/Projects/ping-gopher/cmd/pinggopher/main.go#L44-L69)<br>[main.go:72-87](file:///c:/Users/HP/Projects/ping-gopher/cmd/pinggopher/main.go#L72-L87) | Confirmed Defect | **Broken Process Lifecycle & Graceful Shutdown**: `startAPIRole` runs `server.ListenAndServe()` synchronously on main thread, bypassing signal handler. In `all` mode, SIGINT/SIGTERM terminates immediately without `server.Shutdown(ctx)` or flushing probes. | Run HTTP server in a goroutine across all roles, register `signal.Notify`, and execute `server.Shutdown(ctx)` with a 10s timeout before exiting. |
| **HIGH** | Security | [config.go:25](file:///c:/Users/HP/Projects/ping-gopher/internal/config/config.go#L25)<br>[docker-compose.yml:28](file:///c:/Users/HP/Projects/ping-gopher/docker-compose.yml#L28) | Security Risk | **Insecure Default JWT Secret**: Default secret `"pinggopher-secret-key-change-in-prod"` is hardcoded in fallback and compose. Startup allows running with weak keys, enabling JWT token forgery. | Require `JWT_SECRET` in production. Add startup check in `main.go` that panics if `JWT_SECRET` is shorter than 32 chars or equals default strings. |
| **HIGH** | Security | [auth_handlers.go:41](file:///c:/Users/HP/Projects/ping-gopher/internal/api/auth_handlers.go#L41)<br>[monitor_handlers.go:34](file:///c:/Users/HP/Projects/ping-gopher/internal/api/monitor_handlers.go#L34) | Confirmed Vulnerability | **Unbounded Request Body Size (OOM DoS)**: JSON endpoints decode raw request bodies directly via `json.NewDecoder(r.Body)` without size limits, permitting multi-gigabyte payload requests to crash the service. | Wrap request bodies with `http.MaxBytesReader(w, r.Body, 1<<20)` (1 MB limit) in middleware or API handler entry points. |
| **HIGH** | Engineering | [worker.go:74-93](file:///c:/Users/HP/Projects/ping-gopher/internal/worker/worker.go#L74-L93)<br>[notifier.go:19-63](file:///c:/Users/HP/Projects/ping-gopher/internal/notifier/notifier.go#L19-L63) | Confirmed Defect | **Disconnected Alerting Pipeline**: `WorkerEngine.ProcessHTTPCheck` creates and resolves incidents in the DB but never invokes `NotificationEngine`. Webhook notifications are never sent. | Inject `*notifier.NotificationEngine` into `WorkerEngine` and invoke `NotifyIncidentCreated` / `NotifyIncidentResolved` during state transitions. |
| **HIGH** | Engineering | [scheduler.go:75-99](file:///c:/Users/HP/Projects/ping-gopher/internal/scheduler/scheduler.go#L75-L99) | Confirmed Defect | **Ignored `check_interval_seconds` & Unbounded Goroutines**: `RunCheckCycle` queries all non-paused monitors every 10s regardless of configured interval. Launches untracked goroutines for every monitor every tick. | Filter monitors in DB query by `last_checked_at + check_interval_seconds <= NOW()`. Use a worker pool or semaphore to limit concurrent probe executions. |
| **HIGH** | Engineering | [probe.go:34-40](file:///c:/Users/HP/Projects/ping-gopher/internal/worker/probe.go#L34-L40)<br>[webhook.go:37](file:///c:/Users/HP/Projects/ping-gopher/internal/notifier/webhook.go#L37) | Resource Leak | **HTTP Transport Socket Exhaustion**: `ExecuteHTTPProbe` and `SendWebhookAlert` create `&http.Client{}` per request instead of reusing a shared `http.Transport`, leading to TCP socket exhaustion (`TIME_WAIT`). | Instantiate a global shared `http.Client` with custom `http.Transport` (connection pooling, `MaxIdleConnsPerHost: 100`) and reuse it across probes. |
| **HIGH** | Engineering | [models.go:74-82](file:///c:/Users/HP/Projects/ping-gopher/internal/db/models.go#L74-L82)<br>[monitor_handlers.go:212](file:///c:/Users/HP/Projects/ping-gopher/internal/api/monitor_handlers.go#L212) | DB / Performance | **Missing Composite Index & Unbounded Log Growth**: `PingLog` lacks composite index `(monitor_id, created_at DESC)`. Telemetry logs grow endlessly without a retention pruning policy. | Add `gorm:"index:idx_monitor_created,priority:1"` on `MonitorID` and `priority:2` on `CreatedAt`. Add a daily cron cleanup job purging logs older than 30 days. |
| **HIGH** | Resource Limit | [monitor_handlers.go:205-209](file:///c:/Users/HP/Projects/ping-gopher/internal/api/monitor_handlers.go#205-L209) | Confirmed Defect | **Unbounded Log Query Limit**: `GetMonitorLogsHandler` accepts `?limit=` query parameter without max limit validation. `?limit=10000000` loads millions of database records into RAM. | Enforce `if limit > 500 { limit = 500 }` in `GetMonitorLogsHandler`. |
| **HIGH** | Operations | [probe.go:92-99](file:///c:/Users/HP/Projects/ping-gopher/internal/worker/probe.go#L92-L99) | Goroutine Leak | **Unbounded TCP Dial Timeout in SSL Probe**: `ExecuteSSLProbe` creates `dialer := &tls.Dialer{}` without configuring `NetDialer.Timeout`. Unresponsive hosts cause permanent goroutine leaks. | Pass `context.WithTimeout(ctx, timeout)` to `dialer.DialContext(ctx, "tcp", targetAddr)`. |
| **HIGH** | Maintainability | [README.md:3-10](file:///c:/Users/HP/Projects/ping-gopher/README.md#L3-L10)<br>[go.mod:3](file:///c:/Users/HP/Projects/ping-gopher/go.mod#L3)<br>[Dockerfile:2](file:///c:/Users/HP/Projects/ping-gopher/Dockerfile#L2) | Documentation Defect | **Misleading Architectural Documentation & Go Version**: README claims `gopher-queue` and Postgres support. `go.mod` specifies non-existent Go 1.25.0 (`go 1.25.0`), creating build drift with CI (`Go 1.22`). | Align `go.mod` and `Dockerfile` to stable `go 1.22`. Update README to accurately state current architecture or implement promised features. |
| **MEDIUM** | Security | [auth_handlers.go:46-58](file:///c:/Users/HP/Projects/ping-gopher/internal/api/auth_handlers.go#L46-L58)<br>[auth.go:21-26](file:///c:/Users/HP/Projects/ping-gopher/internal/auth/auth.go#L21-L26) | Security Vulnerability | **Bcrypt Password Processing CPU DoS**: Password string length is unbounded. Submitting megabyte-sized password strings forces heavy CPU computation in `bcrypt.GenerateFromPassword`. | Restrict password inputs to max 72 bytes (Bcrypt truncation limit) before passing to hashing routines. |
| **MEDIUM** | Security | [router.go:22-26](file:///c:/Users/HP/Projects/ping-gopher/internal/api/router.go#L22-L26) | Best-Practice Deviation | **Missing Rate Limiting**: Authentication (`/v1/auth/login`, `/v1/auth/signup`) and public status endpoints lack rate limiting, permitting password brute-forcing. | Implement token-bucket or sliding-window rate limiting middleware (e.g. 5 req/sec per IP on auth endpoints). |
| **MEDIUM** | Engineering | [monitor_handlers.go:212](file:///c:/Users/HP/Projects/ping-gopher/internal/api/monitor_handlers.go#L212)<br>[status_handlers.go:70](file:///c:/Users/HP/Projects/ping-gopher/internal/api/status_handlers.go#L70) | Defect | **Ignored GORM Query Errors**: Several HTTP handlers execute `DB.Find(&slice)` and ignore returned `.Error`, sending HTTP 200 responses when DB queries fail. | Check `if err := dbQuery.Error; err != nil` and respond with `JSONError(w, 500, ...)` across all handlers. |
| **MEDIUM** | Operations | [scheduler.go:61-68](file:///c:/Users/HP/Projects/ping-gopher/internal/scheduler/scheduler.go#L61-L68) | Defect | **Goroutine Leak on Scheduler Shutdown**: `Scheduler.Stop()` only waits for the promoter loop. Spawns detached probe goroutines that remain active during shutdown. | Track probe execution goroutines in `sync.WaitGroup` or context cancellation in `Scheduler`. |
| **MEDIUM** | Observability | [main.go:29](file:///c:/Users/HP/Projects/ping-gopher/cmd/pinggopher/main.go#L29)<br>[db.go:17](file:///c:/Users/HP/Projects/ping-gopher/internal/db/db.go#L17) | Observability Defect | **Unstructured Logging & Verbose SQL Output**: System uses standard `fmt.Printf`. GORM logger is set to `logger.Info` level, dumping millions of SQL queries to stdout. | Replace `fmt.Printf` with `log/slog`. Set GORM log level to `logger.Warn` in production. Add `/metrics` Prometheus endpoint. |
| **LOW** | Security | [middleware.go:23-25](file:///c:/Users/HP/Projects/ping-gopher/internal/api/middleware.go#L23-L25) | Best-Practice Deviation | **Permissive CORS & Missing Security Headers**: `Access-Control-Allow-Origin: *` set globally. Lacks `X-Content-Type-Options`, `X-Frame-Options`, `HSTS`, `CSP`. | Restrict CORS origin to configured domain. Add standard security header middleware. |
| **LOW** | Operations | [Dockerfile:22-35](file:///c:/Users/HP/Projects/ping-gopher/Dockerfile#L22-L35) | Security Best Practice | **Container Root Execution & Missing Healthcheck**: Container runs as `root`. Dockerfile lacks `HEALTHCHECK` instruction. | Create non-root user (`USER appuser`) in Dockerfile and add `HEALTHCHECK --interval=30s CMD wget -qO- http://localhost:8080/health \|\| exit 1`. |

---

## 3. What Would Actually Break in Production (Top 3 Incident Walkthroughs)

### 💥 Scenario 1: SQLite `SQLITE_BUSY` Deadlock Cascade Under Concurrent Traffic
* **Trigger**: The application is deployed to production with 100 configured monitors. User traffic arrives on the Web Dashboard while the Scheduler runs its 10-second check cycle.
* **Failure Chain Walkthrough**:
  1. `Scheduler.RunCheckCycle()` fires every 10 seconds and spawns 100 concurrent goroutines executing `s.WorkerEngine.ProcessHTTPCheck`.
  2. Each goroutine executes `ExecuteHTTPProbe`, followed by `w.DB.Create(&pingLog)` and `w.DB.Model(&monitor).Update(...)`.
  3. Simultaneously, multiple authenticated users issue HTTP GET/POST requests to `/v1/monitors` and `/v1/monitors/:id/logs`.
  4. Because SQLite is initialized in `db.go` without `WAL` mode (`_pragma=journal_mode(WAL)`), without `busy_timeout` (`_pragma=busy_timeout(5000)`), and without `SetMaxOpenConns(1)`, Go's `database/sql` pool opens multiple write connections to the single SQLite file.
  5. SQLite immediately fails concurrent writes with `database is locked` (`SQLITE_BUSY`).
  6. GORM does not automatically retry `SQLITE_BUSY` errors. As a result, probe result insertions fail, monitor status updates are dropped, and API requests return `500 Internal Server Error`.
  7. The application becomes completely unresponsive, dropping uptime checks and failing user requests.

---

### 💥 Scenario 2: Container OOM Crash via Uncapped Telemetry Queries & Unbounded JSON Payloads
* **Trigger**: A user or malicious actor issues an HTTP GET request to `/v1/monitors/:id/logs?limit=50000000` or POSTs a 200MB JSON payload to `/v1/auth/signup`.
* **Failure Chain Walkthrough**:
  1. In `GetMonitorLogsHandler` (`monitor_handlers.go:205`), `strconv.Atoi(limitStr)` parses `50000000`. The handler executes `h.DB.Limit(50000000).Find(&logs)` without imposing a maximum limit ceiling.
  2. GORM allocates memory to deserialize tens of millions of `PingLog` structs into heap memory.
  3. For POST requests, `json.NewDecoder(r.Body).Decode(&req)` reads the entire 200MB body directly into memory without `http.MaxBytesReader` enforcement.
  4. Process RSS memory usage spikes exponentially within milliseconds, exceeding container cgroup memory limits (e.g. 512MB).
  5. The Linux kernel OOM killer sends `SIGKILL` to `pinggopher`, terminating the application instantly and dropping all in-flight requests and background monitoring cycles.

---

### 💥 Scenario 3: Complete Process Failure on SIGTERM & Zombie Probe Goroutines
* **Trigger**: Kubernetes or Docker Swarm issues a `SIGTERM` signal to perform a rolling update or node maintenance.
* **Failure Chain Walkthrough**:
  1. When deployed in API mode (`--role=api`), `cmd/pinggopher/main.go:50` executes `startAPIRole(cfg, apiHandler)`.
  2. Inside `startAPIRole` (`main.go:84`), `server.ListenAndServe()` blocks synchronously on the main goroutine.
  3. Code execution never reaches line 61 (`sigChan := make(chan os.Signal, 1)`), so OS signal handlers are never registered.
  4. The container runtime waits 10 seconds for graceful exit, receives no response, and forcefully sends `SIGKILL`. In-flight client HTTP requests are abruptly terminated with `ECONNRESET`.
  5. When deployed in `all` mode, main receives `SIGTERM` and calls `sched.Stop()`. `sched.Stop()` only waits for the promoter loop ticker (`s.wg`), leaving dozens of background probe goroutines (`go ProcessHTTPCheck`) remaining active detached.
  6. `main()` exits immediately while background goroutines are mid-write to the SQLite file, leaving database transactions uncommitted and telemetry corrupted.

---

## 4. Prioritized Remediation List (Ordered by Severity × Effort)

### 🔴 Phase 1: Immediate Critical Hotfixes (Day 1 - Blockers)
1. **Fix SQLite Concurrency Configuration** (`db.go`):
   - Update SQLite DSN: `fmt.Sprintf("%s?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)", dbPath)`
   - Set connection pool limit: `sqlDB.SetMaxOpenConns(1)` for SQLite connections.
2. **Patch SSRF Vulnerability in Probers** (`probe.go`, `monitor_handlers.go`):
   - Add URL validation helper `validateTargetURL(targetURL string) error` checking scheme (`http`/`https`) and resolving host IP against private/loopback/metadata CIDRs (`127.0.0.0/8`, `10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16`, `169.254.169.254/32`).
3. **Fix Public Status Page Data Leak** (`status_handlers.go`):
   - Require tenant identification slug/ID on public status pages (`/v1/status/public?tenant_id=...`) and scope queries strictly to `user_id`.
4. **Implement HTTP Server Graceful Shutdown** (`main.go`):
   - Run `server.ListenAndServe()` in a goroutine across all role types.
   - Execute `server.Shutdown(ctx)` with a 10s context timeout upon receiving `SIGINT`/`SIGTERM`.

### 🟠 Phase 2: High-Severity Security & Reliability Fixes (Week 1)
5. **Connect Alerting Pipeline** (`worker.go`):
   - Pass `NotificationEngine` into `WorkerEngine` and invoke `NotifyIncidentCreated` / `NotifyIncidentResolved` during monitor state transitions.
6. **Cap Log Query Limits & Enforce Body Size Restrictions** (`middleware.go`, `monitor_handlers.go`, `auth_handlers.go`):
   - Add `MaxBytesMiddleware` capping request bodies at 1 MB (`http.MaxBytesReader`).
   - Enforce hard limit `limit = min(parsedLimit, 500)` in `GetMonitorLogsHandler`.
7. **Fix Scheduler Interval & Worker Execution** (`scheduler.go`, `main.go`):
   - Respect `check_interval_seconds` when querying monitors to check.
   - Implement actual Redis queue task publishing or a bounded worker goroutine pool.
8. **Reuse Shared HTTP Client** (`probe.go`, `webhook.go`):
   - Replace per-request `&http.Client{}` instantiations with a package-level shared `http.Client` equipped with connection pooling.
9. **Fix Insecure Default JWT Key Handling** (`config.go`, `main.go`):
   - Add startup check in `main.go` requiring `JWT_SECRET` in non-development environments.

### 🟡 Phase 3: Operational & Engineering Hardening (Week 2)
10. **Database Indexing & Retention Policy** (`models.go`, `db.go`):
    - Add composite GORM index `idx_monitor_created` on `PingLog(monitor_id, created_at DESC)`.
    - Implement a daily background pruner purging `PingLog` records older than 30 days.
11. **Structured Logging & Observability** (`cmd/pinggopher`, `internal/*`):
    - Migrate from `fmt.Printf` to Go standard `log/slog`. Set GORM log level to `logger.Warn`.
12. **Container & CI Cleanups** (`Dockerfile`, `go.mod`, `README.md`):
    - Align `go.mod` to Go 1.22. Create non-root user in `Dockerfile` and add `HEALTHCHECK`.
    - Update `README.md` to match actual architecture and features.
