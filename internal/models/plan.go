package models

import (
	"time"

	"github.com/google/uuid"
)

// WeeklyPlan represents a structured workout split plan
type WeeklyPlan struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Categories  []string   `json:"categories"`
	UserID      *uuid.UUID `json:"user_id,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
