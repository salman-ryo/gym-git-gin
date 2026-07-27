# Gym-Git Authentication Architecture & Flow

This document details the complete authentication flow across the Next.js TypeScript frontend and Go/Gin backend using Supabase Auth, along with all files involved in the process.

---

## 1. System Architecture Diagram

```
+-----------------------------------------------------------------------------------+
|                                 WEB FRONTEND                                      |
|  1. User enters credentials -> supabase.auth.signInWithPassword()                 |
|  2. Session created in Supabase Auth                                             |
|  3. Next.js / @supabase/ssr manages HttpOnly cookies for Web                     |
|  4. apiFetch attaches Authorization: Bearer <access_token> + credentials:include  |
+----------------------------------------+------------------------------------------+
                                         |
                                         | HTTP Requests (Bearer header / Cookies)
                                         v
+-----------------------------------------------------------------------------------+
|                                 GO / GIN BACKEND                                  |
|  5. AuthMiddleware intercepts request                                             |
|  6. extractToken() checks Authorization header -> Chunked/Web Cookie fallbacks    |
|  7. cleanToken() unpacks Base64/JSON wrapped tokens safely                         |
|  8. parseSupabaseJWT() validates HMAC signature & claims via SUPABASE_JWT_SECRET  |
|  9. Extracted claims stored in Gin Context (auth_user_id UUID, email, metadata)   |
| 10. Handler executes business logic                                               |
+-----------------------------------------------------------------------------------+
```

---

## 2. All Files Involved in the Authentication Flow

### A. Frontend Files (`frontend/`)

- **[utils/supabase/client.ts](file:///c:/Users/salma/Development/Jiyu/CodingAgent/gymgit/frontend/utils/supabase/client.ts)**
  Creates the browser-side Supabase client (`createBrowserClient()`) from `@supabase/ssr` to manage client-side authentication sessions and browser cookie storage.
- **[utils/supabase/server.ts](file:///c:/Users/salma/Development/Jiyu/CodingAgent/gymgit/frontend/utils/supabase/server.ts)**
  Creates the server-side Supabase client (`createServerClient()`) for Next.js Server Components, reading and setting cookies via `next/headers`.
- **[utils/supabase/middleware.ts](file:///c:/Users/salma/Development/Jiyu/CodingAgent/gymgit/frontend/utils/supabase/middleware.ts)**
  Implements `updateSession(request)` to refresh active Supabase sessions on route navigation and enforce route protection rules.
- **[middleware.ts](file:///c:/Users/salma/Development/Jiyu/CodingAgent/gymgit/frontend/middleware.ts)**
  Next.js root middleware. Intercepts incoming requests and delegates session updates and route protection to `utils/supabase/middleware.ts`.
- **[utils/api.ts](file:///c:/Users/salma/Development/Jiyu/CodingAgent/gymgit/frontend/utils/api.ts)**
  Centralized fetch client (`apiFetch`, `api.get`, `api.post`). Retrieves active Supabase token, attaches `Authorization: Bearer <token>`, sets `credentials: 'include'`, and formats request/error envelopes.
- **[lib/auth-context.tsx](file:///c:/Users/salma/Development/Jiyu/CodingAgent/gymgit/frontend/lib/auth-context.tsx)**
  React `AuthProvider` context. Manages user state, triggers `login()`, `signup()`, `loginWithGoogle()`, `logout()`, executes backend profile bootstrapping (`bootstrapBackend`), and listens to Supabase `onAuthStateChange`.
- **[app/login/page.tsx](file:///c:/Users/salma/Development/Jiyu/CodingAgent/gymgit/frontend/app/login/page.tsx)**
  UI page component for Sign In, Sign Up, and Google OAuth triggers.

---

### B. Backend Files (`backend/`)

- **[config/config.go](file:///c:/Users/salma/Development/Jiyu/CodingAgent/gymgit/backend/config/config.go)**
  Loads application environment variables (`SUPABASE_JWT_SECRET`, `ALLOWED_ORIGINS`, `DATABASE_URL`, `PORT`).
- **[cmd/server/main.go](file:///c:/Users/salma/Development/Jiyu/CodingAgent/gymgit/backend/cmd/server/main.go)**
  Main entrypoint. Initializes database connections, repositories, services, handlers, configures CORS middleware, and attaches `AuthMiddleware` to protected route groups (`/auth`, `/logs`, `/stats`).
- **[internal/middleware/cors.go](file:///c:/Users/salma/Development/Jiyu/CodingAgent/gymgit/backend/internal/middleware/cors.go)**
  CORS middleware. Validates origin, sets `Access-Control-Allow-Origin`, `Access-Control-Allow-Credentials: true`, and `Access-Control-Allow-Headers` (including `Authorization`).
- **[internal/middleware/auth.go](file:///c:/Users/salma/Development/Jiyu/CodingAgent/gymgit/backend/internal/middleware/auth.go)**
  Auth middleware (`AuthMiddleware`, `extractToken`, `cleanToken`, `parseSupabaseJWT`). Extracts tokens from Bearer headers or HttpOnly cookies, cleans/decodes Base64/JSON tokens, validates JWT claims, and injects user context into Gin.
- **[internal/handler/auth_handler.go](file:///c:/Users/salma/Development/Jiyu/CodingAgent/gymgit/backend/internal/handler/auth_handler.go)**
  HTTP handlers for auth routes:
  - `POST /auth/bootstrap`: Idempotently creates user profile in Postgres `users` table upon initial login.
  - `GET /auth/me`: Returns user profile and active weekly plan.
  - `PUT /auth/plan`: Updates user's active weekly plan.
  - `POST /auth/logout`: Expires and clears all authentication cookies.
- **[internal/service/auth_service.go](file:///c:/Users/salma/Development/Jiyu/CodingAgent/gymgit/backend/internal/service/auth_service.go)**
  Business service layer for user profile bootstrapping, profile queries, and weekly plan updates.
- **[internal/repository/user_repository.go](file:///c:/Users/salma/Development/Jiyu/CodingAgent/gymgit/backend/internal/repository/user_repository.go)**
  Database layer performing SQL queries (`SELECT`, `INSERT`, `UPDATE`) against the PostgreSQL `users` table.

---

## 3. Client Authentication Protocols

### A. Web Client (Next.js + `@supabase/ssr`)
- **Session Persistence:** `@supabase/ssr` stores session tokens in HttpOnly cookies (`sb-<project-ref>-auth-token`, chunked `.0`, `.1`) for seamless Next.js SSR middleware session refreshing.
- **API Requests:** `apiFetch` attaches `Authorization: Bearer <access_token>` in headers and passes `credentials: 'include'` for CORS cookie transmission to the Go backend.
- **Bootstrap Requirement:** Immediately after successful login/signup, the frontend calls `POST /api/v1/auth/bootstrap` to ensure the application user profile exists in Postgres `users` table.

### B. Mobile / External API Clients
- **Session Persistence:** Mobile apps (React Native, iOS, Android, cURL) store the Supabase access token in secure native storage.
- **API Requests:** Mobile apps send `Authorization: Bearer <access_token>` directly in HTTP headers.

---

## 4. End-to-End Authentication Sequences

### Sign In / Sign Up Flow:
1. Client calls `supabase.auth.signInWithPassword()` or `supabase.auth.signUp()`.
2. Supabase Auth returns session containing `access_token` and `user` object.
3. Client executes `POST /api/v1/auth/bootstrap` with `{ selectedPlanId }` and Bearer token header.
4. Go backend validates JWT, creates or verifies user profile in `users` table, and returns user profile.
5. Client calls `GET /api/v1/auth/me` to load full user profile and active weekly plan.
6. Client redirects user to `/dashboard`.

### Logout Flow:
1. Client calls `supabase.auth.signOut()`.
2. Client calls `POST /api/v1/auth/logout`.
3. Backend clears all active auth cookies (`sb-access-token`, `sb-<project-ref>-auth-token`, etc.) with `MaxAge: -1`.
4. Client clears local React state and redirects to `/login`.

---

## 5. Backend Authentication Middleware Details (`AuthMiddleware`)

The Go middleware ([internal/middleware/auth.go](file:///c:/Users/salma/Development/Jiyu/CodingAgent/gymgit/backend/internal/middleware/auth.go)) handles token extraction, cleaning, decoding, and JWT validation.

### Step 1: Token Extraction (`extractToken`)
`extractToken()` searches incoming requests in order of priority:
1. **Header Authorization:** Checks `Authorization: Bearer <token>`.
2. **Chunked Cookies:** Concatenates `@supabase/ssr` chunked cookies (`auth-token.0`, `auth-token.1`, etc.).
3. **Explicit & Dynamic Cookies:** Searches for cookies named `sb-access-token`, `access_token`, or matching `sb-*-auth-token`.

### Step 2: Token Cleaning & Unpacking (`cleanToken`)
`cleanToken()` normalizes token strings:
1. **Direct JWT Pass:** If string is already a 3-part JWT (`header.payload.signature`), it returns immediately.
2. **Path Unescaping:** Applies `url.PathUnescape` to resolve `%` hex sequences while preserving `+` characters in JWT signatures.
3. **Base64 Unwrapping:** If prefixed with `base64-`, decodes payload using standard or URL-safe Base64 decoders.
4. **JSON Unpacking:** If payload is a JSON array (`["access_token", "refresh_token"]`) or object (`{"access_token": "..."}`), extracts the JWT string.

### Step 3: JWT Verification & Claims Extraction (`parseSupabaseJWT`)
1. Validates HMAC signature using `SUPABASE_JWT_SECRET` (supporting both raw string bytes and Base64-decoded secret bytes).
2. Applies `jwt.WithLeeway(5 * time.Minute)` to tolerate clock skew.
3. Parses `sub` claim to `uuid.UUID`.

### Step 4: Gin Context Injection
Injects authenticated user context into Gin request context:
- `auth_user_id`: User's Supabase UUID (`uuid.UUID`)
- `user_email`: Email address (`string`)
- `user_name`: Full Name from metadata (`string`)
- `user_avatar_url`: Avatar URL (`string`)
- `user_provider`: Auth Provider (e.g., `"email"`, `"google"`)
