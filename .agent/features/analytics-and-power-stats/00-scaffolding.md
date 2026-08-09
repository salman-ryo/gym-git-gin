# Scaffolding & DTO Definitions: Analytics & Power Stats

> **Feature:** `analytics-and-power-stats`  
> **Phase:** `00-scaffolding`

---

### Task 0.1: Stats Models & Scientific Power Score DTOs

* **Context Bundle:**
  1. [.agent/rules/01-architecture.md](file:///.agent/rules/01-architecture.md)
  2. [.agent/rules/02-code-style.md](file:///.agent/rules/02-code-style.md)
* **Owns:**
  - `internal/models/stats.go`
* **Forbidden:**
  - `internal/service/**`
  - `internal/handler/**`
* **Acceptance Criteria:**
  - **WHEN** stats DTOs are defined, **THE SYSTEM SHALL** include `StreakInfo` (current, longest, active status), `StatsSummary` (total workouts, total hours, average duration, attendance rate), `WorkoutDistribution` map, and `PowerScore` structure.
  - **WHEN** `PowerScore` is mapped, **THE SYSTEM SHALL** contain `Score`, `Tier`, `TierName`, `Level`, `ConsistencyScore`, `IntensityScore`, `VolumeScore`, `RecoveryScore`, and `RadarData`.
