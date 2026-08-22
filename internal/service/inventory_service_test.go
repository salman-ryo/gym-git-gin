package service_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"gymgit/backend/internal/models"
	"gymgit/backend/internal/service"
	"gymgit/backend/internal/timezone"

	"github.com/google/uuid"
)

type mockItemRepo struct {
	items map[string]*models.Item
}

func (m *mockItemRepo) GetAll(ctx context.Context) ([]models.Item, error) {
	var list []models.Item
	for _, it := range m.items {
		list = append(list, *it)
	}
	return list, nil
}
func (m *mockItemRepo) GetByID(ctx context.Context, id string) (*models.Item, error) {
	if it, ok := m.items[id]; ok {
		return it, nil
	}
	return nil, nil
}
func (m *mockItemRepo) Create(ctx context.Context, item *models.Item) error {
	m.items[item.ID] = item
	return nil
}

type mockInventoryRepo struct {
	quantities map[string]int
	effects    []models.UserActiveEffect
}

func (m *mockInventoryRepo) GetInventory(ctx context.Context, userID uuid.UUID) ([]models.UserInventoryItem, error) {
	var list []models.UserInventoryItem
	for itemID, qty := range m.quantities {
		if qty > 0 {
			list = append(list, models.UserInventoryItem{
				ItemID:   itemID,
				Name:     itemID,
				Quantity: qty,
			})
		}
	}
	return list, nil
}
func (m *mockInventoryRepo) GetItemQuantity(ctx context.Context, userID uuid.UUID, itemID string) (int, error) {
	return m.quantities[itemID], nil
}
func (m *mockInventoryRepo) AddItemQuantity(ctx context.Context, userID uuid.UUID, itemID string, delta int) (int, error) {
	m.quantities[itemID] += delta
	if m.quantities[itemID] > 9 {
		m.quantities[itemID] = 9
	}
	return m.quantities[itemID], nil
}
func (m *mockInventoryRepo) DeductItemQuantity(ctx context.Context, userID uuid.UUID, itemID string, delta int) (int, error) {
	curr := m.quantities[itemID]
	if curr < delta {
		return curr, fmt.Errorf("insufficient quantity")
	}
	m.quantities[itemID] -= delta
	return m.quantities[itemID], nil
}
func (m *mockInventoryRepo) CreateActiveEffect(ctx context.Context, effect *models.UserActiveEffect) error {
	m.effects = append(m.effects, *effect)
	return nil
}
func (m *mockInventoryRepo) GetActiveEffects(ctx context.Context, userID uuid.UUID) ([]models.UserActiveEffect, error) {
	return m.effects, nil
}
func (m *mockInventoryRepo) GetLatestActiveEffectByItem(ctx context.Context, userID uuid.UUID, itemID string) (*models.UserActiveEffect, error) {
	for i := len(m.effects) - 1; i >= 0; i-- {
		if m.effects[i].ItemID == itemID && m.effects[i].ExpiresAt.After(time.Now()) {
			return &m.effects[i], nil
		}
	}
	return nil, nil
}
func (m *mockInventoryRepo) DeactivateExpiredEffects(ctx context.Context, userID uuid.UUID) error {
	return nil
}


func TestInventoryService_UseTimeBasedItem(t *testing.T) {
	userID := uuid.New()
	itemRepo := &mockItemRepo{
		items: map[string]*models.Item{
			"STREAK_FREEZE_TOKEN": {
				ID:              "STREAK_FREEZE_TOKEN",
				Name:            "Streak Freeze Token",
				EffectType:      models.EffectTypeTimeBased,
				DurationSeconds: 86400,
			},
		},
	}
	invRepo := &mockInventoryRepo{
		quantities: map[string]int{"STREAK_FREEZE_TOKEN": 3},
	}
	logRepo := &mockLogRepo{}
	streakRepo := &mockStreakRepo{}
	userRepo := &mockUserRepo{}

	invService := service.NewInventoryService(itemRepo, invRepo, logRepo, streakRepo, userRepo)
	loc := timezone.LoadLocation("UTC")

	result, err := invService.UseItem(context.Background(), userID, "STREAK_FREEZE_TOKEN", 1, nil, loc)
	if err != nil {
		t.Fatalf("unexpected error using item: %v", err)
	}

	if result.RemainingQuantity != 2 {
		t.Errorf("expected 2 remaining tokens, got %d", result.RemainingQuantity)
	}
	if result.ActiveUntil == nil {
		t.Fatalf("expected active_until to be non-nil for time-based item")
	}
}

func TestInventoryService_RedeemRestoreShield_MultiDayValidation(t *testing.T) {
	userID := uuid.New()
	itemRepo := &mockItemRepo{
		items: map[string]*models.Item{
			"RESTORE_SHIELD": {
				ID:         "RESTORE_SHIELD",
				Name:       "Restore Shield",
				EffectType: models.EffectTypeInstantUse,
			},
		},
	}
	invRepo := &mockInventoryRepo{
		quantities: map[string]int{"RESTORE_SHIELD": 2},
	}
	logRepo := &mockLogRepo{}
	streakRepo := &mockStreakRepo{}
	userRepo := &mockUserRepo{}

	invService := service.NewInventoryService(itemRepo, invRepo, logRepo, streakRepo, userRepo)
	loc := timezone.LoadLocation("UTC")
	userToday := timezone.GetUserToday(loc)

	// 1. Test future/today rejection
	todayStr := userToday.Format("2006-01-02")
	_, errToday := invService.RedeemRestoreShield(context.Background(), userID, []string{todayStr}, "Push", 1.0, loc)
	if errToday == nil {
		t.Errorf("expected error for today's date (%s), but got none", todayStr)
	}

	// 2. Test insufficient shields: 3 days missed, only 2 available
	dates3 := []string{
		userToday.AddDate(0, 0, -3).Format("2006-01-02"),
		userToday.AddDate(0, 0, -2).Format("2006-01-02"),
		userToday.AddDate(0, 0, -1).Format("2006-01-02"),
	}
	_, errInsuff := invService.RedeemRestoreShield(context.Background(), userID, dates3, "Push", 1.0, loc)
	if errInsuff == nil {
		t.Errorf("expected error for insufficient shields (3 needed, 2 available), got none")
	}

	// 3. Test successful 2-day restore with 2 shields
	dates2 := []string{
		userToday.AddDate(0, 0, -2).Format("2006-01-02"),
		userToday.AddDate(0, 0, -1).Format("2006-01-02"),
	}
	res, errValid := invService.RedeemRestoreShield(context.Background(), userID, dates2, "Push", 1.0, loc)
	if errValid != nil {
		t.Fatalf("unexpected error redeeming restore shields: %v", errValid)
	}

	if !res.Success {
		t.Errorf("expected successful redemption result")
	}
	if res.ShieldsConsumed != 2 {
		t.Errorf("expected 2 shields consumed, got %d", res.ShieldsConsumed)
	}
	if res.ShieldsRemaining != 0 {
		t.Errorf("expected 0 remaining shields, got %d", res.ShieldsRemaining)
	}
}
