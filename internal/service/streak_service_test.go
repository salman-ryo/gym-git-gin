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

	if resp.StreakWarningEvent == nil {
		t.Fatalf("expected StreakWarningEvent to be non-nil when rest tokens are exhausted")
	}

	if !resp.StreakWarningEvent.IsAtRisk {
		t.Errorf("expected IsAtRisk to be true")
	}
}

