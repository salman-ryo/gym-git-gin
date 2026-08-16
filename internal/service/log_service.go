package service

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"gymgit/backend/internal/models"
	"gymgit/backend/internal/repository"
	"gymgit/backend/internal/timezone"

	"github.com/google/uuid"
)

type gymLogService struct {
	logRepo    repository.GymLogRepository
	userRepo   repository.UserRepository
	planRepo   repository.PlanRepository
	streakRepo repository.StreakRepository
}

// NewGymLogService creates a new GymLogService instance
func NewGymLogService(
	logRepo repository.GymLogRepository,
	userRepo repository.UserRepository,
	planRepo repository.PlanRepository,
	streakRepo repository.StreakRepository,
) GymLogService {
	return &gymLogService{
		logRepo:    logRepo,
		userRepo:   userRepo,
		planRepo:   planRepo,
		streakRepo: streakRepo,
	}
}

func (s *gymLogService) checkPastLogRestriction(ctx context.Context, userID uuid.UUID, date string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil || user == nil || user.WeeklyPlanID == nil || *user.WeeklyPlanID == "" {
		return nil
	}

	loc := time.UTC
	if user.Timezone != "" {
		loc = timezone.LoadLocation(user.Timezone)
	}
	userToday := timezone.GetUserToday(loc)
	todayStr := userToday.Format("2006-01-02")

	if date >= todayStr {
		return nil
	}

	state, err := s.streakRepo.GetByUserID(ctx, userID)
	if err != nil || state == nil {
		return nil
	}

	// Restrict if the date is within the active plan cycle (current week/cycle)
	if date >= state.CycleStartDate {
		return fmt.Errorf("past workout logs within the active plan cycle cannot be added, edited, or deleted normally. please use a Restore Shield to recover missed days")
	}

	return nil
}

func (s *gymLogService) GetLogs(ctx context.Context, userID uuid.UUID, startDate, endDate *time.Time, workoutType *string) ([]models.GymLog, error) {
	logs, err := s.logRepo.GetLogs(ctx, userID, startDate, endDate, workoutType)
	if err != nil {
		return nil, fmt.Errorf("failed fetching logs: %w", err)
	}
	return logs, nil
}

func (s *gymLogService) SaveLog(ctx context.Context, userID uuid.UUID, date string, hours float64, workoutType string, notes *string) (*models.GymLog, error) {
	if err := s.checkPastLogRestriction(ctx, userID, date); err != nil {
		return nil, err
	}

	// Rule: If hours <= 0 and workoutType is not a Rest day, delete the log for that date
	if hours <= 0 && strings.ToLower(workoutType) != "rest" {
		if err := s.logRepo.DeleteByDate(ctx, userID, date); err != nil {
			return nil, fmt.Errorf("failed deleting log with zero hours: %w", err)
		}
		return nil, nil
	}

	if workoutType == "" {
		workoutType = "General"
	}

	log := &models.GymLog{
		UserID:      userID,
		Date:        date,
		Hours:       hours,
		WorkoutType: workoutType,
		Notes:       notes,
	}

	if err := s.logRepo.UpsertLog(ctx, log); err != nil {
		return nil, fmt.Errorf("failed saving log: %w", err)
	}

	// Clear any active checkin snooze once user logs session or rest day
	_ = s.userRepo.ClearCheckinSnooze(ctx, userID)

	return log, nil
}

func (s *gymLogService) DeleteLog(ctx context.Context, userID uuid.UUID, date string) error {
	if err := s.checkPastLogRestriction(ctx, userID, date); err != nil {
		return err
	}

	if err := s.logRepo.DeleteByDate(ctx, userID, date); err != nil {
		return fmt.Errorf("failed deleting log: %w", err)
	}
	return nil
}

func (s *gymLogService) ResetDemoLogs(ctx context.Context, userID uuid.UUID) error {
	// Retrieve active plan categories if possible
	var categories []string = []string{"Push", "Pull", "Legs", "Rest"}
	user, err := s.userRepo.GetByID(ctx, userID)
	if err == nil && user != nil && user.WeeklyPlanID != nil {
		plan, planErr := s.planRepo.GetByID(ctx, *user.WeeklyPlanID)
		if planErr == nil && plan != nil && len(plan.Categories) > 0 {
			categories = plan.Categories
		}
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	
	loc := time.UTC
	if user != nil && user.Timezone != "" {
		loc = timezone.LoadLocation(user.Timezone)
	}
	today := timezone.GetUserToday(loc)
	demoLogs := make([]models.GymLog, 0, 365)

	sampleNotes := map[string][]string{
		"Push":      {"Focused on heavy bench press", "Solid chest & triceps pump", "Overhead press progression feel great"},
		"Pull":      {"Heavy deadlifts and lat pulldowns", "Biceps & upper back focus", "Great grip strength today"},
		"Legs":      {"Squats 4x8 felt smooth", "Leg press & calf raises to finish", "Hamstrings on fire"},
		"Upper":     {"Upper body power day", "Incline dumbbell press & chest supported rows"},
		"Lower":     {"Romanian deadlifts and Bulgarian split squats", "Quad focus heavy day"},
		"Full Body": {"Compound lifts: Squats, Bench, Pullups", "High efficiency full body circuit"},
	}

	// Generate 365 days of demo historical logs
	for i := 364; i >= 0; i-- {
		logDate := today.AddDate(0, 0, -i).Format("2006-01-02")
		categoryIdx := (364 - i) % len(categories)
		workoutType := categories[categoryIdx]

		// 75% chance of logging a session, rest days marked or skipped
		if workoutType == "Rest" || r.Float32() < 0.20 {
			continue // Skip or rest day
		}

		// Random duration between 0.75h and 1.75h
		hours := 0.75 + r.Float64()*(1.75-0.75)
		hours = float64(int(hours*100)) / 100.0 // Round to 2 decimal places

		var note *string
		notesList, ok := sampleNotes[workoutType]
		if ok && len(notesList) > 0 {
			n := notesList[r.Intn(len(notesList))]
			note = &n
		}

		demoLogs = append(demoLogs, models.GymLog{
			ID:          uuid.New(),
			UserID:      userID,
			Date:        logDate,
			Hours:       hours,
			WorkoutType: workoutType,
			Notes:       note,
		})
	}

	if err := s.logRepo.ResetDemoLogs(ctx, userID, demoLogs); err != nil {
		return fmt.Errorf("failed resetting demo logs: %w", err)
	}

	return nil
}
