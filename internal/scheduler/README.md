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

## 🧪 Testing & Verification

### Automated Unit Test Suite (`scheduler_test.go`)

We use an isolated SQLite test database (`test_scheduler.db`) and mock HTTP server to test monitor check polling and dispatching:

| Test Method | Test Focus & Verification |
| :--- | :--- |
| **`TestSchedulerRunCheckCycle`** | - Creates an active monitor record (`status = 'UP'`) in an isolated database.<br>- Executes `RunOnce(ctx)` synchronously to trigger a promoter check cycle.<br>- Verifies that the scheduler selects active monitors, dispatches check payloads to `WorkerEngine`, and generates a `PingLog` entry. |

### Running Scheduler Unit Tests

```bash
go test -v ./internal/scheduler
```
