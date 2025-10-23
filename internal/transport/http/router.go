package http

import (
	"errandShop/internal/domain/auth"
	"errandShop/internal/domain/products"
	"errandShop/internal/middleware"
	v1 "errandShop/internal/transport/http/v1"

	"github.com/gofiber/fiber/v2"
)

type Deps struct {
	Auth       *auth.Handler
	Products   *products.Handler
	JWT        fiber.Handler
	AdminOnly  fiber.Handler
	CustOnly   fiber.Handler
	SuperAdmin fiber.Handler
}

func Build(app *fiber.App, d *Deps) {
	api := app.Group("/api")
	v := api.Group("/v1")

	v.Get("/system/health", func(c *fiber.Ctx) error { return c.SendString("ok") })

	ag := v.Group("/auth")
	ag.Post("/register", d.Auth.Register)
	ag.Post("/login", d.Auth.Login)
	ag.Post("/refresh", d.Auth.RefreshToken)
	ag.Post("/logout", d.JWT, d.Auth.Logout)
	ag.Get("/me", d.JWT, d.Auth.Me)

	// Public routes
	v1.MountProductRoutes(v, d.Products)

	// Admin group
	admin := v.Group("/admin", d.JWT, d.AdminOnly)
	v1.MountAdminProductRoutes(admin, d.Products)

	// Superadmin-only subgroup under admin for category management
	superAdmin := admin.Group("", d.SuperAdmin)
	v1.MountSuperAdminCategoryRoutes(superAdmin, d.Products)

	_ = v.Group("/mobile", d.JWT, d.CustOnly)
}

func DefaultDeps(authH *auth.Handler, prodH *products.Handler) *Deps {
	return &Deps{
		Auth:       authH,
		Products:   prodH,
		JWT:        middleware.JWTMiddleware(nil), // TODO: Pass proper config
		AdminOnly:  middleware.AdminOnly(),
		CustOnly:   middleware.CustomerOnly(),
		SuperAdmin: middleware.SuperAdminMiddleware(),
	}
}
