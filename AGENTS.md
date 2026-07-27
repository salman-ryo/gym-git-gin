# Gym-Git Backend System Rules

You are an expert Go backend developer. Follow these core rules for every file you generate:

1. **Tech Stack:** 
   - Language/Framework: Go with Gin.
   - Database: PostgreSQL (hosted on Supabase) accessed via standard Go repository patterns (e.g., `database/sql` + `lib/pq` or `pgx`, or `gorm` if preferred, but keep it clean).
   - Auth: Supabase Auth (JWT verification).

2. **Architecture:** 
   - Use a modular layered structure: `/cmd`, `/config`, `/internal/handler`, `/internal/service`, `/internal/repository`, `/internal/middleware`, `/internal/models`.
   - Keep business logic in services, HTTP parsing in handlers, and DB queries in repositories.

3. **Authentication Rule:** 
   - NEVER implement password hashing or local login flows. 
   - Rely strictly on verifying Supabase JWTs. 
   - The auth middleware must seamlessly accept BOTH an `HttpOnly` cookie (for web) AND an `Authorization: Bearer <token>` header (for mobile).

4. **Response Envelope:** 
   - All HTTP 200/201 responses MUST follow: `{ "success": true, "data": { ... }, "message": "..." }`
   - All HTTP 4xx/5xx responses MUST follow: `{ "success": false, "error": { "code": "...", "message": "...", "details": [...] }, "timestamp": "..." }`