package system

import (
	"time"

	"github.com/gin-gonic/gin"
)

// SetupSystemRoutes sets up all system-related routes
func SetupSystemRoutes(router *gin.Engine, handler *Handler) {
	// Admin routes - require admin role
	admin := router.Group("/api/v1/admin/system")
	// TODO: Add proper Gin-compatible JWT and Admin middleware
	{
		// System configuration
		admin.GET("/config", handler.GetSystemConfig)
		admin.PUT("/config", handler.UpdateSystemConfig)

		// Settings management
		admin.GET("/settings", handler.GetSettings)
		admin.GET("/settings/:key", handler.GetSetting)
		admin.PUT("/settings/:key", handler.UpdateSetting)

		// System health
		admin.GET("/health", handler.GetSystemHealth)
		admin.PUT("/health", handler.UpdateSystemHealth)

		// System statistics
		admin.GET("/stats", handler.GetSystemStats)

		// Audit logs
		admin.GET("/audit-logs", handler.GetAuditLogs)
		admin.DELETE("/audit-logs/cleanup", handler.CleanupAuditLogs)
	}

	// Super admin routes - require superadmin role
	// TODO: Add proper Gin-compatible JWT and SuperAdmin middleware
	// superAdmin := router.Group("/api/v1/superadmin/system")
	// {
	//	// Advanced system operations (if needed)
	//	// These could include more sensitive operations
	// }

	// Public health check endpoint
	public := router.Group("/api/v1")
	{
		public.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status":    "healthy",
				"timestamp": time.Now().Unix(),
			})
		})
	}
}
