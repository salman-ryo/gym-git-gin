package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents the user profile mapped to Supabase Auth
type User struct {
	ID                 uuid.UUID  `json:"id"`
	AuthUserID         uuid.UUID  `json:"auth_user_id"`
	Email              string     `json:"email"`
	Name               *string    `json:"name,omitempty"`
	AvatarURL          *string    `json:"avatar_url,omitempty"`
	Provider           string     `json:"provider"`
	Timezone           string     `json:"timezone"`
	WeeklyPlanID       *string    `json:"weekly_plan_id,omitempty"`
	QueuedWeeklyPlanID *string    `json:"queued_weekly_plan_id,omitempty"`
	CheckinSnoozedDate *string    `json:"checkin_snoozed_date,omitempty"`
	CheckinSnoozedAt   *time.Time `json:"checkin_snoozed_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// CheckinSnoozeStatus represents the dynamic snooze state returned to the frontend
type CheckinSnoozeStatus struct {
	Date             string     `json:"date"`
	SnoozedAt        *time.Time `json:"snoozed_at,omitempty"`
	IsSnoozed        bool       `json:"is_snoozed"`
	RemainingSeconds int        `json:"remaining_seconds"`
}

