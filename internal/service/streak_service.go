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
	streakRepo    repository.StreakRepository
	userRepo      repository.UserRepository
	planRepo      repository.PlanRepository
	logRepo       repository.GymLogRepository
	inventoryRepo repository.InventoryRepository
}

// NewStreakService creates a new StreakService instance
func NewStreakService(
	streakRepo repository.StreakRepository,
	userRepo repository.UserRepository,
	planRepo repository.PlanRepository,
	logRepo repository.GymLogRepository,
	inventoryRepo repository.InventoryRepository,
) StreakService {
	return &streakService{
		streakRepo:    streakRepo,
		userRepo:      userRepo,
		planRepo:      planRepo,
		logRepo:       logRepo,
		inventoryRepo: inventoryRepo,
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

	// Normalize date strings to 10 characters (YYYY-MM-DD) to prevent parsing/comparison errors
	if len(state.CycleStartDate) > 10 {
		state.CycleStartDate = state.CycleStartDate[:10]
	}
	if len(state.CycleEndDate) > 10 {
		state.CycleEndDate = state.CycleEndDate[:10]
	}
	if state.LastLoggedDate != nil && len(*state.LastLoggedDate) > 10 {
		cleaned := (*state.LastLoggedDate)[:10]
		state.LastLoggedDate = &cleaned
	}

	// 4. Check Cycle Rollover while userToday > state.CycleEndDate
	for todayStr > state.CycleEndDate {
		// If a plan was queued, activate it on rollover!
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
					if totalRestTokens < 1 {
						totalRestTokens = 1
					}
				}
			}
		}

		// Roll over to next 7-day cycle
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

	// Always ensure target and total rest tokens are in sync with the active plan
	state.WorkoutsTargetInCycle = targetDays
	state.RestTokensTotal = totalRestTokens

	// 5. Fetch logs for current cycle window
	cStart, _ := time.ParseInLocation("2006-01-02", state.CycleStartDate, loc)
	cEnd, _ := time.ParseInLocation("2006-01-02", state.CycleEndDate, loc)

	logs, err := s.logRepo.GetLogs(ctx, userID, &cStart, &cEnd, nil)
	if err == nil {
		workoutDates := make(map[string]bool)
		restDates := make(map[string]bool)
		for _, l := range logs {
			if l.Hours > 0 && !strings.EqualFold(l.WorkoutType, "Rest") {
				workoutDates[l.Date] = true
			} else if strings.EqualFold(l.WorkoutType, "Rest") {
				restDates[l.Date] = true
			}
		}
		state.WorkoutsCompletedInCycle = len(workoutDates)

		// Calculate rest tokens used in cycle up to today:
		// - Past closed days in cycle with no workout count as 1 rest day.
		// - Today counts as rest only if explicitly logged as Rest.
		restUsed := 0
		for d := cStart; !d.After(userToday) && !d.After(cEnd); d = d.AddDate(0, 0, 1) {
			dStr := d.Format("2006-01-02")
			if workoutDates[dStr] {
				continue
			}
			if dStr == todayStr {
				if restDates[dStr] {
					restUsed++
				}
			} else {
				restUsed++
			}
		}

		if restUsed > state.RestTokensTotal {
			restUsed = state.RestTokensTotal
		}
		state.RestTokensUsed = restUsed

		// Compute Split Accuracy Score for current cycle
		accuracyBreakdown := CalculateSplitAccuracy(logs, activePlan, state.CycleStartDate, state.CycleEndDate)
		state.AccuracyScore = accuracyBreakdown.AccuracyScore
	}

	// 6. Recalculate dynamic rolling streak & latest workout log date
	allLogs, errAll := s.logRepo.GetLogs(ctx, userID, nil, nil, nil)
	if errAll == nil {
		scientificStreak := CalculateScientificStreak(allLogs, targetDays, 30, userToday)
		state.CurrentStreak = scientificStreak.CurrentStreak
		if state.CurrentStreak > state.LongestStreak {
			state.LongestStreak = state.CurrentStreak
		}

		var latestLogDate *string
		for _, l := range allLogs {
			if l.Hours > 0 && !strings.EqualFold(l.WorkoutType, "Rest") {
				if latestLogDate == nil || l.Date > *latestLogDate {
					d := l.Date
					latestLogDate = &d
				}
			}
		}
		state.LastLoggedDate = latestLogDate
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

	// 8. Calculate Streak Warning Event (Is At Risk?)
	var warningEvent *models.StreakWarningEvent
	var brokenEvent *models.StreakBrokenEvent

	todayLogLogged := false
	if logs != nil {
		for _, l := range logs {
			if l.Date == todayStr && l.Hours > 0 && !strings.EqualFold(l.WorkoutType, "Rest") {
				todayLogLogged = true
				break
			}
		}
	}

	userMidnight := time.Date(userToday.Year(), userToday.Month(), userToday.Day(), 23, 59, 59, 0, loc)
	userNow := time.Now().In(loc)
	hoursRemaining := int(userMidnight.Sub(userNow).Hours())
	if hoursRemaining < 0 {
		hoursRemaining = 0
	}

	if !state.IsFrozen && !todayLogLogged && remainingRest == 0 && state.CurrentStreak > 0 {
		warningEvent = &models.StreakWarningEvent{
			IsAtRisk:       true,
			HoursRemaining: hoursRemaining,
			RestTokensLeft: remainingRest,
			Message:        fmt.Sprintf("Streak at risk! Complete your workout within %d hours to maintain your %d-day streak.", hoursRemaining, state.CurrentStreak),
		}
	}

	// 9. Calculate Streak Broken Event
	if state.CurrentStreak == 0 && state.LongestStreak > 0 && !state.IsFrozen {
		gapAnalysis := FindLastStreakDateAndGap(allLogs, targetDays, userToday)
		if gapAnalysis.MissedDaysCount > 0 {
			shieldsCount := 0
			if s.inventoryRepo != nil {
				shieldsCount, _ = s.inventoryRepo.GetItemQuantity(ctx, userID, "RESTORE_SHIELD")
			}

			canRestore := shieldsCount >= gapAnalysis.MissedDaysCount && gapAnalysis.MissedDaysCount <= 9
			firstBrokenDate := gapAnalysis.MissedDates[0]
			canRestoreUntil := userToday.Format("2006-01-02")

			prevStreak := gapAnalysis.PreviousStreak
			if prevStreak == 0 {
				prevStreak = state.LongestStreak
			}

			brokenEvent = &models.StreakBrokenEvent{
				PreviousStreak:         prevStreak,
				LastStreakDate:         gapAnalysis.LastStreakDate,
				BrokenOn:               firstBrokenDate,
				MissedDaysCount:        gapAnalysis.MissedDaysCount,
				RequiredShields:        gapAnalysis.MissedDaysCount,
				RestoreShieldAvailable: canRestore,
				RestoreShieldsCount:    shieldsCount,
				MissedDates:            gapAnalysis.MissedDates,
				CanRestoreUntil:        canRestoreUntil,
			}
		}
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
		StreakBrokenEvent:  brokenEvent,
		StreakWarningEvent: warningEvent,
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

