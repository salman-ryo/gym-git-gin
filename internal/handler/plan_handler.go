package handler

import (
	"net/http"

	"gymgit/backend/internal/middleware"
	"gymgit/backend/internal/models"
	"gymgit/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PlanHandler struct {
	planService service.PlanService
}

// NewPlanHandler initializes a new PlanHandler
func NewPlanHandler(planService service.PlanService) *PlanHandler {
	return &PlanHandler{planService: planService}
}

// GetPlans returns all available weekly plans
func (h *PlanHandler) GetPlans(c *gin.Context) {
	plans, err := h.planService.GetAllPlans(c.Request.Context())
	if err != nil {
		models.SendError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve plans", []string{err.Error()})
		return
	}

	models.SendSuccess(c, http.StatusOK, plans, "Weekly plans retrieved successfully")
}

type QueuePlanRequest struct {
	WeeklyPlanID string `json:"weekly_plan_id" binding:"required"`
}

// QueuePlan queues a weekly plan change to activate on the next 7-day cycle
func (h *PlanHandler) QueuePlan(c *gin.Context) {
	authUserIDVal, exists := c.Get(middleware.ContextAuthUserIDKey)
	if !exists {
		models.SendError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Missing authenticated user context", nil)
		return
	}
	authUserID := authUserIDVal.(uuid.UUID)

	var req QueuePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.WeeklyPlanID == "" {
		models.SendError(c, http.StatusBadRequest, "BAD_REQUEST", "Invalid request payload; weekly_plan_id is required", nil)
		return
	}

	if err := h.planService.QueuePlanChange(c.Request.Context(), authUserID, req.WeeklyPlanID); err != nil {
		models.SendError(c, http.StatusBadRequest, "QUEUE_FAILED", err.Error(), nil)
		return
	}

	models.SendSuccess(c, http.StatusOK, gin.H{
		"queued_weekly_plan_id": req.WeeklyPlanID,
	}, "Weekly plan queued successfully for the next 7-day cycle")
}
