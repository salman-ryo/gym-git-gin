# Code Style & Layered Conventions

> **Rule ID:** `02-code-style`  
> **Applicable Globs:** `internal/handler/**`, `internal/service/**`, `internal/repository/**`, `internal/models/**`, `internal/middleware/**`

---

## 1. Layered Separation of Concerns

Every Go package in `internal/` has a distinct and strictly bounded scope:

```text
HTTP Request
     │
     ▼
[ internal/middleware ] ── (JWT Verification, Context Injection, CORS)
     │
     ▼
[ internal/handler ]    ── (Parse JSON/Params, Validate DTOs, Call Service, Send Envelope)
     │
     ▼
[ internal/service ]    ── (Business Rules, Power Score Algorithms, Validate Logic)
     │
     ▼
[ internal/repository ] ── (SQL Execution, Parameterized Queries, Scan Rows)
     │
     ▼
[ PostgreSQL / Supabase ]
```

### Layer Rules:
1. **Handlers (`internal/handler`):**
   - Must **NOT** execute SQL queries directly or import `database/sql`.
   - Extract parameters from `*gin.Context` (URL params, query strings, body JSON, and context values like `auth_user_id`).
   - Call appropriate `service` methods.
   - Format and return responses using `models.SendSuccess(c, ...)` or `models.SendError(c, ...)`.

2. **Services (`internal/service`):**
   - Must **NOT** import `github.com/gin-gonic/gin`. Keep services decoupled from the HTTP transport layer.
   - Contain all business validation, workout power score calculations, streak computations, and orchestration across multiple repositories.
   - Depend on repository interfaces defined in `interfaces.go`, enabling test mocking.

3. **Repositories (`internal/repository`):**
   - Must **NOT** perform HTTP operations or complex domain calculations.
   - Execute parameterized SQL queries (`$1, $2, ...`) against `*sql.DB`.
   - Gracefully handle `sql.ErrNoRows` and scan rows into `models` structs.

4. **Models (`internal/models`):**
   - Define data structures, DTOs, and JSON annotations (`json:"fieldName"`).
   - Maintain pure data definitions without business logic or database drivers.

---

## 2. Idiomatic Go Conventions

1. **Error Handling:**
   - Always check errors explicitly: `if err != nil { return nil, fmt.Errorf("context message: %w", err) }`.
   - Wrap errors with descriptive context using `%w`.
   - Differentiate between expected not-found states (`sql.ErrNoRows`) and unexpected system errors.

2. **Interface Segregation:**
   - Define small, focused interfaces in `interfaces.go` for repositories and services.
   - Constructors follow standard Go naming: `NewUserRepository(db *sql.DB) UserRepository`.

3. **UUID & Time Handling:**
   - Use `github.com/google/uuid` for all entity IDs and foreign keys.
   - Use `time.Time` for timestamps, serialized as RFC3339 in JSON.
   - Store and compute all timestamps in UTC.

4. **Parameterized SQL Queries:**
   - **Never** concatenate strings into SQL queries.
   - Always use Postgres positional parameters: `$1, $2, $3`.
   - Use `ON CONFLICT (...) DO UPDATE` for idempotent upsert operations.

---

## 3. Naming Conventions

| Entity Type | Convention | Example |
| :--- | :--- | :--- |
| **Go Packages** | Lowercase single-word | `handler`, `service`, `repository`, `models` |
| **Struct Types** | PascalCase | `GymLog`, `WeeklyPlan`, `AuthService` |
| **Interfaces** | PascalCase | `UserRepository`, `StatsService` |
| **Functions / Methods** | PascalCase (exported), camelCase (internal) | `UpsertLog`, `calculateConsistency` |
| **JSON Fields** | snake_case | `{"workout_type": "Push", "created_at": "..."}` |
| **DB Columns** | snake_case | `user_id`, `workout_type`, `created_at` |

---

## 4. No Ghosting Policy

* **Zero Placeholder Code:** Stubs, fake dummy returns, or comments such as `// TODO: implement logic` are forbidden.
* Every handler, service, and repository function must be fully implemented, with complete error branches, input sanitation, and database interaction.
