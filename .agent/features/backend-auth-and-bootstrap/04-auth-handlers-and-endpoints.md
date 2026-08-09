# Auth Handlers & Route Integration

> **Feature:** `backend-auth-and-bootstrap`  
> **Phase:** `04-auth-handlers-and-endpoints`

---

### Task 4.1: Auth HTTP Handlers & Route Registration

* **Context Bundle:**
  1. [.agent/rules/01-architecture.md](file:///.agent/rules/01-architecture.md)
  2. [.agent/rules/03-testing-and-errors.md](file:///.agent/rules/03-testing-and-errors.md)
* **Owns:**
  - `internal/handler/auth_handler.go`
  - `cmd/server/main.go`
* **Forbidden:**
  - `internal/repository/**`
  - `migrations/**`
* **Acceptance Criteria:**
  - **WHEN** `POST /api/v1/auth/bootstrap` is called with `{ selectedPlanId: "ppl-standard" }` by an authenticated user, **THE SYSTEM SHALL** return HTTP 200/201 with `{ success: true, data: { user: ... } }`.
  - **WHEN** `GET /api/v1/auth/me` is called, **THE SYSTEM SHALL** return the authenticated user's profile and active plan.
  - **WHEN** `PUT /api/v1/auth/plan` is called with a new `plan_id` or custom plan details, **THE SYSTEM SHALL** update the user's plan and return the updated user object.
  - **WHEN** `POST /api/v1/auth/logout` is called, **THE SYSTEM SHALL** return a success envelope confirming session termination.
