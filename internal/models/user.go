package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents the user profile mapped to Supabase Auth
type User struct {
	ID           uuid.UUID `json:"id"`
	AuthUserID   uuid.UUID `json:"auth_user_id"`
	Email        string    `json:"email"`
	Name         *string   `json:"name,omitempty"`
	AvatarURL    *string   `json:"avatar_url,omitempty"`
	Provider     string    `json:"provider"`
	Timezone           string    `json:"timezone"`
	WeeklyPlanID       *string   `json:"weekly_plan_id,omitempty"`
	QueuedWeeklyPlanID *string   `json:"queued_weekly_plan_id,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}
