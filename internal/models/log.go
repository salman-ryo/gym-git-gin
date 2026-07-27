package models

import (
	"time"

	"github.com/google/uuid"
)

// GymLog represents an individual workout session record
type GymLog struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	Date        string    `json:"date"` // YYYY-MM-DD
	Hours       float64   `json:"hours"`
	WorkoutType string    `json:"workout_type"`
	Notes       *string   `json:"notes,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
