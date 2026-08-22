# 🌐 REST API Server & Router (`internal/api`)

`internal/api` provides HTTP RESTful API routing, JSON response formatting, CORS middleware, JWT authentication middleware, and tenant endpoints for **PingGopher**.

---

## 📡 API Endpoints Reference

### 1. System Health
- **`GET /health`** (Public): Returns API status.

### 2. Tenant Authentication
- **`POST /v1/auth/signup`** (Public): Registers a new tenant user. Expects `{"email": "...", "password": "..."}`. Returns JWT token.
- **`POST /v1/auth/login`** (Public): Authenticates user credentials. Returns JWT token.

### 3. Monitor Management (JWT Authenticated)
Requires `Authorization: Bearer <TOKEN>` header.

- **`POST /v1/monitors`**: Creates a new target monitor.
- **`GET /v1/monitors`**: Lists user's monitors.
- **`GET /v1/monitors/{id}`**: Gets monitor details.
- **`PUT /v1/monitors/{id}`**: Updates monitor target URL, check interval, or status (`UP` / `PAUSED`).
- **`DELETE /v1/monitors/{id}`**: Deletes monitor target.

### 4. Telemetry Logs & Incidents (JWT Authenticated)
- **`GET /v1/monitors/{id}/logs`**: Returns recent probe execution logs (`PingLog`). Optional query param: `?limit=50`.
- **`GET /v1/monitors/{id}/incidents`**: Returns outage incident history.

---

## 🔒 Security & Middleware

- **`CORSMiddleware`**: Injects headers for cross-origin browser dashboards.
- **`AuthMiddleware`**: Validates JWT Bearer tokens and attaches tenant `UserID` to request contexts.

---

## 🧪 Testing & Verification

### Automated Integration Test Suite (`api_test.go`)

| Test Method | Test Focus & Verification |
| :--- | :--- |
| **`TestSignupAndLoginAPI`** | Tests user signup (`POST /v1/auth/signup`), duplicate email rejection (`409 Conflict`), and login authentication (`POST /v1/auth/login`). |
| **`TestMonitorCRUDAPI`** | Tests authenticated monitor creation (`POST /v1/monitors`), listing (`GET /v1/monitors`), updating (`PUT`), and deletion (`DELETE`) using JWT Bearer headers. |
| **`TestTelemetryAndIncidentsAPI`** | Tests querying execution `PingLog` entries (`GET /v1/monitors/{id}/logs`) and outage `Incident` records. |

### Running API Integration Tests

```bash
go test -v ./internal/api
```
