# Auth Service & Repository Layer

> **Feature:** `backend-auth-and-bootstrap`  
> **Phase:** `03-auth-service-and-repository`

---

### Task 3.1: User Repository & Auth Service Implementation

* **Context Bundle:**
  1. [.agent/rules/01-architecture.md](file:///.agent/rules/01-architecture.md)
  2. [.agent/rules/02-code-style.md](file:///.agent/rules/02-code-style.md)
* **Owns:**
  - `internal/repository/interfaces.go`
  - `internal/repository/user_repository.go`
  - `internal/service/interfaces.go`
  - `internal/service/auth_service.go`
* **Forbidden:**
  - `internal/handler/**`
  - `cmd/server/main.go`
* **Acceptance Criteria:**
  - **WHEN** `FindOrCreateUser` is invoked, **THE SYSTEM SHALL** look up the user by `auth_user_id` or create a new user profile record with the specified or fallback weekly plan (`ppl-standard`).
  - **WHEN** `UpdateWeeklyPlan` is called, **THE SYSTEM SHALL** update the user's active `weekly_plan_id` in PostgreSQL.
  - **WHEN** user profile is requested, **THE SYSTEM SHALL** return the populated `models.User` object alongside their linked `WeeklyPlan`.
