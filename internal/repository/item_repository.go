package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"gymgit/backend/internal/models"
)

type postgresItemRepository struct {
	db *sql.DB
}

// NewItemRepository creates a new PostgreSQL implementation of ItemRepository
func NewItemRepository(db *sql.DB) ItemRepository {
	return &postgresItemRepository{db: db}
}

func (r *postgresItemRepository) GetAll(ctx context.Context) ([]models.Item, error) {
	query := `
		SELECT id, name, description, effect_type, duration_seconds, rarity, icon_slug, metadata, created_at, updated_at
		FROM items
		ORDER BY name ASC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed querying items: %w", err)
	}
	defer rows.Close()

	var items []models.Item
	for rows.Next() {
		var item models.Item
		if err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Description,
			&item.EffectType,
			&item.DurationSeconds,
			&item.Rarity,
			&item.IconSlug,
			&item.Metadata,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed scanning item row: %w", err)
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *postgresItemRepository) GetByID(ctx context.Context, id string) (*models.Item, error) {
	query := `
		SELECT id, name, description, effect_type, duration_seconds, rarity, icon_slug, metadata, created_at, updated_at
		FROM items
		WHERE id = $1
	`
	item := &models.Item{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&item.ID,
		&item.Name,
		&item.Description,
		&item.EffectType,
		&item.DurationSeconds,
		&item.Rarity,
		&item.IconSlug,
		&item.Metadata,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed querying item %s: %w", id, err)
	}
	return item, nil
}

func (r *postgresItemRepository) Create(ctx context.Context, item *models.Item) error {
	query := `
		INSERT INTO items (id, name, description, effect_type, duration_seconds, rarity, icon_slug, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, COALESCE($8, '{}'::jsonb))
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			effect_type = EXCLUDED.effect_type,
			duration_seconds = EXCLUDED.duration_seconds,
			rarity = EXCLUDED.rarity,
			icon_slug = EXCLUDED.icon_slug,
			metadata = EXCLUDED.metadata,
			updated_at = NOW()
		RETURNING created_at, updated_at
	`
	err := r.db.QueryRowContext(
		ctx,
		query,
		item.ID,
		item.Name,
		item.Description,
		item.EffectType,
		item.DurationSeconds,
		item.Rarity,
		item.IconSlug,
		item.Metadata,
	).Scan(&item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed creating item definition: %w", err)
	}
	return nil
}
