# Gym-Git Frontend Developer Specification & Integration Guide (Specs.md)

> **Document Version:** 1.0.0  
> **Target Audience:** Frontend Developers, UI/UX Engineers & Coding Agents  
> **Scope:** Comprehensive specification of backend features, streak engine rules, inventory mechanics, reward roadmap system, API contracts, JSON envelopes, and frontend implementation flows.

---

## 1. System Overview & Core Architecture

Gym-Git transforms traditional gym consistency tracking into a gamified, timezone-aware split engine. It provides physiological flexibility with **7-Day Plan Cycles**, **Rest Tokens**, **Restore Shields**, **Sickness Freeze Vault ("Ice Pause")**, and an interactive **Streak Reward Roadmap System**.

```text
                                  ┌──────────────────────────────┐
                                  │      User App Visit          │
                                  └──────────────┬───────────────┘
                                                 │
                                                 ▼
                                     [ GET /api/v1/auth/me ]
                                                 │
                         ┌───────────────────────┴───────────────────────┐
                         ▼                                               ▼
               [ Active Profile & Plan ]                      [ Streak Lifecycle Check ]
                         │                                               │
                         │                        ┌──────────────────────┴──────────────────────┐
                         │                        ▼                                             ▼
                         │               [ streak_broken_event ]                       [ streak_warning_event ]
                         │                        │                                             │
                         │                        ▼                                             ▼
                         │             (Show Streak Broken Modal)                   (Show Risk Warning Banner)
                         │                        │
                         │            [ Consume Restore Shield ]
                         │                        │
                         ▼                        ▼
      ┌──────────────────────────────────────────────────────────────────────┐
      │                         Main Dashboard UI                            │
      │  • 7-Day Cycle Progress & Rest Tokens                                │
      │  • Git-Style Contribution Heatmap Matrix (Green / Icy Blue)          │
      │  • Gamified Power Score & Anime Tier Badge                           │
      │  • Interactive Streak Reward Roadmap (LOCKED / CLAIMABLE / CLAIMED)  │
      │  • Inventory Drawer & Active Items                                   │
      └──────────────────────────────────────────────────────────────────────┘
```

---

## 2. Feature Specifications & Architectural Rules

### Phase 1: Global Timezone Engine & Grace Period Anchor
* **User Timezone Profile:** Every user has an IANA timezone (e.g. `America/New_York`, `Asia/Kolkata`, `Europe/London`) stored in profile.
* **Context Resolution:** The frontend must send `X-Timezone: <IANA_STRING>` header (e.g., `America/New_York`) or relies on database fallback.
* **Localized Wall-Clock Day:** A day runs strictly from `00:00:00` to `23:59:59` in the user's localized timezone. A day is **never** marked missed until local midnight arrives.
* **Strict Historical Segregation:**
  - **Today's Log (`date == user_today`):** Updates workout volume AND increments active streak.
  - **Historical Log (`date < user_today`):** Updates historical volume and heatmap matrix ONLY. Does **NOT** revive a dead streak unless a **Restore Shield** power-up is explicitly redeemed.

---

### Phase 2: 7-Day Plan Cycles, Rest Tokens & Accuracy Scoring
* **Discrete 7-Day Cycle Windows:**
  - A plan runs in fixed 7-day cycle windows (`cycle_start_date` to `cycle_end_date`).
  - Target workout days = count of workout categories in plan (e.g. 4-day split $\rightarrow$ 4 Target Workouts, 3 Rest Tokens per cycle).
* **Rest Token Mechanics:**
  - Rest days consume Rest Tokens automatically.
  - Rest days can be taken consecutively or staggered without breaking active streak, as long as target workouts are met within the 7-day cycle.
* **Split Accuracy Score (0-100%):**
  - Evaluates compliance order against active plan. Swapping workout days or taking out-of-order rest days preserves the streak but adjusts the Accuracy Score.
* **Queued Plan Changes:**
  - Plan changes (`PUT /api/v1/plans/queue`) do NOT disrupt the current 7-day cycle. They activate strictly at the start of the next 7-day cycle.

---

### Phase 3: Gamified Inventory & Master Item Catalog
Master power-up definitions available in the system:

| Item ID | Item Name | Effect Type | Duration | Description |
| :--- | :--- | :--- | :--- | :--- |
| `RESTORE_SHIELD` | Restore Shield | `INSTANT_USE` | Instant | Restores a lost streak from a missed day within the 3-day lookback window. |
| `STREAK_FREEZE_TOKEN` | Streak Freeze Token | `TIME_BASED` | 86400s (24h) | Pauses streak decay without breaking streak ("Ice Pause"). |
| `XP_BOOST` | XP Boost Token | `TIME_BASED` | 604800s (7d) | Grants 1.5x multiplier to earned power points. |
| `ACCURACY_CHARM` | Accuracy Charm | `INSTANT_USE` | Instant | Protects workout split accuracy score from penalty for 1 cycle. |

---

### Phase 4: Sickness Freeze Vault, Item Usage & Dynamic Reward Roadmap
* **Sickness / Injury Freeze Vault ("Ice Pause"):**
  - Triggered by consuming `STREAK_FREEZE_TOKEN`s via `POST /api/v1/streak/freeze` or `POST /api/v1/inventory/use`.
  - Streak status becomes `is_frozen = true`.
  - **STRICT NO AUTO-UNFREEZE RULE:** Opening the app, logging in, or checking profile data does **NEVER** automatically unfreeze a streak.
  - Unfreezing occurs **ONLY** via explicit manual toggle (`POST /api/v1/streak/unfreeze`) or natural freeze token duration expiration.
* **Dynamic Streak Reward Roadmap System:**
  - Presented as an interactive progression roadmap on the frontend (`GET /api/v1/rewards/roadmap`).
  - Milestone statuses:
    - `CLAIMED`: Reward already claimed into user inventory.
    - `CLAIMABLE`: Qualified (`user_max_streak >= streak_target`) and not yet claimed. Displays an interactive **"CLAIM REWARD"** button.
    - `LOCKED`: Streak target not yet achieved (`user_max_streak < streak_target`).
  - Claiming a reward (`POST /api/v1/rewards/claim`) adds items directly to `user_inventories` and records claim history.
  - **Dynamic Milestone CRUD:** Administrators can insert milestones at arbitrary streak targets (e.g. Day 11 $\rightarrow$ +5 Restore Shields), modify, or delete milestone steps easily.

---

### Phase 5: Streak Break Detection & Lifecycle Events
* **Streak Broken Event (`streak_broken_event`):**
  - Returned in `GET /api/v1/auth/me` and `GET /api/v1/streak` when a streak break occurs (`current_streak == 0` and `longest_streak > 0`).
  - Contains: `previous_streak`, `broken_on`, `restore_shield_available`, `restore_shields_count`, `can_restore_until`.
  - Frontend triggers a dedicated **"Streak Broken" Modal** offering instant Restore Shield redemption.
* **Streak At Risk Warning (`streak_warning_event`):**
  - Returned when today's workout target is unmet, rest tokens are exhausted (`rest_tokens_remaining == 0`), and user is not frozen.
  - Contains: `is_at_risk: true`, `hours_remaining` (wall-clock hours left until local midnight), `rest_tokens_left`, and warning message.
  - Frontend displays a high-priority **"Streak At Risk!" Warning Banner**.

---

## 3. Standard API Envelopes & Error Formatting

All backend responses use standard JSON envelopes:

### Success Response (HTTP 200 / 201)
```json
{
  "success": true,
  "data": { ... },
  "message": "Operation completed successfully"
}
```

### Error Response (HTTP 4xx / 5xx)
```json
{
  "success": false,
  "error": {
    "code": "BAD_REQUEST",
    "message": "Invalid request payload; streak_target is required",
    "details": null
  },
  "timestamp": "2026-08-11T14:10:00Z"
}
```

---

## 4. Complete REST API Endpoint Specification

### A. Authentication & Session Bootstrap

#### `GET /api/v1/auth/me`
* **Headers:** `Authorization: Bearer <token>`, `X-Timezone: America/New_York`
* **Description:** Retrieves authenticated user profile, active plan, and streak lifecycle status.
* **Response `data`:**
```json
{
  "user": {
    "id": "u1111111-1111-1111-1111-111111111111",
    "auth_user_id": "a2222222-2222-2222-2222-222222222222",
    "email": "alex@gymgit.com",
    "name": "Alex Mercer",
    "avatar_url": "https://example.com/avatar.jpg",
    "weekly_plan_id": "ppl-standard",
    "queued_weekly_plan_id": null,
    "timezone": "America/New_York"
  },
  "plan": {
    "id": "ppl-standard",
    "name": "PPL Standard 4-Day Split",
    "categories": ["Push", "Pull", "Legs", "Cardio"]
  },
  "streak": {
    "current_streak": 0,
    "longest_streak": 14,
    "compliance_rate": 85,
    "cycle_info": {
      "cycle_start_date": "2026-08-08",
      "cycle_end_date": "2026-08-14",
      "workouts_completed_in_cycle": 1,
      "workouts_target_in_cycle": 4,
      "rest_tokens_total": 3,
      "rest_tokens_used": 3,
      "rest_tokens_remaining": 0,
      "days_remaining_in_cycle": 3
    },
    "accuracy_score": 85,
    "is_frozen": false,
    "streak_broken_event": {
      "previous_streak": 14,
      "broken_on": "2026-08-10",
      "restore_shield_available": true,
      "restore_shields_count": 2,
      "can_restore_until": "2026-08-11"
    },
    "streak_warning_event": {
      "is_at_risk": true,
      "hours_remaining": 6,
      "rest_tokens_left": 0,
      "message": "Streak at risk! Complete your workout within 6 hours to maintain your streak."
    }
  }
}
```

---

### B. Streak & Cycle Engine

#### `GET /api/v1/streak`
* **Headers:** `Authorization: Bearer <token>`, `X-Timezone: America/New_York`
* **Description:** Returns active streak, cycle details, rest tokens, freeze state, and lifecycle events.

#### `POST /api/v1/streak/restore`
* **Headers:** `Authorization: Bearer <token>`
* **Payload:**
```json
{
  "target_date": "2026-08-10",
  "workout_type": "Restored Push Session",
  "hours": 1.0
}
```
* **Description:** Redeems 1 Restore Shield to revive a missed streak day within the 3-day lookback window.
* **Response `data`:**
```json
{
  "success": true,
  "restored_date": "2026-08-10",
  "new_current_streak": 15,
  "shields_remaining": 1,
  "message": "Restore Shield redeemed successfully for 2026-08-10! Active streak updated to 15 days."
}
```

#### `POST /api/v1/streak/freeze`
* **Headers:** `Authorization: Bearer <token>`
* **Payload:**
```json
{
  "duration_days": 2,
  "reason": "Severe flu recovery"
}
```
* **Description:** Consumes `STREAK_FREEZE_TOKEN`s from inventory and sets streak status to frozen ("Ice Pause").

#### `POST /api/v1/streak/unfreeze`
* **Headers:** `Authorization: Bearer <token>`
* **Description:** Manually deactivates an active streak freeze ("Ice Pause").

---

### C. Inventory & Items

#### `GET /api/v1/items`
* **Description:** Public catalog of all master items definitions.

#### `GET /api/v1/inventory`
* **Headers:** `Authorization: Bearer <token>`
* **Description:** Returns user item quantities and active time-based item effects.

#### `POST /api/v1/inventory/use`
* **Headers:** `Authorization: Bearer <token>`
* **Payload:**
```json
{
  "item_id": "STREAK_FREEZE_TOKEN",
  "quantity": 1,
  "payload": {
    "reason": "Resting"
  }
}
```
* **Description:** Consumes/activates an item from user inventory.

---

### D. Streak Reward Roadmap & Milestones

#### `GET /api/v1/rewards/roadmap`
* **Headers:** `Authorization: Bearer <token>`
* **Query Params:** `plan_id` (optional, defaults to `default-streak-roadmap`)
* **Description:** Returns ordered roadmap milestones with real-time status (`LOCKED`, `CLAIMABLE`, `CLAIMED`).
* **Response `data`:**
```json
[
  {
    "milestone_id": "m7777777-7777-7777-7777-777777777777",
    "plan_id": "default-streak-roadmap",
    "streak_target": 7,
    "item_id": "RESTORE_SHIELD",
    "item_name": "Restore Shield",
    "item_icon": "shield-icon",
    "rarity": "rare",
    "quantity": 1,
    "title": "7-Day Shield Anchor",
    "description": "Claim 1x Restore Shield to protect your consistency",
    "badge_slug": "shield-badge-bronze",
    "status": "CLAIMED",
    "claimed_at": "2026-08-09T10:30:00Z"
  },
  {
    "milestone_id": "m1010101-1010-1010-1010-101010101010",
    "plan_id": "default-streak-roadmap",
    "streak_target": 10,
    "item_id": "STREAK_FREEZE_TOKEN",
    "item_name": "Streak Freeze Token",
    "item_icon": "snowflake-icon",
    "rarity": "rare",
    "quantity": 1,
    "title": "10-Day Ice Defender",
    "description": "Claim 1x Streak Freeze Token for sickness/rest days",
    "badge_slug": "freeze-badge-bronze",
    "status": "CLAIMABLE"
  },
  {
    "milestone_id": "m1111111-1111-1111-1111-111111111111",
    "plan_id": "default-streak-roadmap",
    "streak_target": 11,
    "item_id": "RESTORE_SHIELD",
    "item_name": "Restore Shield",
    "item_icon": "shield-icon",
    "rarity": "rare",
    "quantity": 5,
    "title": "11-Day Shield Power",
    "description": "Claim 5x Restore Shields",
    "badge_slug": "shield-badge-silver",
    "status": "LOCKED"
  }
]
```

#### `POST /api/v1/rewards/claim`
* **Headers:** `Authorization: Bearer <token>`
* **Payload:**
```json
{
  "plan_id": "default-streak-roadmap",
  "streak_target": 10,
  "item_id": "STREAK_FREEZE_TOKEN"
}
```
* **Description:** Claims an unlocked milestone reward (`CLAIMABLE`), adds item quantity to inventory, and returns updated balance.

---

### E. Admin Control Panel Endpoints

#### `POST /api/v1/admin/rewards/plans/:id/milestones`
* **Headers:** `Authorization: Bearer <token>`, `X-Admin-Secret: <ADMIN_SECRET>`
* **Payload:**
```json
{
  "streak_target": 11,
  "item_id": "RESTORE_SHIELD",
  "quantity": 5,
  "title": "Day 11 Shield Surge",
  "description": "Custom Day 11 milestone reward",
  "badge_slug": "shield-badge-11"
}
```
* **Description:** Dynamic milestone CRUD endpoint to insert or update any milestone target in a plan.

#### `DELETE /api/v1/admin/rewards/plans/:id/milestones/:milestone_id`
* **Description:** Admin endpoint to remove a milestone target.

---

## 5. Frontend UI/UX Integration Guidelines

### A. Contribution Heatmap Rendering
- **Normal Active Day:** Dark to vibrant green gradient (`#166534` to `#22c55e`).
- **Frozen Day ("Ice Pause"):** Visually distinct icy blue frost tile (`#38bdf8` or frost overlay icon).
- **Rest Token Day:** Neutral subtle indicator (`#334155`).

### B. Streak Broken Modal & Restoration Flow
1. Upon `GET /api/v1/auth/me` or `GET /api/v1/streak`, check if `streak_broken_event != null`.
2. If present, open **"Streak Broken!" Modal**:
   - Display previous streak length (e.g. `"14-Day Streak Broken"`).
   - If `restore_shield_available == true`:
     - Show **"Redeem Restore Shield (Available: X)"** button.
     - Clicking button calls `POST /api/v1/streak/restore` with `target_date: broken_on`.
     - On success, play streak restoration celebration animation and refresh dashboard.
   - If `restore_shield_available == false`:
     - Offer shortcut to view **Streak Reward Roadmap** to unlock Restore Shields.

### C. Streak Risk Warning Banner
1. Check if `streak_warning_event != null` and `is_at_risk == true`.
2. Render a high-priority warning banner at the top of the dashboard:
   - Text: `"Streak at Risk! You have 0 rest tokens left. Log today's workout within X hours before midnight to protect your streak."`
   - Primary Action: **"Log Workout Now"**.

### D. Interactive Streak Reward Roadmap Component
1. Render milestones as a visual vertical or horizontal progression timeline node path.
2. For each node:
   - Node status `LOCKED`: Dimmed/greyed out with lock icon.
   - Node status `CLAIMABLE`: Glowing animated border with prominent **"CLAIM"** button.
   - Node status `CLAIMED`: Checkmark icon with timestamp tooltip.
3. Clicking **"CLAIM"**:
   - Calls `POST /api/v1/rewards/claim`.
   - Triggers celebratory item drop animation and increments user's inventory badge balance.
