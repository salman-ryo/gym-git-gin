package handler

import (
	"net/http"

	"gymgit/backend/internal/middleware"
	"gymgit/backend/internal/models"
	"gymgit/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthHandler struct {
	authService service.AuthService
}

// NewAuthHandler initializes a new AuthHandler
func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Bootstrap creates or retrieves user profile idempotently
func (h *AuthHandler) Bootstrap(c *gin.Context) {
	authUserIDVal, exists := c.Get(middleware.ContextAuthUserIDKey)
	if !exists {
		models.SendError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Missing authenticated user context", nil)
		return
	}
	authUserID := authUserIDVal.(uuid.UUID)

	email, _ := c.Get(middleware.ContextUserEmailKey)
	name, _ := c.Get(middleware.ContextUserNameKey)
	avatarURL, _ := c.Get(middleware.ContextAvatarURLKey)
	provider, _ := c.Get(middleware.ContextProviderKey)

	user, err := h.authService.BootstrapProfile(
		c.Request.Context(),
		authUserID,
		email.(string),
		name.(string),
		avatarURL.(string),
		provider.(string),
	)
	if err != nil {
		models.SendError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to bootstrap profile", []string{err.Error()})
		return
	}

	models.SendSuccess(c, http.StatusOK, user, "Profile bootstrapped successfully")
}

// GetMe returns authenticated user profile and active weekly plan
func (h *AuthHandler) GetMe(c *gin.Context) {
	authUserIDVal, exists := c.Get(middleware.ContextAuthUserIDKey)
	if !exists {
		models.SendError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Missing authenticated user context", nil)
		return
	}
	authUserID := authUserIDVal.(uuid.UUID)

	user, activePlan, err := h.authService.GetProfile(c.Request.Context(), authUserID)
	if err != nil {
		models.SendError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve user profile", []string{err.Error()})
		return
	}

	if user == nil {
		// Auto-bootstrap profile if not found
		email, _ := c.Get(middleware.ContextUserEmailKey)
		name, _ := c.Get(middleware.ContextUserNameKey)
		avatarURL, _ := c.Get(middleware.ContextAvatarURLKey)
		provider, _ := c.Get(middleware.ContextProviderKey)

		var errBootstrap error
		user, errBootstrap = h.authService.BootstrapProfile(
			c.Request.Context(),
			authUserID,
			email.(string),
			name.(string),
			avatarURL.(string),
			provider.(string),
		)
		if errBootstrap != nil {
			models.SendError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "User profile not found and auto-bootstrap failed", []string{errBootstrap.Error()})
			return
		}
		_, activePlan, _ = h.authService.GetProfile(c.Request.Context(), authUserID)
	}

	models.SendSuccess(c, http.StatusOK, gin.H{
		"user": user,
		"plan": activePlan,
	}, "User profile retrieved successfully")
}

type UpdatePlanRequest struct {
	PlanID string `json:"plan_id" binding:"required"`
}

// UpdatePlan updates the user's selected weekly plan
func (h *AuthHandler) UpdatePlan(c *gin.Context) {
	authUserIDVal, exists := c.Get(middleware.ContextAuthUserIDKey)
	if !exists {
		models.SendError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Missing authenticated user context", nil)
		return
	}
	authUserID := authUserIDVal.(uuid.UUID)

	var req UpdatePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.PlanID == "" {
		models.SendError(c, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body; plan_id is required", nil)
		return
	}

	user, _, err := h.authService.GetProfile(c.Request.Context(), authUserID)
	if err != nil || user == nil {
		models.SendError(c, http.StatusNotFound, "NOT_FOUND", "User profile not found", nil)
		return
	}

	if err := h.authService.UpdatePlan(c.Request.Context(), user.ID, req.PlanID); err != nil {
		models.SendError(c, http.StatusBadRequest, "UPDATE_FAILED", err.Error(), nil)
		return
	}

	models.SendSuccess(c, http.StatusOK, gin.H{
		"user_id": user.ID,
		"plan_id": req.PlanID,
	}, "Weekly plan updated successfully")
}

// Logout performs backend logout cleanup
func (h *AuthHandler) Logout(c *gin.Context) {
	models.SendSuccess(c, http.StatusOK, nil, "Logged out successfully")
}
