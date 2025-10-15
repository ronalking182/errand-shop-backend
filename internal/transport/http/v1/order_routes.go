package v1

import (
	"errandShop/internal/domain/orders"

	"github.com/gofiber/fiber/v2"
)

// Customer order routes
func MountOrderRoutes(r fiber.Router, h *orders.Handler) {
	r.Get("/orders", h.List)
	r.Get("/orders/:id", h.Get)
	r.Post("/orders", h.Create)
	r.Put("/orders/:id/status", h.UpdateStatus)
}

// Admin order routes
func MountAdminOrderRoutes(r fiber.Router, h *orders.Handler) {
	// Order management
	r.Get("/orders", h.AdminList)
	r.Get("/orders/:id", h.AdminGet)
	r.Put("/orders/:id/status", h.AdminUpdateStatus)
	r.Put("/orders/:id/payment-status", h.AdminUpdatePaymentStatus)

	// Statistics
	r.Get("/orders/stats", h.GetStats)
}
