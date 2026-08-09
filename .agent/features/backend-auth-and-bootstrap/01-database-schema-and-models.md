# Database Schema & Model Definition: Auth & Users

> **Feature:** `backend-auth-and-bootstrap`  
> **Phase:** `01-database-schema-and-models`

---

### Task 1.1: Users & Weekly Plans Schema and Domain Models

* **Context Bundle:**
  1. [.agent/rules/01-architecture.md](file:///.agent/rules/01-architecture.md)
  2. [.agent/rules/02-code-style.md](file:///.agent/rules/02-code-style.md)
* **Owns:**
  - `migrations/000001_create_tables.up.sql`
  - `migrations/embed.go`
  - `internal/models/user.go`
  - `internal/models/plan.go`
* **Forbidden:**
  - `internal/handler/**`
  - `internal/middleware/**`
* **Acceptance Criteria:**
  - **WHEN** database migrations run, **THE SYSTEM SHALL** ensure `weekly_plans` and `users` tables exist with appropriate foreign keys and unique constraints on `auth_user_id` and `email`.
  - **WHEN** user profile data is queried, **THE SYSTEM SHALL** map database rows to `models.User` with fields `ID`, `AuthUserID`, `Email`, `Name`, `AvatarURL`, `Provider`, and `WeeklyPlanID`.
