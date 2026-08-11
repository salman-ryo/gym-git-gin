package repository

import (
	"context"
	"database/sql"
	"fmt"

	"gymgit/backend/internal/models"

	"github.com/google/uuid"
)

type rewardRepository struct {
	db *sql.DB
}

// NewRewardRepository creates a new RewardRepository implementation
func NewRewardRepository(db *sql.DB) RewardRepository {
	return &rewardRepository{db: db}
}

func (r *rewardRepository) GetActiveRewardPlan(ctx context.Context, planID string) (*models.RewardPlan, error) {
	if r.db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	query := `SELECT id, name, description, is_active, created_at, updated_at FROM reward_plans WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, planID)

	var plan models.RewardPlan
	var desc sql.NullString
	if err := row.Scan(&plan.ID, &plan.Name, &desc, &plan.IsActive, &plan.CreatedAt, &plan.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed fetching reward plan %s: %w", planID, err)
	}
	if desc.Valid {
		plan.Description = desc.String
	}

	milestones, err := r.GetRewardMilestones(ctx, planID)
	if err != nil {
		return nil, err
	}
	plan.Milestones = milestones

	return &plan, nil
}

func (r *rewardRepository) GetAllRewardPlans(ctx context.Context) ([]models.RewardPlan, error) {
	if r.db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	query := `SELECT id, name, description, is_active, created_at, updated_at FROM reward_plans ORDER BY id ASC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed querying reward plans: %w", err)
	}
	defer rows.Close()

	var plans []models.RewardPlan
	for rows.Next() {
		var plan models.RewardPlan
		var desc sql.NullString
		if err := rows.Scan(&plan.ID, &plan.Name, &desc, &plan.IsActive, &plan.CreatedAt, &plan.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed scanning reward plan row: %w", err)
		}
		if desc.Valid {
			plan.Description = desc.String
		}
		plans = append(plans, plan)
	}

	return plans, nil
}

func (r *rewardRepository) CreateRewardPlan(ctx context.Context, plan *models.RewardPlan) error {
	if r.db == nil {
		return fmt.Errorf("database connection is nil")
	}

	query := `
		INSERT INTO reward_plans (id, name, description, is_active, updated_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			is_active = EXCLUDED.is_active,
			updated_at = NOW()
	`
	_, err := r.db.ExecContext(ctx, query, plan.ID, plan.Name, plan.Description, plan.IsActive)
	if err != nil {
		return fmt.Errorf("failed creating/updating reward plan %s: %w", plan.ID, err)
	}
	return nil
}

func (r *rewardRepository) DeleteRewardPlan(ctx context.Context, planID string) error {
	if r.db == nil {
		return fmt.Errorf("database connection is nil")
	}

	query := `DELETE FROM reward_plans WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, planID)
	if err != nil {
		return fmt.Errorf("failed deleting reward plan %s: %w", planID, err)
	}
	return nil
}

func (r *rewardRepository) GetRewardMilestones(ctx context.Context, planID string) ([]models.RewardPlanMilestone, error) {
	if r.db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	query := `
		SELECT id, plan_id, streak_target, item_id, quantity, title, description, badge_slug, created_at, updated_at
		FROM reward_plan_milestones
		WHERE plan_id = $1
		ORDER BY streak_target ASC, item_id ASC
	`
	rows, err := r.db.QueryContext(ctx, query, planID)
	if err != nil {
		return nil, fmt.Errorf("failed querying milestones for plan %s: %w", planID, err)
	}
	defer rows.Close()

	var milestones []models.RewardPlanMilestone
	for rows.Next() {
		var m models.RewardPlanMilestone
		var desc, badge sql.NullString
		if err := rows.Scan(&m.ID, &m.PlanID, &m.StreakTarget, &m.ItemID, &m.Quantity, &m.Title, &desc, &badge, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed scanning milestone row: %w", err)
		}
		if desc.Valid {
			m.Description = desc.String
		}
		if badge.Valid {
			m.BadgeSlug = badge.String
		}
		milestones = append(milestones, m)
	}

	return milestones, nil
}

func (r *rewardRepository) UpsertMilestone(ctx context.Context, milestone *models.RewardPlanMilestone) error {
	if r.db == nil {
		return fmt.Errorf("database connection is nil")
	}

	query := `
		INSERT INTO reward_plan_milestones (plan_id, streak_target, item_id, quantity, title, description, badge_slug, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		ON CONFLICT (plan_id, streak_target, item_id) DO UPDATE SET
			quantity = EXCLUDED.quantity,
			title = EXCLUDED.title,
			description = EXCLUDED.description,
			badge_slug = EXCLUDED.badge_slug,
			updated_at = NOW()
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRowContext(
		ctx, query,
		milestone.PlanID, milestone.StreakTarget, milestone.ItemID,
		milestone.Quantity, milestone.Title, milestone.Description, milestone.BadgeSlug,
	).Scan(&milestone.ID, &milestone.CreatedAt, &milestone.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed upserting milestone for plan %s (streak target %d): %w", milestone.PlanID, milestone.StreakTarget, err)
	}
	return nil
}

func (r *rewardRepository) DeleteMilestone(ctx context.Context, milestoneID uuid.UUID) error {
	if r.db == nil {
		return fmt.Errorf("database connection is nil")
	}

	query := `DELETE FROM reward_plan_milestones WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, milestoneID)
	if err != nil {
		return fmt.Errorf("failed deleting milestone %s: %w", milestoneID.String(), err)
	}
	return nil
}

func (r *rewardRepository) GetUserClaimedRewards(ctx context.Context, userID uuid.UUID, planID string) ([]models.UserClaimedReward, error) {
	if r.db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	query := `
		SELECT id, user_id, plan_id, streak_target, item_id, claimed_at
		FROM user_claimed_rewards
		WHERE user_id = $1 AND plan_id = $2
	`
	rows, err := r.db.QueryContext(ctx, query, userID, planID)
	if err != nil {
		return nil, fmt.Errorf("failed querying claimed rewards: %w", err)
	}
	defer rows.Close()

	var claims []models.UserClaimedReward
	for rows.Next() {
		var c models.UserClaimedReward
		if err := rows.Scan(&c.ID, &c.UserID, &c.PlanID, &c.StreakTarget, &c.ItemID, &c.ClaimedAt); err != nil {
			return nil, fmt.Errorf("failed scanning claimed reward row: %w", err)
		}
		claims = append(claims, c)
	}
	return claims, nil
}

func (r *rewardRepository) IsRewardClaimed(ctx context.Context, userID uuid.UUID, planID string, streakTarget int, itemID string) (bool, error) {
	if r.db == nil {
		return false, fmt.Errorf("database connection is nil")
	}

	query := `SELECT COUNT(*) FROM user_claimed_rewards WHERE user_id = $1 AND plan_id = $2 AND streak_target = $3 AND item_id = $4`
	var count int
	err := r.db.QueryRowContext(ctx, query, userID, planID, streakTarget, itemID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed checking claim status: %w", err)
	}
	return count > 0, nil
}

func (r *rewardRepository) ClaimReward(ctx context.Context, claim *models.UserClaimedReward) error {
	if r.db == nil {
		return fmt.Errorf("database connection is nil")
	}

	query := `
		INSERT INTO user_claimed_rewards (user_id, plan_id, streak_target, item_id, claimed_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (user_id, plan_id, streak_target, item_id) DO NOTHING
		RETURNING id, claimed_at
	`
	err := r.db.QueryRowContext(ctx, query, claim.UserID, claim.PlanID, claim.StreakTarget, claim.ItemID).Scan(&claim.ID, &claim.ClaimedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("reward already claimed")
		}
		return fmt.Errorf("failed recording claimed reward: %w", err)
	}
	return nil
}
