# Plan Handlers & Endpoint Routing

> **Feature:** `weekly-plans`  
> **Phase:** `02-plan-handlers-and-endpoints`

---

### Task 2.1: Plan HTTP Handlers & Route Registration

* **Context Bundle:**
  1. [.agent/rules/01-architecture.md](file:///.agent/rules/01-architecture.md)
  2. [.agent/rules/03-testing-and-errors.md](file:///.agent/rules/03-testing-and-errors.md)
* **Owns:**
  - `internal/handler/plan_handler.go`
  - `cmd/server/main.go`
* **Forbidden:**
  - `internal/repository/**`
  - `migrations/**`
* **Acceptance Criteria:**
  - **WHEN** `GET /api/v1/plans` is called, **THE SYSTEM SHALL** return all available weekly plans in a `{ success: true, data: { plans: [...] } }` JSON envelope.
  - **WHEN** custom plans are requested, **THE SYSTEM SHALL** format category lists accurately as JSON arrays.
