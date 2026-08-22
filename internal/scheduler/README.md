# ⏰ Monitor Scheduler Promoter (`internal/scheduler`)

`internal/scheduler` manages periodic polling of active monitors from the database and dispatches check tasks to probe execution workers.

---

## 📦 Key Exports

### `Scheduler` Struct
Manages periodic background ticker loops:
- Queries database for active targets (`status != 'PAUSED'`).
- Builds `CheckPayload` items (`monitor_id`, `target_url`).
- Dispatches execution tasks to worker pools asynchronously.

### `Start(pollInterval time.Duration)`
Boots background promoter loop at specified interval (default: `10s`).

### `Stop()`
Gracefully halts the promoter ticker loop and waits for active dispatches to settle.

---

## 🧪 Testing

Run scheduler unit tests:
```bash
go test -v ./internal/scheduler
```
