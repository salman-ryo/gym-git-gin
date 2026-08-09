# Scientific Power Score & Stats Calculation Engine

> **Feature:** `analytics-and-power-stats`  
> **Phase:** `01-stats-calculator-engine`

---

### Task 1.1: Pure Algorithmic Computation Engine & Unit Tests

* **Context Bundle:**
  1. [.agent/rules/01-architecture.md](file:///.agent/rules/01-architecture.md)
  2. [.agent/rules/03-testing-and-errors.md](file:///.agent/rules/03-testing-and-errors.md)
* **Owns:**
  - `internal/service/stats_calculator.go`
  - `internal/service/stats_calculator_test.go`
* **Forbidden:**
  - `internal/repository/**`
  - `internal/handler/**`
* **Acceptance Criteria:**
  - **WHEN** `CalculatePowerScore` runs over log history, **THE SYSTEM SHALL** compute four weighted sub-scores:
    1. **Consistency (35%):** Target workouts per week vs actual workouts.
    2. **Intensity (25%):** Average workout duration relative to 1.25h optimal baseline.
    3. **Volume (25%):** Total weekly active hours.
    4. **Recovery (15%):** Penalizing streaks without rest days or excessive overtraining.
  - **WHEN** power tier is evaluated, **THE SYSTEM SHALL** assign appropriate Anime Tiers: Human, Genin, Hunter, Hashira, Special Grade, Super Saiyan, Monarch, God Level.
  - **WHEN** `CalculateStreak` runs, **THE SYSTEM SHALL** accurately compute current streak (respecting grace days) and all-time longest streak.
