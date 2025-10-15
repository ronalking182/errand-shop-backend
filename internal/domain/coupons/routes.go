package coupons

import (
	"errandShop/config"
	"errandShop/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

// SetupRoutes sets up all coupon-related routes
func SetupRoutes(app *fiber.App, handler *Handler, cfg *config.Config) {
	// API v1 group
	api := app.Group("/api/v1")
	
	// Admin routes (protected)
	adminRoutes := api.Group("/admin")
	adminRoutes.Use(middleware.JWTMiddleware(cfg))
	adminRoutes.Use(middleware.AdminMiddleware()) // Ensure user has admin privileges
	
	// Admin Coupon Management
	couponAdmin := adminRoutes.Group("/coupons")
	couponAdmin.Get("/", handler.ListCoupons)              // GET /api/v1/admin/coupons
	couponAdmin.Post("/", handler.CreateCoupon)             // POST /api/v1/admin/coupons
	couponAdmin.Get("/stats", handler.GetCouponStats)       // GET /api/v1/admin/coupons/stats
	couponAdmin.Get("/:id", handler.GetCoupon)              // GET /api/v1/admin/coupons/:id
	couponAdmin.Put("/:id", handler.UpdateCoupon)           // PUT /api/v1/admin/coupons/:id
	couponAdmin.Delete("/:id", handler.DeleteCoupon)        // DELETE /api/v1/admin/coupons/:id
	couponAdmin.Post("/:id/toggle", handler.ToggleCouponActive) // POST /api/v1/admin/coupons/:id/toggle
	
	// User routes (protected)
	userRoutes := api.Group("/user")
	userRoutes.Use(middleware.JWTMiddleware(cfg))
	
	// User Coupon Operations
	couponUser := userRoutes.Group("/coupons")
	couponUser.Get("/available", handler.GetAvailableCoupons) // GET /api/v1/user/coupons/available
	couponUser.Post("/validate", handler.ValidateCoupon)      // POST /api/v1/user/coupons/validate
	couponUser.Post("/apply", handler.ApplyCoupon)           // POST /api/v1/user/coupons/apply
	couponUser.Post("/generate", handler.MobileAutoGenerateCoupon) // POST /api/v1/user/coupons/generate
	
	// User Refund Credits
	refundUser := userRoutes.Group("/refund-credits")
	refundUser.Get("/", handler.GetUserRefundCredits)        // GET /api/v1/user/refund-credits
	refundUser.Post("/convert", handler.ConvertRefundToCredit) // POST /api/v1/user/refund-credits/convert
	
	// System routes (admin only - for internal operations)
	systemRoutes := api.Group("/system")
	systemRoutes.Use(middleware.JWTMiddleware(cfg))
	systemRoutes.Use(middleware.AdminMiddleware()) // Only admins can access system operations
	
	// System Coupon Operations
	couponSystem := systemRoutes.Group("/coupons")
	couponSystem.Post("/generate-refund", handler.GenerateRefundCoupon)   // POST /api/v1/system/coupons/generate-refund
	couponSystem.Post("/auto-generate", handler.AutoGenerateUserCoupon)   // POST /api/v1/system/coupons/auto-generate
	
	// System Refund Credits
	refundSystem := systemRoutes.Group("/refund-credits")
	refundSystem.Post("/create", handler.CreateRefundCredit) // POST /api/v1/system/refund-credits/create
}

// SetupPublicRoutes sets up public coupon routes (no authentication required)
func SetupPublicRoutes(app *fiber.App, handler *Handler) {
	// Public coupon operations - using /public path to avoid /api/v1 middleware
	publicCoupons := app.Group("/public/coupons")
	publicCoupons.Post("/validate", handler.ValidatePublicCoupon) // POST /public/coupons/validate
	publicCoupons.Get("/", handler.GetPublicCoupons)              // GET /public/coupons
}