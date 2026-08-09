# Gym Log Service & Validation Logic

> **Feature:** `gym-logs-and-history`  
> **Phase:** `02-log-service-and-validation`

---

### Task 2.1: GymLogService Validation & Demo Reset Logic

* **Context Bundle:**
  1. [.agent/rules/01-architecture.md](file:///.agent/rules/01-architecture.md)
  2. [.agent/rules/02-code-style.md](file:///.agent/rules/02-code-style.md)
* **Owns:**
  - `internal/service/interfaces.go`
  - `internal/service/log_service.go`
* **Forbidden:**
  - `internal/handler/**`
  - `cmd/server/main.go`
* **Acceptance Criteria:**
  - **WHEN** a log is created or updated, **THE SYSTEM SHALL** validate that `hours` is positive (between 0.1 and 24.0) and `workout_type` is non-empty.
  - **WHEN** date format is parsed, **THE SYSTEM SHALL** enforce `YYYY-MM-DD` standard and reject future dates beyond acceptable tolerance.
  - **WHEN** `ResetDemoLogs` is triggered, **THE SYSTEM SHALL** clear the user's logs and seed a realistic 90-day workout history aligned with their active weekly plan.
