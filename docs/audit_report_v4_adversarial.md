# 🛡️ Adversarial Re-Audit Report: PingGopher

**Auditor Role:** Independent Senior Security Engineer (Adversarial Re-Audit)  
**Prior Report Under Review:** `docs/audit_report_v3.md` (dated August 23, 2026, verdict: "APPROVED FOR PRODUCTION")  
**Date:** August 23, 2026  
**Verdict:** ❌ **NOT APPROVED FOR PRODUCTION** — Remaining Critical SSL Probe SSRF, Incident Privacy Leak, and Container Healthcheck Defects.

---

## 1. Executive Summary

An independent adversarial security re-audit was performed on Release `v1.2.3`. While previous phases resolved the primary HTTP probe SSRF vectors and database WAL locking, this re-audit identified critical unmitigated SSRF vectors in SSL probes, private incident data leaks on public status pages, failing container healthchecks, and rate-limiter proxy bypasses.

---

## 2. Key Adversarial Findings

### 🔴 Critical & High Severity Findings

1. **SSL Probe TOCTOU SSRF Bypass** (`internal/worker/probe.go:127-137`):
   `ExecuteSSLProbe` calls `ValidateSafeURL()` at validation time, but executes TLS dialing via a standard `net.Dialer` rather than `validator.SafeDialContext`. Under DNS TTL=0 rebinding, the TLS handshake connects to the rebound internal IP (`127.0.0.1` / `169.254.169.254`).
2. **Public Status Private Incident Data Leak** (`internal/api/status_handlers.go:92-96`):
   The incident query in `PublicStatusHandler` filters monitors by `user_id`, but omits `.Where("monitors.is_public = ?", true)`. Outage logs, monitor names, and causes for **private monitors** are leaked publicly on `/v1/status/public`.
3. **Always-Failing Docker Container HEALTHCHECK** (`Dockerfile:42`):
   The Dockerfile healthcheck queries `http://localhost:8080/v1/status/public`, which returns `HTTP 400 Bad Request` because `tenant_id` is required. The container is permanently marked `unhealthy` by Docker orchestrators.
4. **Rate Limiter Proxy Bypass & Memory Leak** (`internal/api/middleware.go:35-37`):
   Rate limiting uses raw `r.RemoteAddr`. Behind a reverse proxy, all clients share a single proxy IP bucket. Additionally, the IP map never evicts stale entries, causing memory bloat over time.
5. **IPv4-Mapped IPv6 CIDR Bypass** (`internal/validator/ssrf.go:144`):
   `isPrivateIP` checks `IsLoopback()` but omits `ip.To4()` normalization before CIDR evaluation, potentially missing IPv4-mapped IPv6 private addresses (e.g. `::ffff:169.254.169.254`).

---

## 3. Prior Audit Verdict: NOT APPROVED

Production deployment is paused until Phase 4 remediations (v1.2.4) resolve the SSL probe SSRF vector, private incident leak, and Docker container healthcheck URL.
