package service

import (
	"context"
	"fmt"

	"gymgit/backend/internal/models"
	"gymgit/backend/internal/repository"

	"github.com/google/uuid"
)

type authService struct {
	userRepo repository.UserRepository
	planRepo repository.PlanRepository
}

// NewAuthService creates a new AuthService instance
func NewAuthService(userRepo repository.UserRepository, planRepo repository.PlanRepository) AuthService {
	return &authService{
		userRepo: userRepo,
		planRepo: planRepo,
	}
}

func (s *authService) BootstrapProfile(ctx context.Context, authUserID uuid.UUID, email, name, avatarURL, provider string) (*models.User, error) {
	// Check if profile already exists
	existingUser, err := s.userRepo.GetByAuthUserID(ctx, authUserID)
	if err != nil {
		return nil, fmt.Errorf("failed checking existing user: %w", err)
	}

	if existingUser != nil {
		return existingUser, nil
	}

	// Default weekly plan
	defaultPlanID := "ppl"
	var namePtr *string
	if name != "" {
		namePtr = &name
	}
	var avatarPtr *string
	if avatarURL != "" {
		avatarPtr = &avatarURL
	}

	newUser := &models.User{
		AuthUserID:   authUserID,
		Email:        email,
		Name:         namePtr,
		AvatarURL:    avatarPtr,
		Provider:     provider,
		WeeklyPlanID: &defaultPlanID,
	}

	if err := s.userRepo.Create(ctx, newUser); err != nil {
		return nil, fmt.Errorf("failed creating user profile: %w", err)
	}

	return newUser, nil
}

func (s *authService) GetProfile(ctx context.Context, authUserID uuid.UUID) (*models.User, *models.WeeklyPlan, error) {
	user, err := s.userRepo.GetByAuthUserID(ctx, authUserID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed retrieving user: %w", err)
	}
	if user == nil {
		return nil, nil, nil
	}

	var plan *models.WeeklyPlan
	if user.WeeklyPlanID != nil && *user.WeeklyPlanID != "" {
		plan, err = s.planRepo.GetByID(ctx, *user.WeeklyPlanID)
		if err != nil {
			return user, nil, fmt.Errorf("failed retrieving user active plan: %w", err)
		}
	}

	// Fallback to default plan if none assigned or not found
	if plan == nil {
		plan, _ = s.planRepo.GetByID(ctx, "ppl")
	}

	return user, plan, nil
}

func (s *authService) UpdatePlan(ctx context.Context, userID uuid.UUID, planID string) error {
	// Verify target plan exists
	plan, err := s.planRepo.GetByID(ctx, planID)
	if err != nil {
		return fmt.Errorf("error verifying plan ID: %w", err)
	}
	if plan == nil {
		return fmt.Errorf("invalid weekly plan ID: %s", planID)
	}

	if err := s.userRepo.UpdateWeeklyPlan(ctx, userID, planID); err != nil {
		return fmt.Errorf("failed to update plan: %w", err)
	}
	return nil
}
