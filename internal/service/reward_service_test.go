package service_test

import (
	"context"
	"testing"
	"time"

	"gymgit/backend/internal/models"
	"gymgit/backend/internal/service"

	"github.com/google/uuid"
)

// mockRewardRepo simulates database operations for reward plans and milestones
type mockRewardRepo struct {
	plans      map[string]*models.RewardPlan
	milestones map[string][]models.RewardPlanMilestone
	claims     map[string][]models.UserClaimedReward
}

func newMockRewardRepo() *mockRewardRepo {
	r := &mockRewardRepo{
		plans:      make(map[string]*models.RewardPlan),
		milestones: make(map[string][]models.RewardPlanMilestone),
		claims:     make(map[string][]models.UserClaimedReward),
	}

	// Seed default plan
	defaultPlanID := "default-streak-roadmap"
	r.plans[defaultPlanID] = &models.RewardPlan{
		ID:          defaultPlanID,
		Name:        "Default Roadmap",
		Description: "Standard roadmap",
		IsActive:    true,
	}

	m7ID := uuid.New()
	m10ID := uuid.New()
	m11ID := uuid.New()

	r.milestones[defaultPlanID] = []models.RewardPlanMilestone{
		{ID: m7ID, PlanID: defaultPlanID, StreakTarget: 7, ItemID: "RESTORE_SHIELD", Quantity: 1, Title: "7-Day Shield Anchor"},
		{ID: m10ID, PlanID: defaultPlanID, StreakTarget: 10, ItemID: "STREAK_FREEZE_TOKEN", Quantity: 1, Title: "10-Day Ice Defender"},
		{ID: m11ID, PlanID: defaultPlanID, StreakTarget: 11, ItemID: "RESTORE_SHIELD", Quantity: 5, Title: "11-Day Shield Power"},
	}

	return r
}

func (m *mockRewardRepo) GetActiveRewardPlan(ctx context.Context, planID string) (*models.RewardPlan, error) {
	plan, exists := m.plans[planID]
	if !exists {
		return nil, nil
	}
	cp := *plan
	cp.Milestones = m.milestones[planID]
	return &cp, nil
}

func (m *mockRewardRepo) GetAllRewardPlans(ctx context.Context) ([]models.RewardPlan, error) {
	var result []models.RewardPlan
	for _, p := range m.plans {
		result = append(result, *p)
	}
	return result, nil
}

func (m *mockRewardRepo) CreateRewardPlan(ctx context.Context, plan *models.RewardPlan) error {
	m.plans[plan.ID] = plan
	return nil
}

func (m *mockRewardRepo) DeleteRewardPlan(ctx context.Context, planID string) error {
	delete(m.plans, planID)
	delete(m.milestones, planID)
	return nil
}

func (m *mockRewardRepo) GetRewardMilestones(ctx context.Context, planID string) ([]models.RewardPlanMilestone, error) {
	return m.milestones[planID], nil
}

func (m *mockRewardRepo) UpsertMilestone(ctx context.Context, milestone *models.RewardPlanMilestone) error {
	if milestone.ID == uuid.Nil {
		milestone.ID = uuid.New()
	}
	list := m.milestones[milestone.PlanID]
	updated := false
	for i, existing := range list {
		if existing.StreakTarget == milestone.StreakTarget && existing.ItemID == milestone.ItemID {
			list[i] = *milestone
			updated = true
			break
		}
	}
	if !updated {
		list = append(list, *milestone)
	}
	m.milestones[milestone.PlanID] = list
	return nil
}

func (m *mockRewardRepo) DeleteMilestone(ctx context.Context, milestoneID uuid.UUID) error {
	for planID, list := range m.milestones {
		var newList []models.RewardPlanMilestone
		for _, item := range list {
			if item.ID != milestoneID {
				newList = append(newList, item)
			}
		}
		m.milestones[planID] = newList
	}
	return nil
}

func (m *mockRewardRepo) GetUserClaimedRewards(ctx context.Context, userID uuid.UUID, planID string) ([]models.UserClaimedReward, error) {
	key := userID.String() + "_" + planID
	return m.claims[key], nil
}

func (m *mockRewardRepo) IsRewardClaimed(ctx context.Context, userID uuid.UUID, planID string, streakTarget int, itemID string) (bool, error) {
	key := userID.String() + "_" + planID
	claims := m.claims[key]
	for _, c := range claims {
		if c.StreakTarget == streakTarget && c.ItemID == itemID {
			return true, nil
		}
	}
	return false, nil
}

func (m *mockRewardRepo) ClaimReward(ctx context.Context, claim *models.UserClaimedReward) error {
	key := claim.UserID.String() + "_" + claim.PlanID
	claim.ID = uuid.New()
	claim.ClaimedAt = time.Now()
	m.claims[key] = append(m.claims[key], *claim)
	return nil
}

func createTestItemRepo() *mockItemRepo {
	icon1 := "shield-icon"
	icon2 := "snowflake-icon"
	icon3 := "zap-icon"
	return &mockItemRepo{
		items: map[string]*models.Item{
			"RESTORE_SHIELD":      {ID: "RESTORE_SHIELD", Name: "Restore Shield", EffectType: models.EffectTypeInstantUse, Rarity: "rare", IconSlug: &icon1},
			"STREAK_FREEZE_TOKEN": {ID: "STREAK_FREEZE_TOKEN", Name: "Streak Freeze Token", EffectType: models.EffectTypeTimeBased, DurationSeconds: 86400, Rarity: "rare", IconSlug: &icon2},
			"XP_BOOST":            {ID: "XP_BOOST", Name: "XP Boost Token", EffectType: models.EffectTypeTimeBased, DurationSeconds: 604800, Rarity: "epic", IconSlug: &icon3},
		},
	}
}

func createTestInventoryRepo() *mockInventoryRepo {
	return &mockInventoryRepo{
		quantities: make(map[string]int),
	}
}

func createTestStreakRepo() *mockStreakRepo {
	return &mockStreakRepo{}
}

func createTestLogRepo() *mockLogRepo {
	return &mockLogRepo{}
}

func TestGetRoadmap_StatusEvaluation(t *testing.T) {
	rewardRepo := newMockRewardRepo()
	itemRepo := createTestItemRepo()
	inventoryRepo := createTestInventoryRepo()
	streakRepo := createTestStreakRepo()
	logRepo := createTestLogRepo()

	svc := service.NewRewardService(rewardRepo, itemRepo, inventoryRepo, streakRepo, logRepo)

	userID := uuid.New()
	// Set user streak state to 10 days
	_ = streakRepo.UpsertState(context.Background(), &models.UserStreakState{
		UserID:        userID,
		CurrentStreak: 10,
		LongestStreak: 10,
	})

	roadmap, err := svc.GetRoadmap(context.Background(), userID, "default-streak-roadmap")
	if err != nil {
		t.Fatalf("expected no error fetching roadmap, got: %v", err)
	}

	if len(roadmap) != 3 {
		t.Fatalf("expected 3 milestones in roadmap, got %d", len(roadmap))
	}

	// 7 days -> CLAIMABLE
	if roadmap[0].StreakTarget != 7 || roadmap[0].Status != models.MilestoneStatusClaimable {
		t.Errorf("expected 7-day milestone to be CLAIMABLE, got status: %s", roadmap[0].Status)
	}

	// 10 days -> CLAIMABLE
	if roadmap[1].StreakTarget != 10 || roadmap[1].Status != models.MilestoneStatusClaimable {
		t.Errorf("expected 10-day milestone to be CLAIMABLE, got status: %s", roadmap[1].Status)
	}

	// 11 days -> LOCKED
	if roadmap[2].StreakTarget != 11 || roadmap[2].Status != models.MilestoneStatusLocked {
		t.Errorf("expected 11-day milestone to be LOCKED, got status: %s", roadmap[2].Status)
	}
}

func TestClaimReward_SuccessAndDoubleClaim(t *testing.T) {
	rewardRepo := newMockRewardRepo()
	itemRepo := createTestItemRepo()
	inventoryRepo := createTestInventoryRepo()
	streakRepo := createTestStreakRepo()
	logRepo := createTestLogRepo()

	svc := service.NewRewardService(rewardRepo, itemRepo, inventoryRepo, streakRepo, logRepo)

	userID := uuid.New()
	_ = streakRepo.UpsertState(context.Background(), &models.UserStreakState{
		UserID:        userID,
		CurrentStreak: 12,
		LongestStreak: 12,
	})

	// 1. Claim Day 11 milestone (+5 Restore Shields)
	res, err := svc.ClaimReward(context.Background(), userID, "default-streak-roadmap", 11, "RESTORE_SHIELD")
	if err != nil {
		t.Fatalf("expected successful claim for Day 11, got error: %v", err)
	}

	if !res.Success || res.QuantityAwarded != 5 {
		t.Errorf("expected 5 Restore Shields awarded, got: %d", res.QuantityAwarded)
	}

	// Check inventory balance
	qty, _ := inventoryRepo.GetItemQuantity(context.Background(), userID, "RESTORE_SHIELD")
	if qty != 5 {
		t.Errorf("expected inventory quantity of 5, got %d", qty)
	}

	// 2. Attempt double claim for Day 11 -> Should fail
	_, errDouble := svc.ClaimReward(context.Background(), userID, "default-streak-roadmap", 11, "RESTORE_SHIELD")
	if errDouble == nil {
		t.Fatalf("expected error on double claim, got nil")
	}

	// 3. Verify roadmap status is now CLAIMED
	roadmap, _ := svc.GetRoadmap(context.Background(), userID, "default-streak-roadmap")
	for _, m := range roadmap {
		if m.StreakTarget == 11 {
			if m.Status != models.MilestoneStatusClaimed {
				t.Errorf("expected Day 11 status to be CLAIMED, got %s", m.Status)
			}
		}
	}
}

func TestAdminMilestoneCRUD(t *testing.T) {
	rewardRepo := newMockRewardRepo()
	itemRepo := createTestItemRepo()
	inventoryRepo := createTestInventoryRepo()
	streakRepo := createTestStreakRepo()
	logRepo := createTestLogRepo()

	svc := service.NewRewardService(rewardRepo, itemRepo, inventoryRepo, streakRepo, logRepo)

	// Admin adds Day 15 milestone: +2 XP_BOOST
	m, err := svc.UpsertMilestone(context.Background(), "default-streak-roadmap", models.UpsertMilestoneRequest{
		StreakTarget: 15,
		ItemID:       "XP_BOOST",
		Quantity:     2,
		Title:        "15-Day XP Overdrive",
	})

	if err != nil {
		t.Fatalf("expected success adding milestone, got: %v", err)
	}
	if m.StreakTarget != 15 || m.Quantity != 2 {
		t.Errorf("milestone creation mismatch: got target %d, quantity %d", m.StreakTarget, m.Quantity)
	}

	// Admin deletes the Day 15 milestone
	errDel := svc.DeleteMilestone(context.Background(), m.ID)
	if errDel != nil {
		t.Fatalf("expected success deleting milestone, got: %v", errDel)
	}

	milestones, _ := rewardRepo.GetRewardMilestones(context.Background(), "default-streak-roadmap")
	for _, item := range milestones {
		if item.StreakTarget == 15 {
			t.Errorf("expected Day 15 milestone to be deleted, but found in list")
		}
	}
}
