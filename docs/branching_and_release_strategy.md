# 🌿 PingGopher Git Flow & Production Release Safety Architecture

PingGopher enforces an Enterprise-Grade **Git Flow & Semantic Versioning Strategy** designed to guarantee **100% production uptime**, zero accidental deployments, and isolated feature engineering.

---

## 🛡️ Production Safety & Deployment Shield

The production environment (hosted on Render / Cloud Infrastructure) is configured to deploy **EXCLUSIVELY** from the `main` branch upon official release tags (`v1.0.0`, `v1.1.0`).

```text
 🚀 PRODUCTION DEPLOYMENT SHIELD (Render / Cloud)
  main ─────────────────────────────────────────────────────────────● MERGE ONLY WHEN AUDITED (v1.1.0)
                                                                    ▲
 🧪 STAGING & RELEASE CANDIDATE (Pre-Release Testing)                │
  release/v1.1.0 ──────────────────────────────────● Tag: v1.1.0-rc1 ┤
                                                   ▲                │
 ⚙️ INTEGRATION & TESTING (Integration Sandbox)    │                │
  develop ─────────────────────────────────────────┴────────────────┼──────────────
                                                   ▲                │
 🛠️ ISOLATED FEATURE DEVELOPMENT                   │                │
  feature/multi-channel-alerting ──────────────────┤                │
  feature/tcp-dns-probes ──────────────────────────┘                │
```

---

## 🌿 Branch Roles & Engineering Rules

### 1. 🚀 `main` — Production General Availability Branch
- **Status**: Stable & Production-Ready.
- **Access Rule**: **DIRECT COMMITS FORBIDDEN.**
- **Deployment Trigger**: Linked directly to production cloud infrastructure. Automatically deploys upon merging an audited Release Candidate or Hotfix PR.
- **Tagging Obligation**: Every commit on `main` MUST carry a signed Semantic Version tag (`v1.0.0`, `v1.1.0`).

### 2. ⚙️ `develop` — Active Integration Branch
- **Status**: Pre-Release Integration.
- **Purpose**: Serves as the primary integration branch for upcoming feature cycles (`v1.1.0-dev`).
- **Access Rule**: Feature branches merge into `develop` ONLY after passing local test suites (`go test -count=1 ./...`).

### 3. 🛠️ `feature/*` — Isolated Feature Sandbox
- **Naming Pattern**: `feature/<feature-name>` (e.g. `feature/multi-channel-alerting`, `feature/tcp-dns-probes`).
- **Purpose**: Sandbox for experimental development. Keeps unverified code completely isolated from `develop` and `main`.
- **Lifecycle**: Created off `develop`, merged into `develop` via Pull Request, deleted upon successful merge.

### 4. 🧪 `release/*` — Release Candidate (RC) Branch
- **Naming Pattern**: `release/v<major>.<minor>.<patch>` (e.g. `release/v1.1.0`).
- **Purpose**: Feature freeze and release validation.
- **Pre-Release Tagging**: Tagged with pre-release tags (`v1.1.0-rc1`, `v1.1.0-beta.1`) for beta testing before merging into `main`.

### 5. 🚑 `hotfix/*` — Emergency Production Patch Branch
- **Naming Pattern**: `hotfix/v<major>.<minor>.<patch>` (e.g. `hotfix/v1.0.1`).
- **Purpose**: Urgent security patches branched directly off `main` and merged back into both `main` and `develop`.

---

## 🚦 Release Gate & Acceptance Criteria

Before any code is permitted to merge into `main` for a production release, it must satisfy all 4 Release Gates:

1. ✅ **Automated Test Suite Gating**: `go test -count=1 ./...` passes 100% with zero panics.
2. ✅ **SSRF & Security Gating**: All new socket dialers or webhook dispatchers enforce `validator.SafeDialContext` and `ValidateSafeURL`.
3. ✅ **Multi-Tenant Privacy Gating**: Database queries enforce tenant scoping (`user_id` / `tenant_id`).
4. ✅ **Release Candidate Sanity Verification**: Both CLI and Web UI verified against a running local server instance.

---

## 🏷️ Semantic Versioning Standards (`MAJOR.MINOR.PATCH`)

- **`MAJOR` (e.g. `v1.0.0` -> `v2.0.0`)**: Incompatible API breaking changes or major architecture overhaul.
- **`MINOR` (e.g. `v1.0.0` -> `v1.1.0`)**: Backwards-compatible feature additions (e.g. Slack/Discord alerting, TCP probes).
- **`PATCH` (e.g. `v1.0.0` -> `v1.0.1`)**: Backwards-compatible security patches and bug fixes.
