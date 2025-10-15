package delivery

import (
	"errandShop/config"
	"errandShop/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

// SetupDeliveryRoutes sets up delivery routes (simplified for third-party logistics)
func SetupDeliveryRoutes(app *fiber.App, handler *DeliveryHandler, cfg *config.Config) {
	// Public routes
	public := app.Group("/api/v1/delivery")
	public.Post("/quote", handler.GetDeliveryQuote)
	public.Get("/track/:tracking_number", handler.GetDeliveryByTrackingNumber)
	
	// Order confirmation route (separate group for correct path)
	orders := app.Group("/api/v1/orders")
	orders.Post("/confirm", handler.ConfirmOrder)

	// Delivery costing routes (public for React Native compatibility)
	public.Post("/estimate", handler.EstimateDelivery)

	// Protected routes (authenticated users)
	protected := app.Group("/api/v1/delivery")
	protected.Use(middleware.JWTMiddleware(cfg))

	// Customer delivery endpoints
	protected.Post("/", handler.CreateDelivery)
	protected.Get("/:id", handler.GetDelivery)
	protected.Get("/order/:order_id", handler.GetDeliveryByOrderID)
	protected.Get("/:id/tracking", handler.GetTrackingUpdates)

	// Admin routes (for managing third-party logistics)
	admin := app.Group("/api/v1/delivery/admin")
	admin.Use(middleware.JWTMiddleware(cfg))
	admin.Use(middleware.RBACMiddleware("admin", "superadmin")) // Only admins and super admins

	// Delivery management
	admin.Get("/deliveries", handler.ListDeliveries)
	admin.Put("/deliveries/:id/status", handler.UpdateDeliveryStatus)
	admin.Put("/deliveries/:id/assign-provider", handler.AssignLogisticsProvider)
	admin.Put("/deliveries/:id/cancel", handler.CancelDelivery)

	// Analytics
	admin.Get("/stats", handler.GetDeliveryStats)
	admin.Get("/providers", handler.GetLogisticsProviders)
	// Replace RBAC middleware line with:
	protected.Post("/", handler.CreateDelivery) // Remove RBAC middleware
	// Or use: deliveryRoutes.Post("/", middleware.JWTAuth(), handler.CreateDelivery)
	protected.Post("/drivers", handler.CreateDriver)
}
