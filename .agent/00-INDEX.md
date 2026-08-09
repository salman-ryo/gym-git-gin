# Gym-Git Backend Context Index & Feature Router

> **Central Context Gateway**: Map Go backend file paths, globs, task types, and features directly to rule definitions and phase files.

---

## 1. Project Rules Router

| Task Category / Path Glob | Applicable Rule File | Core Responsibility |
| :--- | :--- | :--- |
| **Global Architecture & Config** (`cmd/**`, `config/**`, `internal/database/**`, `migrations/**`, `.air.toml`) | [.agent/rules/01-architecture.md](file:///.agent/rules/01-architecture.md) | Go 1.22+ runtime, Gin server setup, PostgreSQL pooling, embedded SQL migrations, Supabase Auth integration |
| **Code Style & Layered Patterns** (`internal/handler/**`, `internal/service/**`, `internal/repository/**`, `internal/models/**`) | [.agent/rules/02-code-style.md](file:///.agent/rules/02-code-style.md) | Layered architecture separation, interface definitions, SQL query practices, idiomatic Go patterns, naming |
| **Testing, Error Handling & Response Envelope** (`internal/**/*_test.go`, `internal/models/response.go`, `internal/middleware/**`) | [.agent/rules/03-testing-and-errors.md](file:///.agent/rules/03-testing-and-errors.md) | Standard JSON success/error envelopes, HTTP status codes, table-driven unit tests, mock repositories |

---

## 2. Feature Workspaces & State Trackers

| Feature | Directory | Description & Current State |
| :--- | :--- | :--- |
| **1. Backend Auth & User Bootstrap** | [.agent/features/backend-auth-and-bootstrap/](file:///.agent/features/backend-auth-and-bootstrap/) | Supabase JWT authentication middleware (RS256/ES256 JWKS & HS256), idempotent user bootstrap (`POST /api/v1/auth/bootstrap`), user profile hydration (`GET /api/v1/auth/me`), plan switching (`PUT /api/v1/auth/plan`), and logout. [View STATE.md](file:///.agent/features/backend-auth-and-bootstrap/STATE.md) |
| **2. Gym Logs & Workout History** | [.agent/features/gym-logs-and-history/](file:///.agent/features/gym-logs-and-history/) | Daily workout logs CRUD, user-isolated range queries (`GET /api/v1/logs`), atomic upserting (`POST /api/v1/logs`, `PUT /api/v1/logs/:date`), log deletion (`DELETE /api/v1/logs/:date`), and demo seed reset (`POST /api/v1/logs/reset`). [View STATE.md](file:///.agent/features/gym-logs-and-history/STATE.md) |
| **3. Analytics & Scientific Power Stats** | [.agent/features/analytics-and-power-stats/](file:///.agent/features/analytics-and-power-stats/) | Contribution heatmap matrix calculation, streak tracker, workout distribution, and scientific Power Score engine (Consistency, Intensity, Volume, Recovery, Anime Tiers) (`GET /api/v1/stats`, `GET /api/v1/stats/power`). [View STATE.md](file:///.agent/features/analytics-and-power-stats/STATE.md) |
| **4. Weekly Plans & Workout Splits** | [.agent/features/weekly-plans/](file:///.agent/features/weekly-plans/) | Preset split catalog (`ppl-standard`, `ppl-core`, `upper-lower`, `full-body`, `ppl`) and custom user-created workout routines (`GET /api/v1/plans`). [View STATE.md](file:///.agent/features/weekly-plans/STATE.md) |

---

## 3. Master Operating Procedures (MOP) Quick Reference

1. **Context Read:** Read ONLY this index and the target feature's `STATE.md` to identify the required 2–3 context files.
2. **No Ghosting:** Write full, production-ready Go code. No `TODO` comments or empty stubs.
3. **Circuit Breaker:** If 3 consecutive attempts fail, revert to last known working state, log in `STATE.md`, and request human user assistance.
4. **Post-Execution Hook (MANDATORY):** Mark tasks `[x]` in `STATE.md` and append a timestamped log to the execution table before finishing.
