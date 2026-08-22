# 🚀 Main Entrypoint (`cmd/pinggopher`)

`cmd/pinggopher` contains the primary application entry point (`main.go`) for **PingGopher**.

---

## 🎭 "Deploy Any Role" Execution Strategy

PingGopher uses a single binary compiled from `cmd/pinggopher/main.go` that inspects the `--role` CLI flag (or `ROLE` environment variable) to conditionally initialize and run runtime roles:

```bash
# Run all roles in a single process (API + Worker + Scheduler)
./bin/pinggopher.exe --role=all

# Run dedicated HTTP REST API server
./bin/pinggopher.exe --role=api --port=8080

# Run gopher-queue task execution worker pool
./bin/pinggopher.exe --role=worker --redis=localhost:6379

# Run monitor check promoter loop
./bin/pinggopher.exe --role=scheduler
```

---

## 🛠️ CLI Flags Reference

| Flag | Default | Description |
| :--- | :--- | :--- |
| `--role` | `all` | Deployment role: `all`, `api`, `worker`, `scheduler` |
| `--port` | `8080` | HTTP API server port |
| `--db` | `pinggopher.db` | SQLite database file path or Postgres DSN |
| `--redis` | `localhost:6379` | Redis broker address for `gopher-queue` |
| `--jwt-secret` | `pinggopher-secret...` | Secret key for signing JWT tokens |
