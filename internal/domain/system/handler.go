package system

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GetSystemConfig gets the current system configuration
func (h *Handler) GetSystemConfig(c *gin.Context) {
	config, err := h.service.GetSystemConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get system configuration",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    ToSystemConfigResponse(config),
	})
}

// UpdateSystemConfig updates the system configuration
func (h *Handler) UpdateSystemConfig(c *gin.Context) {
	var req UpdateSystemConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	if err := h.service.UpdateSystemConfig(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to update system configuration",
		})
		return
	}

	// Log the action
	userID := c.GetUint("userID")
	h.service.LogSystemAction(userID, "UPDATE_CONFIG", "System configuration updated", c.ClientIP(), c.GetHeader("User-Agent"))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "System configuration updated successfully",
	})
}

// GetSettings gets all system settings
func (h *Handler) GetSettings(c *gin.Context) {
	settings, err := h.service.GetAllSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get settings",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    settings,
	})
}

// GetSetting gets a specific setting
func (h *Handler) GetSetting(c *gin.Context) {
	key := c.Param("key")
	setting, err := h.service.GetSetting(key)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Setting not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    setting,
	})
}

// UpdateSetting updates a specific setting
func (h *Handler) UpdateSetting(c *gin.Context) {
	key := c.Param("key")
	var value map[string]interface{}
	if err := c.ShouldBindJSON(&value); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	if err := h.service.UpdateSetting(key, value); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to update setting",
		})
		return
	}

	// Log the action
	userID := c.GetUint("userID")
	h.service.LogSystemAction(userID, "UPDATE_SETTING", fmt.Sprintf("Setting '%s' updated", key), c.ClientIP(), c.GetHeader("User-Agent"))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Setting updated successfully",
	})
}

// GetAuditLogs gets audit logs with pagination
func (h *Handler) GetAuditLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	userIDStr := c.Query("userId")

	var userID *uint
	if userIDStr != "" {
		if id, err := strconv.ParseUint(userIDStr, 10, 32); err == nil {
			uid := uint(id)
			userID = &uid
		}
	}

	logs, total, err := h.service.GetAuditLogs(page, limit, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get audit logs",
		})
		return
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))
	response := AuditLogListResponse{
		Logs:       ToAuditLogResponses(logs),
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// GetSystemHealth gets system health status
func (h *Handler) GetSystemHealth(c *gin.Context) {
	health, err := h.service.GetSystemHealth()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get system health",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    health,
	})
}

// GetSystemStats gets system statistics
func (h *Handler) GetSystemStats(c *gin.Context) {
	stats, err := h.service.GetSystemStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get system stats",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// UpdateSystemHealth updates system health status
func (h *Handler) UpdateSystemHealth(c *gin.Context) {
	var req SystemHealthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	if err := h.service.UpdateSystemHealth(req.Component, req.Status, req.Message); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to update system health",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "System health updated successfully",
	})
}

// CleanupAuditLogs cleans up old audit logs
func (h *Handler) CleanupAuditLogs(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "90"))

	if err := h.service.CleanupOldAuditLogs(days); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to cleanup audit logs",
		})
		return
	}

	// Log the action
	userID := c.GetUint("userID")
	h.service.LogSystemAction(userID, "CLEANUP_LOGS", fmt.Sprintf("Cleaned up audit logs older than %d days", days), c.ClientIP(), c.GetHeader("User-Agent"))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Audit logs cleaned up successfully",
	})
}
