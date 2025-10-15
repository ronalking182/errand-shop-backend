package users

import "time"

// Request DTOs
type CreateUserRequest struct {
	Name        string   `json:"name" binding:"required"`
	Email       string   `json:"email" binding:"required,email"`
	Phone       string   `json:"phone"`
	Password    string   `json:"password" binding:"required,min=8"`
	Role        string   `json:"role" binding:"required,oneof=admin superadmin customer"`
	Permissions []string `json:"permissions"`
}

type UpdateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email" binding:"omitempty,email"`
	Phone string `json:"phone"`
	Role  string `json:"role" binding:"omitempty,oneof=admin superadmin customer"`
}

type UpdateProfileRequest struct {
	Name   string  `json:"name" validate:"omitempty,min=2"`
	Email  string  `json:"email" validate:"omitempty,email"`
	Phone  *string `json:"phone" validate:"omitempty,min=8"`
	Gender *string `json:"gender" validate:"omitempty"`
	Avatar *string `json:"avatar" validate:"omitempty"`
}

type UpdatePasswordRequest struct {
	CurrentPassword string `json:"currentPassword" binding:"required"`
	NewPassword     string `json:"newPassword" binding:"required,min=8"`
}

type UpdatePermissionsRequest struct {
	Permissions []string `json:"permissions" binding:"required"`
}

type TogglePermissionRequest struct {
	Permission string `json:"permission" binding:"required"`
	Grant      bool   `json:"grant"`
}

type ToggleUserStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=active inactive suspended"`
}

// Response DTOs
type UserResponse struct {
	ID          uint       `json:"id"`
	Name        string     `json:"name"`
	Email       string     `json:"email"`
	Phone       *string    `json:"phone"`
	Gender      *string    `json:"gender"`
	Avatar      *string    `json:"avatar"`
	Role        string     `json:"role"`
	Permissions []string   `json:"permissions"`
	Status      string     `json:"status"`
	IsVerified  bool       `json:"isVerified"`
	LastLoginAt *time.Time `json:"lastLoginAt"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

type UserListResponse struct {
	Users      []UserResponse `json:"users"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	Limit      int            `json:"limit"`
	TotalPages int            `json:"totalPages"`
}

type PermissionResponse struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
}

// Helper function to convert User to UserResponse
func ToUserResponse(user *User) UserResponse {
	return UserResponse{
		ID:          user.ID,
		Name:        user.Name,
		Email:       user.Email,
		Phone:       user.Phone,
		Gender:      user.Gender,
		Avatar:      user.Avatar,
		Role:        user.Role,
		Permissions: user.Permissions,
		Status:      user.Status,
		IsVerified:  user.IsVerified,
		LastLoginAt: user.LastLoginAt,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}
}

// Helper function to convert multiple Users to UserResponses
func ToUserResponses(users []User) []UserResponse {
	responses := make([]UserResponse, len(users))
	for i, user := range users {
		responses[i] = ToUserResponse(&user)
	}
	return responses
}
