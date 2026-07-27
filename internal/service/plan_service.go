package service

import (
	"context"
	"fmt"
	"gymgit/backend/internal/models"
	"gymgit/backend/internal/repository"
)

type planService struct {
	planRepo repository.PlanRepository
}

// NewPlanService creates a new PlanService instance
func NewPlanService(planRepo repository.PlanRepository) PlanService {
	return &planService{planRepo: planRepo}
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
