package handler

import (
	"net/http"

	"gymgit/backend/internal/middleware"
	"gymgit/backend/internal/models"
	"gymgit/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type InventoryHandler struct {
	inventoryService service.InventoryService
}

// NewInventoryHandler initializes a new InventoryHandler
func NewInventoryHandler(inventoryService service.InventoryService) *InventoryHandler {
	return &InventoryHandler{inventoryService: inventoryService}
}

// GetCatalog returns available item definitions from master catalog
func (h *InventoryHandler) GetCatalog(c *gin.Context) {
	catalog, err := h.inventoryService.GetCatalog(c.Request.Context())
	if err != nil {
		models.SendError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve item catalog", []string{err.Error()})
		return
	}
	models.SendSuccess(c, http.StatusOK, catalog, "Item catalog retrieved successfully")
}

// GetInventory returns user item balances and active item effects
func (h *InventoryHandler) GetInventory(c *gin.Context) {
	authUserIDVal, exists := c.Get(middleware.ContextAuthUserIDKey)
	if !exists {
		models.SendError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Missing authenticated user context", nil)
		return
	}
	authUserID := authUserIDVal.(uuid.UUID)

	inventory, activeEffects, err := h.inventoryService.GetUserInventory(c.Request.Context(), authUserID)
	if err != nil {
		models.SendError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve user inventory", []string{err.Error()})
		return
	}

	models.SendSuccess(c, http.StatusOK, gin.H{
		"inventory":      inventory,
		"active_effects": activeEffects,
	}, "User inventory retrieved successfully")
}

// UseItem consumes or activates an item from user inventory
func (h *InventoryHandler) UseItem(c *gin.Context) {
	authUserIDVal, exists := c.Get(middleware.ContextAuthUserIDKey)
	if !exists {
		models.SendError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Missing authenticated user context", nil)
		return
	}
	authUserID := authUserIDVal.(uuid.UUID)

	var req models.UseItemRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.ItemID == "" {
		models.SendError(c, http.StatusBadRequest, "BAD_REQUEST", "Invalid request payload; item_id is required", nil)
		return
	}

	loc := middleware.GetUserLocationFromContext(c)

	result, err := h.inventoryService.UseItem(c.Request.Context(), authUserID, req.ItemID, req.Quantity, req.Payload, loc)
	if err != nil {
		models.SendError(c, http.StatusBadRequest, "ITEM_USE_FAILED", err.Error(), nil)
		return
	}

	models.SendSuccess(c, http.StatusOK, result, "Item used successfully")
}
