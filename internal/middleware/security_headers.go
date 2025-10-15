package middleware

import (
    "github.com/gofiber/fiber/v2"
)

// APISecurityHeaders adds security and caching headers for /api/* routes
func APISecurityHeaders() fiber.Handler {
    return func(c *fiber.Ctx) error {
        // Security headers
        c.Set("X-Content-Type-Options", "nosniff")
        c.Set("Referrer-Policy", "no-referrer")

        // Caching policy
        switch c.Method() {
        case fiber.MethodPost, fiber.MethodPut, fiber.MethodPatch, fiber.MethodDelete:
            // Do not cache mutating requests
            c.Set("Cache-Control", "no-store")
        default:
            // For GET: do not override; respect any route-specific Cache-Control
        }

        return c.Next()
    }
}