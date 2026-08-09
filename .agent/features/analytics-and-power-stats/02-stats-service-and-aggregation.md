# Stats Service & Metric Aggregation Layer

> **Feature:** `analytics-and-power-stats`  
> **Phase:** `02-stats-service-and-aggregation`

---

### Task 2.1: StatsService Orchestration & Range Aggregation

* **Context Bundle:**
  1. [.agent/rules/01-architecture.md](file:///.agent/rules/01-architecture.md)
  2. [.agent/rules/02-code-style.md](file:///.agent/rules/02-code-style.md)
* **Owns:**
  - `internal/service/interfaces.go`
  - `internal/service/stats_service.go`
* **Forbidden:**
  - `internal/handler/**`
  - `cmd/server/main.go`
* **Acceptance Criteria:**
  - **WHEN** `GetStats` is invoked, **THE SYSTEM SHALL** fetch all user logs and user profile, compute streaks, workout distribution, total hours, and return `models.StatsSummary`.
  - **WHEN** `GetPowerStats` is invoked with a custom day window (e.g. 30, 90, 365 days), **THE SYSTEM SHALL** filter logs, fetch active weekly plan targets, and compute full `models.PowerScore` breakdown.
