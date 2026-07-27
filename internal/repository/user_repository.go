package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"gymgit/backend/internal/models"

	"github.com/google/uuid"
)

type postgresUserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new PostgreSQL implementation of UserRepository
func NewUserRepository(db *sql.DB) UserRepository {
	return &postgresUserRepository{db: db}
}

func (r *postgresUserRepository) GetByAuthUserID(ctx context.Context, authUserID uuid.UUID) (*models.User, error) {
	query := `
		SELECT id, auth_user_id, email, name, avatar_url, provider, weekly_plan_id, created_at, updated_at
		FROM users
		WHERE auth_user_id = $1
	`
	u := &models.User{}
	err := r.db.QueryRowContext(ctx, query, authUserID).Scan(
		&u.ID,
		&u.AuthUserID,
		&u.Email,
		&u.Name,
		&u.AvatarURL,
		&u.Provider,
		&u.WeeklyPlanID,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query user by auth_user_id: %w", err)
	}
	return u, nil
}

func (r *postgresUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := `
		SELECT id, auth_user_id, email, name, avatar_url, provider, weekly_plan_id, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	u := &models.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&u.ID,
		&u.AuthUserID,
		&u.Email,
		&u.Name,
		&u.AvatarURL,
		&u.Provider,
		&u.WeeklyPlanID,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query user by id: %w", err)
	}
	return u, nil
}

func (r *postgresUserRepository) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (auth_user_id, email, name, avatar_url, provider, weekly_plan_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (auth_user_id) DO UPDATE SET
			email = EXCLUDED.email,
			name = COALESCE(EXCLUDED.name, users.name),
			avatar_url = COALESCE(EXCLUDED.avatar_url, users.avatar_url),
			provider = COALESCE(EXCLUDED.provider, users.provider),
			updated_at = NOW()
		RETURNING id, auth_user_id, email, name, avatar_url, provider, weekly_plan_id, created_at, updated_at
	`
	err := r.db.QueryRowContext(
		ctx,
		query,
		user.AuthUserID,
		user.Email,
		user.Name,
		user.AvatarURL,
		user.Provider,
		user.WeeklyPlanID,
	).Scan(
		&user.ID,
		&user.AuthUserID,
		&user.Email,
		&user.Name,
		&user.AvatarURL,
		&user.Provider,
		&user.WeeklyPlanID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert or update user profile: %w", err)
	}
	return nil
}

func (r *postgresUserRepository) UpdateWeeklyPlan(ctx context.Context, userID uuid.UUID, planID string) error {
	query := `
		UPDATE users
		SET weekly_plan_id = $1, updated_at = NOW()
		WHERE id = $2
	`
	result, err := r.db.ExecContext(ctx, query, planID, userID)
	if err != nil {
		return fmt.Errorf("failed to update user weekly plan: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("user not found for id %s", userID)
	}
	return nil
}

func (r *postgresUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, auth_user_id, email, name, avatar_url, provider, weekly_plan_id, created_at, updated_at
		FROM users
		WHERE email = $1
	`
	u := &models.User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&u.ID,
		&u.AuthUserID,
		&u.Email,
		&u.Name,
		&u.AvatarURL,
		&u.Provider,
		&u.WeeklyPlanID,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query user by email: %w", err)
	}
	return u, nil
}

func (r *postgresUserRepository) UpdateAuthUserID(ctx context.Context, id uuid.UUID, authUserID uuid.UUID) error {
	query := `
		UPDATE users
		SET auth_user_id = $1, updated_at = NOW()
		WHERE id = $2
	`
	result, err := r.db.ExecContext(ctx, query, authUserID, id)
	if err != nil {
		return fmt.Errorf("failed to update user auth_user_id: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("user not found for id %s", id)
	}
	return nil
}
