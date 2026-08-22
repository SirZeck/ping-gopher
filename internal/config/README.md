# ⚙️ Configuration Manager (`internal/config`)

`internal/config` provides centralized configuration loading from environment variables and CLI flags.

---

## 📦 Key Exports

### `Config` Struct
```go
type Config struct {
    Role         string // Deployment role (all, api, worker, scheduler)
    Port         string // HTTP API port (default: "8080")
    DatabasePath string // SQLite file path or Postgres DSN (default: "pinggopher.db")
    RedisAddr    string // Redis broker address for gopher-queue (default: "localhost:6379")
    JWTSecret    string // Secret key for JWT signing
}
```

### `LoadConfig() *Config`
Parses command-line flags and falls back to environment variables (`ROLE`, `PORT`, `DB_PATH`, `REDIS_ADDR`, `JWT_SECRET`) if set.

---

## 💡 Usage Example

```go
package main

import (
    "fmt"
    "github.com/pinggopher/ping-gopher/internal/config"
)

func main() {
    cfg := config.LoadConfig()
    fmt.Printf("Starting PingGopher on port %s (Role: %s)\n", cfg.Port, cfg.Role)
}
```
