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
	// 1. Check if profile already exists by AuthUserID
	existingUser, err := s.userRepo.GetByAuthUserID(ctx, authUserID)
	if err != nil {
		return nil, fmt.Errorf("failed checking existing user: %w", err)
	}

	if existingUser != nil {
		return existingUser, nil
	}

	// 2. Check if profile already exists by email (to handle Supabase user re-creation conflict)
	existingUserByEmail, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed checking user by email: %w", err)
	}

	if existingUserByEmail != nil {
		// Link the existing profile to the new Supabase AuthUserID
		if err := s.userRepo.UpdateAuthUserID(ctx, existingUserByEmail.ID, authUserID); err != nil {
			return nil, fmt.Errorf("failed to link existing user email to new auth ID: %w", err)
		}
		existingUserByEmail.AuthUserID = authUserID
		return existingUserByEmail, nil
	}

	// 3. Otherwise, create a new user profile
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
		WeeklyPlanID: nil, // Starts with no plan assigned to trigger onboarding prompt
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

	return user, plan, nil
}

func (s *authService) UpdatePlan(ctx context.Context, userID uuid.UUID, planID string, customName string, customDesc string, categories []string) error {
	var targetPlanID string
	if planID == "custom-plan" {
		targetPlanID = fmt.Sprintf("custom-%s", userID.String())

		if customName == "" {
			customName = "My Custom Weekly Plan"
		}
		if customDesc == "" {
			customDesc = "Personalized workout categories."
		}
		if len(categories) == 0 {
			categories = []string{"Push", "Pull", "Legs", "Custom"}
		}

		plan := &models.WeeklyPlan{
			ID:          targetPlanID,
			Name:        customName,
			Description: customDesc,
			Categories:  categories,
			UserID:      &userID,
		}
		if err := s.planRepo.Create(ctx, plan); err != nil {
			return fmt.Errorf("failed to save custom weekly plan: %w", err)
		}
	} else {
		targetPlanID = planID
		// Verify target plan exists
		plan, err := s.planRepo.GetByID(ctx, targetPlanID)
		if err != nil {
			return fmt.Errorf("error verifying plan ID: %w", err)
		}
		if plan == nil {
			return fmt.Errorf("invalid weekly plan ID: %s", targetPlanID)
		}
	}

	if err := s.userRepo.UpdateWeeklyPlan(ctx, userID, targetPlanID); err != nil {
		return fmt.Errorf("failed to update user weekly plan: %w", err)
	}
	return nil
}
