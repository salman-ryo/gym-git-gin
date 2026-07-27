package models

// StreakStats represents scientific streak analytics
type StreakStats struct {
	CurrentStreak      int `json:"current_streak"`
	TargetDaysPerWeek  int `json:"target_days_per_week"`
	ComplianceRate     int `json:"compliance_rate"` // 0-100%
	TotalCompliantDays int `json:"total_compliant_days"`
	TotalTrackedDays   int `json:"total_tracked_days"`
	TotalActiveDays    int `json:"total_active_days"`
}

// PowerScoreBreakdown represents the 4 components of the Gym Power Score
type PowerScoreBreakdown struct {
	Consistency     int `json:"consistency"`      // 0-45 pts
	DurationQuality int `json:"duration_quality"` // 0-25 pts
	Variety         int `json:"variety"`          // 0-20 pts
	Momentum        int `json:"momentum"`         // 0-10 pts
	TotalScore      int `json:"total_score"`      // 0-100 pts
}

// AnimeTier represents the gamified tier mapping based on Gym Power Score
type AnimeTier struct {
	Tier      string `json:"tier"`      // "D", "C", "B", "A", "S", "S+", "SS"
	Character string `json:"character"` // e.g. "Satoru Gojo"
	Anime     string `json:"anime"`     // e.g. "Jujutsu Kaisen"
	Title     string `json:"title"`     // e.g. "The Honored One"
}

// DashboardStatsResponse represents response for GET /api/v1/stats
type DashboardStatsResponse struct {
	Streak             StreakStats `json:"streak"`
	TotalSessions      int         `json:"total_sessions"`
	TotalHours         float64     `json:"total_hours"`
	AvgSessionDuration float64     `json:"avg_session_duration"`
	PowerScore         int         `json:"power_score"`
	AnimeTier          AnimeTier   `json:"anime_tier"`
	PeriodDays         int         `json:"period_days"`
}

// PowerStatsResponse represents response for GET /api/v1/stats/power
type PowerStatsResponse struct {
	PowerScore         PowerScoreBreakdown `json:"power_score"`
	AnimeTier          AnimeTier           `json:"anime_tier"`
	PeriodDays         int                 `json:"period_days"`
	TargetDaysPerWeek  int                 `json:"target_days_per_week"`
	ActiveDays         int                 `json:"active_days"`
	UniqueWorkoutTypes int                 `json:"unique_workout_types"`
}
