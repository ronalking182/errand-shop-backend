package v1

import (
	"errandShop/internal/domain/products"

	"github.com/gofiber/fiber/v2"
)

// Public product routes
func MountProductRoutes(r fiber.Router, h *products.Handler) {
	r.Get("/products", h.List)
	r.Get("/products/categories", h.GetCategories)
	r.Get("/categories", h.GetCategories) // Direct categories endpoint for frontend compatibility
	r.Get("/products/:id", h.Get)
}

// Admin product routes
func MountAdminProductRoutes(r fiber.Router, h *products.Handler) {
	// Admin product management
	r.Get("/products", h.AdminList)
	r.Post("/products", h.Create)
	
	// Categories (before parameterized routes)
	r.Get("/products/categories", h.GetCategories)
	
	// Stock management (before parameterized routes)
	r.Post("/products/stock/bulk-update", h.BulkUpdateStock)
	r.Get("/products/low-stock", h.GetLowStock)

	// Parameterized routes (must come last)
	r.Get("/products/:id", h.Get)
	r.Put("/products/:id", h.Update)
	r.Delete("/products/:id", h.Delete)
}

// Superadmin-only category CRUD routes (mount under an admin group with SuperAdmin middleware)
func MountSuperAdminCategoryRoutes(r fiber.Router, h *products.Handler) {
	// CRUD on categories table
	r.Post("/categories", h.CreateCategory)
	r.Get("/categories", h.ListCategories)
	r.Get("/categories/:id", h.GetCategory)
	r.Put("/categories/:id", h.UpdateCategory)
	r.Delete("/categories/:id", h.DeleteCategory)
}
