package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"gymgit/backend/internal/models"

	"github.com/google/uuid"
)

type postgresGymLogRepository struct {
	db *sql.DB
}

// NewGymLogRepository creates a new PostgreSQL implementation of GymLogRepository
func NewGymLogRepository(db *sql.DB) GymLogRepository {
	return &postgresGymLogRepository{db: db}
}

func (r *postgresGymLogRepository) GetLogs(ctx context.Context, userID uuid.UUID, startDate, endDate *time.Time, workoutType *string) ([]models.GymLog, error) {
	queryBuilder := strings.Builder{}
	queryBuilder.WriteString(`
		SELECT id, user_id, TO_CHAR(date, 'YYYY-MM-DD') as date, hours, workout_type, notes, created_at, updated_at
		FROM gym_logs
		WHERE user_id = $1
	`)

	args := []interface{}{userID}
	paramIdx := 2

	if startDate != nil {
		queryBuilder.WriteString(fmt.Sprintf(" AND date >= $%d", paramIdx))
		args = append(args, startDate.Format("2006-01-02"))
		paramIdx++
	}

	if endDate != nil {
		queryBuilder.WriteString(fmt.Sprintf(" AND date <= $%d", paramIdx))
		args = append(args, endDate.Format("2006-01-02"))
		paramIdx++
	}

	if workoutType != nil && *workoutType != "" {
		queryBuilder.WriteString(fmt.Sprintf(" AND LOWER(workout_type) = LOWER($%d)", paramIdx))
		args = append(args, *workoutType)
		paramIdx++
	}

	queryBuilder.WriteString(" ORDER BY date DESC")

	rows, err := r.db.QueryContext(ctx, queryBuilder.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query gym logs: %w", err)
	}
	defer rows.Close()

	logs := []models.GymLog{}
	for rows.Next() {
		var l models.GymLog
		if err := rows.Scan(
			&l.ID,
			&l.UserID,
			&l.Date,
			&l.Hours,
			&l.WorkoutType,
			&l.Notes,
			&l.CreatedAt,
			&l.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed scanning gym log row: %w", err)
		}
		logs = append(logs, l)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return logs, nil
}

func (r *postgresGymLogRepository) GetByDate(ctx context.Context, userID uuid.UUID, date string) (*models.GymLog, error) {
	query := `
		SELECT id, user_id, TO_CHAR(date, 'YYYY-MM-DD') as date, hours, workout_type, notes, created_at, updated_at
		FROM gym_logs
		WHERE user_id = $1 AND date = $2
	`
	l := &models.GymLog{}
	err := r.db.QueryRowContext(ctx, query, userID, date).Scan(
		&l.ID,
		&l.UserID,
		&l.Date,
		&l.Hours,
		&l.WorkoutType,
		&l.Notes,
		&l.CreatedAt,
		&l.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed querying gym log by date: %w", err)
	}
	return l, nil
}

func (r *postgresGymLogRepository) UpsertLog(ctx context.Context, log *models.GymLog) error {
	if log.ID == uuid.Nil {
		log.ID = uuid.New()
	}

	query := `
		INSERT INTO gym_logs (id, user_id, date, hours, workout_type, notes)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (user_id, date) DO UPDATE SET
			hours = EXCLUDED.hours,
			workout_type = EXCLUDED.workout_type,
			notes = EXCLUDED.notes,
			updated_at = NOW()
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		log.ID,
		log.UserID,
		log.Date,
		log.Hours,
		log.WorkoutType,
		log.Notes,
	).Scan(&log.ID, &log.CreatedAt, &log.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed upserting gym log: %w", err)
	}
	return nil
}

func (r *postgresGymLogRepository) DeleteByDate(ctx context.Context, userID uuid.UUID, date string) error {
	query := `DELETE FROM gym_logs WHERE user_id = $1 AND date = $2`
	_, err := r.db.ExecContext(ctx, query, userID, date)
	if err != nil {
		return fmt.Errorf("failed deleting gym log for date %s: %w", date, err)
	}
	return nil
}

func (r *postgresGymLogRepository) ResetDemoLogs(ctx context.Context, userID uuid.UUID, logs []models.GymLog) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed starting reset transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. Clear existing logs for user
	deleteQuery := `DELETE FROM gym_logs WHERE user_id = $1`
	if _, err := tx.ExecContext(ctx, deleteQuery, userID); err != nil {
		return fmt.Errorf("failed clearing old user logs: %w", err)
	}

	// 2. Bulk insert demo logs
	insertQuery := `
		INSERT INTO gym_logs (id, user_id, date, hours, workout_type, notes)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	stmt, err := tx.PrepareContext(ctx, insertQuery)
	if err != nil {
		return fmt.Errorf("failed preparing bulk insert statement: %w", err)
	}
	defer stmt.Close()

	for _, l := range logs {
		logID := l.ID
		if logID == uuid.Nil {
			logID = uuid.New()
		}
		if _, err := stmt.ExecContext(ctx, logID, userID, l.Date, l.Hours, l.WorkoutType, l.Notes); err != nil {
			return fmt.Errorf("failed inserting demo log for date %s: %w", l.Date, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed committing demo logs transaction: %w", err)
	}
	return nil
}
