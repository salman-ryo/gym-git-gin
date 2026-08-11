package handler

import (
	"net/http"

	"gymgit/backend/internal/middleware"
	"gymgit/backend/internal/models"
	"gymgit/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RewardHandler struct {
	rewardService service.RewardService
}

// NewRewardHandler initializes a new RewardHandler instance
func NewRewardHandler(rewardService service.RewardService) *RewardHandler {
	return &RewardHandler{rewardService: rewardService}
}

// GetRoadmap returns user streak progression roadmap with milestone statuses
func (h *RewardHandler) GetRoadmap(c *gin.Context) {
	userID, ok := middleware.GetResolvedUserID(c)
	if !ok {
		return
	}
	planID := c.Query("plan_id")

	roadmap, err := h.rewardService.GetRoadmap(c.Request.Context(), userID, planID)
	if err != nil {
		models.SendError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve streak reward roadmap", []string{err.Error()})
		return
	}

	models.SendSuccess(c, http.StatusOK, roadmap, "Streak reward roadmap retrieved successfully")
}

// ClaimReward handles user reward claims for unlocked milestone targets
func (h *RewardHandler) ClaimReward(c *gin.Context) {
	userID, ok := middleware.GetResolvedUserID(c)
	if !ok {
		return
	}

	var req models.ClaimRewardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		models.SendError(c, http.StatusBadRequest, "BAD_REQUEST", "Invalid request payload; streak_target and item_id are required", nil)
		return
	}

	result, err := h.rewardService.ClaimReward(c.Request.Context(), userID, req.PlanID, req.StreakTarget, req.ItemID)
	if err != nil {
		models.SendError(c, http.StatusBadRequest, "CLAIM_FAILED", err.Error(), nil)
		return
	}

	models.SendSuccess(c, http.StatusOK, result, "Reward claimed successfully")
}

// GetAllPlans returns all available reward plans
func (h *RewardHandler) GetAllPlans(c *gin.Context) {
	plans, err := h.rewardService.GetAllPlans(c.Request.Context())
	if err != nil {
		models.SendError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve reward plans", []string{err.Error()})
		return
	}

	models.SendSuccess(c, http.StatusOK, plans, "Reward plans retrieved successfully")
}

// Admin Endpoints for Plan & Milestone CRUD

// CreateRewardPlan allows admins to create or update a reward plan
func (h *RewardHandler) CreateRewardPlan(c *gin.Context) {
	var req models.CreateRewardPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		models.SendError(c, http.StatusBadRequest, "BAD_REQUEST", "Invalid request payload; id and name are required", nil)
		return
	}

	plan, err := h.rewardService.CreateRewardPlan(c.Request.Context(), req)
	if err != nil {
		models.SendError(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}

	models.SendSuccess(c, http.StatusCreated, plan, "Reward plan created successfully")
}

// DeleteRewardPlan allows admins to delete a reward plan
func (h *RewardHandler) DeleteRewardPlan(c *gin.Context) {
	planID := c.Param("id")
	if planID == "" {
		models.SendError(c, http.StatusBadRequest, "BAD_REQUEST", "Plan ID parameter is required", nil)
		return
	}

	if err := h.rewardService.DeleteRewardPlan(c.Request.Context(), planID); err != nil {
		models.SendError(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}

	models.SendSuccess(c, http.StatusOK, gin.H{"deleted_plan_id": planID}, "Reward plan deleted successfully")
}

// UpsertMilestone allows admins to add or update a milestone target (e.g., Day 11 -> RESTORE_SHIELD x5)
func (h *RewardHandler) UpsertMilestone(c *gin.Context) {
	planID := c.Param("id")
	if planID == "" {
		models.SendError(c, http.StatusBadRequest, "BAD_REQUEST", "Plan ID parameter is required", nil)
		return
	}

	var req models.UpsertMilestoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		models.SendError(c, http.StatusBadRequest, "BAD_REQUEST", "Invalid request payload", nil)
		return
	}

	milestone, err := h.rewardService.UpsertMilestone(c.Request.Context(), planID, req)
	if err != nil {
		models.SendError(c, http.StatusBadRequest, "UPSERT_FAILED", err.Error(), nil)
		return
	}

	models.SendSuccess(c, http.StatusOK, milestone, "Roadmap milestone saved successfully")
}

// DeleteMilestone allows admins to delete a milestone target from a plan
func (h *RewardHandler) DeleteMilestone(c *gin.Context) {
	milestoneIDStr := c.Param("milestone_id")
	milestoneID, err := uuid.Parse(milestoneIDStr)
	if err != nil {
		models.SendError(c, http.StatusBadRequest, "BAD_REQUEST", "Invalid milestone UUID parameter", nil)
		return
	}

	if err := h.rewardService.DeleteMilestone(c.Request.Context(), milestoneID); err != nil {
		models.SendError(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}

	models.SendSuccess(c, http.StatusOK, gin.H{"deleted_milestone_id": milestoneIDStr}, "Roadmap milestone deleted successfully")
}
