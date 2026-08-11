package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// EffectType constants
const (
	EffectTypeInstantUse = "INSTANT_USE"
	EffectTypeTimeBased  = "TIME_BASED"
)

// Item represents a master item definition in the catalog
type Item struct {
	ID              string          `json:"id"`
	Name            string          `json:"name"`
	Description     string          `json:"description,omitempty"`
	EffectType      string          `json:"effect_type"` // 'INSTANT_USE' or 'TIME_BASED'
	DurationSeconds int             `json:"duration_seconds"`
	Rarity          string          `json:"rarity"`
	IconSlug        *string         `json:"icon_slug,omitempty"`
	Metadata        json.RawMessage `json:"metadata,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// UserInventoryItem represents a user's balance of a specific item
type UserInventoryItem struct {
	ItemID   string `json:"item_id"`
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
	Item     Item   `json:"item"`
}

// UserActiveEffect represents an active time-based item effect for a user
type UserActiveEffect struct {
	ID          uuid.UUID       `json:"id"`
	UserID      uuid.UUID       `json:"user_id"`
	ItemID      string          `json:"item_id"`
	ActivatedAt time.Time       `json:"activated_at"`
	ExpiresAt   time.Time       `json:"expires_at"`
	IsActive    bool            `json:"is_active"`
	Metadata    json.RawMessage `json:"metadata,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// UseItemRequest represents the HTTP payload for POST /api/v1/inventory/use
type UseItemRequest struct {
	ItemID   string                 `json:"item_id" binding:"required"`
	Quantity int                    `json:"quantity"`
	Payload  map[string]interface{} `json:"payload,omitempty"`
}

// UseItemResult represents the result of applying/using an item
type UseItemResult struct {
	ItemID             string            `json:"item_id"`
	QuantityConsumed   int               `json:"quantity_consumed"`
	RemainingQuantity  int               `json:"remaining_quantity"`
	EffectType         string            `json:"effect_type"`
	ActiveUntil        *time.Time        `json:"active_until,omitempty"`
	Details            string            `json:"details"`
	RestoredStreakDate *string           `json:"restored_streak_date,omitempty"`
}

// RestoreShieldRequest represents the HTTP payload for POST /api/v1/streak/restore
type RestoreShieldRequest struct {
	TargetDate  string     `json:"target_date" binding:"required"` // YYYY-MM-DD
	WorkoutType string     `json:"workout_type,omitempty"`
	Hours       float64    `json:"hours,omitempty"`
}

// RestoreShieldResult represents the result of redeeming a Restore Shield
type RestoreShieldResult struct {
	Success            bool   `json:"success"`
	RestoredDate       string `json:"restored_date"`
	NewCurrentStreak   int    `json:"new_current_streak"`
	ShieldsRemaining   int    `json:"shields_remaining"`
	Message            string `json:"message"`
}

// MilestoneReward represents an item reward granted for achieving a streak milestone
type MilestoneReward struct {
	StreakDays int    `json:"streak_days"`
	ItemID     string `json:"item_id"`
	Quantity   int    `json:"quantity"`
	ItemName   string `json:"item_name"`
}
