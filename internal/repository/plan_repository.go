package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"gymgit/backend/internal/models"
)

type postgresPlanRepository struct {
	db *sql.DB
}

// NewPlanRepository creates a new PostgreSQL implementation of PlanRepository
func NewPlanRepository(db *sql.DB) PlanRepository {
	return &postgresPlanRepository{db: db}
}

func (r *postgresPlanRepository) GetAll(ctx context.Context) ([]models.WeeklyPlan, error) {
	query := `
		SELECT id, name, description, categories, user_id, created_at, updated_at
		FROM weekly_plans
		WHERE user_id IS NULL
		ORDER BY id ASC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query weekly plans: %w", err)
	}
	defer rows.Close()

	var plans []models.WeeklyPlan
	for rows.Next() {
		var p models.WeeklyPlan
		var categoriesRaw []byte
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &categoriesRaw, &p.UserID, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan weekly plan: %w", err)
		}
		if len(categoriesRaw) > 0 {
			if err := json.Unmarshal(categoriesRaw, &p.Categories); err != nil {
				p.Categories = []string{}
			}
		} else {
			p.Categories = []string{}
		}
		plans = append(plans, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return plans, nil
}

func (r *postgresPlanRepository) GetByID(ctx context.Context, id string) (*models.WeeklyPlan, error) {
	query := `
		SELECT id, name, description, categories, user_id, created_at, updated_at
		FROM weekly_plans
		WHERE id = $1
	`
	p := &models.WeeklyPlan{}
	var categoriesRaw []byte
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID,
		&p.Name,
		&p.Description,
		&categoriesRaw,
		&p.UserID,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query weekly plan by id: %w", err)
	}
	if len(categoriesRaw) > 0 {
		if err := json.Unmarshal(categoriesRaw, &p.Categories); err != nil {
			p.Categories = []string{}
		}
	} else {
		p.Categories = []string{}
	}
	return p, nil
}

func (r *postgresPlanRepository) Create(ctx context.Context, plan *models.WeeklyPlan) error {
	query := `
		INSERT INTO weekly_plans (id, name, description, categories, user_id)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			categories = EXCLUDED.categories,
			updated_at = NOW()
	`
	categoriesRaw, err := json.Marshal(plan.Categories)
	if err != nil {
		return fmt.Errorf("failed to marshal categories: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query, plan.ID, plan.Name, plan.Description, categoriesRaw, plan.UserID)
	if err != nil {
		return fmt.Errorf("failed to save weekly plan: %w", err)
	}
	return nil
}
