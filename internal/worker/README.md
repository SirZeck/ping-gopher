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

## 🧪 Testing

Run probe unit tests:
```bash
go test -v ./internal/worker
```
