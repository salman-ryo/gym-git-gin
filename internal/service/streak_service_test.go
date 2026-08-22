package service_test

import (
	"context"
	"testing"
	"time"

	"gymgit/backend/internal/models"
	"gymgit/backend/internal/service"
	"gymgit/backend/internal/timezone"

	"github.com/google/uuid"
)

func strPtr(s string) *string {
	return &s
}

type mockUserRepo struct {
	user *models.User
}

func (m *mockUserRepo) GetByAuthUserID(ctx context.Context, authUserID uuid.UUID) (*models.User, error) {
	return m.user, nil
}
func (m *mockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	return m.user, nil
}
func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	return m.user, nil
}
func (m *mockUserRepo) Create(ctx context.Context, user *models.User) error {
	m.user = user
	return nil
}
func (m *mockUserRepo) UpdateWeeklyPlan(ctx context.Context, userID uuid.UUID, planID string) error {
	if m.user != nil {
		m.user.WeeklyPlanID = &planID
	}
	return nil
}
func (m *mockUserRepo) SetQueuedPlan(ctx context.Context, userID uuid.UUID, planID *string) error {
	if m.user != nil {
		m.user.QueuedWeeklyPlanID = planID
	}
	return nil
}
func (m *mockUserRepo) UpdateTimezone(ctx context.Context, userID uuid.UUID, tz string) error {
	if m.user != nil {
		m.user.Timezone = tz
	}
	return nil
}
func (m *mockUserRepo) UpdateAuthUserID(ctx context.Context, id uuid.UUID, authUserID uuid.UUID) error {
	return nil
}
func (m *mockUserRepo) SetCheckinSnooze(ctx context.Context, userID uuid.UUID, dateStr string, snoozedAt time.Time) error {
	if m.user != nil {
		m.user.CheckinSnoozedDate = &dateStr
		m.user.CheckinSnoozedAt = &snoozedAt
	}
	return nil
}
func (m *mockUserRepo) ClearCheckinSnooze(ctx context.Context, userID uuid.UUID) error {
	if m.user != nil {
		m.user.CheckinSnoozedDate = nil
		m.user.CheckinSnoozedAt = nil
	}
	return nil
}

type mockPlanRepo struct {
	plans map[string]*models.WeeklyPlan
}

func (m *mockPlanRepo) GetAll(ctx context.Context) ([]models.WeeklyPlan, error) {
	var result []models.WeeklyPlan
	for _, p := range m.plans {
		result = append(result, *p)
	}
	return result, nil
}
func (m *mockPlanRepo) GetByID(ctx context.Context, id string) (*models.WeeklyPlan, error) {
	if p, ok := m.plans[id]; ok {
		return p, nil
	}
	return nil, nil
}
func (m *mockPlanRepo) Create(ctx context.Context, plan *models.WeeklyPlan) error {
	m.plans[plan.ID] = plan
	return nil
}

type mockLogRepo struct {
	logs []models.GymLog
}

func (m *mockLogRepo) GetLogs(ctx context.Context, userID uuid.UUID, startDate, endDate *time.Time, workoutType *string) ([]models.GymLog, error) {
	if startDate != nil && endDate != nil {
		var filtered []models.GymLog
		sStr := startDate.Format("2006-01-02")
		eStr := endDate.Format("2006-01-02")
		for _, l := range m.logs {
			if l.Date >= sStr && l.Date <= eStr {
				filtered = append(filtered, l)
			}
		}
		return filtered, nil
	}
	return m.logs, nil
}
func (m *mockLogRepo) GetByDate(ctx context.Context, userID uuid.UUID, date string) (*models.GymLog, error) {
	return nil, nil
}
func (m *mockLogRepo) UpsertLog(ctx context.Context, log *models.GymLog) error {
	m.logs = append(m.logs, *log)
	return nil
}
func (m *mockLogRepo) DeleteByDate(ctx context.Context, userID uuid.UUID, date string) error {
	return nil
}
func (m *mockLogRepo) ResetDemoLogs(ctx context.Context, userID uuid.UUID, logs []models.GymLog) error {
	m.logs = logs
	return nil
}

type mockStreakRepo struct {
	state *models.UserStreakState
}

func (m *mockStreakRepo) GetByUserID(ctx context.Context, userID uuid.UUID) (*models.UserStreakState, error) {
	return m.state, nil
}
func (m *mockStreakRepo) UpsertState(ctx context.Context, state *models.UserStreakState) error {
	m.state = state
	return nil
}

func TestStreakService_GetStreakState(t *testing.T) {
	userID := uuid.New()
	planID := "ppl-standard"

	user := &models.User{
		ID:           userID,
		AuthUserID:   userID,
		Email:        "test@gymgit.com",
		WeeklyPlanID: &planID,
		Timezone:     "America/New_York",
	}

	userRepo := &mockUserRepo{user: user}
	planRepo := &mockPlanRepo{
		plans: map[string]*models.WeeklyPlan{
			"ppl-standard": {
				ID:          "ppl-standard",
				Name:        "PPL Standard",
				Categories: []string{"Push", "Pull", "Legs", "Cardio"},
			},
		},
	}
	logRepo := &mockLogRepo{}
	streakRepo := &mockStreakRepo{}

	streakSvc := service.NewStreakService(streakRepo, userRepo, planRepo, logRepo, nil)
	loc := timezone.LoadLocation("America/New_York")

	resp, err := streakSvc.GetStreakState(context.Background(), userID, loc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp == nil {
		t.Fatalf("expected non-nil response")
	}

	if resp.CycleInfo.WorkoutsTargetInCycle != 4 {
		t.Errorf("expected 4 workout targets for 4-category plan, got %d", resp.CycleInfo.WorkoutsTargetInCycle)
	}

	if resp.CycleInfo.RestTokensTotal != 3 {
		t.Errorf("expected 3 rest tokens for 4-day plan, got %d", resp.CycleInfo.RestTokensTotal)
	}
}

func TestStreakService_StreakWarningAndBrokenEvents(t *testing.T) {
	userID := uuid.New()
	planID := "ppl-standard"

	user := &models.User{
		ID:           userID,
		AuthUserID:   userID,
		Email:        "test@gymgit.com",
		WeeklyPlanID: &planID,
		Timezone:     "America/New_York",
	}

	userRepo := &mockUserRepo{user: user}
	planRepo := &mockPlanRepo{
		plans: map[string]*models.WeeklyPlan{
			"ppl-standard": {
				ID:          "ppl-standard",
				Name:        "PPL Standard",
				Categories: []string{"Push", "Pull", "Legs", "Cardio"},
			},
		},
	}
	logRepo := &mockLogRepo{}
	// User streak was broken: current_streak 0, longest_streak 15
	streakRepo := &mockStreakRepo{
		state: &models.UserStreakState{
			UserID:          userID,
			CurrentStreak:   0,
			LongestStreak:   15,
			CycleStartDate:  "2026-08-01",
			CycleEndDate:    "2026-08-07",
			RestTokensTotal: 3,
			RestTokensUsed:  3,
			IsFrozen:        false,
		},
	}

	streakSvc := service.NewStreakService(streakRepo, userRepo, planRepo, logRepo, nil)
	loc := timezone.LoadLocation("America/New_York")

	resp, err := streakSvc.GetStreakState(context.Background(), userID, loc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.StreakBrokenEvent == nil {
		t.Fatalf("expected StreakBrokenEvent to be non-nil for broken streak")
	}

	if resp.StreakBrokenEvent.PreviousStreak != 15 {
		t.Errorf("expected previous streak 15, got %d", resp.StreakBrokenEvent.PreviousStreak)
	}

	// When CurrentStreak is 0, StreakWarningEvent must be nil (no streak to decay)
	if resp.StreakWarningEvent != nil {
		t.Fatalf("expected StreakWarningEvent to be nil when streak is already 0")
	}

	// 2. Test active streak at risk (CurrentStreak > 0 and 0 rest tokens remaining)
	today := timezone.GetUserToday(loc)
	var activeLogs []models.GymLog
	for i := 1; i <= 10; i++ {
		activeLogs = append(activeLogs, models.GymLog{
			UserID:      userID,
			Date:        today.AddDate(0, 0, -i).Format("2006-01-02"),
			Hours:       1.0,
			WorkoutType: "Push",
		})
	}
	activeLogRepo := &mockLogRepo{logs: activeLogs}
	activeStreakRepo := &mockStreakRepo{
		state: &models.UserStreakState{
			UserID:          userID,
			CurrentStreak:   10,
			LongestStreak:   15,
			CycleStartDate:  today.AddDate(0, 0, -3).Format("2006-01-02"),
			CycleEndDate:    today.AddDate(0, 0, 3).Format("2006-01-02"),
			RestTokensTotal: 1,
			RestTokensUsed:  1,
			IsFrozen:        false,
		},
	}
	activeStreakSvc := service.NewStreakService(activeStreakRepo, userRepo, planRepo, activeLogRepo, nil)
	activeResp, errActive := activeStreakSvc.GetStreakState(context.Background(), userID, loc)
	if errActive != nil {
		t.Fatalf("unexpected error: %v", errActive)
	}

	if activeResp.StreakWarningEvent == nil {
		t.Fatalf("expected StreakWarningEvent to be non-nil when active streak has exhausted rest tokens")
	}

	if !activeResp.StreakWarningEvent.IsAtRisk {
		t.Errorf("expected IsAtRisk to be true for active streak")
	}
}

func TestStreak_TodayUnlogged_DoesNotBreakStreak(t *testing.T) {
	userID := uuid.New()
	user := &models.User{
		ID:           userID,
		WeeklyPlanID: strPtr("ppl-standard"),
		Timezone:     "UTC",
	}
	userRepo := &mockUserRepo{user: user}
	planRepo := &mockPlanRepo{
		plans: map[string]*models.WeeklyPlan{
			"ppl-standard": {
				ID:         "ppl-standard",
				Name:       "PPL Standard",
				Categories: []string{"Push", "Pull", "Legs", "Rest"},
			},
		},
	}

	loc := time.UTC
	today := timezone.GetUserToday(loc)

	// User logged yesterday and 4 preceding days. Today is completely unlogged.
	var logs []models.GymLog
	for i := 1; i <= 5; i++ {
		logs = append(logs, models.GymLog{
			UserID:      userID,
			Date:        today.AddDate(0, 0, -i).Format("2006-01-02"),
			Hours:       1.0,
			WorkoutType: "Push",
		})
	}
	logRepo := &mockLogRepo{logs: logs}
	streakRepo := &mockStreakRepo{
		state: &models.UserStreakState{
			UserID:          userID,
			CurrentStreak:   5,
			LongestStreak:   5,
			CycleStartDate:  today.AddDate(0, 0, -3).Format("2006-01-02"),
			CycleEndDate:    today.AddDate(0, 0, 3).Format("2006-01-02"),
			RestTokensTotal: 3,
			RestTokensUsed:  0,
			IsFrozen:        false,
		},
	}

	streakSvc := service.NewStreakService(streakRepo, userRepo, planRepo, logRepo, nil)
	resp, err := streakSvc.GetStreakState(context.Background(), userID, loc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Active streak must NOT be 0 simply because today is not yet logged!
	if resp.CurrentStreak < 5 {
		t.Errorf("expected CurrentStreak to be at least 5 (alive from yesterday), got %d", resp.CurrentStreak)
	}

	// Must NOT fire a StreakBrokenEvent
	if resp.StreakBrokenEvent != nil {
		t.Errorf("expected StreakBrokenEvent to be nil during in-progress today, but got non-nil")
	}
}

func TestStreak_LastStreakDate_5DaysGap(t *testing.T) {
	userID := uuid.New()
	user := &models.User{
		ID:           userID,
		WeeklyPlanID: strPtr("ppl-standard"),
		Timezone:     "UTC",
	}
	userRepo := &mockUserRepo{user: user}
	planRepo := &mockPlanRepo{
		plans: map[string]*models.WeeklyPlan{
			"ppl-standard": {
				ID:         "ppl-standard",
				Name:       "PPL Standard",
				Categories: []string{"Push", "Pull", "Legs", "Cardio", "Upper", "Lower"},
			},
		},
	}

	loc := time.UTC
	today := timezone.GetUserToday(loc)

	// User last logged a workout 5 days ago (and 9 days before that = 10 day streak)
	var logs []models.GymLog
	for i := 5; i <= 14; i++ {
		logs = append(logs, models.GymLog{
			UserID:      userID,
			Date:        today.AddDate(0, 0, -i).Format("2006-01-02"),
			Hours:       1.0,
			WorkoutType: "Push",
		})
	}
	logRepo := &mockLogRepo{logs: logs}
	streakRepo := &mockStreakRepo{
		state: &models.UserStreakState{
			UserID:          userID,
			CurrentStreak:   0,
			LongestStreak:   10,
			CycleStartDate:  today.AddDate(0, 0, -3).Format("2006-01-02"),
			CycleEndDate:    today.AddDate(0, 0, 3).Format("2006-01-02"),
			RestTokensTotal: 1,
			RestTokensUsed:  1,
			IsFrozen:        false,
		},
	}

	streakSvc := service.NewStreakService(streakRepo, userRepo, planRepo, logRepo, nil)
	resp, err := streakSvc.GetStreakState(context.Background(), userID, loc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.CurrentStreak != 0 {
		t.Errorf("expected CurrentStreak to be 0 for 5-day inactive gap, got %d", resp.CurrentStreak)
	}

	if resp.StreakBrokenEvent == nil {
		t.Fatalf("expected StreakBrokenEvent to be non-nil for broken streak")
	}

	// Scientific rolling window evaluates days -5, -6... as active.
	// Day -4 and day -3 qualify as compliant rest days because active sessions in [D-6, D] >= required (5).
	// On day -2, active sessions in window drop to 4 (< 5), so day -2 is the first non-compliant day.
	// Hence, last compliant date is day -3 (2026-08-19), and missed dates are days -2 and -1 (2 missed days).
	expectedLastStreakDate := today.AddDate(0, 0, -3).Format("2006-01-02")
	if resp.StreakBrokenEvent.LastStreakDate != expectedLastStreakDate {
		t.Errorf("expected LastStreakDate %s, got %s", expectedLastStreakDate, resp.StreakBrokenEvent.LastStreakDate)
	}

	if resp.StreakBrokenEvent.MissedDaysCount != 2 {
		t.Errorf("expected MissedDaysCount 2 (days -2, -1), got %d", resp.StreakBrokenEvent.MissedDaysCount)
	}

	if resp.StreakBrokenEvent.RequiredShields != resp.StreakBrokenEvent.MissedDaysCount {
		t.Errorf("expected RequiredShields to equal MissedDaysCount (%d), got %d", resp.StreakBrokenEvent.MissedDaysCount, resp.StreakBrokenEvent.RequiredShields)
	}
}

func TestInventory_MaxItemLimit_9(t *testing.T) {
	invRepo := &mockInventoryRepo{
		quantities: map[string]int{"RESTORE_SHIELD": 8},
	}

	newQty, err := invRepo.AddItemQuantity(context.Background(), uuid.New(), "RESTORE_SHIELD", 3)
	if err != nil {
		t.Fatalf("unexpected error adding item quantity: %v", err)
	}

	if newQty != 9 {
		t.Errorf("expected newQty to be capped at 9, got %d", newQty)
	}
}

