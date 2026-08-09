# Scaffolding & Environment Setup: Auth & Bootstrap

> **Feature:** `backend-auth-and-bootstrap`  
> **Phase:** `00-scaffolding`

---

### Task 0.1: Configuration Loading & PostgreSQL Connection Pooling

* **Context Bundle:**
  1. [.agent/rules/01-architecture.md](file:///.agent/rules/01-architecture.md)
  2. [.agent/rules/02-code-style.md](file:///.agent/rules/02-code-style.md)
* **Owns:**
  - `config/config.go`
  - `internal/database/database.go`
  - `.env` / `.env.example`
* **Forbidden:**
  - `internal/service/**`
  - `internal/handler/**`
  - `internal/models/**`
* **Acceptance Criteria:**
  - **WHEN** the server boots, **THE SYSTEM SHALL** load configuration variables (`PORT`, `DATABASE_URL`, `SUPABASE_JWT_SECRET`, `SUPABASE_JWKS_URL`, `GIN_MODE`, `ALLOWED_ORIGINS`) via `config.LoadConfig()`.
  - **WHEN** `DATABASE_URL` is provided, **THE SYSTEM SHALL** initialize a PostgreSQL pool with max 25 open connections and verify connectivity via `Ping()`.
