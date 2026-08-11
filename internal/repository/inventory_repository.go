package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"gymgit/backend/internal/models"

	"github.com/google/uuid"
)

type postgresInventoryRepository struct {
	db *sql.DB
}

// NewInventoryRepository creates a new PostgreSQL implementation of InventoryRepository
func NewInventoryRepository(db *sql.DB) InventoryRepository {
	return &postgresInventoryRepository{db: db}
}

func (r *postgresInventoryRepository) GetInventory(ctx context.Context, userID uuid.UUID) ([]models.UserInventoryItem, error) {
	query := `
		SELECT ui.item_id, i.name, ui.quantity,
		       i.description, i.effect_type, i.duration_seconds, i.rarity, i.icon_slug, i.metadata
		FROM user_inventories ui
		JOIN items i ON ui.item_id = i.id
		WHERE ui.user_id = $1 AND ui.quantity > 0
		ORDER BY i.name ASC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed querying user inventory: %w", err)
	}
	defer rows.Close()

	var result []models.UserInventoryItem
	for rows.Next() {
		var item models.UserInventoryItem
		if err := rows.Scan(
			&item.ItemID,
			&item.Name,
			&item.Quantity,
			&item.Item.Description,
			&item.Item.EffectType,
			&item.Item.DurationSeconds,
			&item.Item.Rarity,
			&item.Item.IconSlug,
			&item.Item.Metadata,
		); err != nil {
			return nil, fmt.Errorf("failed scanning inventory row: %w", err)
		}
		item.Item.ID = item.ItemID
		item.Item.Name = item.Name
		result = append(result, item)
	}
	return result, nil
}

func (r *postgresInventoryRepository) GetItemQuantity(ctx context.Context, userID uuid.UUID, itemID string) (int, error) {
	query := `
		SELECT quantity
		FROM user_inventories
		WHERE user_id = $1 AND item_id = $2
	`
	var qty int
	err := r.db.QueryRowContext(ctx, query, userID, itemID).Scan(&qty)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, fmt.Errorf("failed querying item quantity: %w", err)
	}
	return qty, nil
}

func (r *postgresInventoryRepository) AddItemQuantity(ctx context.Context, userID uuid.UUID, itemID string, delta int) (int, error) {
	if delta <= 0 {
		return 0, fmt.Errorf("delta must be positive for AddItemQuantity")
	}
	query := `
		INSERT INTO user_inventories (user_id, item_id, quantity)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, item_id) DO UPDATE SET
			quantity = user_inventories.quantity + EXCLUDED.quantity,
			updated_at = NOW()
		RETURNING quantity
	`
	var newQty int
	err := r.db.QueryRowContext(ctx, query, userID, itemID, delta).Scan(&newQty)
	if err != nil {
		return 0, fmt.Errorf("failed adding item quantity: %w", err)
	}
	return newQty, nil
}

func (r *postgresInventoryRepository) DeductItemQuantity(ctx context.Context, userID uuid.UUID, itemID string, delta int) (int, error) {
	if delta <= 0 {
		return 0, fmt.Errorf("delta must be positive for DeductItemQuantity")
	}

	currQty, err := r.GetItemQuantity(ctx, userID, itemID)
	if err != nil {
		return 0, err
	}
	if currQty < delta {
		return currQty, fmt.Errorf("insufficient quantity of item %s: available %d, requested %d", itemID, currQty, delta)
	}

	query := `
		UPDATE user_inventories
		SET quantity = quantity - $1, updated_at = NOW()
		WHERE user_id = $2 AND item_id = $3 AND quantity >= $1
		RETURNING quantity
	`
	var newQty int
	err = r.db.QueryRowContext(ctx, query, delta, userID, itemID).Scan(&newQty)
	if err != nil {
		return 0, fmt.Errorf("failed deducting item quantity: %w", err)
	}
	return newQty, nil
}

func (r *postgresInventoryRepository) CreateActiveEffect(ctx context.Context, effect *models.UserActiveEffect) error {
	query := `
		INSERT INTO user_active_effects (user_id, item_id, activated_at, expires_at, is_active, metadata)
		VALUES ($1, $2, $3, $4, $5, COALESCE($6, '{}'::jsonb))
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRowContext(
		ctx,
		query,
		effect.UserID,
		effect.ItemID,
		effect.ActivatedAt,
		effect.ExpiresAt,
		effect.IsActive,
		effect.Metadata,
	).Scan(&effect.ID, &effect.CreatedAt, &effect.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed creating active effect: %w", err)
	}
	return nil
}

func (r *postgresInventoryRepository) GetActiveEffects(ctx context.Context, userID uuid.UUID) ([]models.UserActiveEffect, error) {
	query := `
		SELECT id, user_id, item_id, activated_at, expires_at, is_active, metadata, created_at, updated_at
		FROM user_active_effects
		WHERE user_id = $1 AND is_active = TRUE AND expires_at > NOW()
		ORDER BY expires_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed querying active effects: %w", err)
	}
	defer rows.Close()

	var effects []models.UserActiveEffect
	for rows.Next() {
		var eff models.UserActiveEffect
		if err := rows.Scan(
			&eff.ID,
			&eff.UserID,
			&eff.ItemID,
			&eff.ActivatedAt,
			&eff.ExpiresAt,
			&eff.IsActive,
			&eff.Metadata,
			&eff.CreatedAt,
			&eff.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed scanning active effect: %w", err)
		}
		effects = append(effects, eff)
	}
	return effects, nil
}

func (r *postgresInventoryRepository) GetLatestActiveEffectByItem(ctx context.Context, userID uuid.UUID, itemID string) (*models.UserActiveEffect, error) {
	query := `
		SELECT id, user_id, item_id, activated_at, expires_at, is_active, metadata, created_at, updated_at
		FROM user_active_effects
		WHERE user_id = $1 AND item_id = $2 AND is_active = TRUE AND expires_at > NOW()
		ORDER BY expires_at DESC
		LIMIT 1
	`
	eff := &models.UserActiveEffect{}
	err := r.db.QueryRowContext(ctx, query, userID, itemID).Scan(
		&eff.ID,
		&eff.UserID,
		&eff.ItemID,
		&eff.ActivatedAt,
		&eff.ExpiresAt,
		&eff.IsActive,
		&eff.Metadata,
		&eff.CreatedAt,
		&eff.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed querying latest active effect for item %s: %w", itemID, err)
	}
	return eff, nil
}

func (r *postgresInventoryRepository) DeactivateExpiredEffects(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE user_active_effects
		SET is_active = FALSE, updated_at = NOW()
		WHERE user_id = $1 AND is_active = TRUE AND expires_at <= NOW()
	`
	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed deactivating expired effects: %w", err)
	}
	return nil
}
