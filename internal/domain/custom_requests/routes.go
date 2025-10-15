package custom_requests

import (
	"github.com/gofiber/fiber/v2"
	"errandShop/config"
	"errandShop/internal/middleware"
)

// SetupRoutes configures user custom request routes
func SetupRoutes(api fiber.Router, handler *Handler, cfg *config.Config) {
	// Protected user routes (JWT required)
	customRequestRoutes := api.Group("/custom-requests", middleware.JWTMiddleware(cfg))
	
	// User endpoints
	customRequestRoutes.Post("/", handler.CreateCustomRequest)                    // Create custom request
	customRequestRoutes.Get("/:id", handler.GetCustomRequest)                     // Get custom request by ID
	customRequestRoutes.Put("/:id", handler.UpdateCustomRequest)                  // Update custom request
	customRequestRoutes.Delete("/:id", handler.DeleteCustomRequest)               // Delete custom request
	customRequestRoutes.Delete("/:id/permanent-delete", handler.PermanentlyDeleteCustomRequest) // Permanently delete cancelled custom request
	customRequestRoutes.Post("/:id/cancel", handler.CancelCustomRequest)          // Cancel custom request
	customRequestRoutes.Get("/", handler.ListUserCustomRequests)                 // List user's custom requests
	customRequestRoutes.Post("/accept-quote", handler.AcceptQuote)               // Accept quote
	customRequestRoutes.Post("/:id/accept", handler.AcceptQuoteByRequestID)        // Accept quote by request ID
	customRequestRoutes.Post("/:id/messages", handler.SendMessage)                // Send message
}

// SetupAdminRoutes configures admin custom request routes
func SetupAdminRoutes(adminRoutes fiber.Router, handler *Handler, cfg *config.Config) {
	// Admin custom request routes
	customRequestAdminRoutes := adminRoutes.Group("/custom-requests")
	
	// Admin endpoints
	customRequestAdminRoutes.Get("/", handler.ListCustomRequestsAdmin)                    // List all custom requests
	customRequestAdminRoutes.Get("/:id", handler.GetCustomRequestAdmin)                   // Get custom request by ID (admin)
	customRequestAdminRoutes.Put("/:id/status", handler.UpdateCustomRequestStatus)        // Update request status
	customRequestAdminRoutes.Put("/:id/assign/:assignee_id", handler.AssignCustomRequest) // Assign request
	customRequestAdminRoutes.Post("/quotes", handler.CreateQuote)                        // Create quote
	customRequestAdminRoutes.Put("/quotes/:id", handler.UpdateQuote)                     // Update quote
	customRequestAdminRoutes.Put("/quotes/:id/send", handler.SendQuote)                 // Send quote
	customRequestAdminRoutes.Post("/:id/messages", handler.SendMessageAdmin)             // Send admin message
	customRequestAdminRoutes.Get("/stats", handler.GetCustomRequestStats)               // Get statistics
	customRequestAdminRoutes.Post("/bulk/status", handler.BulkUpdateStatus)             // Bulk update status
	customRequestAdminRoutes.Post("/bulk/assign", handler.BulkAssign)                   // Bulk assign
	
	// Admin cancel and delete endpoints (no status restrictions)
	customRequestAdminRoutes.Post("/:id/cancel", handler.CancelCustomRequestAdmin)                    // Cancel any request (admin)
	customRequestAdminRoutes.Delete("/:id/permanent-delete", handler.PermanentlyDeleteCustomRequestAdmin) // Permanently delete any request (admin)
}