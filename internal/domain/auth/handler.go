package auth

import (
	"errandShop/internal/presenter"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Handler struct {
	Service   *Service
	Validator *validator.Validate
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		Service:   service,
		Validator: validator.New(),
	}
}

// Register handles user registration
func (h *Handler) Register(c *fiber.Ctx) error {
	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return presenter.Err(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if err := h.Validator.Struct(req); err != nil {
		return presenter.Err(c, fiber.StatusBadRequest, "Validation failed: "+err.Error())
	}

	response, err := h.Service.Register(c.Context(), req)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return presenter.Err(c, fiber.StatusConflict, err.Error())
		}
		return presenter.Err(c, fiber.StatusInternalServerError, "Failed to register user")
	}

	return presenter.Created(c, response)
}

// Login handles user authentication
func (h *Handler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return presenter.Err(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if err := h.Validator.Struct(req); err != nil {
		return presenter.Err(c, fiber.StatusBadRequest, "Validation failed: "+err.Error())
	}

	response, err := h.Service.Login(c.Context(), req)
	if err != nil {
		if strings.Contains(err.Error(), "invalid credentials") || strings.Contains(err.Error(), "not found") {
			return presenter.Err(c, fiber.StatusUnauthorized, "Invalid credentials")
		}
		if strings.Contains(err.Error(), "not verified") {
			return presenter.Err(c, fiber.StatusForbidden, "Email not verified")
		}
		return presenter.Err(c, fiber.StatusInternalServerError, "Login failed")
	}

	return presenter.OK(c, response, nil)
}

// VerifyEmail handles email verification
func (h *Handler) VerifyEmail(c *fiber.Ctx) error {
	var req VerifyEmailRequest
	if err := c.BodyParser(&req); err != nil {
		return presenter.Err(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if err := h.Validator.Struct(req); err != nil {
		return presenter.Err(c, fiber.StatusBadRequest, "Validation failed: "+err.Error())
	}

	response, err := h.Service.VerifyEmailWithCode(c.Context(), req.Code)
	if err != nil {
		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "expired") {
			return presenter.Err(c, fiber.StatusBadRequest, err.Error())
		}
		return presenter.Err(c, fiber.StatusInternalServerError, "Email verification failed")
	}

	return presenter.OK(c, response, nil)
}

// RefreshToken handles token refresh
func (h *Handler) RefreshToken(c *fiber.Ctx) error {
	refreshToken := c.Get("Authorization")
	if refreshToken == "" {
		return presenter.Err(c, fiber.StatusBadRequest, "Refresh token required")
	}

	// Remove "Bearer " prefix if present
	if strings.HasPrefix(refreshToken, "Bearer ") {
		refreshToken = strings.TrimPrefix(refreshToken, "Bearer ")
	}

	response, err := h.Service.RefreshToken(c.Context(), refreshToken)
	if err != nil {
		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "expired") {
			return presenter.Err(c, fiber.StatusUnauthorized, "Invalid or expired refresh token")
		}
		return presenter.Err(c, fiber.StatusInternalServerError, "Token refresh failed")
	}

	return presenter.OK(c, response, nil)
}

// Logout handles user logout
func (h *Handler) Logout(c *fiber.Ctx) error {
	userID := c.Locals("userID")
	if userID == nil {
		return presenter.Err(c, fiber.StatusUnauthorized, "User not authenticated")
	}

	userIDUUID, ok := userID.(uuid.UUID)
	if !ok {
		return presenter.Err(c, fiber.StatusInternalServerError, "Invalid user ID")
	}

	err := h.Service.Logout(c.Context(), userIDUUID)
	if err != nil {
		return presenter.Err(c, fiber.StatusInternalServerError, "Logout failed")
	}

	return presenter.OK(c, fiber.Map{"message": "Logged out successfully"}, nil)
}

// ForgotPassword handles password reset request
func (h *Handler) ForgotPassword(c *fiber.Ctx) error {
	var req ForgotPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return presenter.Err(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if err := h.Validator.Struct(req); err != nil {
		return presenter.Err(c, fiber.StatusBadRequest, "Validation failed: "+err.Error())
	}

	err := h.Service.ForgotPassword(c.Context(), req.Email)
	if err != nil {
		return presenter.Err(c, fiber.StatusInternalServerError, "Failed to send reset email")
	}

	return presenter.OK(c, fiber.Map{"message": "Password reset email sent"}, nil)
}

// ResetPassword handles password reset
func (h *Handler) ResetPassword(c *fiber.Ctx) error {
	var req ResetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return presenter.Err(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if err := h.Validator.Struct(req); err != nil {
		return presenter.Err(c, fiber.StatusBadRequest, "Validation failed: "+err.Error())
	}

	err := h.Service.ResetPassword(c.Context(), req.Email, req.OTP, req.NewPassword)
	if err != nil {
		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "expired") {
			return presenter.Err(c, fiber.StatusBadRequest, err.Error())
		}
		return presenter.Err(c, fiber.StatusInternalServerError, "Password reset failed")
	}

	return presenter.OK(c, fiber.Map{"message": "Password reset successfully"}, nil)
}

// Me handles getting current user info
func (h *Handler) Me(c *fiber.Ctx) error {
	userID := c.Locals("userID")
	if userID == nil {
		return presenter.Err(c, fiber.StatusUnauthorized, "User not authenticated")
	}

	userIDUUID, ok := userID.(uuid.UUID)
	if !ok {
		return presenter.Err(c, fiber.StatusInternalServerError, "Invalid user ID")
	}

	user, err := h.Service.GetUserByID(c.Context(), userIDUUID)
	if err != nil {
		return presenter.Err(c, fiber.StatusInternalServerError, "Failed to get user info")
	}

	return presenter.OK(c, user, nil)
}

// Admin endpoints

// GetUsers handles getting all users (admin only)
func (h *Handler) GetUsers(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	search := c.Query("search", "")
	role := c.Query("role", "")
	status := c.Query("status", "")

	users, total, err := h.Service.GetUsers(c.Context(), page, limit, search, role, status)
	if err != nil {
		return presenter.Err(c, fiber.StatusInternalServerError, "Failed to get users")
	}

	// Calculate pagination
	totalPages := int((total + int64(limit) - 1) / int64(limit))
	pagination := &presenter.PageMeta{
		Page:       page,
		Limit:      limit,
		Total:      int(total),
		TotalPages: totalPages,
	}

	return presenter.OK(c, fiber.Map{"users": users}, pagination)
}

// UpdateUserStatus handles updating user status (admin only)
func (h *Handler) UpdateUserStatus(c *fiber.Ctx) error {
	userIDStr := c.Params("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return presenter.Err(c, fiber.StatusBadRequest, "Invalid user ID")
	}

	var req struct {
		Status string `json:"status" validate:"required,oneof=active inactive suspended"`
	}

	if err := c.BodyParser(&req); err != nil {
		return presenter.Err(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if err := h.Validator.Struct(&req); err != nil {
		return presenter.Err(c, fiber.StatusBadRequest, err.Error())
	}

	err = h.Service.UpdateUserStatus(c.Context(), userID, req.Status)
	if err != nil {
		return presenter.Err(c, fiber.StatusInternalServerError, "Failed to update user status")
	}

	return presenter.OK(c, fiber.Map{"message": "User status updated successfully"}, nil)
}

// Add these methods to the Handler struct

// CreateUser handles creating a new user (admin only)
func (h *Handler) CreateUser(c *fiber.Ctx) error {
	var req CreateUserRequest // Changed from RegisterRequest to CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return presenter.Err(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if err := h.Validator.Struct(req); err != nil {
		return presenter.Err(c, fiber.StatusBadRequest, "Validation failed: "+err.Error())
	}

	user, err := h.Service.CreateUser(c.Context(), req)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return presenter.Err(c, fiber.StatusConflict, err.Error())
		}
		return presenter.Err(c, fiber.StatusInternalServerError, "Failed to create user")
	}

	return presenter.Created(c, user)
}

// GetUserByID handles getting a specific user by ID (admin only)
func (h *Handler) GetUserByID(c *fiber.Ctx) error {
	userIDStr := c.Params("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return presenter.Err(c, fiber.StatusBadRequest, "Invalid user ID")
	}

	user, err := h.Service.GetUserByID(c.Context(), userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return presenter.Err(c, fiber.StatusNotFound, "User not found")
		}
		return presenter.Err(c, fiber.StatusInternalServerError, "Failed to get user")
	}

	return presenter.OK(c, user, nil)
}

// UpdateUser handles updating user details (admin only)
func (h *Handler) UpdateUser(c *fiber.Ctx) error {
	userIDStr := c.Params("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return presenter.Err(c, fiber.StatusBadRequest, "Invalid user ID")
	}

	var req UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return presenter.Err(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if err := h.Validator.Struct(&req); err != nil {
		return presenter.Err(c, fiber.StatusBadRequest, err.Error())
	}

	user, err := h.Service.UpdateUser(c.Context(), userID, req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return presenter.Err(c, fiber.StatusNotFound, "User not found")
		}
		if strings.Contains(err.Error(), "phone number already exists") {
			return presenter.Err(c, fiber.StatusConflict, "Phone number is already in use by another user")
		}
		if strings.Contains(err.Error(), "email already exists") {
			return presenter.Err(c, fiber.StatusConflict, "Email is already in use by another user")
		}
		return presenter.Err(c, fiber.StatusInternalServerError, "Failed to update user")
	}

	return presenter.OK(c, user, nil)
}

// DeleteUser handles deleting a user (admin only)
func (h *Handler) DeleteUser(c *fiber.Ctx) error {
	userIDStr := c.Params("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return presenter.Err(c, fiber.StatusBadRequest, "Invalid user ID")
	}

	err = h.Service.DeleteUser(c.Context(), userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return presenter.Err(c, fiber.StatusNotFound, "User not found")
		}
		return presenter.Err(c, fiber.StatusInternalServerError, "Failed to delete user")
	}

	return presenter.OK(c, fiber.Map{"message": "User deleted successfully"}, nil)
}

// GetAvailablePermissions returns all available permissions (admin only)
func (h *Handler) GetAvailablePermissions(c *fiber.Ctx) error {
	permissions := []string{
		"products:read",
		"products:write",
		"products:delete",
		"orders:read",
		"orders:write",
		"orders:cancel",
		"chat:read",
		"chat:write",
		"coupons:read",
		"coupons:create",
		"reports:read",
		"users:read",
		"users:write",
		"users:delete",
	}

	return presenter.OK(c, fiber.Map{"permissions": permissions}, nil)
}

// UpdateUserPermissions handles updating user permissions (admin only)
func (h *Handler) UpdateUserPermissions(c *fiber.Ctx) error {
	userIDStr := c.Params("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return presenter.Err(c, fiber.StatusBadRequest, "Invalid user ID")
	}

	var req UpdatePermissionsRequest
	if err := c.BodyParser(&req); err != nil {
		return presenter.Err(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if err := h.Validator.Struct(&req); err != nil {
		return presenter.Err(c, fiber.StatusBadRequest, err.Error())
	}

	err = h.Service.UpdateUserPermissions(c.Context(), userID, req.Permissions)
	if err != nil {
		return presenter.Err(c, fiber.StatusInternalServerError, "Failed to update permissions")
	}

	return presenter.OK(c, fiber.Map{"message": "User permissions updated successfully"}, nil)
}

// ForcePasswordReset handles forcing password reset (admin only)
func (h *Handler) ForcePasswordReset(c *fiber.Ctx) error {
	userIDStr := c.Params("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return presenter.Err(c, fiber.StatusBadRequest, "Invalid user ID")
	}

	err = h.Service.ForcePasswordReset(c.Context(), userID)
	if err != nil {
		return presenter.Err(c, fiber.StatusInternalServerError, "Failed to force password reset")
	}

	return presenter.OK(c, fiber.Map{"message": "Password reset forced successfully"}, nil)
}

// ChangePassword handles password change for authenticated users
func (h *Handler) ChangePassword(c *fiber.Ctx) error {
	userID := c.Locals("userID")
	if userID == nil {
		return presenter.Err(c, fiber.StatusUnauthorized, "User not authenticated")
	}

	userIDUUID, ok := userID.(uuid.UUID)
	if !ok {
		return presenter.Err(c, fiber.StatusInternalServerError, "Invalid user ID")
	}

	var req ChangePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return presenter.Err(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if err := h.Validator.Struct(&req); err != nil {
		return presenter.Err(c, fiber.StatusBadRequest, err.Error())
	}

	err := h.Service.ChangePassword(c.Context(), userIDUUID, req.CurrentPassword, req.NewPassword)
	if err != nil {
		if strings.Contains(err.Error(), "incorrect") {
			return presenter.Err(c, fiber.StatusBadRequest, "Current password is incorrect")
		}
		return presenter.Err(c, fiber.StatusInternalServerError, "Failed to change password")
	}

	return presenter.OK(c, fiber.Map{"message": "Password changed successfully"}, nil)
}

// ResendOTP handles resending OTP for email verification
func (h *Handler) ResendOTP(c *fiber.Ctx) error {
	var req ResendOTPRequest
	if err := c.BodyParser(&req); err != nil {
		return presenter.Err(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if err := h.Validator.Struct(&req); err != nil {
		return presenter.Err(c, fiber.StatusBadRequest, err.Error())
	}

	err := h.Service.ResendOTP(c.Context(), req.Email)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return presenter.Err(c, fiber.StatusNotFound, "User not found")
		}
		if strings.Contains(err.Error(), "already verified") {
			return presenter.Err(c, fiber.StatusBadRequest, "Email already verified")
		}
		return presenter.Err(c, fiber.StatusInternalServerError, "Failed to resend OTP")
	}

	return presenter.OK(c, fiber.Map{"message": "OTP sent successfully"}, nil)
}
