package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"gymgit/backend/internal/models"

	"github.com/google/uuid"
)

type postgresStreakRepository struct {
	db *sql.DB
}

// NewStreakRepository creates a new PostgreSQL implementation of StreakRepository
func NewStreakRepository(db *sql.DB) StreakRepository {
	return &postgresStreakRepository{db: db}
}

func (r *postgresStreakRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*models.UserStreakState, error) {
	query := `
		SELECT id, user_id, current_streak, longest_streak,
		       TO_CHAR(cycle_start_date, 'YYYY-MM-DD') as cycle_start_date,
		       TO_CHAR(cycle_end_date, 'YYYY-MM-DD') as cycle_end_date,
		       workouts_completed_in_cycle, workouts_target_in_cycle, rest_tokens_total, rest_tokens_used,
		       accuracy_score, TO_CHAR(last_logged_date, 'YYYY-MM-DD') as last_logged_date, is_frozen, updated_at
		FROM user_streak_states
		WHERE user_id = $1
	`
	s := &models.UserStreakState{}
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&s.ID,
		&s.UserID,
		&s.CurrentStreak,
		&s.LongestStreak,
		&s.CycleStartDate,
		&s.CycleEndDate,
		&s.WorkoutsCompletedInCycle,
		&s.WorkoutsTargetInCycle,
		&s.RestTokensTotal,
		&s.RestTokensUsed,
		&s.AccuracyScore,
		&s.LastLoggedDate,
		&s.IsFrozen,
		&s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query user streak state: %w", err)
	}
	return s, nil
}

func (r *postgresStreakRepository) UpsertState(ctx context.Context, state *models.UserStreakState) error {
	query := `
		INSERT INTO user_streak_states (
			user_id, current_streak, longest_streak, cycle_start_date, cycle_end_date,
			workouts_completed_in_cycle, workouts_target_in_cycle, rest_tokens_total, rest_tokens_used,
			accuracy_score, last_logged_date, is_frozen, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW())
		ON CONFLICT (user_id) DO UPDATE SET
			current_streak = EXCLUDED.current_streak,
			longest_streak = EXCLUDED.longest_streak,
			cycle_start_date = EXCLUDED.cycle_start_date,
			cycle_end_date = EXCLUDED.cycle_end_date,
			workouts_completed_in_cycle = EXCLUDED.workouts_completed_in_cycle,
			workouts_target_in_cycle = EXCLUDED.workouts_target_in_cycle,
			rest_tokens_total = EXCLUDED.rest_tokens_total,
			rest_tokens_used = EXCLUDED.rest_tokens_used,
			accuracy_score = EXCLUDED.accuracy_score,
			last_logged_date = EXCLUDED.last_logged_date,
			is_frozen = EXCLUDED.is_frozen,
			updated_at = NOW()
		RETURNING id, updated_at
	`
	err := r.db.QueryRowContext(
		ctx,
		query,
		state.UserID,
		state.CurrentStreak,
		state.LongestStreak,
		state.CycleStartDate,
		state.CycleEndDate,
		state.WorkoutsCompletedInCycle,
		state.WorkoutsTargetInCycle,
		state.RestTokensTotal,
		state.RestTokensUsed,
		state.AccuracyScore,
		state.LastLoggedDate,
		state.IsFrozen,
	).Scan(&state.ID, &state.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to upsert user streak state: %w", err)
	}
	return nil
}
