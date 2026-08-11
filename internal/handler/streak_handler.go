package handler

import (
	"net/http"

	"gymgit/backend/internal/middleware"
	"gymgit/backend/internal/models"
	"gymgit/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type StreakHandler struct {
	streakService    service.StreakService
	inventoryService service.InventoryService
}

// NewStreakHandler initializes a new StreakHandler
func NewStreakHandler(streakService service.StreakService, inventoryService service.InventoryService) *StreakHandler {
	return &StreakHandler{
		streakService:    streakService,
		inventoryService: inventoryService,
	}
}

// GetStreak returns active 7-day cycle status, rest tokens, accuracy score, and streak
func (h *StreakHandler) GetStreak(c *gin.Context) {
	authUserIDVal, exists := c.Get(middleware.ContextAuthUserIDKey)
	if !exists {
		models.SendError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Missing authenticated user context", nil)
		return
	}
	authUserID := authUserIDVal.(uuid.UUID)

	loc := middleware.GetUserLocationFromContext(c)

	streakState, err := h.streakService.GetStreakState(c.Request.Context(), authUserID, loc)
	if err != nil {
		models.SendError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve streak state", []string{err.Error()})
		return
	}

	models.SendSuccess(c, http.StatusOK, streakState, "Streak state retrieved successfully")
}

// RestoreStreak redeems a Restore Shield to revive a missed streak day within 3 days lookback
func (h *StreakHandler) RestoreStreak(c *gin.Context) {
	authUserIDVal, exists := c.Get(middleware.ContextAuthUserIDKey)
	if !exists {
		models.SendError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Missing authenticated user context", nil)
		return
	}
	authUserID := authUserIDVal.(uuid.UUID)

	var req models.RestoreShieldRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.TargetDate == "" {
		models.SendError(c, http.StatusBadRequest, "BAD_REQUEST", "Invalid request payload; target_date is required", nil)
		return
	}

	loc := middleware.GetUserLocationFromContext(c)

	result, err := h.inventoryService.RedeemRestoreShield(c.Request.Context(), authUserID, req.TargetDate, req.WorkoutType, req.Hours, loc)
	if err != nil {
		models.SendError(c, http.StatusBadRequest, "RESTORE_FAILED", err.Error(), nil)
		return
	}

	models.SendSuccess(c, http.StatusOK, result, "Streak restored successfully")
}

// FreezeStreak consumes STREAK_FREEZE_TOKEN(s) to set streak status to frozen (Ice Pause)
func (h *StreakHandler) FreezeStreak(c *gin.Context) {
	authUserIDVal, exists := c.Get(middleware.ContextAuthUserIDKey)
	if !exists {
		models.SendError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Missing authenticated user context", nil)
		return
	}
	authUserID := authUserIDVal.(uuid.UUID)

	var req struct {
		DurationDays int    `json:"duration_days"`
		Reason       string `json:"reason"`
	}
	_ = c.ShouldBindJSON(&req)
	if req.DurationDays <= 0 {
		req.DurationDays = 1
	}

	loc := middleware.GetUserLocationFromContext(c)

	// Consume STREAK_FREEZE_TOKEN from inventory
	useResult, err := h.inventoryService.UseItem(c.Request.Context(), authUserID, "STREAK_FREEZE_TOKEN", req.DurationDays, map[string]interface{}{"reason": req.Reason}, loc)
	if err != nil {
		models.SendError(c, http.StatusBadRequest, "FREEZE_FAILED", err.Error(), nil)
		return
	}

	state, err := h.streakService.FreezeStreak(c.Request.Context(), authUserID, req.DurationDays, req.Reason)
	if err != nil {
		models.SendError(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}

	models.SendSuccess(c, http.StatusOK, gin.H{
		"is_frozen":          state.IsFrozen,
		"tokens_consumed":    useResult.QuantityConsumed,
		"remaining_tokens":  useResult.RemainingQuantity,
		"active_until":       useResult.ActiveUntil,
		"details":            useResult.Details,
	}, "Streak successfully paused in ice")
}

// UnfreezeStreak manually deactivates an active streak freeze state
func (h *StreakHandler) UnfreezeStreak(c *gin.Context) {
	authUserIDVal, exists := c.Get(middleware.ContextAuthUserIDKey)
	if !exists {
		models.SendError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Missing authenticated user context", nil)
		return
	}
	authUserID := authUserIDVal.(uuid.UUID)

	state, err := h.streakService.UnfreezeStreak(c.Request.Context(), authUserID)
	if err != nil {
		models.SendError(c, http.StatusBadRequest, "UNFREEZE_FAILED", err.Error(), nil)
		return
	}

	models.SendSuccess(c, http.StatusOK, gin.H{
		"is_frozen": state.IsFrozen,
		"message":   "Streak freeze manually deactivated. Workout tracking resumed.",
	}, "Streak un-frozen successfully")
}

