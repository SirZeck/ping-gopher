# Contributing to PingGopher 🐹

Thank you for your interest in contributing to **PingGopher**! We welcome bug fixes, documentation improvements, feature suggestions, and code contributions.

---

## 🛠️ Development Setup

### 1. Prerequisites
- **Go 1.22+**
- **Docker & Docker Compose** (or local Redis server)
- **Git**

### 2. Fork & Clone Repository
```bash
git clone https://github.com/YOUR_USERNAME/ping-gopher.git
cd ping-gopher
```

### 3. Build & Run Tests
Ensure all existing tests pass before making modifications:

```bash
# Run unit tests
make test

# Build local binary
make build
```

---

## 📁 Package Architecture & Rules

PingGopher follows a **Modular Monolith** pattern. Each functional domain package contains its own localized `README.md` detailing its responsibilities:

- [`cmd/pinggopher`](cmd/pinggopher/README.md): Multi-role application launcher.
- [`internal/config`](internal/config/README.md): Environment variables and CLI configuration loader.
- [`internal/db`](internal/db/README.md): GORM models, ERD, and database initializers.

When adding new domain packages, please include a `README.md` in the new folder following our workspace guidelines.

---

## 🚀 Submitting Pull Requests

1. **Create a Feature Branch**:
   ```bash
   git checkout -b feat/your-feature-name
   ```
2. **Commit Changes**: Follow conventional commit messages (e.g. `feat: ...`, `fix: ...`, `docs: ...`).
3. **Run Code Formatting & Tests**:
   ```bash
   go fmt ./...
   go test -v ./...
   ```
4. **Push Branch & Open PR**: Push to your fork and submit a Pull Request targeting the `main` branch.
