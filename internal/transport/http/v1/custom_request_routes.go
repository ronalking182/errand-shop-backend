package v1

import (
	"github.com/gofiber/fiber/v2"
	custom_requests "errandShop/internal/domain/custom_requests"
)

// MountCustomRequestRoutes mounts custom request routes
func MountCustomRequestRoutes(router fiber.Router, handler *custom_requests.Handler, authMiddleware, adminMiddleware fiber.Handler) {
	// User routes (require authentication)
	userRoutes := router.Group("/custom-requests", authMiddleware)
	{
		// Custom request CRUD
		userRoutes.Post("/", handler.CreateCustomRequest)
		userRoutes.Get("/", handler.ListUserCustomRequests)
		userRoutes.Get("/:id", handler.GetCustomRequest)
		userRoutes.Put("/:id", handler.UpdateCustomRequest)
		userRoutes.Delete("/:id", handler.DeleteCustomRequest)
		userRoutes.Delete("/:id/permanent-delete", handler.PermanentlyDeleteCustomRequest)
		userRoutes.Post("/:id/cancel", handler.CancelCustomRequest)

		// Quote acceptance
		userRoutes.Post("/accept-quote", handler.AcceptQuote)

		// Messages
		userRoutes.Post("/:id/messages", handler.SendMessage)
	}

	// Admin routes (require admin authentication)
	adminRoutes := router.Group("/admin/custom-requests", authMiddleware, adminMiddleware)
	{
		// Custom request management
		adminRoutes.Get("/", handler.ListCustomRequestsAdmin)
		adminRoutes.Get("/:id", handler.GetCustomRequestAdmin)
		adminRoutes.Put("/:id/status", handler.UpdateCustomRequestStatus)
		adminRoutes.Put("/:id/assign/:assignee_id", handler.AssignCustomRequest)

		// Quote management
		adminRoutes.Post("/quotes", handler.CreateQuote)
		adminRoutes.Put("/quotes/:id/send", handler.SendQuote)

		// Messages
		adminRoutes.Post("/:id/messages", handler.SendMessageAdmin)

		// Statistics
		adminRoutes.Get("/stats", handler.GetCustomRequestStats)
		
		// Admin cancel and delete endpoints (no status restrictions)
		adminRoutes.Post("/:id/cancel", handler.CancelCustomRequestAdmin)
		adminRoutes.Delete("/:id/permanent-delete", handler.PermanentlyDeleteCustomRequestAdmin)
	}
}