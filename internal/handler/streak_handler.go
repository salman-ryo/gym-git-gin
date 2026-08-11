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
	streakService service.StreakService
}

// NewStreakHandler initializes a new StreakHandler
func NewStreakHandler(streakService service.StreakService) *StreakHandler {
	return &StreakHandler{streakService: streakService}
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
