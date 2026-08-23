# PingGopher 🐹📡

> **Modern, High-Performance Uptime & Synthetic Monitoring SaaS powered by [`gopher-queue`](https://github.com/SirZeck/gopher-queue)**

[![CI](https://github.com/SirZeck/ping-gopher/actions/workflows/ci.yml/badge.svg?style=for-the-badge)](https://github.com/SirZeck/ping-gopher/actions)
[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=for-the-badge&logo=go)](https://golang.org)
[![gopher-queue](https://img.shields.io/badge/Engine-gopher--queue_v1.0.0-FF6B6B?style=for-the-badge)](https://github.com/SirZeck/gopher-queue)
[![License](https://img.shields.io/badge/License-MIT-blue.svg?style=for-the-badge)](LICENSE)

**PingGopher** is a standalone, full-stack **Uptime & Synthetic Monitoring SaaS** built with Go and powered by the distributed task engine `gopher-queue`. Engineered with a **Modular Monolith ("Deploy Any Role")** architecture, PingGopher can run as an all-in-one process or scale horizontally by deploying dedicated API, Worker, and Scheduler roles from a single binary.

---

## 📸 Screenshots & Product Previews

### 1. Main Web Dashboard (`/`)
![PingGopher Dashboard Preview](assets/dashboard.png)

### 2. Probe Response Latency & Execution Logs Modal
![PingGopher Telemetry & SLA Logs](assets/logsAndSLAs.png)

### 3. Hosted Public Status Page (`/status.html`)
![PingGopher Public Status Page](assets/status.png)

---

## 📂 Project Structure

```text
ping-gopher/
├── .github/workflows/    # Automated CI/CD build & test workflows
├── .agents/              # Workspace rules & agent instructions
├── assets/               # Repository preview screenshots & media assets
├── bin/                  # Pre-compiled cross-platform static binaries
├── cmd/
│   ├── pinggopher/       # Main server binary entrypoint & multi-role launcher
│   └── pinggopher-cli/   # Official CLI tool client (terminal control)
├── internal/
│   ├── api/              # REST API handlers, router & middleware
│   ├── auth/             # Password security (Bcrypt) & JWT token engine
│   ├── config/           # Environment variables & CLI flags parser
│   ├── db/               # GORM database layer, models & migrations
│   ├── notifier/         # Webhook & multi-channel alert pipeline
│   ├── scheduler/        # Background monitor check promoter loop
│   └── worker/           # Synthetic probe execution engine (HTTP/SSL)
├── web/                  # Glassmorphic SPA frontend & Go static embedder
├── Dockerfile            # Multi-stage production container build
├── docker-compose.yml    # Multi-container orchestration (PingGopher + Redis)
├── Makefile              # Development build & execution shortcuts
└── README.md             # Master project documentation
```

---

## 🏛️ Architecture & System Design

PingGopher uses a **Modular Monolith** pattern with decoupling between domain packages (`internal/db`, `internal/config`, `internal/api`, `internal/worker`, `internal/notifier`).

```
                    +------------------------------------+
                    |  REST API / Web Dashboard Client   |
                    +------------------------------------+
                                      |
                                      v
                    +------------------------------------+
                    |    pinggopher (--role=api)         |
                    +------------------------------------+
                                      |
                                      v
     +------------------------------------------------------------------+
     |                 Database Layer (SQLite / PostgreSQL)              |
     |           Users | Monitors | PingLogs | Incidents                |
     +------------------------------------------------------------------+
                                      ^
                                      |
    +---------------------------------+---------------------------------+
    |                                                                   |
    v                                                                   v
+------------------------------------+              +------------------------------------+
|   pinggopher (--role=scheduler)    |              |     pinggopher (--role=worker)     |
|   Enqueues Monitor Probes          |              |     Executes Probes via           |
|   into gopher-queue                |              |     gopher-queue Task Pools      |
+------------------------------------+              +------------------------------------+
                  \                                                   /
                   \                                                 /
                    +-----------------------------------------------+
                    |     gopher-queue Broker (Redis 7.0+)          |
                    |     Atomic Leasing, Retries, Watchdog, DLQ    |
                    +-----------------------------------------------+
```

### 🎭 Deploy Any Role Strategy
Run any node type using runtime CLI flags or environment variables:
- **`pinggopher --role=all`**: All-in-one development node running API, Scheduler, and Worker.
- **`pinggopher --role=api`**: Dedicated HTTP REST API & Web Dashboard node.
- **`pinggopher --role=worker`**: Dedicated background probe execution worker.
- **`pinggopher --role=scheduler`**: Dedicated monitor promoter loop enqueuing jobs to `gopher-queue`.

---

## ✨ Key Features

- 🌐 **External Synthetic Uptime Monitoring**: Measures HTTP response codes, latency (ms), and SSL certificate expiration.
- ⚡ **Powered by `gopher-queue`**: Leverages server-side Redis Lua scripts, atomic leasing, watchdog supervisor failover, and exponential retries.
- 🛡️ **Anti-Alert Fatigue**: Smart escalation logic using retry windows before dispatching alerts.
- 📊 **Multi-Tenant Data Models**: Built-in GORM schema for Users, Monitors, Latency Logs, and Outage Incidents.
- 🚀 **Zero-Dependency Local Storage**: Runs out of the box with embedded CGO-free SQLite for local development and PostgreSQL support for production.

---

## 💻 Official CLI Tool (`pinggopher-cli`)

PingGopher includes an official command-line client for terminal-based tenant authentication, monitor CRUD management, public status checking, and probe latency inspection.

### 📥 Installation via `go install`
```bash
go install github.com/SirZeck/ping-gopher/cmd/pinggopher-cli@latest
```

> [!TIP]
> **Executing `pinggopher-cli` from any terminal window:**
> Go installs binaries into `$(go env GOPATH)/bin` (`%USERPROFILE%\go\bin` on Windows).
> - **Windows (PowerShell)**: `[Environment]::SetEnvironmentVariable("Path", $env:Path + ";$env:USERPROFILE\go\bin", "User")`
> - **Linux / macOS**: `export PATH=$PATH:$(go env GOPATH)/bin` (add to `~/.bashrc` or `~/.zshrc`).

### ⚡ Terminal Usage Examples

```bash
# 1. Register a new tenant account and save session token
pinggopher-cli signup --url http://localhost:8080 --email newuser@pinggopher.io --password SecretPassword123!

# 2. Or authenticate an existing account
pinggopher-cli login --url http://localhost:8080 --email admin@pinggopher.io --password SecretPassword123!

# 2. View public system operational health status
pinggopher-cli status

# 3. List active monitor targets in formatted ASCII table
pinggopher-cli monitor list

# 4. Add a new target monitor directly from the terminal
pinggopher-cli monitor add --name "Production API" --url "https://api.example.com/health" --interval 30

# 5. Inspect probe latency response history logs
pinggopher-cli logs --id <MONITOR_ID> --limit 20
```

Read the full [`cmd/pinggopher-cli/README.md`](cmd/pinggopher-cli/README.md) for detailed CLI documentation.

---

## 🚀 Quickstart

### 1. Run via Docker Compose (Recommended)
Launch PingGopher and Redis broker with a single command:

```bash
docker compose up -d --build
```

Verify application status:
```bash
docker compose logs -f pinggopher
```

---

### 2. Build & Run Locally

#### Prerequisites
- **Go 1.22+**
- **Redis 7.0+** running on `localhost:6379` (required for `gopher-queue` worker pool)

#### Build Binary
```bash
make build
# Produces bin/pinggopher.exe
```

#### Run All-in-One Node
```bash
make run-all
# Or run manually:
./bin/pinggopher.exe --role=all
```

---

## ⚙️ Configuration

Configurations can be set via command-line flags or environment variables:

| Flag | Environment Variable | Default | Description |
| :--- | :--- | :--- | :--- |
| `--role` | `ROLE` | `all` | Service role (`all`, `api`, `worker`, `scheduler`) |
| `--port` | `PORT` | `8080` | HTTP API server port |
| `--db` | `DB_PATH` | `pinggopher.db` | SQLite database file path or Postgres DSN |
| `--redis` | `REDIS_ADDR` | `localhost:6379` | Redis broker address for `gopher-queue` |
| `--jwt-secret` | `JWT_SECRET` | `pinggopher-secret...` | Secret key for JWT token signing |

---

## 🤝 Contributing

Contributions are welcome! Please check our [Contributing Guide](CONTRIBUTING.md) for setup steps and pull request guidelines.

---

## 📄 License

Distributed under the MIT License. See `LICENSE` for details.
