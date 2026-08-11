package models

import (
	"time"

	"github.com/google/uuid"
)

// UserStreakState represents the database entity tracking a user's streak & 7-day plan cycle state
type UserStreakState struct {
	ID                        uuid.UUID  `json:"id"`
	UserID                    uuid.UUID  `json:"user_id"`
	CurrentStreak             int        `json:"current_streak"`
	LongestStreak             int        `json:"longest_streak"`
	CycleStartDate            string     `json:"cycle_start_date"` // YYYY-MM-DD
	CycleEndDate              string     `json:"cycle_end_date"`   // YYYY-MM-DD
	WorkoutsCompletedInCycle int        `json:"workouts_completed_in_cycle"`
	WorkoutsTargetInCycle    int        `json:"workouts_target_in_cycle"`
	RestTokensTotal           int        `json:"rest_tokens_total"`
	RestTokensUsed            int        `json:"rest_tokens_used"`
	AccuracyScore             int        `json:"accuracy_score"` // 0-100%
	LastLoggedDate            *string    `json:"last_logged_date,omitempty"`
	IsFrozen                  bool       `json:"is_frozen"`
	UpdatedAt                 time.Time  `json:"updated_at"`
}

// StreakCycleInfo represents the 7-day cycle status DTO
type StreakCycleInfo struct {
	CycleStartDate            string `json:"cycle_start_date"`
	CycleEndDate              string `json:"cycle_end_date"`
	WorkoutsCompletedInCycle int    `json:"workouts_completed_in_cycle"`
	WorkoutsTargetInCycle    int    `json:"workouts_target_in_cycle"`
	RestTokensTotal           int    `json:"rest_tokens_total"`
	RestTokensUsed            int    `json:"rest_tokens_used"`
	RestTokensRemaining       int    `json:"rest_tokens_remaining"`
	DaysRemainingInCycle      int    `json:"days_remaining_in_cycle"`
}

// AccuracyBreakdown provides details on how split accuracy was evaluated
type AccuracyBreakdown struct {
	AccuracyScore   int      `json:"accuracy_score"` // 0-100%
	TotalEvaluated  int      `json:"total_evaluated"`
	MatchedCount    int      `json:"matched_count"`
	Deviations      []string `json:"deviations,omitempty"`
}

// StreakResponse represents the response envelope data for GET /api/v1/streak
type StreakResponse struct {
	CurrentStreak      int                `json:"current_streak"`
	LongestStreak      int                `json:"longest_streak"`
	ComplianceRate     int                `json:"compliance_rate"`
	CycleInfo          StreakCycleInfo    `json:"cycle_info"`
	AccuracyScore      int                `json:"accuracy_score"`
	QueuedWeeklyPlanID *string            `json:"queued_weekly_plan_id,omitempty"`
	IsFrozen           bool               `json:"is_frozen"`
	LastLoggedDate     *string            `json:"last_logged_date,omitempty"`
}
