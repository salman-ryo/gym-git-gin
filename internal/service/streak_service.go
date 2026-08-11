package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gymgit/backend/internal/models"
	"gymgit/backend/internal/repository"
	"gymgit/backend/internal/timezone"

	"github.com/google/uuid"
)

type streakService struct {
	streakRepo repository.StreakRepository
	userRepo   repository.UserRepository
	planRepo   repository.PlanRepository
	logRepo    repository.GymLogRepository
}

// NewStreakService creates a new StreakService instance
func NewStreakService(streakRepo repository.StreakRepository, userRepo repository.UserRepository, planRepo repository.PlanRepository, logRepo repository.GymLogRepository) StreakService {
	return &streakService{
		streakRepo: streakRepo,
		userRepo:   userRepo,
		planRepo:   planRepo,
		logRepo:    logRepo,
	}
}

func (s *streakService) GetStreakState(ctx context.Context, userID uuid.UUID, loc *time.Location) (*models.StreakResponse, error) {
	if loc == nil {
		loc = time.UTC
	}
	userToday := timezone.GetUserToday(loc)
	todayStr := userToday.Format("2006-01-02")

	// 1. Retrieve user profile
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed retrieving user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user profile not found")
	}

	// 2. Retrieve user active plan
	var activePlan *models.WeeklyPlan
	if user.WeeklyPlanID != nil && *user.WeeklyPlanID != "" {
		activePlan, _ = s.planRepo.GetByID(ctx, *user.WeeklyPlanID)
	}
	targetDays := 4
	if activePlan != nil && len(activePlan.Categories) > 0 {
		targetDays = CalculateTargetDaysPerWeek(activePlan.Categories)
	}
	totalRestTokens := 7 - targetDays
	if totalRestTokens < 1 {
		totalRestTokens = 1
	}

	// 3. Retrieve or initialize user_streak_state
	state, err := s.streakRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed retrieving streak state: %w", err)
	}

	if state == nil {
		// Initialize state for new user
		cycleStart := userToday
		cycleEnd := cycleStart.AddDate(0, 0, 6)

		state = &models.UserStreakState{
			UserID:                    userID,
			CurrentStreak:             0,
			LongestStreak:             0,
			CycleStartDate:            cycleStart.Format("2006-01-02"),
			CycleEndDate:              cycleEnd.Format("2006-01-02"),
			WorkoutsCompletedInCycle: 0,
			WorkoutsTargetInCycle:    targetDays,
			RestTokensTotal:           totalRestTokens,
			RestTokensUsed:            0,
			AccuracyScore:             100,
			IsFrozen:                  false,
		}
	}

	// 4. Check Cycle Rollover if userToday > state.CycleEndDate
	if todayStr > state.CycleEndDate {
		// If a plan was queued, activate it now!
		if user.QueuedWeeklyPlanID != nil && *user.QueuedWeeklyPlanID != "" {
			newPlanID := *user.QueuedWeeklyPlanID
			if err := s.userRepo.UpdateWeeklyPlan(ctx, userID, newPlanID); err == nil {
				_ = s.userRepo.SetQueuedPlan(ctx, userID, nil)
				user.WeeklyPlanID = &newPlanID
				user.QueuedWeeklyPlanID = nil
				activePlan, _ = s.planRepo.GetByID(ctx, newPlanID)
				if activePlan != nil && len(activePlan.Categories) > 0 {
					targetDays = CalculateTargetDaysPerWeek(activePlan.Categories)
					totalRestTokens = 7 - targetDays
				}
			}
		}

		// Roll over to new 7-day cycle
		cycleStart, parseErr := time.ParseInLocation("2006-01-02", state.CycleEndDate, loc)
		if parseErr == nil {
			cycleStart = cycleStart.AddDate(0, 0, 1)
		} else {
			cycleStart = userToday
		}
		cycleEnd := cycleStart.AddDate(0, 0, 6)

		state.CycleStartDate = cycleStart.Format("2006-01-02")
		state.CycleEndDate = cycleEnd.Format("2006-01-02")
		state.WorkoutsCompletedInCycle = 0
		state.WorkoutsTargetInCycle = targetDays
		state.RestTokensTotal = totalRestTokens
		state.RestTokensUsed = 0
	}

	// 5. Fetch logs for current cycle window
	cStart, _ := time.ParseInLocation("2006-01-02", state.CycleStartDate, loc)
	cEnd, _ := time.ParseInLocation("2006-01-02", state.CycleEndDate, loc)

	logs, err := s.logRepo.GetLogs(ctx, userID, &cStart, &cEnd, nil)
	if err == nil {
		completed := 0
		for _, l := range logs {
			if l.Hours > 0 && !strings.EqualFold(l.WorkoutType, "Rest") {
				completed++
			}
		}
		state.WorkoutsCompletedInCycle = completed

		// Calculate days elapsed in cycle up to today
		daysElapsed := 0
		if todayStr >= state.CycleStartDate && todayStr <= state.CycleEndDate {
			tCur, _ := time.ParseInLocation("2006-01-02", todayStr, loc)
			daysElapsed = int(tCur.Sub(cStart).Hours()/24) + 1
		} else if todayStr > state.CycleEndDate {
			daysElapsed = 7
		}

		// Rest tokens used = daysElapsed - completed
		restUsed := daysElapsed - completed
		if restUsed < 0 {
			restUsed = 0
		}
		if restUsed > state.RestTokensTotal {
			restUsed = state.RestTokensTotal
		}
		state.RestTokensUsed = restUsed

		// Compute Split Accuracy Score for current cycle
		accuracyBreakdown := CalculateSplitAccuracy(logs, activePlan, state.CycleStartDate, state.CycleEndDate)
		state.AccuracyScore = accuracyBreakdown.AccuracyScore
	}

	// 6. Recalculate dynamic rolling streak
	allLogs, errAll := s.logRepo.GetLogs(ctx, userID, nil, nil, nil)
	if errAll == nil {
		scientificStreak := CalculateScientificStreak(allLogs, targetDays, 30, userToday)
		state.CurrentStreak = scientificStreak.CurrentStreak
		if state.CurrentStreak > state.LongestStreak {
			state.LongestStreak = state.CurrentStreak
		}
	}

	// 7. Persist updated streak state
	_ = s.streakRepo.UpsertState(ctx, state)

	// Calculate remaining rest tokens & days remaining in cycle
	remainingRest := state.RestTokensTotal - state.RestTokensUsed
	if remainingRest < 0 {
		remainingRest = 0
	}

	daysRemainingInCycle := 0
	if cEnd.After(userToday) || cEnd.Equal(userToday) {
		daysRemainingInCycle = int(cEnd.Sub(userToday).Hours()/24) + 1
	}

	return &models.StreakResponse{
		CurrentStreak:  state.CurrentStreak,
		LongestStreak:  state.LongestStreak,
		ComplianceRate: state.AccuracyScore,
		CycleInfo: models.StreakCycleInfo{
			CycleStartDate:            state.CycleStartDate,
			CycleEndDate:              state.CycleEndDate,
			WorkoutsCompletedInCycle: state.WorkoutsCompletedInCycle,
			WorkoutsTargetInCycle:    state.WorkoutsTargetInCycle,
			RestTokensTotal:           state.RestTokensTotal,
			RestTokensUsed:            state.RestTokensUsed,
			RestTokensRemaining:       remainingRest,
			DaysRemainingInCycle:      daysRemainingInCycle,
		},
		AccuracyScore:      state.AccuracyScore,
		QueuedWeeklyPlanID: user.QueuedWeeklyPlanID,
		IsFrozen:           state.IsFrozen,
		LastLoggedDate:     state.LastLoggedDate,
	}, nil
}

func (s *streakService) FreezeStreak(ctx context.Context, userID uuid.UUID, durationDays int, reason string) (*models.UserStreakState, error) {
	state, err := s.streakRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed retrieving streak state: %w", err)
	}
	if state == nil {
		now := time.Now().UTC()
		state = &models.UserStreakState{
			UserID:         userID,
			CycleStartDate: now.Format("2006-01-02"),
			CycleEndDate:   now.AddDate(0, 0, 6).Format("2006-01-02"),
		}
	}

	state.IsFrozen = true
	if err := s.streakRepo.UpsertState(ctx, state); err != nil {
		return nil, fmt.Errorf("failed updating freeze state: %w", err)
	}
	return state, nil
}

func (s *streakService) UnfreezeStreak(ctx context.Context, userID uuid.UUID) (*models.UserStreakState, error) {
	state, err := s.streakRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed retrieving streak state: %w", err)
	}
	if state == nil {
		return nil, fmt.Errorf("streak state not found")
	}

	state.IsFrozen = false
	if err := s.streakRepo.UpsertState(ctx, state); err != nil {
		return nil, fmt.Errorf("failed updating freeze state: %w", err)
	}
	return state, nil
}

