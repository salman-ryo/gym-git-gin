package handler

import (
	"net/http"

	"gymgit/backend/internal/models"
	"gymgit/backend/internal/service"

	"github.com/gin-gonic/gin"
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
