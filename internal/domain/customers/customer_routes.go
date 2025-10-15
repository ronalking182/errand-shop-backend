package customers

import (
	"errandShop/config"
	"errandShop/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app fiber.Router, handler *Handler, cfg *config.Config) {
	// Customer routes (authenticated users)
	customerRoutes := app.Group("/customers")
	customerRoutes.Use(middleware.JWTMiddleware(cfg))
	{
		// Profile management
		customerRoutes.Post("", handler.CreateCustomer)
		customerRoutes.Get("/profile", handler.GetCustomerProfile)
		customerRoutes.Put("/profile", handler.UpdateCustomerProfile)

		// Address management
		customerRoutes.Post("/addresses", handler.CreateAddress)
		customerRoutes.Get("/addresses", handler.GetCustomerAddresses)
		customerRoutes.Put("/addresses/:id", handler.UpdateAddress)
		customerRoutes.Delete("/addresses/:id", handler.DeleteAddress)
		customerRoutes.Put("/addresses/:id/default", handler.SetDefaultAddress)
	}

	// Admin routes (admin only)
	adminRoutes := app.Group("/admin/customers")
	adminRoutes.Use(middleware.JWTMiddleware(cfg), middleware.AdminMiddleware())
	{
		adminRoutes.Get("", handler.ListCustomers)
		adminRoutes.Get("/:id", handler.GetCustomerByID)
		
		// Admin address management
		adminRoutes.Post("/:id/addresses", handler.CreateCustomerAddress)
		adminRoutes.Get("/:id/addresses", handler.GetCustomerAddressesAdmin)
	}
}
