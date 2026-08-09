# Supabase JWT Middleware Specification

> **Feature:** `backend-auth-and-bootstrap`  
> **Phase:** `02-supabase-jwt-middleware`

---

### Task 2.1: Supabase JWT Extraction & Verification Middleware

* **Context Bundle:**
  1. [.agent/rules/01-architecture.md](file:///.agent/rules/01-architecture.md)
  2. [.agent/rules/03-testing-and-errors.md](file:///.agent/rules/03-testing-and-errors.md)
* **Owns:**
  - `internal/middleware/auth.go`
  - `internal/middleware/auth_test.go`
* **Forbidden:**
  - `internal/repository/**`
  - `internal/service/**`
* **Acceptance Criteria:**
  - **WHEN** an HTTP request is received, **THE SYSTEM SHALL** extract the JWT token from `Authorization: Bearer <token>` or cookie values.
  - **WHEN** validating the token, **THE SYSTEM SHALL** verify signatures using cached JWKS (RS256/ES256) or HMAC secret (HS256) with leeway.
  - **WHEN** the token is valid, **THE SYSTEM SHALL** inject `auth_user_id`, `user_email`, `user_name`, and `user_avatar_url` into the Gin context.
  - **WHEN** the token is missing or invalid, **THE SYSTEM SHALL** abort with HTTP 401 and an `UNAUTHORIZED` error envelope.
