# 💻 PingGopher Official CLI Tool (`cmd/pinggopher-cli`)

`pinggopher-cli` is the official command-line client for **PingGopher**, enabling terminal-based tenant authentication, monitor management, real-time status checking, and probe telemetry inspection.

---

## 🚀 Installation

### Via `go install`
```bash
go install github.com/SirZeck/ping-gopher/cmd/pinggopher-cli@latest
```

> [!TIP]
> **Enabling direct `pinggopher-cli` execution from any terminal window:**
> Go installs binaries into `$(go env GOPATH)/bin` (`%USERPROFILE%\go\bin` on Windows).
> - **Windows (PowerShell)**: Run `[Environment]::SetEnvironmentVariable("Path", $env:Path + ";$env:USERPROFILE\go\bin", "User")`
> - **Linux / macOS (Bash/Zsh)**: Add `export PATH=$PATH:$(go env GOPATH)/bin` to your `~/.bashrc` or `~/.zshrc`.

### Local Build
```bash
make build-cli
# Produces bin/pinggopher-cli.exe
```

---

## ⚡ Command Reference

### 1. `signup` (Registration)
Registers a new tenant account on a PingGopher server and saves the JWT Bearer token to `~/.pinggopher/credentials.json`:

```bash
pinggopher-cli signup --url http://localhost:8080 --email newuser@pinggopher.io --password SecretPassword123!
```

### 2. `login` (Authentication)
Authenticates against a PingGopher API server and saves the JWT Bearer token to `~/.pinggopher/credentials.json`:

```bash
pinggopher-cli login --url http://localhost:8080 --email admin@pinggopher.io --password SecretPassword123!
```

### 2. `status` (System Health Summary)
Queries public system operational status without requiring login:

```bash
pinggopher-cli status
```

### 3. `monitor` (Monitor Target Management)
- **List Targets**:
  ```bash
  pinggopher-cli monitor list
  ```
- **Add Target**:
  ```bash
  pinggopher-cli monitor add --name "Production API" --url "https://api.example.com/health" --interval 30
  ```
- **Delete Target**:
  ```bash
  pinggopher-cli monitor delete --id <MONITOR_UUID>
  ```

### 4. `logs` (Probe Latency Telemetry)
Fetches recent probe execution response times and HTTP status codes:

```bash
pinggopher-cli logs --id <MONITOR_UUID> --limit 20
```

---

## 🧪 Testing & Verification

### Automated CLI Unit Test Suite (`cli_test.go`)

| Test Method | Test Focus & Verification |
| :--- | :--- |
| **`TestCLILoginAndCredentials`** | Tests `DoAPIRequest` and credential persistence to temporary config paths. |
| **`TestCLIMonitorManagement`** | Tests `monitor list` and `monitor add` command workflows against an in-memory `httptest.Server`. |

### Running CLI Tests

```bash
go test -v ./cmd/pinggopher-cli
```
