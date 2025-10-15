package middleware

import (
	"errandShop/config"
	"errandShop/internal/pkg/jwt"
	"errandShop/internal/presenter"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	jwtlib "github.com/golang-jwt/jwt/v5"
)

// JWTMiddleware validates JWT tokens
func JWTMiddleware(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Debug logging for POST requests
		if c.Method() == "POST" {
			fmt.Printf("DEBUG JWT - Method: %s, Path: %s\n", c.Method(), c.Path())
			fmt.Printf("DEBUG JWT - Authorization header: %s\n", c.Get("Authorization"))
			fmt.Printf("DEBUG JWT - Cookie token: %s\n", c.Cookies("token"))
		}
		
		token := extractToken(c)
		if token == "" {
			if c.Method() == "POST" {
				fmt.Printf("DEBUG JWT - No token found for POST request\n")
			}
			return presenter.Err(c, fiber.StatusUnauthorized, "Missing or invalid token")
		}

		claims, err := validateToken(token, cfg.JWTSecret)
		if err != nil {
			if c.Method() == "POST" {
				fmt.Printf("DEBUG JWT - Token validation failed for POST: %v\n", err)
			}
			return presenter.Err(c, fiber.StatusUnauthorized, "Invalid token")
		}

		// Set user info in context
		c.Locals("userID", claims.UserID)
		c.Locals("email", claims.Email)
		c.Locals("role", claims.Role)
		c.Locals("permissions", claims.Permissions)

		if c.Method() == "POST" {
			fmt.Printf("DEBUG JWT - POST request authenticated successfully, role: %s\n", claims.Role)
		}

		return c.Next()
	}
}

// OptionalJWTMiddleware validates JWT tokens but doesn't require them
func OptionalJWTMiddleware(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := extractToken(c)
		if token != "" {
			claims, err := validateToken(token, cfg.JWTSecret)
			if err == nil {
				c.Locals("userID", claims.UserID)
				c.Locals("email", claims.Email)
				c.Locals("role", claims.Role)
				c.Locals("permissions", claims.Permissions)
			}
		}
		return c.Next()
	}
}

// AdminMiddleware ensures user has admin role
func AdminMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role := c.Locals("role")
		fmt.Printf("DEBUG - Role from token: %v (type: %T)\n", role, role)
		if role == nil || (role != "admin" && role != "superadmin") {
			return presenter.Err(c, fiber.StatusForbidden, "Admin access required")
		}
		return c.Next()
	}
}

// PermissionMiddleware checks for specific permissions
func PermissionMiddleware(requiredPermission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		permissions := c.Locals("permissions")
		if permissions == nil {
			return presenter.Err(c, fiber.StatusForbidden, "Insufficient permissions")
		}

		permList, ok := permissions.([]string)
		if !ok {
			return presenter.Err(c, fiber.StatusForbidden, "Invalid permissions format")
		}

		for _, perm := range permList {
			if perm == requiredPermission || perm == "*" {
				return c.Next()
			}
		}

		return presenter.Err(c, fiber.StatusForbidden, "Insufficient permissions")
	}
}

// SuperAdminMiddleware ensures user has superadmin role
func SuperAdminMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role := c.Locals("role")
		if role == nil || role != "superadmin" {
			return presenter.Err(c, fiber.StatusForbidden, "Super admin access required")
		}
		return c.Next()
	}
}

func extractToken(c *fiber.Ctx) string {
	// Try Authorization header first
	auth := c.Get("Authorization")
	if auth != "" && strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}

	// Try cookie as fallback
	return c.Cookies("token")
}

func validateToken(tokenString, secret string) (*jwt.JWTClaims, error) {
	token, err := jwtlib.ParseWithClaims(tokenString, &jwt.JWTClaims{}, func(token *jwtlib.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*jwt.JWTClaims)
	if !ok || !token.Valid {
		return nil, jwtlib.ErrInvalidKey
	}

	return claims, nil
}
