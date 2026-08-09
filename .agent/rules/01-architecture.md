# Architecture & Infrastructure Specification

> **Rule ID:** `01-architecture`  
> **Applicable Globs:** `cmd/**`, `config/**`, `internal/database/**`, `internal/middleware/**`, `migrations/**`, `.air.toml`

---

## 1. Technology Stack

| Layer | Technology | Details |
| :--- | :--- | :--- |
| **Language / Runtime** | Go 1.22+ | Standard library, idiomatic Go conventions, strict nil and error checks |
| **Web Framework** | Gin (`github.com/gin-gonic/gin v1.10.0`) | High-performance HTTP routing, middleware pipeline, JSON binding |
| **Database** | PostgreSQL (Supabase Hosted) | Managed via `database/sql` + `github.com/lib/pq` repository layer |
| **Authentication** | Supabase Auth (JWT Verification) | Dynamic RS256 / ES256 JWKS validation with fallback to HS256 HMAC |
| **Dev Tooling** | Air (`.air.toml`) | Live-reloading daemon for fast local Go development |
| **Migrations** | Embedded SQL (`embed.go`) | Idempotent DDL scripts embedded into Go binary for automatic execution on startup |

---

## 2. Layered Modular Architecture

The backend strictly adheres to a clean, decoupled layered architecture:

```text
backend/
├── cmd/
│   └── server/
│       └── main.go          # Dependency injection, router wiring & server boot
├── config/
│   └── config.go            # Environment variable loading & validation
├── migrations/
│   ├── 000001_create_tables.up.sql   # DDL table creation & initial seed data
│   ├── 000001_create_tables.down.sql # DDL rollback script
│   └── embed.go             # Embeds SQL migrations into Go binary
├── internal/
│   ├── database/
│   │   └── database.go      # PostgreSQL pool setup & migration runner
│   ├── models/              # Domain entities, DTOs & response envelopes
│   │   ├── user.go          # User app profile model
│   │   ├── plan.go          # Weekly workout plan entity
│   │   ├── log.go           # Daily workout log entity
│   │   ├── stats.go         # Contribution matrix, streak & power score DTOs
│   │   └── response.go      # Standardized JSON response envelope helpers
│   ├── repository/          # Database queries & SQL execution
│   │   ├── interfaces.go    # UserRepository, PlanRepository, GymLogRepository interfaces
│   │   ├── user_repository.go
│   │   ├── plan_repository.go
│   │   └── log_repository.go
│   ├── service/             # Pure business logic, analytics & orchestration
│   │   ├── interfaces.go    # AuthService, PlanService, GymLogService, StatsService interfaces
│   │   ├── auth_service.go
│   │   ├── plan_service.go
│   │   ├── log_service.go
│   │   ├── stats_service.go
│   │   └── stats_calculator.go # Pure algorithms (Power Score, Streaks, Heatmap)
│   ├── handler/             # HTTP controllers, binding & envelope formatting
│   │   ├── health_handler.go
│   │   ├── auth_handler.go
│   │   ├── plan_handler.go
│   │   ├── log_handler.go
│   │   └── stats_handler.go
│   └── middleware/          # HTTP interceptors
│       ├── auth.go          # Supabase JWT validation & user context injection
│       ├── cors.go          # Cross-Origin Resource Sharing configuration
│       └── auth_test.go     # Unit tests for JWT parsing & validation
```

---

## 3. Environment Variables & Configuration

All environment configuration is loaded via [config/config.go](file:///config/config.go) using `godotenv`:

* `PORT`: Listening HTTP port (Default: `8080`).
* `DATABASE_URL`: Full PostgreSQL connection string (`postgresql://postgres:[password]@db.[ref].supabase.co:5432/postgres?sslmode=require`).
* `SUPABASE_JWT_SECRET`: JWT secret for HMAC token verification.
* `SUPABASE_JWKS_URL`: Supabase JWKS endpoint for asymmetric RS256/ES256 signature verification (e.g., `https://[ref].supabase.co/auth/v1/.well-known/jwks.json`).
* `GIN_MODE`: Gin runtime mode (`debug` or `release`).
* `ALLOWED_ORIGINS`: Comma-separated CORS allowed origins (e.g., `http://localhost:3000,http://127.0.0.1:3000`).

---

## 4. Authentication & Security Rules

### A. Authentication Contract
1. **NEVER** implement password hashing, custom salt generation, or local sign-up/login tables.
2. Rely strictly on verifying **Supabase JWTs** issued by the client application.
3. The auth middleware ([internal/middleware/auth.go](file:///internal/middleware/auth.go)) must seamlessly extract JWTs from:
   - `Authorization: Bearer <token>` HTTP header (Mobile / CLI / API clients).
   - `sb-access-token` / `access_token` cookies (Web clients).
4. On successful validation, the middleware extracts the `sub` claim (as `uuid.UUID`) and user metadata, injecting them into the Gin context (`c.Set("auth_user_id", ...)`).

### B. User Provisioning & Bootstrap Handshake
* When a user logs in on the frontend, they call `POST /api/v1/auth/bootstrap` with their desired `selectedPlanId`.
* The Go backend idempotently verifies or creates the user record in the `users` table, linking `auth_user_id` to Supabase's identity and setting default weekly plan presets.

---

## 5. Standard API Response Envelopes

All endpoints **MUST** return standardized JSON envelopes matching [internal/models/response.go](file:///internal/models/response.go):

### Success Response (HTTP 200 / 201)
```json
{
  "success": true,
  "data": {
    "user": { ... }
  },
  "message": "User bootstrapped successfully"
}
```

### Error Response (HTTP 4xx / 5xx)
```json
{
  "success": false,
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Invalid or expired authentication token",
    "details": ["token is expired by 2m"]
  },
  "timestamp": "2026-08-09T07:15:00Z"
}
```

---

## 6. Database Schema & Pooling Guidelines

* **Connection Pool:** Configured in [internal/database/database.go](file:///internal/database/database.go) with `MaxOpenConns(25)`, `MaxIdleConns(5)`, and `ConnMaxLifetime(15m)`.
* **Primary Tables:**
  - `weekly_plans`: Presets (`ppl-standard`, `ppl-core`, etc.) and user-custom routines.
  - `users`: App profile table mapping `auth_user_id` to Supabase Auth `sub`.
  - `gym_logs`: Daily workout log entries with `UNIQUE (user_id, date)` index.
* **Idempotent Migrations:** Embedded SQL in [migrations/embed.go](file:///migrations/embed.go) automatically runs on server boot, ensuring tables and seeds exist.
