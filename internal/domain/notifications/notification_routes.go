package notifications

import (
	"errandShop/config"
	"errandShop/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

// SetupRoutes sets up notification routes
func SetupRoutes(app *fiber.App, cfg *config.Config, handler *NotificationHandler) {
	api := app.Group("/api/v1")
	notifications := api.Group("/notifications")

	// Protected routes (require authentication)
	protected := notifications.Group("", middleware.JWTMiddleware(cfg))
	protected.Get("/", handler.GetNotifications)
	protected.Put("/:id/read", handler.MarkAsRead)
	protected.Put("/read-all", handler.MarkAllAsRead)
	protected.Post("/push-token", handler.RegisterPushToken)
}

// SetupAdminRoutes sets up admin notification routes
func SetupAdminRoutes(app *fiber.App, cfg *config.Config, handler *NotificationHandler) {
	admin := app.Group("/api/v1/admin/notifications")
	admin.Use(middleware.JWTMiddleware(cfg))
	admin.Use(middleware.AdminMiddleware())

	// Admin routes
	admin.Post("/broadcast", handler.SendBroadcastNotification)
	admin.Get("/templates", handler.GetNotificationTemplates)
	admin.Post("/templates", handler.CreateNotificationTemplate)
	admin.Put("/templates/:id", handler.UpdateNotificationTemplate)
	admin.Delete("/templates/:id", handler.DeleteNotificationTemplate)
	admin.Get("/stats", handler.GetNotificationStats)
}
