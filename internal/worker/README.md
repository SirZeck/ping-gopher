# ⚡ Core Ping & Synthetic Worker Engine (`internal/worker`)

`internal/worker` executes synthetic HTTP/HTTPS response probes, inspects TLS SSL certificate expiration dates, and processes probe telemetries.

---

## 📦 Key Components

### 1. `ExecuteHTTPProbe(targetURL string, timeout time.Duration) *HTTPProbeResult`
Performs HTTP/HTTPS GET requests against target URLs, measuring status code, response latency in milliseconds, and network diagnostic messages.

### 2. `ExecuteSSLProbe(targetURL string, timeout time.Duration) *SSLProbeResult`
Establishes a TLS connection to extract server peer certificates, calculating remaining validity days and issuer authority.

### 3. `WorkerEngine`
Connects task payload handlers to database updates:
- Records execution data to `ping_logs`.
- Updates `monitors.status` (`UP` / `DOWN`).
- Automatically triggers `Incident` workflows:
  - Creates an `OPEN` incident when a target transitions from `UP` -> `DOWN`.
  - Marks active incidents as `RESOLVED` when a target recovers from `DOWN` -> `UP`.

---

## 🧪 Testing & Verification

### Automated Unit Test Suite (`probe_test.go`)

We use Go's standard `net/http/httptest` package to launch isolated in-memory HTTP mock servers and test probe handlers without external network dependencies:

| Test Method | Test Focus & Verification |
| :--- | :--- |
| **`TestExecuteHTTPProbeSuccess`** | - Launches a mock HTTP server returning `200 OK`.<br>- Verifies `ExecuteHTTPProbe` measures non-zero latency and marks `IsUp = true`. |
| **`TestExecuteHTTPProbeFailure`** | - Launches a mock HTTP server returning `500 Internal Server Error`.<br>- Verifies `ExecuteHTTPProbe` records status `500` and marks `IsUp = false` with cause diagnostics. |
| **`TestWorkerEngineProcessHTTPCheck`** | - End-to-end integration test connecting `WorkerEngine` to a mock HTTP server and isolated test database (`test_worker.db`).<br>- Verifies JSON `CheckPayload` parsing, probe execution, `PingLog` record creation, and `Monitor.Status` state transition (`PAUSED` -> `UP`). |

### Running Worker Unit Tests

```bash
go test -v ./internal/worker
```
