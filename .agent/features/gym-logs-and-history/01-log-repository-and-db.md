# Gym Log Repository & SQL Implementation

> **Feature:** `gym-logs-and-history`  
> **Phase:** `01-log-repository-and-db`

---

### Task 1.1: GymLogRepository Interface & PostgreSQL Queries

* **Context Bundle:**
  1. [.agent/rules/01-architecture.md](file:///.agent/rules/01-architecture.md)
  2. [.agent/rules/02-code-style.md](file:///.agent/rules/02-code-style.md)
* **Owns:**
  - `internal/repository/interfaces.go`
  - `internal/repository/log_repository.go`
* **Forbidden:**
  - `internal/handler/**`
  - `internal/middleware/**`
* **Acceptance Criteria:**
  - **WHEN** `Upsert` is executed, **THE SYSTEM SHALL** insert or update the log for `(user_id, date)` atomically using `ON CONFLICT (user_id, date) DO UPDATE`.
  - **WHEN** `GetByDateRange` is called, **THE SYSTEM SHALL** return all logs for a user between `startDate` and `endDate` ordered by date.
  - **WHEN** `DeleteByDate` is called, **THE SYSTEM SHALL** delete the log for the given `(user_id, date)`.
  - **WHEN** `DeleteAllByUserID` is called, **THE SYSTEM SHALL** purge all logs for the user (used for demo resets).
