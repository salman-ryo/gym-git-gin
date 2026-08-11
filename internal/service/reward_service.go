package service

import (
	"context"
	"fmt"

	"gymgit/backend/internal/models"
	"gymgit/backend/internal/repository"

	"github.com/google/uuid"
)

type rewardService struct {
	rewardRepo    repository.RewardRepository
	itemRepo      repository.ItemRepository
	inventoryRepo repository.InventoryRepository
	streakRepo    repository.StreakRepository
	logRepo       repository.GymLogRepository
}

// NewRewardService creates a new RewardService instance
func NewRewardService(
	rewardRepo repository.RewardRepository,
	itemRepo repository.ItemRepository,
	inventoryRepo repository.InventoryRepository,
	streakRepo repository.StreakRepository,
	logRepo repository.GymLogRepository,
) RewardService {
	return &rewardService{
		rewardRepo:    rewardRepo,
		itemRepo:      itemRepo,
		inventoryRepo: inventoryRepo,
		streakRepo:    streakRepo,
		logRepo:       logRepo,
	}
}

func (s *rewardService) GetRoadmap(ctx context.Context, userID uuid.UUID, planID string) ([]models.RoadmapMilestoneResponse, error) {
	var plan *models.RewardPlan
	var err error

	if planID != "" {
		plan, err = s.rewardRepo.GetActiveRewardPlan(ctx, planID)
	}

	if err != nil || plan == nil {
		planID = "default-streak-roadmap"
		plan, err = s.rewardRepo.GetActiveRewardPlan(ctx, planID)
		if err != nil {
			return nil, fmt.Errorf("failed fetching default reward plan: %w", err)
		}
		if plan == nil {
			return nil, fmt.Errorf("default reward plan not found")
		}
	}

	// Fetch user streak state
	streakState, err := s.streakRepo.GetByUserID(ctx, userID)
	userCurrentStreak := 0
	userLongestStreak := 0
	if err == nil && streakState != nil {
		userCurrentStreak = streakState.CurrentStreak
		userLongestStreak = streakState.LongestStreak
	}

	// Max streak achieved by user determines qualification
	userMaxStreak := userCurrentStreak
	if userLongestStreak > userMaxStreak {
		userMaxStreak = userLongestStreak
	}

	// Fetch items catalog for metadata lookups
	allItems, _ := s.itemRepo.GetAll(ctx)
	itemMap := make(map[string]models.Item)
	for _, item := range allItems {
		itemMap[item.ID] = item
	}

	// Fetch user claimed rewards
	claimedRewards, err := s.rewardRepo.GetUserClaimedRewards(ctx, userID, planID)
	if err != nil {
		return nil, fmt.Errorf("failed fetching claimed rewards: %w", err)
	}

	claimedMap := make(map[string]models.UserClaimedReward)
	for _, claim := range claimedRewards {
		key := fmt.Sprintf("%d_%s", claim.StreakTarget, claim.ItemID)
		claimedMap[key] = claim
	}

	var roadmap []models.RoadmapMilestoneResponse
	for _, m := range plan.Milestones {
		itemName := m.ItemID
		itemIcon := ""
		rarity := ""
		if item, exists := itemMap[m.ItemID]; exists {
			itemName = item.Name
			if item.IconSlug != nil {
				itemIcon = *item.IconSlug
			}
			rarity = item.Rarity
		}


		key := fmt.Sprintf("%d_%s", m.StreakTarget, m.ItemID)
		claim, isClaimed := claimedMap[key]

		var status models.MilestoneStatus
		var claimedAt *models.UserClaimedReward

		if isClaimed {
			status = models.MilestoneStatusClaimed
			claimedAt = &claim
		} else if userMaxStreak >= m.StreakTarget {
			status = models.MilestoneStatusClaimable
		} else {
			status = models.MilestoneStatusLocked
		}

		resp := models.RoadmapMilestoneResponse{
			MilestoneID:  m.ID,
			PlanID:       m.PlanID,
			StreakTarget: m.StreakTarget,
			ItemID:       m.ItemID,
			ItemName:     itemName,
			ItemIcon:     itemIcon,
			Rarity:       rarity,
			Quantity:     m.Quantity,
			Title:        m.Title,
			Description:  m.Description,
			BadgeSlug:    m.BadgeSlug,
			Status:       status,
		}
		if claimedAt != nil {
			resp.ClaimedAt = &claimedAt.ClaimedAt
		}

		roadmap = append(roadmap, resp)
	}

	return roadmap, nil
}

func (s *rewardService) ClaimReward(ctx context.Context, userID uuid.UUID, planID string, streakTarget int, itemID string) (*models.ClaimRewardResult, error) {
	if planID == "" {
		planID = "default-streak-roadmap"
	}
	if streakTarget <= 0 || itemID == "" {
		return nil, fmt.Errorf("streak_target and item_id are required to claim a reward")
	}

	// 1. Verify item exists in catalog
	item, err := s.itemRepo.GetByID(ctx, itemID)
	if err != nil || item == nil {
		return nil, fmt.Errorf("invalid item_id '%s'", itemID)
	}

	// 2. Verify milestone target exists in plan
	milestones, err := s.rewardRepo.GetRewardMilestones(ctx, planID)
	if err != nil {
		return nil, fmt.Errorf("failed fetching plan milestones: %w", err)
	}

	var targetMilestone *models.RewardPlanMilestone
	for i := range milestones {
		if milestones[i].StreakTarget == streakTarget && milestones[i].ItemID == itemID {
			targetMilestone = &milestones[i]
			break
		}
	}
	if targetMilestone == nil {
		return nil, fmt.Errorf("reward milestone for streak target %d days with item '%s' not found in plan '%s'", streakTarget, itemID, planID)
	}

	// 3. Verify user streak eligibility
	streakState, err := s.streakRepo.GetByUserID(ctx, userID)
	userCurrentStreak := 0
	userLongestStreak := 0
	if err == nil && streakState != nil {
		userCurrentStreak = streakState.CurrentStreak
		userLongestStreak = streakState.LongestStreak
	}

	userMaxStreak := userCurrentStreak
	if userLongestStreak > userMaxStreak {
		userMaxStreak = userLongestStreak
	}

	if userMaxStreak < streakTarget {
		return nil, fmt.Errorf("streak requirement not met: milestone requires %d days streak, current max streak is %d days", streakTarget, userMaxStreak)
	}

	// 4. Check if already claimed
	alreadyClaimed, err := s.rewardRepo.IsRewardClaimed(ctx, userID, planID, streakTarget, itemID)
	if err != nil {
		return nil, fmt.Errorf("failed checking claim status: %w", err)
	}
	if alreadyClaimed {
		return nil, fmt.Errorf("milestone reward for %d days (%s) has already been claimed", streakTarget, item.Name)
	}

	// 5. Claim reward: record claim and add item quantity to inventory
	claim := &models.UserClaimedReward{
		UserID:       userID,
		PlanID:       planID,
		StreakTarget: streakTarget,
		ItemID:       itemID,
	}
	if err := s.rewardRepo.ClaimReward(ctx, claim); err != nil {
		return nil, fmt.Errorf("failed executing reward claim: %w", err)
	}

	newQty, err := s.inventoryRepo.AddItemQuantity(ctx, userID, itemID, targetMilestone.Quantity)
	if err != nil {
		return nil, fmt.Errorf("failed granting item reward to inventory: %w", err)
	}

	return &models.ClaimRewardResult{
		Success:            true,
		StreakTarget:       streakTarget,
		ItemID:             itemID,
		ItemName:           item.Name,
		QuantityAwarded:    targetMilestone.Quantity,
		RemainingInventory: newQty,
		ClaimedAt:          claim.ClaimedAt,
		Message:            fmt.Sprintf("Successfully claimed %dx %s for hitting %d days streak!", targetMilestone.Quantity, item.Name, streakTarget),
	}, nil
}

func (s *rewardService) GetAllPlans(ctx context.Context) ([]models.RewardPlan, error) {
	plans, err := s.rewardRepo.GetAllRewardPlans(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed fetching reward plans: %w", err)
	}
	for i := range plans {
		milestones, _ := s.rewardRepo.GetRewardMilestones(ctx, plans[i].ID)
		plans[i].Milestones = milestones
	}
	return plans, nil
}

func (s *rewardService) CreateRewardPlan(ctx context.Context, req models.CreateRewardPlanRequest) (*models.RewardPlan, error) {
	plan := &models.RewardPlan{
		ID:          req.ID,
		Name:        req.Name,
		Description: req.Description,
		IsActive:    req.IsActive,
	}
	if err := s.rewardRepo.CreateRewardPlan(ctx, plan); err != nil {
		return nil, err
	}
	return plan, nil
}

func (s *rewardService) DeleteRewardPlan(ctx context.Context, planID string) error {
	return s.rewardRepo.DeleteRewardPlan(ctx, planID)
}

func (s *rewardService) UpsertMilestone(ctx context.Context, planID string, req models.UpsertMilestoneRequest) (*models.RewardPlanMilestone, error) {
	if planID == "" {
		return nil, fmt.Errorf("plan_id is required")
	}
	if req.StreakTarget <= 0 {
		return nil, fmt.Errorf("streak_target must be greater than 0")
	}
	if req.ItemID == "" {
		return nil, fmt.Errorf("item_id is required")
	}
	if req.Quantity <= 0 {
		req.Quantity = 1
	}

	// Verify item exists
	item, err := s.itemRepo.GetByID(ctx, req.ItemID)
	if err != nil || item == nil {
		return nil, fmt.Errorf("item definition '%s' not found", req.ItemID)
	}

	milestone := &models.RewardPlanMilestone{
		PlanID:       planID,
		StreakTarget: req.StreakTarget,
		ItemID:       req.ItemID,
		Quantity:     req.Quantity,
		Title:        req.Title,
		Description:  req.Description,
		BadgeSlug:    req.BadgeSlug,
	}

	if err := s.rewardRepo.UpsertMilestone(ctx, milestone); err != nil {
		return nil, err
	}

	return milestone, nil
}

func (s *rewardService) DeleteMilestone(ctx context.Context, milestoneID uuid.UUID) error {
	return s.rewardRepo.DeleteMilestone(ctx, milestoneID)
}
