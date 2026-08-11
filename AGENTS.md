# Gym-Git Backend Agent Canonical Entry Point

> **Central Context & Router for AI Software Engineers, Claude Code & Coding Agents**  
> All agents operating on the Gym-Git Go Backend must start here and follow the [Master Operating Procedures](#master-operating-procedures).

---

## 1. Fast Index & Router
* **Central Index & Router:** [.agent/00-INDEX.md](file:///.agent/00-INDEX.md) (or [.agent/rules/00-INDEX.md](file:///.agent/rules/00-INDEX.md))
* **Architecture & Stack Spec:** [.agent/rules/01-architecture.md](file:///.agent/rules/01-architecture.md)
* **Code Style & Layered Conventions:** [.agent/rules/02-code-style.md](file:///.agent/rules/02-code-style.md)
* **Testing, Errors & Logging:** [.agent/rules/03-testing-and-errors.md](file:///.agent/rules/03-testing-and-errors.md)

### Active Feature Directories & State Trackers
1. **Backend Auth & User Bootstrap:** [.agent/features/backend-auth-and-bootstrap/STATE.md](file:///.agent/features/backend-auth-and-bootstrap/STATE.md)
2. **Gym Logs & Workout History:** [.agent/features/gym-logs-and-history/STATE.md](file:///.agent/features/gym-logs-and-history/STATE.md)
3. **Analytics & Scientific Power Stats:** [.agent/features/analytics-and-power-stats/STATE.md](file:///.agent/features/analytics-and-power-stats/STATE.md)
4. **Weekly Plans & Workout Splits:** [.agent/features/weekly-plans/STATE.md](file:///.agent/features/weekly-plans/STATE.md)

---

## 2. Tooling & Development Commands

* **Go Version:** Go 1.22+
* **Live Reload Dev Server:** `air` (or `air -c .air.toml`)
* **Standard Run:** `go run ./cmd/server/main.go` (Server runs on `http://localhost:8080`)
* **Build Binary:** `go build -o ./bin/server ./cmd/server/main.go`
* **Run Tests:** `go test -v ./...`
* **Dependency Sync:** `go mod tidy`
* **Terminal** You are using Windows terminal so make sure any other commands are suitably formatted to work on it.

---

## 3. Master Operating Procedures (MOP)

Every agent operating in this repository **must strictly abide** by the following four core rules:

### 1. Context Read (Context Economy)
* At the start of ANY task, read **ONLY** [.agent/rules/00-INDEX.md](file:///.agent/rules/00-INDEX.md) and the target feature's `STATE.md` file.
* Identify the exact phase file and the 2–3 required context files from the task's **Context Bundle**.
* **Do NOT** blindly load entire specs, large directories, or unrelated feature packages into context.

### 2. No Ghosting
* **Never** leave logic unimplemented or insert placeholder comments like `// TODO: implement logic`, `// left as exercise`, or stubbed dummy return values.
* Deliver complete, production-grade, working Go code with proper error handling (`if err != nil`), context propagation, and boundary validation.

### 3. Circuit Breaker Rule
* If a task, build, or test fix fails **3 consecutive times**, **STOP IMMEDIATELY**.
* Revert code back to the last working Git commit/state.
* Document the blocker, reproduction steps, and root cause in the target feature's `STATE.md`.
* Ask the human engineer for architectural guidance.

### 4. Post-Execution Hook (MANDATORY)
* Before completing any task or ending your turn, **AUTONOMOUSLY** open the target feature's `STATE.md`.
* Mark finished tasks as `[x]`.
* Append a timestamped entry to the **Execution Log** detailing modified files, verification commands executed, and current status.

---

## 4. Project Context Summary

| Area | Choice / Standard |
| :--- | :--- |
| **Language / Runtime** | Go 1.22+ |
| **Web Framework** | Gin Web Framework (`github.com/gin-gonic/gin v1.10.0`) |
| **Database** | PostgreSQL (hosted on Supabase) via `database/sql` + `github.com/lib/pq` |
| **Authentication** | Supabase Auth (JWT verification supporting RS256/ES256 JWKS & HS256) |
| **Auth Transport** | Dual support: `Authorization: Bearer <token>` (Mobile/Client) & HttpOnly cookie (Web) |
| **Layered Architecture** | Modular separation: `/cmd`, `/config`, `/internal/handler`, `/internal/service`, `/internal/repository`, `/internal/middleware`, `/internal/models` |
| **Success Envelope** | `{ "success": true, "data": { ... }, "message": "..." }` |
| **Error Envelope** | `{ "success": false, "error": { "code": "...", "message": "...", "details": [...] }, "timestamp": "..." }` |
| **Live Reloading** | Air (`.air.toml`) |