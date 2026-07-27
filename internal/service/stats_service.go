package service

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"gymgit/backend/internal/models"
	"gymgit/backend/internal/repository"

	"github.com/google/uuid"
)

type statsService struct {
	userRepo repository.UserRepository
	planRepo repository.PlanRepository
	logRepo  repository.GymLogRepository
}

// NewStatsService creates a new StatsService instance
func NewStatsService(userRepo repository.UserRepository, planRepo repository.PlanRepository, logRepo repository.GymLogRepository) StatsService {
	return &statsService{
		userRepo: userRepo,
		planRepo: planRepo,
		logRepo:  logRepo,
	}
}

func (s *statsService) GetDashboardStats(ctx context.Context, userID uuid.UUID, days int) (*models.DashboardStatsResponse, error) {
	if days <= 0 {
		days = 30
	}

	targetDaysPerWeek := s.getUserTargetDaysPerWeek(ctx, userID)

	// Fetch logs up to 365 days back to support streak calculation across rolling window
	startDate := time.Now().UTC().AddDate(0, 0, -365)
	endDate := time.Now().UTC()

	logs, err := s.logRepo.GetLogs(ctx, userID, &startDate, &endDate, nil)
	if err != nil {
		return nil, fmt.Errorf("failed fetching user logs for dashboard stats: %w", err)
	}

	streak := CalculateScientificStreak(logs, targetDaysPerWeek, days)
	powerBreakdown := CalculatePowerScore(logs, targetDaysPerWeek, days)
	animeTier := MapPowerScoreToAnimeTier(powerBreakdown.TotalScore)

	// Period specific metrics
	periodStartDate := time.Now().UTC().AddDate(0, 0, -days+1).Format("2006-01-02")
	todayStr := time.Now().UTC().Format("2006-01-02")

	totalSessions := 0
	totalHours := 0.0

	for _, l := range logs {
		if l.Date >= periodStartDate && l.Date <= todayStr && l.Hours > 0 {
			totalSessions++
			totalHours += l.Hours
		}
	}

	avgSessionDuration := 0.0
	if totalSessions > 0 {
		avgSessionDuration = math.Round((totalHours/float64(totalSessions))*100.0) / 100.0
	}
	totalHours = math.Round(totalHours*100.0) / 100.0

	return &models.DashboardStatsResponse{
		Streak:             streak,
		TotalSessions:      totalSessions,
		TotalHours:         totalHours,
		AvgSessionDuration: avgSessionDuration,
		PowerScore:         powerBreakdown.TotalScore,
		AnimeTier:          animeTier,
		PeriodDays:         days,
	}, nil
}

func (s *statsService) GetPowerStats(ctx context.Context, userID uuid.UUID, days int) (*models.PowerStatsResponse, error) {
	if days <= 0 {
		days = 30
	}

	targetDaysPerWeek := s.getUserTargetDaysPerWeek(ctx, userID)

	startDate := time.Now().UTC().AddDate(0, 0, -365)
	endDate := time.Now().UTC()

	logs, err := s.logRepo.GetLogs(ctx, userID, &startDate, &endDate, nil)
	if err != nil {
		return nil, fmt.Errorf("failed fetching user logs for power stats: %w", err)
	}

	powerBreakdown := CalculatePowerScore(logs, targetDaysPerWeek, days)
	animeTier := MapPowerScoreToAnimeTier(powerBreakdown.TotalScore)

	// Count active days & unique workout types in window
	periodStartDate := time.Now().UTC().AddDate(0, 0, -days+1).Format("2006-01-02")
	todayStr := time.Now().UTC().Format("2006-01-02")

	activeDaysMap := make(map[string]bool)
	uniqueTypesMap := make(map[string]bool)

	for _, l := range logs {
		if l.Date >= periodStartDate && l.Date <= todayStr && l.Hours > 0 {
			activeDaysMap[l.Date] = true
			if l.WorkoutType != "" && !strings.EqualFold(l.WorkoutType, "Rest") {
				uniqueTypesMap[strings.ToLower(l.WorkoutType)] = true
			}
		}
	}

	return &models.PowerStatsResponse{
		PowerScore:         powerBreakdown,
		AnimeTier:          animeTier,
		PeriodDays:         days,
		TargetDaysPerWeek:  targetDaysPerWeek,
		ActiveDays:         len(activeDaysMap),
		UniqueWorkoutTypes: len(uniqueTypesMap),
	}, nil
}

func (s *statsService) getUserTargetDaysPerWeek(ctx context.Context, userID uuid.UUID) int {
	categories := []string{"Push", "Pull", "Legs"}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err == nil && user != nil && user.WeeklyPlanID != nil {
		plan, planErr := s.planRepo.GetByID(ctx, *user.WeeklyPlanID)
		if planErr == nil && plan != nil && len(plan.Categories) > 0 {
			categories = plan.Categories
		}
	}

	return CalculateTargetDaysPerWeek(categories)
}
