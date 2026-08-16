package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"gymgit/backend/internal/models"
	"time"

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
		SELECT id, auth_user_id, email, name, avatar_url, provider, timezone, weekly_plan_id, queued_weekly_plan_id, checkin_snoozed_date, checkin_snoozed_at, created_at, updated_at
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
		&u.Timezone,
		&u.WeeklyPlanID,
		&u.QueuedWeeklyPlanID,
		&u.CheckinSnoozedDate,
		&u.CheckinSnoozedAt,
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
		SELECT id, auth_user_id, email, name, avatar_url, provider, timezone, weekly_plan_id, queued_weekly_plan_id, checkin_snoozed_date, checkin_snoozed_at, created_at, updated_at
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
		&u.Timezone,
		&u.WeeklyPlanID,
		&u.QueuedWeeklyPlanID,
		&u.CheckinSnoozedDate,
		&u.CheckinSnoozedAt,
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
	tz := user.Timezone
	if tz == "" {
		tz = "UTC"
	}
	query := `
		INSERT INTO users (auth_user_id, email, name, avatar_url, provider, timezone, weekly_plan_id, queued_weekly_plan_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (auth_user_id) DO UPDATE SET
			email = EXCLUDED.email,
			name = COALESCE(EXCLUDED.name, users.name),
			avatar_url = COALESCE(EXCLUDED.avatar_url, users.avatar_url),
			provider = COALESCE(EXCLUDED.provider, users.provider),
			updated_at = NOW()
		RETURNING id, auth_user_id, email, name, avatar_url, provider, timezone, weekly_plan_id, queued_weekly_plan_id, checkin_snoozed_date, checkin_snoozed_at, created_at, updated_at
	`
	err := r.db.QueryRowContext(
		ctx,
		query,
		user.AuthUserID,
		user.Email,
		user.Name,
		user.AvatarURL,
		user.Provider,
		tz,
		user.WeeklyPlanID,
		user.QueuedWeeklyPlanID,
	).Scan(
		&user.ID,
		&user.AuthUserID,
		&user.Email,
		&user.Name,
		&user.AvatarURL,
		&user.Provider,
		&user.Timezone,
		&user.WeeklyPlanID,
		&user.QueuedWeeklyPlanID,
		&user.CheckinSnoozedDate,
		&user.CheckinSnoozedAt,
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

func (r *postgresUserRepository) SetQueuedPlan(ctx context.Context, userID uuid.UUID, planID *string) error {
	query := `
		UPDATE users
		SET queued_weekly_plan_id = $1, updated_at = NOW()
		WHERE id = $2
	`
	result, err := r.db.ExecContext(ctx, query, planID, userID)
	if err != nil {
		return fmt.Errorf("failed to update user queued weekly plan: %w", err)
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

func (r *postgresUserRepository) UpdateTimezone(ctx context.Context, userID uuid.UUID, tz string) error {
	query := `
		UPDATE users
		SET timezone = $1, updated_at = NOW()
		WHERE id = $2
	`
	result, err := r.db.ExecContext(ctx, query, tz, userID)
	if err != nil {
		return fmt.Errorf("failed to update user timezone: %w", err)
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
		SELECT id, auth_user_id, email, name, avatar_url, provider, timezone, weekly_plan_id, queued_weekly_plan_id, checkin_snoozed_date, checkin_snoozed_at, created_at, updated_at
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
		&u.Timezone,
		&u.WeeklyPlanID,
		&u.QueuedWeeklyPlanID,
		&u.CheckinSnoozedDate,
		&u.CheckinSnoozedAt,
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

func (r *postgresUserRepository) SetCheckinSnooze(ctx context.Context, userID uuid.UUID, dateStr string, snoozedAt time.Time) error {
	query := `
		UPDATE users
		SET checkin_snoozed_date = $1, checkin_snoozed_at = $2, updated_at = NOW()
		WHERE id = $3
	`
	result, err := r.db.ExecContext(ctx, query, dateStr, snoozedAt, userID)
	if err != nil {
		return fmt.Errorf("failed to update user checkin snooze: %w", err)
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

func (r *postgresUserRepository) ClearCheckinSnooze(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE users
		SET checkin_snoozed_date = NULL, checkin_snoozed_at = NULL, updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to clear user checkin snooze: %w", err)
	}
	return nil
}

