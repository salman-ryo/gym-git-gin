# Stats HTTP Handlers & Route Integration

> **Feature:** `analytics-and-power-stats`  
> **Phase:** `03-stats-handlers-and-endpoints`

---

### Task 3.1: Stats Handler Endpoints & Routing

* **Context Bundle:**
  1. [.agent/rules/01-architecture.md](file:///.agent/rules/01-architecture.md)
  2. [.agent/rules/03-testing-and-errors.md](file:///.agent/rules/03-testing-and-errors.md)
* **Owns:**
  - `internal/handler/stats_handler.go`
  - `cmd/server/main.go`
* **Forbidden:**
  - `internal/repository/**`
  - `migrations/**`
* **Acceptance Criteria:**
  - **WHEN** `GET /api/v1/stats` is called by an authenticated user, **THE SYSTEM SHALL** return overall summary statistics, streaks, and attendance rate in a success envelope.
  - **WHEN** `GET /api/v1/stats/power` is called with optional `?days=30`, **THE SYSTEM SHALL** return scientific power score, sub-scores, and tier progression.
