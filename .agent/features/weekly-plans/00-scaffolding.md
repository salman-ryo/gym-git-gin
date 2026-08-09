# Scaffolding & Seed Catalog: Weekly Workout Plans

> **Feature:** `weekly-plans`  
> **Phase:** `00-scaffolding`

---

### Task 0.1: Weekly Plans Schema, Seed Data & Models

* **Context Bundle:**
  1. [.agent/rules/01-architecture.md](file:///.agent/rules/01-architecture.md)
  2. [.agent/rules/02-code-style.md](file:///.agent/rules/02-code-style.md)
* **Owns:**
  - `migrations/000001_create_tables.up.sql`
  - `internal/models/plan.go`
* **Forbidden:**
  - `internal/service/**`
  - `internal/handler/**`
* **Acceptance Criteria:**
  - **WHEN** database migrations execute, **THE SYSTEM SHALL** seed standard presets: `ppl-standard`, `ppl-core`, `upper-lower`, `full-body`, and `ppl`.
  - **WHEN** plans are represented in Go, **THE SYSTEM SHALL** use `models.WeeklyPlan` with fields `ID`, `UserID`, `Name`, `Description`, and `Categories` (string slice).
