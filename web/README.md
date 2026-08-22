# 🖥️ Web Dashboard Application (`web`)

`web` provides the glassmorphic, single-page web dashboard frontend for **PingGopher**.

---

## 🎨 Design System & Aesthetics

- **Theme**: Glassmorphic Dark Mode (`background: rgba(18, 24, 38, 0.75)`, `backdrop-filter: blur(16px)`).
- **Typography**: [Inter](https://fonts.google.com/specimen/Inter) Google Font.
- **Palette Tokens**:
  - `Emerald` (`#10b981`): Operational targets (`UP`), high SLA metrics.
  - `Rose` (`#f43f5e`): Outage targets (`DOWN`), error diagnostics.
  - `Amber` (`#f59e0b`): Paused target monitors.
  - `Cyan & Blue`: Primary call-to-action buttons & background radial glows.

---

## 📦 Component Structure

- **[`index.html`](index.html)**: Semantic HTML5 SPA container including navigation header, tabbed Auth Modal, SLA Stats summary bar, Monitor Cards grid, Add Target Modal, and Telemetry Execution Logs Modal.
- **[`style.css`](style.css)**: Glassmorphism CSS design system, flexbox/grid layouts, pulse dots, and responsive breakpoints.
- **[`app.js`](app.js)**: SPA state management, JWT token persistence (`localStorage`), REST API integration (`fetchAPI`), polling timer (every 10s), and interactive latency timeline charts.
- **[`embed.go`](embed.go)**: Go 1.16+ `embed.FS` wrapper exporting `StaticHandler()` for zero-dependency single-binary serving.

---

## 🧪 Testing & Verification

1. **Static Serving Check**: Run `go test -v ./internal/api` to verify root route `/` responds with `index.html`.
2. **Browser Verification**: Navigate to `http://localhost:8080/` to test interactive login, target creation, status toggling, and telemetry charts.
