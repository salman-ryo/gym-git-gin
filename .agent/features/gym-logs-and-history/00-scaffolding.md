# Scaffolding & Database Schema: Gym Logs & Workout History

> **Feature:** `gym-logs-and-history`  
> **Phase:** `00-scaffolding`

---

### Task 0.1: Gym Logs Schema & Model Structure

* **Context Bundle:**
  1. [.agent/rules/01-architecture.md](file:///.agent/rules/01-architecture.md)
  2. [.agent/rules/02-code-style.md](file:///.agent/rules/02-code-style.md)
* **Owns:**
  - `migrations/000001_create_tables.up.sql`
  - `internal/models/log.go`
* **Forbidden:**
  - `internal/service/**`
  - `internal/handler/**`
* **Acceptance Criteria:**
  - **WHEN** migrations run, **THE SYSTEM SHALL** ensure `gym_logs` table exists with columns `id`, `user_id`, `date`, `hours`, `workout_type`, `notes`, `created_at`, `updated_at`.
  - **WHEN** constraints are created, **THE SYSTEM SHALL** enforce a unique constraint on `(user_id, date)` and index on `(user_id, date DESC)`.
  - **WHEN** logs are mapped to structs, **THE SYSTEM SHALL** use `models.GymLog` with proper JSON tags.
