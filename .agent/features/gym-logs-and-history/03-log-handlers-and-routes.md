# Gym Log Handlers & Endpoint Routing

> **Feature:** `gym-logs-and-history`  
> **Phase:** `03-log-handlers-and-routes`

---

### Task 3.1: Log HTTP Handlers & Route Integration

* **Context Bundle:**
  1. [.agent/rules/01-architecture.md](file:///.agent/rules/01-architecture.md)
  2. [.agent/rules/03-testing-and-errors.md](file:///.agent/rules/03-testing-and-errors.md)
* **Owns:**
  - `internal/handler/log_handler.go`
  - `cmd/server/main.go`
* **Forbidden:**
  - `internal/repository/**`
  - `migrations/**`
* **Acceptance Criteria:**
  - **WHEN** `GET /api/v1/logs` is called with optional `startDate`, `endDate`, and `workoutType`, **THE SYSTEM SHALL** return `{ success: true, data: { logs: [...] } }`.
  - **WHEN** `POST /api/v1/logs` or `PUT /api/v1/logs/:date` is called with valid payload, **THE SYSTEM SHALL** upsert the log and return the saved log entity.
  - **WHEN** `DELETE /api/v1/logs/:date` is called, **THE SYSTEM SHALL** delete the entry and return a success envelope.
  - **WHEN** `POST /api/v1/logs/reset` is called, **THE SYSTEM SHALL** re-seed demo logs and return success confirmation.
