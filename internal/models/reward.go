package models

import (
	"time"

	"github.com/google/uuid"
)

// RewardPlan defines a streak reward plan entity
type RewardPlan struct {
	ID          string               `json:"id"`
	Name        string               `json:"name"`
	Description string               `json:"description,omitempty"`
	IsActive    bool                 `json:"is_active"`
	Milestones  []RewardPlanMilestone `json:"milestones,omitempty"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
}

// RewardPlanMilestone defines a single reward milestone target in a reward plan
type RewardPlanMilestone struct {
	ID          uuid.UUID `json:"id"`
	PlanID      string    `json:"plan_id"`
	StreakTarget int       `json:"streak_target"`
	ItemID      string    `json:"item_id"`
	Quantity    int       `json:"quantity"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	BadgeSlug   string    `json:"badge_slug,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// UserClaimedReward records a claimed milestone reward by a user
type UserClaimedReward struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	PlanID       string    `json:"plan_id"`
	StreakTarget int       `json:"streak_target"`
	ItemID       string    `json:"item_id"`
	ClaimedAt    time.Time `json:"claimed_at"`
}

// MilestoneStatus represents the claim status of a milestone for a user
type MilestoneStatus string

const (
	MilestoneStatusLocked    MilestoneStatus = "LOCKED"
	MilestoneStatusClaimable MilestoneStatus = "CLAIMABLE"
	MilestoneStatusClaimed   MilestoneStatus = "CLAIMED"
)

// RoadmapMilestoneResponse defines the DTO returned to frontend UI for roadmap presentation
type RoadmapMilestoneResponse struct {
	MilestoneID  uuid.UUID       `json:"milestone_id"`
	PlanID       string          `json:"plan_id"`
	StreakTarget int             `json:"streak_target"`
	ItemID       string          `json:"item_id"`
	ItemName     string          `json:"item_name"`
	ItemIcon     string          `json:"item_icon,omitempty"`
	Rarity       string          `json:"rarity,omitempty"`
	Quantity     int             `json:"quantity"`
	Title        string          `json:"title"`
	Description  string          `json:"description,omitempty"`
	BadgeSlug    string          `json:"badge_slug,omitempty"`
	Status       MilestoneStatus `json:"status"` // "LOCKED", "CLAIMABLE", "CLAIMED"
	ClaimedAt    *time.Time      `json:"claimed_at,omitempty"`
}

// ClaimRewardRequest defines the payload sent by user to claim a reward
type ClaimRewardRequest struct {
	PlanID       string `json:"plan_id,omitempty"`
	StreakTarget int    `json:"streak_target" binding:"required"`
	ItemID       string `json:"item_id" binding:"required"`
}

// ClaimRewardResult defines the output of a reward claim action
type ClaimRewardResult struct {
	Success           bool      `json:"success"`
	StreakTarget      int       `json:"streak_target"`
	ItemID            string    `json:"item_id"`
	ItemName          string    `json:"item_name"`
	QuantityAwarded   int       `json:"quantity_awarded"`
	RemainingInventory int      `json:"remaining_inventory"`
	ClaimedAt         time.Time `json:"claimed_at"`
	Message           string    `json:"message"`
}

// Admin DTOs for Milestone CRUD

type CreateRewardPlanRequest struct {
	ID          string `json:"id" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
}

type UpsertMilestoneRequest struct {
	StreakTarget int    `json:"streak_target" binding:"required"`
	ItemID       string `json:"item_id" binding:"required"`
	Quantity     int    `json:"quantity" binding:"required"`
	Title        string `json:"title" binding:"required"`
	Description  string `json:"description"`
	BadgeSlug    string `json:"badge_slug"`
}
