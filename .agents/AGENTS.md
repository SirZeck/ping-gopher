# Workspace Agent Guidelines & Project Rules: PingGopher

## 📚 Package Documentation Rule
Every functional domain package directory (e.g., `cmd/pinggopher`, `internal/config`, `internal/db`, `internal/api`, `internal/worker`, `internal/notifier`, `web`) **must maintain a dedicated `README.md`**.

Each package `README.md` must include:
1. **Responsibility & Purpose**: What this subsystem does within the Modular Monolith.
2. **Key Exported Symbols & Interfaces**: Core structs, functions, or middleware.
3. **Dependencies & Configuration**: External packages, environment variables, or CLI flags consumed.
4. **Usage Code Examples & Commands**: Snippets showing how to initialize or test the package.
