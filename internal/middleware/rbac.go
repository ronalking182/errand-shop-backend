package middleware

import (
	"errandShop/internal/pkg/models"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gofiber/fiber/v2"
)

// Make sure you have these functions exported
func AdminOnly() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role := c.Locals("role")
		fmt.Printf("[RBAC DEBUG] AdminOnly - Role: %v (type: %T)\n", role, role)

		// Allow both admin and superadmin access
		if role != "admin" && role != "superadmin" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Admin access required",
			})
		}

		return c.Next()
	}
}

// RBACMiddleware checks if user has any of the specified roles
func RBACMiddleware(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRole := c.Locals("role")
		if userRole == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		roleStr, ok := userRole.(string)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid role type",
			})
		}

		// Check if user role matches any of the allowed roles
		for _, allowedRole := range roles {
			if roleStr == allowedRole {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Insufficient permissions",
		})
	}
}

func CustomerOnly() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role := c.Locals("role")

		// Allow customer, admin, and superadmin access
		if role != "customer" && role != "admin" && role != "superadmin" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Customer access required",
			})
		}

		return c.Next()
	}
}

// RequirePermission middleware checks if user has specific permission
func RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userInterface, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "User not found in context"})
			c.Abort()
			return
		}

		user, ok := userInterface.(*models.User)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Invalid user type"})
			c.Abort()
			return
		}

		if !user.HasPermission(permission) {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyPermission checks if user has any of the specified permissions
func RequireAnyPermission(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userInterface, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "User not found in context"})
			c.Abort()
			return
		}

		user, ok := userInterface.(*models.User)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Invalid user type"})
			c.Abort()
			return
		}

		for _, permission := range permissions {
			if user.HasPermission(permission) {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "Insufficient permissions"})
		c.Abort()
	}
}
