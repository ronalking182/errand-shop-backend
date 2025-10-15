package auth



// Mobile App DTOs
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type RegisterRequest struct {
	Name      string `json:"name" validate:"required,min=2"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=6"`
	Phone     string `json:"phone" validate:"required,min=8"`
}

type AuthResponse struct {
	User                 UserResponse `json:"user"`
	Token                string       `json:"token"`
	RefreshToken         string       `json:"refreshToken"`
	RequirePasswordReset bool         `json:"requirePasswordReset,omitempty"` // Add this line
}

type UserResponse struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	FirstName   string   `json:"first_name"`
	LastName    string   `json:"last_name"`
	Email       string   `json:"email"`
	Phone       string   `json:"phone"`
	Avatar      *string  `json:"avatar"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
	Status      string   `json:"status"`
	CreatedAt   string   `json:"createdAt"`
}



// OTP DTOs
type VerifyEmailRequest struct {
	Code string `json:"code" validate:"required,len=6"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResendOTPRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
	Email       string `json:"email" validate:"required,email"`
	OTP         string `json:"otp" validate:"required,len=6"`
	NewPassword string `json:"newPassword" validate:"required,min=6"`
}

// Token DTOs
type RefreshTokenResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refreshToken"`
}



// Admin DTOs - ADD THESE
type CreateUserRequest struct {
	Name     string  `json:"name" validate:"required,min=2"`
	Email    string  `json:"email" validate:"required,email"`
	Password string  `json:"password" validate:"required,min=6"`
	Phone    *string `json:"phone,omitempty" validate:"omitempty,min=8"`
	Role     string  `json:"role" validate:"required,oneof=customer admin superadmin"`
	Status   string  `json:"status" validate:"required,oneof=active inactive suspended"`
}

type UpdateUserRequest struct {
	Name   *string `json:"name,omitempty" validate:"omitempty,min=2"`
	Email  *string `json:"email,omitempty" validate:"omitempty,email"`
	Phone  *string `json:"phone,omitempty" validate:"omitempty,min=8"`
	Role   *string `json:"role,omitempty" validate:"omitempty,oneof=customer admin superadmin"`
	Status *string `json:"status,omitempty" validate:"omitempty,oneof=active inactive suspended"`
}

type UpdatePermissionsRequest struct {
	Permissions []string `json:"permissions" validate:"required"`
}

type ForceResetRequest struct {
	ForceReset bool `json:"forceReset"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword" validate:"required,min=6"`
	NewPassword     string `json:"newPassword" validate:"required,min=6"`
}
