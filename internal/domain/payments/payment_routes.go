package payments

import (
	"errandShop/config"
	"errandShop/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

// Type alias for the service interface
type PaymentService = Service

type PaymentHandler = Handler

func NewPaymentHandler(service PaymentService) *PaymentHandler {
	return NewHandler(service)
}

// SetupRoutes sets up payment routes
func SetupRoutes(app *fiber.App, cfg *config.Config, handler *PaymentHandler) {
	api := app.Group("/api/v1")
	payments := api.Group("/payments")

	// Paystack routes (public, no authentication required) - MUST be registered first
	paystack := payments.Group("/paystack")
	paystack.Post("/initialize", handler.InitializePaystackPayment)
	paystack.Get("/verify/:reference", handler.VerifyPaystackPayment)

	// Webhook routes (no authentication required)
	webhooks := app.Group("/api/v1/webhooks")
	webhooks.Post("/paystack", handler.PaystackWebhook)

	// Protected routes (require authentication) - registered after specific routes
	protected := payments.Group("", middleware.JWTMiddleware(cfg))
	protected.Post("/initialize", handler.InitializePayment)
	protected.Post("/process", handler.ProcessPayment)
	protected.Get("/:id", handler.GetPayment)
	protected.Get("/transaction/:ref", handler.GetPaymentByTransactionRef)
	protected.Post("/:id/refund", handler.InitiateRefund)
	protected.Get("/refund/:id", handler.GetRefund)
	protected.Get("/:payment_id/refunds", handler.GetPaymentRefunds)
	protected.Post("/webhook", handler.ProcessWebhook)
}

// SetupAdminRoutes sets up admin payment routes
func SetupAdminRoutes(app *fiber.App, cfg *config.Config, handler *PaymentHandler) {
	admin := app.Group("/api/v1/admin/payments")
	admin.Use(middleware.JWTMiddleware(cfg))
	admin.Use(middleware.AdminOnly())

	// Admin routes
	admin.Get("/", handler.GetAllPayments)
	admin.Get("/stats", handler.GetPaymentStats)
	admin.Post("/refund/:id/process", handler.ProcessRefund)
}
