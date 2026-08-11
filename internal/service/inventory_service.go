package service

import (
	"context"
	"fmt"

	"gymgit/backend/internal/models"
	"gymgit/backend/internal/repository"
	"gymgit/backend/internal/timezone"
	"time"

	"github.com/google/uuid"
)

type inventoryService struct {
	itemRepo      repository.ItemRepository
	inventoryRepo repository.InventoryRepository
	logRepo       repository.GymLogRepository
	streakRepo    repository.StreakRepository
	userRepo      repository.UserRepository
}

// NewInventoryService creates a new InventoryService instance
func NewInventoryService(
	itemRepo repository.ItemRepository,
	inventoryRepo repository.InventoryRepository,
	logRepo repository.GymLogRepository,
	streakRepo repository.StreakRepository,
	userRepo repository.UserRepository,
) InventoryService {
	return &inventoryService{
		itemRepo:      itemRepo,
		inventoryRepo: inventoryRepo,
		logRepo:       logRepo,
		streakRepo:    streakRepo,
		userRepo:      userRepo,
	}
}

func (s *inventoryService) GetCatalog(ctx context.Context) ([]models.Item, error) {
	items, err := s.itemRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed fetching item catalog: %w", err)
	}
	return items, nil
}

func (s *inventoryService) GetUserInventory(ctx context.Context, userID uuid.UUID) ([]models.UserInventoryItem, []models.UserActiveEffect, error) {
	_ = s.inventoryRepo.DeactivateExpiredEffects(ctx, userID)

	inventory, err := s.inventoryRepo.GetInventory(ctx, userID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed fetching user inventory: %w", err)
	}

	activeEffects, err := s.inventoryRepo.GetActiveEffects(ctx, userID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed fetching active effects: %w", err)
	}

	return inventory, activeEffects, nil
}

func (s *inventoryService) UseItem(ctx context.Context, userID uuid.UUID, itemID string, quantity int, payload map[string]interface{}, loc *time.Location) (*models.UseItemResult, error) {
	if quantity <= 0 {
		quantity = 1
	}

	item, err := s.itemRepo.GetByID(ctx, itemID)
	if err != nil {
		return nil, fmt.Errorf("failed verifying item %s: %w", itemID, err)
	}
	if item == nil {
		return nil, fmt.Errorf("item definition not found for %s", itemID)
	}

	currentQty, err := s.inventoryRepo.GetItemQuantity(ctx, userID, itemID)
	if err != nil {
		return nil, err
	}
	if currentQty < quantity {
		return nil, fmt.Errorf("insufficient quantity of item '%s': available %d, requested %d", item.Name, currentQty, quantity)
	}

	// 1. Instant Use Items
	if item.EffectType == models.EffectTypeInstantUse {
		if itemID == "RESTORE_SHIELD" {
			targetDate, _ := payload["target_date"].(string)
			workoutType, _ := payload["workout_type"].(string)
			hours, _ := payload["hours"].(float64)

			res, err := s.RedeemRestoreShield(ctx, userID, targetDate, workoutType, hours, loc)
			if err != nil {
				return nil, err
			}
			remQty, _ := s.inventoryRepo.GetItemQuantity(ctx, userID, itemID)
			return &models.UseItemResult{
				ItemID:             itemID,
				QuantityConsumed:   1,
				RemainingQuantity:  remQty,
				EffectType:         item.EffectType,
				Details:            res.Message,
				RestoredStreakDate: &res.RestoredDate,
			}, nil
		}

		// Generic Instant Use item deduction
		remainingQty, err := s.inventoryRepo.DeductItemQuantity(ctx, userID, itemID, quantity)
		if err != nil {
			return nil, err
		}
		return &models.UseItemResult{
			ItemID:            itemID,
			QuantityConsumed:  quantity,
			RemainingQuantity: remainingQty,
			EffectType:        item.EffectType,
			Details:           fmt.Sprintf("Successfully used %d x %s", quantity, item.Name),
		}, nil
	}

	// 2. Time Based Items
	remainingQty, err := s.inventoryRepo.DeductItemQuantity(ctx, userID, itemID, quantity)
	if err != nil {
		return nil, err
	}

	totalDuration := time.Duration(item.DurationSeconds*quantity) * time.Second

	// Check if user already has an active effect for this item
	latestEffect, _ := s.inventoryRepo.GetLatestActiveEffectByItem(ctx, userID, itemID)
	now := time.Now().UTC()
	activatedAt := now
	expiresAt := now.Add(totalDuration)

	if latestEffect != nil && latestEffect.ExpiresAt.After(now) {
		activatedAt = latestEffect.ActivatedAt
		expiresAt = latestEffect.ExpiresAt.Add(totalDuration)
	}

	activeEffect := &models.UserActiveEffect{
		UserID:      userID,
		ItemID:      itemID,
		ActivatedAt: activatedAt,
		ExpiresAt:   expiresAt,
		IsActive:    true,
	}

	if err := s.inventoryRepo.CreateActiveEffect(ctx, activeEffect); err != nil {
		return nil, fmt.Errorf("failed activating item effect: %w", err)
	}

	if itemID == "STREAK_FREEZE_TOKEN" {
		state, errState := s.streakRepo.GetByUserID(ctx, userID)
		if errState == nil && state != nil {
			state.IsFrozen = true
			_ = s.streakRepo.UpsertState(ctx, state)
		} else if state == nil {
			_ = s.streakRepo.UpsertState(ctx, &models.UserStreakState{
				UserID:         userID,
				CycleStartDate: now.Format("2006-01-02"),
				CycleEndDate:   now.AddDate(0, 0, 6).Format("2006-01-02"),
				IsFrozen:       true,
			})
		}
	}

	return &models.UseItemResult{
		ItemID:            itemID,
		QuantityConsumed:  quantity,
		RemainingQuantity: remainingQty,
		EffectType:        item.EffectType,
		ActiveUntil:       &expiresAt,
		Details:           fmt.Sprintf("Activated %s effect until %s", item.Name, expiresAt.Format(time.RFC3339)),
	}, nil
}


func (s *inventoryService) RedeemRestoreShield(ctx context.Context, userID uuid.UUID, targetDate string, workoutType string, hours float64, loc *time.Location) (*models.RestoreShieldResult, error) {
	if targetDate == "" {
		return nil, fmt.Errorf("target_date is required for Restore Shield redemption")
	}

	if loc == nil {
		loc = time.UTC
	}
	userToday := timezone.GetUserToday(loc)
	todayStr := userToday.Format("2006-01-02")

	// Lookback Window Check: targetDate must be past date within 3 days (Yesterday, 2 days ago, 3 days ago)
	if targetDate >= todayStr {
		return nil, fmt.Errorf("Restore Shield can only be redeemed on past missed days, not today or future dates")
	}

	minAllowedDate := userToday.AddDate(0, 0, -3).Format("2006-01-02")
	if targetDate < minAllowedDate {
		return nil, fmt.Errorf("Restore Shield lookback window expired: target date %s is older than the allowed 3-day window (%s)", targetDate, minAllowedDate)
	}

	// Check Restore Shield inventory balance
	qty, err := s.inventoryRepo.GetItemQuantity(ctx, userID, "RESTORE_SHIELD")
	if err != nil {
		return nil, fmt.Errorf("failed verifying Restore Shield balance: %w", err)
	}
	if qty < 1 {
		return nil, fmt.Errorf("no Restore Shield items available in inventory")
	}

	if workoutType == "" {
		workoutType = "Restored Session"
	}
	if hours <= 0 {
		hours = 1.0
	}

	// Create/upsert historical workout log for targetDate
	log := &models.GymLog{
		UserID:      userID,
		Date:        targetDate,
		Hours:       hours,
		WorkoutType: workoutType,
		IsRestored:  true,
	}
	if err := s.logRepo.UpsertLog(ctx, log); err != nil {
		return nil, fmt.Errorf("failed saving restored workout log: %w", err)
	}

	// Deduct 1 RESTORE_SHIELD from inventory
	remQty, err := s.inventoryRepo.DeductItemQuantity(ctx, userID, "RESTORE_SHIELD", 1)
	if err != nil {
		return nil, fmt.Errorf("failed deducting Restore Shield from inventory: %w", err)
	}

	// Recalculate streak
	allLogs, errLogs := s.logRepo.GetLogs(ctx, userID, nil, nil, nil)
	newStreak := 0
	if errLogs == nil {
		streakStats := CalculateScientificStreak(allLogs, 4, 30, userToday)
		newStreak = streakStats.CurrentStreak

		state, _ := s.streakRepo.GetByUserID(ctx, userID)
		if state != nil {
			state.CurrentStreak = newStreak
			if newStreak > state.LongestStreak {
				state.LongestStreak = newStreak
			}
			_ = s.streakRepo.UpsertState(ctx, state)
		}
	}

	return &models.RestoreShieldResult{
		Success:          true,
		RestoredDate:     targetDate,
		NewCurrentStreak: newStreak,
		ShieldsRemaining: remQty,
		Message:          fmt.Sprintf("Restore Shield redeemed successfully for %s! Active streak updated to %d days.", targetDate, newStreak),
	}, nil
}

func (s *inventoryService) CheckAndGrantMilestones(ctx context.Context, userID uuid.UUID, streakDays int) ([]models.MilestoneReward, error) {
	milestonesMap := map[int][]struct {
		ItemID string
		Qty    int
	}{
		7:   {{"RESTORE_SHIELD", 1}},
		14:  {{"STREAK_FREEZE_TOKEN", 1}},
		21:  {{"RESTORE_SHIELD", 1}},
		30:  {{"RESTORE_SHIELD", 1}, {"XP_BOOST", 1}},
		60:  {{"STREAK_FREEZE_TOKEN", 2}, {"RESTORE_SHIELD", 2}},
		90:  {{"RESTORE_SHIELD", 2}, {"XP_BOOST", 2}},
		180: {{"STREAK_FREEZE_TOKEN", 3}, {"RESTORE_SHIELD", 3}},
		365: {{"STREAK_FREEZE_TOKEN", 5}, {"RESTORE_SHIELD", 5}, {"XP_BOOST", 3}},
	}

	rewardsToGrant, exists := milestonesMap[streakDays]
	if !exists {
		return nil, nil
	}

	var grantedRewards []models.MilestoneReward
	for _, req := range rewardsToGrant {
		_, err := s.inventoryRepo.AddItemQuantity(ctx, userID, req.ItemID, req.Qty)
		if err == nil {
			item, _ := s.itemRepo.GetByID(ctx, req.ItemID)
			itemName := req.ItemID
			if item != nil {
				itemName = item.Name
			}
			grantedRewards = append(grantedRewards, models.MilestoneReward{
				StreakDays: streakDays,
				ItemID:     req.ItemID,
				Quantity:   req.Qty,
				ItemName:   itemName,
			})
		}
	}

	return grantedRewards, nil
}
