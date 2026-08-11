# Gym-Git Streak Engine & Gamification Architecture Plan (CHANGES.md)

> **Document Version:** 1.0.0  
> **Scope:** Complete architectural blueprint for timezone-aware streaks, 7-day plan cycles, rest tokens, inventory power-ups, sickness freeze vault, and PR tracking.

---

## 1. Executive Summary & Core Philosophies

Gym-Git transforms gym consistency tracking by moving beyond simplistic daily calendar streaks. Fitness requires physiological recovery, flexibility for real-world disruptions (sickness, injury, travel), and robust timezone awareness for global users.

```text
                                  ┌──────────────────────────────┐
                                  │      Incoming Action         │
                                  └──────────────┬───────────────┘
                                                 │
                        ┌────────────────────────┴────────────────────────┐
                        ▼                                                 ▼
             [ Log Today's Workout ]                           [ Historical Log Edit ]
                        │                                                 │
                        ▼                                                 ▼
         • Check Timezone (Wall Clock)                     • Target Date < User Today
         • Check 7-Day Cycle & Rest Tokens                 • Update Historical Volume/Stats
         • Evaluate Split Accuracy Score                   • STREAK UNCHANGED (No Cheating)
         • Increment Streak (+1)                                          │
         • Check Milestone Rewards (7d, 14d, etc.)                        ▼
                                                           [ Restore Shield Power-Up? ]
                                                                          │
                                                           ├── YES (<=3 days) ──► Consume Item + Restore Streak
                                                           └── NO             ──► Volume Only
```

---

## 2. Phase-by-Phase Implementation Roadmap

### Phase 1: Global Timezone Engine & Grace Period Anchor
* **Goal:** Eliminate UTC midnight bugs and provide accurate, local-time streak evaluation for users anywhere in the world.
* **Key Components:**
  1. **User Profile Timezone:** Store IANA timezone identifier (e.g. `America/Los_Angeles`, `Asia/Kolkata`, `Europe/London`) in `users.timezone`.
  2. **Context Timezone Resolver:** Request headers (`X-Timezone` or cookie) + database profile fallback.
  3. **Local Wall-Clock Definition of "Today":**
     - Today runs from `00:00:00` to `23:59:59` in the user's localized timezone.
     - A day is NEVER considered missed until `00:00:00` of the following day arrives locally.
  4. **Strict Historical Log Segregation:**
     - **Today’s Log (`date == user_today`):** Updates volume metrics AND increments/verifies current streak.
     - **Historical Log (`date < user_today`):** Updates volume metrics and heatmap history ONLY. Never revives a dead streak unless a Restore Shield power-up is explicitly redeemed.

---

### Phase 2: 7-Day Plan Cycles, Rest Tokens & Accuracy Scoring
* **Goal:** Allow complete schedule flexibility within a 7-day weekly split while rewarding discipline and proper rest management.
* **Key Components:**
  1. **Fixed 7-Day Plan Cycles:**
     - A user's plan runs in discrete 7-day cycle windows relative to their start date.
     - A 4-day workout plan (e.g., Upper/Lower or PPL 4-day) provides **4 Workout Targets** and **3 Rest Tokens** per cycle.
  2. **Rest Token Mechanics:**
     - Rest days consume Rest Tokens.
     - A user can take rest days consecutively or staggered (e.g., 3 rest days in a row) without breaking their streak, provided they complete their 4 workout sessions within the 7-day cycle.
     - Missing a workout after Rest Tokens are exhausted breaks the streak.
  3. **Workout Split Accuracy Score (0-100%):**
     - **100% Accuracy:** Followed the exact scheduled order (e.g., Push on Day 1, Pull on Day 2, Legs on Day 3, Rest on Day 4). This can be used in power level scoring
     - **Partial Deduction:** Swapped workout days or took rest days out of sequence (e.g., did Legs on Push day). Streak is **preserved**, but the Accuracy Score reflects the deviation.
  4. **Queued Plan Changes (No Mid-Cycle Disruption):**
     - Users can change their plan at any time.
     - The change is queued (`queued_weekly_plan_id`) and takes effect strictly at the beginning of the next 7-day cycle.
     - Multiple changes during the week overwrite the queue; only the latest selected plan activates for the next week.
     - Past streak history is protected and never retroactively invalidated.

---

### Phase 3: Gamified Inventory & Milestone Power-Up System
* **Goal:** Reward consistent habit formation with collectible power-up items that give users fair recovery mechanics.
* **Key Components:**
  1. **User Inventory System:**
     - Supports items: `RESTORE_SHIELD`, `STREAK_FREEZE_TOKEN`, `XP_BOOST`, `ACCURACY_CHARM`.
  2. **Streak Milestone Engine:**
     - Automatic item and badge grants on achieving streak milestones (7, 14, 21, 30, 60, 90, 180, 365 days).
     - Reaching a 7-day streak grants a **Restore Shield**.
  3. **Restore Shield Mechanics:**
     - Usable only on historical dates within a 3-day lookback window (`Yesterday`, `2 days ago`, `3 days ago`).
     - Consumes 1 `RESTORE_SHIELD` item from inventory.
     - Validates workout entry for that missed day, recalculates the break, and restores the active streak.
     - Dates older than 3 days cannot be restored under any circumstances.

---

### Phase 4: Sickness & Injury Freeze Vault ("Ice Pause")
* **Goal:** Prevent streak loss during medical emergencies, injury recovery, or severe illness without requiring manual day-by-day maintenance.
* **Key Components:**
  1. **Freeze Activation & Token Duration:**
     - User triggers: *"I'm sick/injured, freeze my streak"*.
     - Optional metadata: Reason (e.g., `"Tore a muscle"`, `"Severe flu"`), expected return window. We can have freeze tokens/items each lasting one day, so for a 3-day freeze the user consumes 3 tokens/items from inventory. Once back, the system checks if the freeze window was exceeded. Staying away past the allocated freeze duration will break the streak.
     - Streak status immediately becomes `"PAUSED_IN_ICE"`.
  2. **Explicit Manual Unfreeze or Expiration Only (No Automatic Unfreeze on Visit/Login):**
     - **No Auto-Unfreeze on App Access / Login:** Coming to the app, logging in, or checking profile data does **NOT** automatically turn off or break the freeze. Users who are sick often log in to inspect their stats or data without working out, so auto-unfreezing on visit is forbidden.
     - **Termination Criteria:** The freeze ends **ONLY** if:
       a) **Expiration:** The freeze token duration or set freeze period expires.
       b) **Manual Unfreeze:** The user explicitly turns off / ends the freeze power (e.g., via `POST /api/v1/streak/unfreeze` or by manually unfreezing).
  3. **Git-Style Heatmap Visual Differentiation:**
     - Frozen days are visually distinct on the contribution graph (e.g., icy blue `#38bdf8` / frost pattern).
     - Tooltip displays freeze duration and user-provided reason.

---

### Phase 5: Streak Break Detection & Lifecycle Events
* **Goal:** Provide clear, motivating feedback when a streak breaks and alert users before it happens.
* **Key Components:**
  1. **Streak Broken Lifecycle Event:**
     - Evaluated upon session bootstrap (`GET /api/v1/auth/me` or login).
     - If the local midnight deadline passed with exhausted rest tokens and no workout:
       - Current streak resets to 0.
       - Returns `streak_broken_event: { previous_streak: 28, broken_on: "2026-08-08", restore_shield_available: true }`.
       - Frontend displays a dedicated `"Streak Broken"` modal offering instant Restore Shield redemption if available in inventory.
  2. **Future Scope — Proactive Alerts:**
     - Push notifications 3 hours and 1 hour before local midnight if today's workout target is unmet.

---

### Phase 6: Future Feature — Personal Records (PR) & Strength Analytics
* **Goal:** Track progressive overload, 1-Rep Max (1RM) records, and exercise volume milestones.
* **Key Components:**
  1. **PR Tracking Engine:**
     - Automatic 1RM estimation via Epley formula: $\text{1RM} = \text{Weight} \times (1 + \frac{\text{Reps}}{30})$.
     - All-time and period-specific PRs for core movements (Bench Press, Squat, Deadlift, Overhead Press, Pull-Ups).
  2. **PR Celebration Feats:**
     - Power Score bonus when hitting verified PRs during a compliant weekly cycle.

---

## 3. Database Schema Blueprint (Draft DDL)

```sql
-- 1. Extend Users table with Timezone & Queued Plan
ALTER TABLE users ADD COLUMN IF NOT EXISTS timezone VARCHAR(100) DEFAULT 'UTC';
ALTER TABLE users ADD COLUMN IF NOT EXISTS queued_weekly_plan_id VARCHAR(50) REFERENCES weekly_plans(id);

-- 2. User Streak State & 7-Day Cycle Table
CREATE TABLE IF NOT EXISTS user_streak_states (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    current_streak INT NOT NULL DEFAULT 0,
    longest_streak INT NOT NULL DEFAULT 0,
    cycle_start_date DATE NOT NULL,
    cycle_end_date DATE NOT NULL,
    workouts_completed_in_cycle INT NOT NULL DEFAULT 0,
    workouts_target_in_cycle INT NOT NULL DEFAULT 4,
    rest_tokens_total INT NOT NULL DEFAULT 3,
    rest_tokens_used INT NOT NULL DEFAULT 0,
    accuracy_score INT NOT NULL DEFAULT 100,
    last_logged_date DATE,
    is_frozen BOOLEAN NOT NULL DEFAULT FALSE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 3. Gamification Item Inventory
CREATE TABLE IF NOT EXISTS user_inventories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    item_type VARCHAR(50) NOT NULL, -- 'RESTORE_SHIELD', 'STREAK_FREEZE', 'ACCURACY_CHARM'
    quantity INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, item_type)
);

-- 4. Sickness / Injury Streak Freezes
CREATE TABLE IF NOT EXISTS streak_freezes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    start_date DATE NOT NULL,
    end_date DATE, -- NULL while active; set when freeze naturally expires or is manually turned off by user
    reason TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 5. Future Scope: Exercise Library & Personal Records
CREATE TABLE IF NOT EXISTS personal_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    exercise_name VARCHAR(150) NOT NULL,
    weight_kg NUMERIC(6,2) NOT NULL,
    reps INT NOT NULL,
    estimated_1rm NUMERIC(6,2) NOT NULL,
    achieved_at DATE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

---

## 4. REST API Contract Extensions

| Method & Endpoint | Payload / Params | Description |
| :--- | :--- | :--- |
| `POST /api/v1/auth/timezone` | `{ "timezone": "America/New_York" }` | Updates user's localized IANA timezone |
| `GET /api/v1/streak` | None | Returns active streak, cycle progress, rest tokens left, freeze state, and broken event |
| `POST /api/v1/streak/freeze` | `{ "reason": "Tore a muscle", "duration_days": 3 }` | Activates sickness/injury freeze vault using freeze token(s) |
| `POST /api/v1/streak/unfreeze` | None | Manually turns off an active freeze power before natural expiration |
| `POST /api/v1/streak/restore` | `{ "target_date": "2026-08-08", "log_id": "uuid" }` | Consumes 1 Restore Shield to revive streak from missed date (up to 3 days ago) |
| `GET /api/v1/inventory` | None | Lists user's power-ups, shields, and items |
| `PUT /api/v1/plans/queue` | `{ "weekly_plan_id": "ppl-core" }` | Queues plan change to activate on next 7-day cycle |
