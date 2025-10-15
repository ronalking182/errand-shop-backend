package orders

import (
	"errandShop/config"
	"errandShop/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

// SetupRoutes configures all order and cart routes
func SetupRoutes(app *fiber.App, cfg *config.Config, orderHandler *Handler, cartHandler *CartHandler, couponHandler *CouponHandler) {
	api := app.Group("/api/v1")

	// Cart routes (protected)
	cart := api.Group("/cart", middleware.JWTMiddleware(cfg))
	cart.Get("/", cartHandler.GetCart)
	cart.Post("/items", cartHandler.AddToCart)
	cart.Put("/items/:id", cartHandler.UpdateCartItem)
	cart.Delete("/items/:id", cartHandler.RemoveFromCart)
	cart.Delete("/clear", cartHandler.ClearCart)

	// Coupon validation (public)
	api.Post("/coupons/validate", couponHandler.ValidateCoupon)

	// Customer order routes (protected) - specific routes to avoid conflicts
	api.Get("/orders", middleware.JWTMiddleware(cfg), orderHandler.List)
	api.Post("/orders", middleware.JWTMiddleware(cfg), orderHandler.Create)
	api.Get("/orders/:id", middleware.JWTMiddleware(cfg), orderHandler.Get)
	api.Put("/orders/:id/status", middleware.JWTMiddleware(cfg), orderHandler.UpdateStatus)
	api.Post("/orders/:id/cancel", middleware.JWTMiddleware(cfg), orderHandler.CancelOrder)

	// Admin routes (require admin role)
	admin := api.Group("/admin", middleware.JWTMiddleware(cfg), middleware.AdminMiddleware())
	adminOrders := admin.Group("/orders")
	adminOrders.Get("/", orderHandler.AdminList)
	adminOrders.Get("/:id", orderHandler.AdminGet)
	adminOrders.Put("/:id/status", orderHandler.AdminUpdateStatus)
	adminOrders.Put("/:id/payment-status", orderHandler.AdminUpdatePaymentStatus)
	adminOrders.Put("/:id/cancel", orderHandler.AdminCancelOrder)
	adminOrders.Get("/stats", orderHandler.GetStats)
}