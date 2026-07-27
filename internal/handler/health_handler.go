package handler

import (
	"net/http"

	"gymgit/backend/internal/models"

	"github.com/gin-gonic/gin"
)

// HealthHandler handles health check requests
type HealthHandler struct{}

// NewHealthHandler initializes a new HealthHandler
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// HealthCheck responds with 200 OK public status envelope
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	models.SendSuccess(c, http.StatusOK, gin.H{
		"status": "healthy",
	}, "Gym-Git backend API is operational")
}
