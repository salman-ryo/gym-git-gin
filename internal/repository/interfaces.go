package repository

import (
	"context"
	"gymgit/backend/internal/models"
	"time"

	"github.com/google/uuid"
)

// UserRepository defines database operations for user profiles
type UserRepository interface {
	GetByAuthUserID(ctx context.Context, authUserID uuid.UUID) (*models.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Create(ctx context.Context, user *models.User) error
	UpdateWeeklyPlan(ctx context.Context, userID uuid.UUID, planID string) error
	SetQueuedPlan(ctx context.Context, userID uuid.UUID, planID *string) error
	UpdateTimezone(ctx context.Context, userID uuid.UUID, timezone string) error
	UpdateAuthUserID(ctx context.Context, id uuid.UUID, authUserID uuid.UUID) error
}

// PlanRepository defines database operations for weekly plans
type PlanRepository interface {
	GetAll(ctx context.Context) ([]models.WeeklyPlan, error)
	GetByID(ctx context.Context, id string) (*models.WeeklyPlan, error)
	Create(ctx context.Context, plan *models.WeeklyPlan) error
}

// GymLogRepository defines database operations for gym logs
type GymLogRepository interface {
	GetLogs(ctx context.Context, userID uuid.UUID, startDate, endDate *time.Time, workoutType *string) ([]models.GymLog, error)
	GetByDate(ctx context.Context, userID uuid.UUID, date string) (*models.GymLog, error)
	UpsertLog(ctx context.Context, log *models.GymLog) error
	DeleteByDate(ctx context.Context, userID uuid.UUID, date string) error
	ResetDemoLogs(ctx context.Context, userID uuid.UUID, logs []models.GymLog) error
}

// StreakRepository defines database operations for 7-day cycle streak state
type StreakRepository interface {
	GetByUserID(ctx context.Context, userID uuid.UUID) (*models.UserStreakState, error)
	UpsertState(ctx context.Context, state *models.UserStreakState) error
}

// ItemRepository defines database operations for the master item catalog
type ItemRepository interface {
	GetAll(ctx context.Context) ([]models.Item, error)
	GetByID(ctx context.Context, id string) (*models.Item, error)
	Create(ctx context.Context, item *models.Item) error
}

// InventoryRepository defines database operations for user item balances & active effects
type InventoryRepository interface {
	GetInventory(ctx context.Context, userID uuid.UUID) ([]models.UserInventoryItem, error)
	GetItemQuantity(ctx context.Context, userID uuid.UUID, itemID string) (int, error)
	AddItemQuantity(ctx context.Context, userID uuid.UUID, itemID string, delta int) (int, error)
	DeductItemQuantity(ctx context.Context, userID uuid.UUID, itemID string, delta int) (int, error)
	CreateActiveEffect(ctx context.Context, effect *models.UserActiveEffect) error
	GetActiveEffects(ctx context.Context, userID uuid.UUID) ([]models.UserActiveEffect, error)
	GetLatestActiveEffectByItem(ctx context.Context, userID uuid.UUID, itemID string) (*models.UserActiveEffect, error)
	DeactivateExpiredEffects(ctx context.Context, userID uuid.UUID) error
}

// RewardRepository defines database operations for reward plans, roadmap milestones, and claimed rewards
type RewardRepository interface {
	GetActiveRewardPlan(ctx context.Context, planID string) (*models.RewardPlan, error)
	GetAllRewardPlans(ctx context.Context) ([]models.RewardPlan, error)
	CreateRewardPlan(ctx context.Context, plan *models.RewardPlan) error
	DeleteRewardPlan(ctx context.Context, planID string) error
	GetRewardMilestones(ctx context.Context, planID string) ([]models.RewardPlanMilestone, error)
	UpsertMilestone(ctx context.Context, milestone *models.RewardPlanMilestone) error
	DeleteMilestone(ctx context.Context, milestoneID uuid.UUID) error
	GetUserClaimedRewards(ctx context.Context, userID uuid.UUID, planID string) ([]models.UserClaimedReward, error)
	IsRewardClaimed(ctx context.Context, userID uuid.UUID, planID string, streakTarget int, itemID string) (bool, error)
	ClaimReward(ctx context.Context, claim *models.UserClaimedReward) error
}

