package users

import (
	"errandShop/config"
	"errandShop/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

// SetupUserRoutes sets up all user-related routes
func SetupUserRoutes(app *fiber.App, cfg *config.Config, handler *Handler) {
	// Public routes
	api := app.Group("/api/v1")
	{
		// User registration (if needed)
		api.Post("/users/register", handler.CreateUser)
	}

	// Protected routes - require authentication
	protected := api.Group("/users")
	protected.Use(middleware.JWTMiddleware(cfg))
	{
		// User profile management
		protected.Get("/profile", handler.GetProfile)
		protected.Put("/profile", handler.UpdateProfile)
		protected.Put("/password", handler.UpdatePassword)
	}

	// Admin routes - require admin role
	admin := api.Group("/admin")
	admin.Use(middleware.JWTMiddleware(cfg))
	admin.Use(middleware.AdminMiddleware())
	{
		// User management
		admin.Get("/users", handler.GetUsers)
		admin.Get("/users/:id", handler.GetUser)
		admin.Post("/users", handler.CreateUser)
		admin.Put("/users/:id", handler.UpdateUser)
		admin.Delete("/users/:id", handler.DeleteUser)
		admin.Put("/users/:id/status", handler.ToggleUserStatus)
		admin.Put("/users/:id/force-password-reset", handler.ForcePasswordReset)
	}

	// Super admin routes - require superadmin role
	superAdmin := api.Group("/superadmin")
	superAdmin.Use(middleware.JWTMiddleware(cfg))
	superAdmin.Use(middleware.SuperAdminMiddleware())
	{
		// Permission management
		superAdmin.Get("/permissions", handler.GetAvailablePermissions)
		superAdmin.Put("/users/:id/permissions", handler.UpdateUserPermissions)
		superAdmin.Put("/users/:id/permissions/toggle", handler.ToggleUserPermission)
	}
}
