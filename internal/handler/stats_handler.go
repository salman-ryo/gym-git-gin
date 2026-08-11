package handler

import (
	"net/http"
	"strconv"

	"gymgit/backend/internal/middleware"
	"gymgit/backend/internal/models"
	"gymgit/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type StatsHandler struct {
	statsService service.StatsService
	authService  service.AuthService
}

// NewStatsHandler initializes a new StatsHandler
func NewStatsHandler(statsService service.StatsService, authService service.AuthService) *StatsHandler {
	return &StatsHandler{
		statsService: statsService,
		authService:  authService,
	}
}

// GetStats returns general dashboard statistics and scientific streak
func (h *StatsHandler) GetStats(c *gin.Context) {
	userID, ok := middleware.GetResolvedUserID(c)
	if !ok {
		return
	}

	days := 30
	if daysStr := c.Query("days"); daysStr != "" {
		if parsedDays, err := strconv.Atoi(daysStr); err == nil && parsedDays > 0 {
			days = parsedDays
		}
	}

	stats, err := h.statsService.GetDashboardStats(c.Request.Context(), userID, days)
	if err != nil {
		models.SendError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to calculate dashboard statistics", []string{err.Error()})
		return
	}

	models.SendSuccess(c, http.StatusOK, stats, "Dashboard statistics retrieved successfully")
}

// GetPowerStats returns Gym Power Score breakdown and gamified Anime Tier mapping
func (h *StatsHandler) GetPowerStats(c *gin.Context) {
	userID, ok := middleware.GetResolvedUserID(c)
	if !ok {
		return
	}

	days := 30
	if daysStr := c.Query("days"); daysStr != "" {
		if parsedDays, err := strconv.Atoi(daysStr); err == nil && parsedDays > 0 {
			days = parsedDays
		}
	}

	powerStats, err := h.statsService.GetPowerStats(c.Request.Context(), userID, days)
	if err != nil {
		models.SendError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to calculate Gym Power Score statistics", []string{err.Error()})
		return
	}

	models.SendSuccess(c, http.StatusOK, powerStats, "Gym Power Score and Anime Tier retrieved successfully")
}
