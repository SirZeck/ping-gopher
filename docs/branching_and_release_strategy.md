# 🌿 PingGopher Git Branching & Release Strategy

PingGopher follows a structured **Git Flow & Semantic Versioning** branching model to ensure production stability, isolated feature development, and reliable release management.

---

## 🌿 Branch Topology & Roles

```text
  main (Production GA Releases) ───────────────────● v1.0.0 ──────────────────────● v1.1.0 (GA)
                                                   │                             ▲
  release/v1.1.0 (Release Candidate) ──────────────┼─────────────● v1.1.0-rc1 ───┤
                                                   │             ▲               │
  develop (Integration) ───────────────────────────●─────────────┼───────────────┤
                                                   │             │               │
  feature/* (Feature Branches) ────────────────────┴──● feature  ┴──● feature    │
```

### 1. `main` (Stable Production Branch)
- **Purpose**: Contains production General Availability (GA) code.
- **Rule**: Every commit on `main` MUST be tagged with a semantic version (`v1.0.0`, `v1.1.0`).
- **Protection**: Direct commits are forbidden. Changes enter `main` exclusively via tested release PRs from `release/*` or `hotfix/*` branches.

### 2. `develop` (Active Development Integration Branch)
- **Purpose**: Primary integration branch for upcoming feature releases (`v1.1.0-dev`).
- **Rule**: All feature branches (`feature/*`) branch off `develop` and merge back into `develop` after code review and automated test suite verification (`go test -count=1 ./...`).

### 3. `feature/*` (Feature Development Branches)
- **Naming Pattern**: `feature/<feature-name>` (e.g. `feature/slack-discord-alerts`, `feature/tcp-dns-probes`).
- **Lifecycle**: Created off `develop`, merged into `develop` via Pull Request, deleted upon merge.

### 4. `release/*` or `beta/*` (Release Candidates)
- **Naming Pattern**: `release/v<major>.<minor>.<patch>` (e.g. `release/v1.1.0`).
- **Purpose**: Preparation for new production releases. Tagged with beta/RC tags (`v1.1.0-rc1`) for community pre-release testing.

### 5. `hotfix/*` (Emergency Production Patches)
- **Naming Pattern**: `hotfix/v<major>.<minor>.<patch>` (e.g. `hotfix/v1.0.1`).
- **Purpose**: Urgent security or production bug fixes branched directly off `main` and merged back into both `main` and `develop`.

---

## 🏷️ Semantic Versioning Rules (`MAJOR.MINOR.PATCH`)

- **`MAJOR` (e.g. `v1.0.0` -> `v2.0.0`)**: Incompatible API breaking changes or major architectural redesigns.
- **`MINOR` (e.g. `v1.0.0` -> `v1.1.0`)**: Backwards-compatible new features (e.g. Slack/Discord alerting, TCP probes).
- **`PATCH` (e.g. `v1.0.0` -> `v1.0.1`)**: Backwards-compatible security patches and bug fixes.
