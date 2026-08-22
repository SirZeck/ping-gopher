# 🔔 Multi-Channel Alerting & Notification Pipeline (`internal/notifier`)

`internal/notifier` dispatches real-time outage and recovery alert notifications to external webhooks when monitor status state transitions occur in **PingGopher**.

---

## 📦 Key Exports & Data Structures

### `WebhookPayload`
Standardized JSON payload structure delivered to configured webhook endpoints:

```json
{
  "event": "incident.open",
  "monitor_id": "8438626d-3b40-4306-9678-680df5a237bb",
  "monitor_name": "Production API",
  "target_url": "https://api.example.com/health",
  "status": "DOWN",
  "cause": "HTTP 503 Service Unavailable",
  "timestamp": "2026-08-22T22:30:00Z"
}
```

### `NotificationEngine`
Orchestrates alert delivery:
- **`NotifyIncidentCreated`**: Dispatches `incident.open` webhook notifications upon outage detection.
- **`NotifyIncidentResolved`**: Dispatches `incident.resolved` webhook notifications upon target recovery.

---

## 🧪 Testing & Verification

### Automated Unit Test Suite (`notifier_test.go`)

We use Go's standard `net/http/httptest` package to launch isolated in-memory HTTP webhook receivers:

| Test Method | Test Focus & Verification |
| :--- | :--- |
| **`TestSendWebhookAlertSuccess`** | Tests JSON payload serialization and HTTP POST dispatch to a mock webhook receiver. |
| **`TestNotificationEngineOutageAndRecovery`** | Tests `NotificationEngine` dispatching `incident.open` on outage creation and `incident.resolved` on recovery. |

### Running Notifier Unit Tests

```bash
go test -v ./internal/notifier
```
