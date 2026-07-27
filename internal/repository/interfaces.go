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
