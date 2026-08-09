# Plan Repository & Plan Service Implementation

> **Feature:** `weekly-plans`  
> **Phase:** `01-plan-repository-and-service`

---

### Task 1.1: PlanRepository & PlanService Logic

* **Context Bundle:**
  1. [.agent/rules/01-architecture.md](file:///.agent/rules/01-architecture.md)
  2. [.agent/rules/02-code-style.md](file:///.agent/rules/02-code-style.md)
* **Owns:**
  - `internal/repository/interfaces.go`
  - `internal/repository/plan_repository.go`
  - `internal/service/interfaces.go`
  - `internal/service/plan_service.go`
* **Forbidden:**
  - `internal/handler/**`
  - `cmd/server/main.go`
* **Acceptance Criteria:**
  - **WHEN** `GetAllPlans` is called, **THE SYSTEM SHALL** return all default preset plans plus any custom plans belonging to the authenticated user.
  - **WHEN** `GetPlanByID` is called, **THE SYSTEM SHALL** return the matching `models.WeeklyPlan` or `sql.ErrNoRows`.
  - **WHEN** `CreateCustomPlan` is called, **THE SYSTEM SHALL** insert a custom plan with `user_id` set to the creating user.
