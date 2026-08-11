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

### Phase 4: Sickness & Injury Freeze Vault, Item Usage & Dynamic Reward Roadmap System
* **Goal:** Provide item usage mechanics, Sickness/Injury Freeze state, and a dynamic Streak Reward Roadmap System presented as a visual progression roadmap on the frontend.
* **Key Components:**
  1. **Dynamic Streak Reward Roadmap Engine:**
     - Computes milestone rewards along a streak roadmap (e.g. 7-day streak $\rightarrow$ 1x Restore Shield, 10-day streak $\rightarrow$ 1x Streak Freeze Token, etc.).
     - **Dynamic Milestone Customization:** Administrators can freely insert, update, or delete milestone targets at arbitrary streak days (e.g. inserting Day 11 $\rightarrow$ +5 Restore Shields, modifying Day 30 rewards, or deleting milestones).
     - Frontend receives an ordered roadmap list via `GET /api/v1/rewards/roadmap` with status tags (`LOCKED`, `CLAIMABLE`, `CLAIMED`).
     - Users explicitly claim unlocked rewards via `POST /api/v1/rewards/claim`, adding the awarded items directly into their inventory.
  2. **Freeze Activation & Token Duration:**
     - User triggers: *"I'm sick/injured, freeze my streak"*.
     - Optional metadata: Reason (e.g., `"Tore a muscle"`, `"Severe flu"`), duration. Consumes 1 or more `STREAK_FREEZE_TOKEN`s.
     - Streak status immediately becomes `"PAUSED_IN_ICE"`.
  3. **Explicit Manual Unfreeze or Expiration Only (No Automatic Unfreeze on Visit/Login):**
     - **No Auto-Unfreeze on App Access / Login:** Coming to the app, logging in, or checking profile data does **NOT** automatically turn off or break the freeze.
     - **Termination Criteria:** The freeze ends **ONLY** if:
       a) **Expiration:** The freeze token duration or set freeze period expires.
       b) **Manual Unfreeze:** The user explicitly turns off / ends the freeze power (e.g., via `POST /api/v1/streak/unfreeze`).
  4. **Git-Style Heatmap Visual Differentiation:**
     - Frozen days are visually distinct on the contribution graph (e.g., icy blue `#38bdf8` / frost pattern).

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

### Phase 6: Personal Records (PR) & Strength Analytics
* **Goal:** Track progressive overload, 1-Rep Max (1RM) records, and exercise volume milestones.
* **Key Components:**
  1. **PR Tracking Engine:**
     - Automatic 1RM estimation via Epley formula: $\text{1RM} = \text{Weight} \times (1 + \frac{\text{Reps}}{30})$.
     - All-time and period-specific PRs for core movements (Bench Press, Squat, Deadlift, Overhead Press, Pull-Ups).
  2. **PR Celebration Feats:**
     - Power Score bonus when hitting verified PRs during a compliant weekly cycle.

---

### Phase 7: Future Feature — Admin Management Panel & Control System
* **Goal:** Provide a comprehensive web-based Administrative Control Suite for full system configuration and user management.
* **Key Components:**
  1. **Reward Plan & Milestone Management:**
     - GUI & REST endpoints to create, modify, activate, or archive Reward Plans.
     - Dynamic milestone CRUD editor: Add custom day milestones (e.g. Day 11 $\rightarrow$ 5 Restore Shields), modify rewards/quantities, or delete milestone steps.
  2. **Items & Power-Ups Catalog Management:**
     - Add new item definitions, adjust duration, effect types, icon slugs, and rarity levels.
  3. **Workout Split & Plan Preset Authoring:**
     - Create and publish standard weekly workout plans (e.g., PPL Standard, Arnold Split) with exercise assignments.
  4. **User Override & Support Tools:**
     - View user streak state, grant compensation power-ups, adjust timezone settings, and manage user plan subscriptions.

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

-- 4. Reward Plans & Roadmap Milestones
CREATE TABLE IF NOT EXISTS reward_plans (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS reward_plan_milestones (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_id VARCHAR(50) NOT NULL REFERENCES reward_plans(id) ON DELETE CASCADE,
    streak_target INT NOT NULL,
    item_id VARCHAR(50) NOT NULL REFERENCES items(id) ON DELETE CASCADE,
    quantity INT NOT NULL DEFAULT 1,
    title VARCHAR(100) NOT NULL,
    description TEXT,
    badge_slug VARCHAR(100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(plan_id, streak_target, item_id)
);

CREATE TABLE IF NOT EXISTS user_claimed_rewards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    plan_id VARCHAR(50) NOT NULL REFERENCES reward_plans(id) ON DELETE CASCADE,
    streak_target INT NOT NULL,
    item_id VARCHAR(50) NOT NULL REFERENCES items(id) ON DELETE CASCADE,
    claimed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, plan_id, streak_target, item_id)
);

-- 5. Sickness / Injury Streak Freezes
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
| `GET /api/v1/rewards/roadmap` | None | Returns interactive streak reward roadmap with status (`LOCKED`, `CLAIMABLE`, `CLAIMED`) |
| `POST /api/v1/rewards/claim` | `{ "streak_target": 7, "item_id": "RESTORE_SHIELD" }` | Claims unlocked milestone reward and grants item into inventory |
| `POST /api/v1/admin/rewards/plans/:id/milestones` | `{ "streak_target": 11, "item_id": "RESTORE_SHIELD", "quantity": 5, "title": "Day 11 Power" }` | Admin: Add or update milestone target in a reward plan |
| `DELETE /api/v1/admin/rewards/plans/:id/milestones/:milestone_id` | None | Admin: Delete a milestone target from a plan |
| `PUT /api/v1/plans/queue` | `{ "weekly_plan_id": "ppl-core" }` | Queues plan change to activate on next 7-day cycle |

