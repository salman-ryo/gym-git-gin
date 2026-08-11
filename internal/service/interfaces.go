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
}

// PlanService defines business logic for weekly plans
type PlanService interface {
	GetAllPlans(ctx context.Context) ([]models.WeeklyPlan, error)
	GetPlanByID(ctx context.Context, id string) (*models.WeeklyPlan, error)
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

