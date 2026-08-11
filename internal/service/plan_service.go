package service

import (
	"context"
	"fmt"

	"gymgit/backend/internal/models"
	"gymgit/backend/internal/repository"

	"github.com/google/uuid"
)

type planService struct {
	planRepo repository.PlanRepository
	userRepo repository.UserRepository
}

// NewPlanService creates a new PlanService instance
func NewPlanService(planRepo repository.PlanRepository, userRepo repository.UserRepository) PlanService {
	return &planService{
		planRepo: planRepo,
		userRepo: userRepo,
	}
}

func (s *planService) GetAllPlans(ctx context.Context) ([]models.WeeklyPlan, error) {
	plans, err := s.planRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed fetching weekly plans: %w", err)
	}
	return plans, nil
}

func (s *planService) GetPlanByID(ctx context.Context, id string) (*models.WeeklyPlan, error) {
	plan, err := s.planRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed fetching weekly plan %s: %w", id, err)
	}
	return plan, nil
}

func (s *planService) QueuePlanChange(ctx context.Context, userID uuid.UUID, planID string) error {
	// Verify target plan exists
	plan, err := s.planRepo.GetByID(ctx, planID)
	if err != nil {
		return fmt.Errorf("failed verifying plan ID: %w", err)
	}
	if plan == nil {
		return fmt.Errorf("invalid weekly plan ID: %s", planID)
	}

	if err := s.userRepo.SetQueuedPlan(ctx, userID, &planID); err != nil {
		return fmt.Errorf("failed setting queued weekly plan: %w", err)
	}
	return nil
}
