package users

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Handler struct {
	service *Service
}

// NewHandler creates a new user handler
func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// Helper function to convert UUID to uint for database compatibility
func (h *Handler) uuidToUint(userUUID uuid.UUID) uint {
	// Convert UUID to string and then hash to uint
	// This is a simple conversion - in production you might want a more sophisticated mapping
	uuidStr := userUUID.String()
	hash := uint(0)
	for _, char := range uuidStr {
		hash = hash*31 + uint(char)
	}
	// Ensure it's a reasonable uint value
	return hash % 1000000 // Limit to 6 digits
}

// GetAvailablePermissions returns all available permissions
func (h *Handler) GetAvailablePermissions(c *fiber.Ctx) error {
	permissions := GetAvailablePermissions()
	return c.JSON(fiber.Map{
		"permissions": permissions,
	})
}

// UpdateUserPermissions updates user permissions (superadmin only)
func (h *Handler) UpdateUserPermissions(c *fiber.Ctx) error {
	userID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid user ID"})
	}

	var req UpdatePermissionsRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	if err := h.service.UpdateUserPermissions(uint(userID), req.Permissions); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": "Failed to update permissions"})
	}

	return c.JSON(fiber.Map{"success": true, "message": "Permissions updated successfully"})
}

// ToggleUserPermission toggles a specific permission
func (h *Handler) ToggleUserPermission(c *fiber.Ctx) error {
	userID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid user ID"})
	}

	var req TogglePermissionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	if err := h.service.ToggleUserPermission(uint(userID), req.Permission, req.Grant); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": "Failed to toggle permission"})
	}

	action := "revoked"
	if req.Grant {
		action = "granted"
	}

	return c.JSON(fiber.Map{"success": true, "message": fmt.Sprintf("Permission %s successfully", action)})
}

// CreateUser creates a new user
func (h *Handler) CreateUser(c *fiber.Ctx) error {
	var req CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	user, err := h.service.CreateUser(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": "Failed to create user"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"success": true, "data": ToUserResponse(user)})
}

// GetProfile gets current user profile
func (h *Handler) GetProfile(c *fiber.Ctx) error {
	// Get user ID from JWT token
	userIDRaw := c.Locals("userID")
	if userIDRaw == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "User not authenticated",
		})
	}

	// Handle different types from JWT claims
	var userIDUUID uuid.UUID
	switch v := userIDRaw.(type) {
	case uuid.UUID:
		userIDUUID = v
	case string:
		parsedUUID, err := uuid.Parse(v)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"error":   "Invalid user ID format",
			})
		}
		userIDUUID = parsedUUID
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid user ID type",
		})
	}

	// Convert UUID to uint for database lookup
	userID := h.uuidToUint(userIDUUID)
	
	user, err := h.service.repo.GetByID(userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "User not found",
		})
	}
	
	// Create response
	response := map[string]interface{}{
		"id":          user.ID,
		"name":        user.Name,
		"email":       user.Email,
		"phone":       user.Phone,
		"gender":      user.Gender,
		"avatar":      user.Avatar,
		"role":        user.Role,
		"permissions": user.Permissions,
		"status":      user.Status,
		"isVerified":  user.IsVerified,
		"lastLoginAt": user.LastLoginAt,
		"createdAt":   user.CreatedAt,
		"updatedAt":   user.UpdatedAt,
	}
	
	return c.JSON(fiber.Map{
		"success": true,
		"data":    response,
	})
}

// UpdateProfile updates current user profile
func (h *Handler) UpdateProfile(c *fiber.Ctx) error {
	// Get user ID from JWT token
	userIDRaw := c.Locals("userID")
	if userIDRaw == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "User not authenticated",
		})
	}

	// Handle different types from JWT claims
	var userIDUUID uuid.UUID
	switch v := userIDRaw.(type) {
	case uuid.UUID:
		userIDUUID = v
	case string:
		parsedUUID, err := uuid.Parse(v)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"error":   "Invalid user ID format",
			})
		}
		userIDUUID = parsedUUID
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid user ID type",
		})
	}

	// Convert UUID to uint for database lookup
	userID := h.uuidToUint(userIDUUID)
	
	var req UpdateProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}
	
	// Get current user
	user, err := h.service.repo.GetByID(userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "User not found",
		})
	}
	
	// Update fields if provided
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Phone != nil {
		user.Phone = req.Phone
	}
	if req.Gender != nil {
		user.Gender = req.Gender
	}
	if req.Avatar != nil {
		user.Avatar = req.Avatar
	}
	
	// Save updated user
	if err := h.service.repo.Update(user); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to update profile",
		})
	}
	
	// Create response
	response := map[string]interface{}{
		"id":          user.ID,
		"name":        user.Name,
		"email":       user.Email,
		"phone":       user.Phone,
		"gender":      user.Gender,
		"avatar":      user.Avatar,
		"role":        user.Role,
		"permissions": user.Permissions,
		"status":      user.Status,
		"isVerified":  user.IsVerified,
		"lastLoginAt": user.LastLoginAt,
		"createdAt":   user.CreatedAt,
		"updatedAt":   user.UpdatedAt,
	}
	
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Profile updated successfully",
		"data":    response,
	})
}

// UpdatePassword updates current user password
func (h *Handler) UpdatePassword(c *fiber.Ctx) error {
	// TODO: Implement password update
	return c.JSON(fiber.Map{"success": true, "message": "Password update endpoint"})
}

// GetUsers retrieves all users (admin only)
func (h *Handler) GetUsers(c *fiber.Ctx) error {
	// TODO: Implement get all users logic
	return c.JSON(fiber.Map{"success": true, "message": "Get users not implemented"})
}

// GetUser retrieves a specific user by ID (admin only)
func (h *Handler) GetUser(c *fiber.Ctx) error {
	// TODO: Implement get user by ID logic
	return c.JSON(fiber.Map{"success": true, "message": "Get user not implemented"})
}

// UpdateUser updates a user (admin only)
func (h *Handler) UpdateUser(c *fiber.Ctx) error {
	// TODO: Implement update user logic
	return c.JSON(fiber.Map{"success": true, "message": "Update user not implemented"})
}

// DeleteUser deletes a user (admin only)
func (h *Handler) DeleteUser(c *fiber.Ctx) error {
	// TODO: Implement delete user logic
	return c.JSON(fiber.Map{"success": true, "message": "Delete user not implemented"})
}

// ToggleUserStatus toggles user active/inactive status (admin only)
func (h *Handler) ToggleUserStatus(c *fiber.Ctx) error {
	// TODO: Implement toggle user status logic
	return c.JSON(fiber.Map{"success": true, "message": "Toggle user status not implemented"})
}

// ForcePasswordReset forces a password reset for a user (admin only)
func (h *Handler) ForcePasswordReset(c *fiber.Ctx) error {
	// TODO: Implement force password reset logic
	return c.JSON(fiber.Map{"success": true, "message": "Force password reset not implemented"})
}
