# Gym-Git Complete Specification

## 1. Core Architecture
- **Web Auth:** `HttpOnly` secure cookies.
- **Mobile Auth:** `Bearer` access tokens in the `Authorization` header.
- **Unified Middleware:** Must extract the Supabase token from either source, verify the JWT, reject invalid/expired tokens (401), and store the user ID in the request context.

## 2. Database Migrations (Postgres)
Write raw SQL migrations (or GORM auto-migrations) for:
1. **users** (App Profile): `id` (UUID PK), `auth_user_id` (UUID, Unique - maps to Supabase), `email` (Unique), `name`, `avatar_url`, `provider`, `weekly_plan_id` (FK to plans), `created_at`, `updated_at`.
2. **weekly_plans**: `id` (String PK), `name`, `description`, `categories` (JSONB).
3. **gym_logs**: `id` (UUID PK), `user_id` (FK to users), `date` (Date, YYYY-MM-DD), `hours` (Numeric), `workout_type` (String), `notes` (Text). 
   - *Indexes:* Composite Unique Index on `(user_id, date)` and Index on `(user_id, date DESC)`.

## 3. API Endpoints
Base URL: `/api/v1`

**Auth & Profile:**
- `GET /health` : Public 200 OK.
- `POST /auth/bootstrap` : Idempotent profile creation. If auth token is valid but no user exists in `users` table, create it. Returns profile.
- `GET /auth/me` : Returns user identity + profile + active weekly plan.
- `PUT /auth/plan` : Updates selected weekly plan.
- `POST /auth/logout` : Clears HttpOnly cookie.

**Gym Logs:**
- `GET /logs` : Query params: `startDate`, `endDate`, `workoutType`.
- `POST /logs` (or PUT `/logs/:date`) : If log for `date` exists, update it. If `hours <= 0`, delete the entry.
- `DELETE /logs/:date` : Deletes log.
- `POST /logs/reset` : Generates 365 days of demo historical data.

**Plans:**
- `GET /plans` : Returns available plans (PPL, Upper/Lower, Full Body).

**Analytics (The Core Logic):**
- `GET /stats` : Dashboard stats.
- `GET /stats/power` : Gym Power Score & Anime Tier breakdown based on a `days` window.

## 4. Scientific Business Logic & Algorithms (Implement in Services)

**A. Scientific Streak**
- Uses a rolling 7-day window. Rest days don't break the streak if frequency is met.
- $$ \text{targetDaysPerWeek} = \min(6, \max(3, |\text{activePlanCategories}|)) $$
- Date $D$ is Compliant if: Logged hours > 0, OR active sessions in window $[D-6, D]$ >= $\max(2, \text{targetDaysPerWeek} - 1)$.
- **Current Streak:** Count consecutive compliant days backward from Today.
- **Compliance Rate:** $$ \text{ComplianceRate} = \text{Round}\left( \frac{\text{TotalCompliantDays}}{\text{TotalTrackedDays}} \times 100 \right) $$

**B. Gym Power Score (0–100)**
$$ \text{PowerScore} = \text{Consistency} + \text{DurationQuality} + \text{Variety} + \text{Momentum} $$
- **Consistency (0-45 pts):** Ratio = `min(1.0, activeDays / targetActiveDays)`. Score = `Round(Ratio * 45)`.
- **Duration Quality (0-25 pts):** Optimal is 0.75h to 1.75h (Multiplier 1.0). Over 1.75h: `max(0.4, 1.0 - (hours - 1.75) * 0.25)`. Under 0.75h: `max(0.2, hours / 0.75)`. Score = `Round(AvgSessionQuality * 25)`.
- **Variety (0-20 pts):** Ratio = `min(1.0, uniqueWorkoutTypesCount / 3)`. Score = `Round(Ratio * 20)`.
- **Momentum (0-10 pts):** Ratio = `min(1.0, activeDays / (periodTotalDays * 0.5))`. Score = `Round(Ratio * 10)`.

**C. Anime Tier Mapping (Return this object in /stats/power)**
- 0-14 (D): Aqua (Konosuba) - Useless Goddess
- 15-34 (C): Mumen Rider (One Punch Man) - Class-C Hero of Justice
- 35-54 (B): Tanjiro Kamado (Demon Slayer) - Water Breathing Swordsman
- 55-69 (A): Izuku Midoriya (My Hero Academia) - One For All Successor
- 70-84 (S): Monkey D. Luffy (One Piece) - Gear 5 Sun God Nika
- 85-94 (S+): Satoru Gojo (Jujutsu Kaisen) - The Honored One
- 95-100 (SS): Saitama (One Punch Man) - One Punch God