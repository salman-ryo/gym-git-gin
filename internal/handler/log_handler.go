package handler

import (
	"net/http"
	"time"

	"gymgit/backend/internal/middleware"
	"gymgit/backend/internal/models"
	"gymgit/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type LogHandler struct {
	logService  service.GymLogService
	authService service.AuthService
}

// NewLogHandler initializes a new LogHandler
func NewLogHandler(logService service.GymLogService, authService service.AuthService) *LogHandler {
	return &LogHandler{
		logService:  logService,
		authService: authService,
	}
}

// GetLogs returns gym logs for the authenticated user with optional filtering
func (h *LogHandler) GetLogs(c *gin.Context) {
	userID, ok := middleware.GetResolvedUserID(c)
	if !ok {
		return
	}

	var startDate, endDate *time.Time
	if startStr := c.Query("startDate"); startStr != "" {
		if t, err := time.Parse("2006-01-02", startStr); err == nil {
			startDate = &t
		}
	}

	if endStr := c.Query("endDate"); endStr != "" {
		if t, err := time.Parse("2006-01-02", endStr); err == nil {
			endDate = &t
		}
	}

	var workoutType *string
	if wt := c.Query("workoutType"); wt != "" {
		workoutType = &wt
	}

	logs, err := h.logService.GetLogs(c.Request.Context(), userID, startDate, endDate, workoutType)
	if err != nil {
		models.SendError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve gym logs", []string{err.Error()})
		return
	}

	models.SendSuccess(c, http.StatusOK, logs, "Gym logs retrieved successfully")
}

type SaveLogRequest struct {
	Date        string  `json:"date" binding:"required"`
	Hours       float64 `json:"hours"`
	WorkoutType string  `json:"workout_type" binding:"required"`
	Notes       *string `json:"notes"`
}

// UpsertLog handles creation/updating or deletion (if hours <= 0) of a gym log
func (h *LogHandler) UpsertLog(c *gin.Context) {
	userID, ok := middleware.GetResolvedUserID(c)
	if !ok {
		return
	}

	var req SaveLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		models.SendError(c, http.StatusBadRequest, "BAD_REQUEST", "Invalid log request payload", []string{err.Error()})
		return
	}

	// Validate date format (YYYY-MM-DD)
	if _, err := time.Parse("2006-01-02", req.Date); err != nil {
		models.SendError(c, http.StatusBadRequest, "BAD_REQUEST", "Invalid date format; expected YYYY-MM-DD", nil)
		return
	}

	// Override date if URL param :date is set (e.g. PUT /logs/:date)
	if dateParam := c.Param("date"); dateParam != "" {
		req.Date = dateParam
	}

	log, err := h.logService.SaveLog(c.Request.Context(), userID, req.Date, req.Hours, req.WorkoutType, req.Notes)
	if err != nil {
		models.SendError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to save gym log", []string{err.Error()})
		return
	}

	if log == nil {
		models.SendSuccess(c, http.StatusOK, gin.H{
			"date": req.Date,
		}, "Gym log removed (hours <= 0)")
		return
	}

	models.SendSuccess(c, http.StatusOK, log, "Gym log saved successfully")
}

// DeleteLog deletes a gym log by date
func (h *LogHandler) DeleteLog(c *gin.Context) {
	userID, ok := middleware.GetResolvedUserID(c)
	if !ok {
		return
	}

	date := c.Param("date")
	if date == "" {
		models.SendError(c, http.StatusBadRequest, "BAD_REQUEST", "Missing date path parameter", nil)
		return
	}

	if _, err := time.Parse("2006-01-02", date); err != nil {
		models.SendError(c, http.StatusBadRequest, "BAD_REQUEST", "Invalid date format; expected YYYY-MM-DD", nil)
		return
	}

	if err := h.logService.DeleteLog(c.Request.Context(), userID, date); err != nil {
		models.SendError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete gym log", []string{err.Error()})
		return
	}

	models.SendSuccess(c, http.StatusOK, gin.H{
		"date": date,
	}, "Gym log deleted successfully")
}

// ResetDemoLogs populates 365 days of demo historical workout logs for testing
func (h *LogHandler) ResetDemoLogs(c *gin.Context) {
	userID, ok := middleware.GetResolvedUserID(c)
	if !ok {
		return
	}

	if err := h.logService.ResetDemoLogs(c.Request.Context(), userID); err != nil {
		models.SendError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed resetting demo logs", []string{err.Error()})
		return
	}

	models.SendSuccess(c, http.StatusOK, gin.H{
		"user_id": userID,
		"days":    365,
	}, "Demo historical workout logs generated successfully")
}
