# Workspace Agent Guidelines & Project Rules: PingGopher

## 📚 Package Documentation Rule
Every functional domain package directory (e.g., `cmd/pinggopher`, `internal/config`, `internal/db`, `internal/api`, `internal/worker`, `internal/notifier`, `web`) **must maintain a dedicated `README.md`**.

Each package `README.md` must include:
1. **Responsibility & Purpose**: What this subsystem does within the Modular Monolith.
2. **Key Exported Symbols & Interfaces**: Core structs, functions, or middleware.
3. **Dependencies & Configuration**: External packages, environment variables, or CLI flags consumed.
4. **Usage Code Examples & Commands**: Snippets showing how to initialize or test the package.

---

## 🔀 Agile Atomic Commit Rule
All code additions, feature implementations, and bug fixes must follow an **Agile Atomic Commit Methodology**:
- **No Monolithic Phase Commits**: Avoid bundling an entire multi-component phase into a single massive commit.
- **Granular Feature Commits**: Commit after completing each atomic story/task (e.g., helper utilities, data handlers, API routers, tests, documentation).
- **Conventional Commit Messages**: Use clear domain-scoped prefixes:
  - `feat(domain)`: New feature or handler (e.g., `feat(auth): implement JWT token generation`)
  - `fix(domain)`: Bug fix or edge-case handling
  - `docs(domain)`: Package documentation or README updates
  - `test(domain)`: Unit or integration test additions
