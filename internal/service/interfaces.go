package service

import (
	"context"
	"gymgit/backend/internal/models"
	"time"

	"github.com/google/uuid"
)

// AuthService defines business logic for authentication and profile bootstrapping
type AuthService interface {
	BootstrapProfile(ctx context.Context, authUserID uuid.UUID, email, name, avatarURL, provider string) (*models.User, error)
	GetProfile(ctx context.Context, userID uuid.UUID) (*models.User, *models.WeeklyPlan, error)
	UpdatePlan(ctx context.Context, userID uuid.UUID, planID string, customName string, customDesc string, categories []string) error
	UpdateTimezone(ctx context.Context, userID uuid.UUID, tz string) (*models.User, error)
	SetCheckinSnooze(ctx context.Context, userID uuid.UUID, dateStr string) (*models.CheckinSnoozeStatus, error)
	GetCheckinSnoozeStatus(ctx context.Context, user *models.User) *models.CheckinSnoozeStatus
	ClearCheckinSnooze(ctx context.Context, userID uuid.UUID) error
}

// PlanService defines business logic for weekly plans
type PlanService interface {
	GetAllPlans(ctx context.Context) ([]models.WeeklyPlan, error)
	GetPlanByID(ctx context.Context, id string) (*models.WeeklyPlan, error)
	QueuePlanChange(ctx context.Context, userID uuid.UUID, planID string) error
}

// GymLogService defines business logic for managing workout logs
type GymLogService interface {
	GetLogs(ctx context.Context, userID uuid.UUID, startDate, endDate *time.Time, workoutType *string) ([]models.GymLog, error)
	SaveLog(ctx context.Context, userID uuid.UUID, date string, hours float64, workoutType string, notes *string) (*models.GymLog, error)
	DeleteLog(ctx context.Context, userID uuid.UUID, date string) error
	ResetDemoLogs(ctx context.Context, userID uuid.UUID) error
}

// StatsService defines business logic for analytics, rolling streak, and Gym Power Score
type StatsService interface {
	GetDashboardStats(ctx context.Context, userID uuid.UUID, days int) (*models.DashboardStatsResponse, error)
	GetPowerStats(ctx context.Context, userID uuid.UUID, days int) (*models.PowerStatsResponse, error)
}

// StreakService defines business logic for 7-day cycle streak states, rest tokens & sickness freezes
type StreakService interface {
	GetStreakState(ctx context.Context, userID uuid.UUID, loc *time.Location) (*models.StreakResponse, error)
	FreezeStreak(ctx context.Context, userID uuid.UUID, durationDays int, reason string) (*models.UserStreakState, error)
	UnfreezeStreak(ctx context.Context, userID uuid.UUID) (*models.UserStreakState, error)
}

// InventoryService defines business logic for item catalog, user inventory, and item activations
type InventoryService interface {
	GetCatalog(ctx context.Context) ([]models.Item, error)
	GetUserInventory(ctx context.Context, userID uuid.UUID) ([]models.UserInventoryItem, []models.UserActiveEffect, error)
	UseItem(ctx context.Context, userID uuid.UUID, itemID string, quantity int, payload map[string]interface{}, loc *time.Location) (*models.UseItemResult, error)
	RedeemRestoreShield(ctx context.Context, userID uuid.UUID, targetDates []string, workoutType string, hours float64, loc *time.Location) (*models.RestoreShieldResult, error)
	CheckAndGrantMilestones(ctx context.Context, userID uuid.UUID, streakDays int) ([]models.MilestoneReward, error)
}

// RewardService defines business logic for reward plans, roadmap milestone claims, and admin CRUD
type RewardService interface {
	GetRoadmap(ctx context.Context, userID uuid.UUID, planID string) ([]models.RoadmapMilestoneResponse, error)
	ClaimReward(ctx context.Context, userID uuid.UUID, planID string, streakTarget int, itemID string) (*models.ClaimRewardResult, error)
	GetAllPlans(ctx context.Context) ([]models.RewardPlan, error)
	CreateRewardPlan(ctx context.Context, req models.CreateRewardPlanRequest) (*models.RewardPlan, error)
	DeleteRewardPlan(ctx context.Context, planID string) error
	UpsertMilestone(ctx context.Context, planID string, req models.UpsertMilestoneRequest) (*models.RewardPlanMilestone, error)
	DeleteMilestone(ctx context.Context, milestoneID uuid.UUID) error
}


