package auth

import (
	"context"
	"crypto/rand"
	"errandShop/config"
	"errandShop/internal/domain/customers"
	"errandShop/internal/services/audit"
	"errandShop/internal/services/email"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// CustomerService interface for creating customer profiles
type CustomerService interface {
	CreateCustomer(req interface{}) (interface{}, error)
}

// Add to Service struct
type Service struct {
	Repo            *Repository
	Cfg             *config.Config
	JWTService      *JWTService
	EmailService    *email.ResendService
	AuditService    *audit.AuditService
	CustomerService customers.Service
}

// Update NewService function
func NewService(repo *Repository, cfg *config.Config, emailService *email.ResendService, auditService *audit.AuditService, customerService customers.Service) *Service {
	return &Service{
		Repo:            repo,
		Cfg:             cfg,
		JWTService:      NewJWTService(cfg.JWTSecret),
		EmailService:    emailService,
		AuditService:    auditService,
		CustomerService: customerService,
	}
}

// Update Register method to include permissions and email
func (s *Service) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	// Normalize email and phone to prevent duplicates
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.Phone = strings.TrimSpace(req.Phone)

	// Check if user already exists
	existingUser, _ := s.Repo.GetByEmail(ctx, req.Email)
	if existingUser != nil {
		return nil, fmt.Errorf("email already exists")
	}

	// Handle name fields - use provided first/last names or split the name field
	firstName := req.FirstName
	lastName := req.LastName
	fullName := req.Name

	if firstName == "" || lastName == "" {
		// Split the name field if first/last names are not provided
		nameParts := strings.Fields(req.Name)
		if len(nameParts) >= 2 {
			firstName = nameParts[0]
			lastName = strings.Join(nameParts[1:], " ")
		} else if len(nameParts) == 1 {
			firstName = nameParts[0]
			lastName = ""
		}
		fullName = req.Name
	} else {
		// Use provided first/last names
		fullName = firstName + " " + lastName
	}

	// Create user
	user := &User{
		Name:      fullName,
		FirstName: firstName,
		LastName:  lastName,
		Email:     req.Email,
		Password:  req.Password,
		Phone:     req.Phone,
		Role:      "customer",
	}

	if err := s.Repo.Create(ctx, user); err != nil {
		return nil, err
	}

	// Create customer profile using registration data
	if s.CustomerService != nil {
		// Create customer profile request using interface{} to avoid circular import
		customerReq := map[string]interface{}{
			"user_id":    user.ID,
			"first_name": firstName,
			"last_name":  lastName,
			"phone":      req.Phone,
		}

		// Create customer profile (don't fail registration if this fails)
		if _, err := s.CustomerService.CreateCustomer(customerReq); err != nil {
			// Log the error but don't fail the registration
			fmt.Printf("Warning: Failed to create customer profile for user %d: %v\n", user.ID, err)
		}
	}

	// Generate tokens with permissions
	token, err := s.JWTService.GenerateToken(user)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.JWTService.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	// Save refresh token
	refreshTokenRecord := &RefreshToken{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	s.Repo.CreateRefreshToken(ctx, refreshTokenRecord)

	// Send verification email
	if err := s.SendVerificationEmail(ctx, user.ID); err != nil {
		fmt.Printf("Warning: Failed to send verification email to %s: %v\n", user.Email, err)
	}

	// Send welcome email
	go s.EmailService.SendWelcomeEmail(context.Background(), user.Email, user.Name)

	// Audit log
	s.AuditService.LogUserAction(ctx, user.ID, "user_registered", "user", map[string]interface{}{
		"email": user.Email,
		"name":  user.Name,
	}, "", "")

	return &AuthResponse{
		User:         s.toUserResponse(user),
		Token:        token,
		RefreshToken: refreshToken,
	}, nil
}

// Update ForgotPassword to use Resend
func (s *Service) ForgotPassword(ctx context.Context, email string) error {
	user, err := s.Repo.GetByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	// Generate OTP
	otp, err := s.generateOTP()
	if err != nil {
		return err
	}

	// Save OTP
	otpRecord := &OTP{
		UserID:    user.ID,
		Email:     email,
		Code:      otp,
		Type:      "password_reset",
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}

	if err := s.Repo.SaveOTP(ctx, otpRecord); err != nil {
		return err
	}

	// Send email using Resend
	if err := s.EmailService.SendOTPEmail(ctx, email, otp, "password_reset"); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	// Audit log
	s.AuditService.LogUserAction(ctx, user.ID, "password_reset_requested", "user", map[string]interface{}{
		"email": email,
	}, "", "")

	return nil
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	user, err := s.Repo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	if !user.CheckPassword(req.Password) {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Check if password reset is required
	if user.ForceReset {
		// Generate a limited token for password reset only
		token, err := s.JWTService.GenerateToken(user)
		if err != nil {
			return nil, err
		}

		return &AuthResponse{
			User:                 s.toUserResponse(user),
			Token:                token,
			RequirePasswordReset: true,
		}, nil
	}

	// Update last login
	now := time.Now()
	user.LastLoginAt = &now
	if err := s.Repo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Generate token
	token, err := s.JWTService.GenerateToken(user)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		User:  s.toUserResponse(user),
		Token: token,
	}, nil
}

func (s *Service) SendVerificationEmail(ctx context.Context, userID uuid.UUID) error {
	user, err := s.Repo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	// Generate 6-digit OTP
	code, err := s.generateOTP()
	if err != nil {
		return err
	}

	// Save OTP
	otp := &OTP{
		UserID:    userID,
		Email:     user.Email,
		Code:      code,
		Type:      "email_verification",
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}

	if err := s.Repo.SaveOTP(ctx, otp); err != nil {
		return err
	}

	// Send OTP via email
	if err := s.EmailService.SendOTPEmail(ctx, user.Email, code, "email_verification"); err != nil {
		return fmt.Errorf("failed to send verification email: %w", err)
	}

	return nil
}

func (s *Service) VerifyEmail(ctx context.Context, userID uuid.UUID, code string) error {
	otp, err := s.Repo.GetOTPByUserIDAndCode(ctx, userID, code)
	if err != nil {
		return fmt.Errorf("invalid or expired code")
	}

	// Mark OTP as used
	otp.Used = true
	s.Repo.UpdateOTP(ctx, otp)

	// Mark user as verified
	user, err := s.Repo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	user.IsVerified = true
	return s.Repo.Update(ctx, user)
}

func (s *Service) generateOTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

// Add missing methods:

// VerifyEmailWithCode verifies email using only the OTP code
func (s *Service) VerifyEmailWithCode(ctx context.Context, code string) (*AuthResponse, error) {
	// Find OTP record by code
	otpRecord, err := s.Repo.GetValidOTPByCode(ctx, code, "email_verification")
	if err != nil {
		return nil, err
	}

	// Get user by email from OTP record
	user, err := s.Repo.GetByEmail(ctx, otpRecord.Email)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Mark OTP as used
	otpRecord.Used = true
	if err := s.Repo.UpdateOTP(ctx, otpRecord); err != nil {
		return nil, err
	}

	// Mark user as verified and set status to active
	user.IsVerified = true
	user.Status = "active"
	if err := s.Repo.Update(ctx, user); err != nil {
		return nil, err
	}

	// Generate JWT tokens
	accessToken, err := s.JWTService.GenerateToken(user)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.JWTService.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	// Store refresh token
	if err := s.Repo.StoreRefreshToken(ctx, user.ID, refreshToken); err != nil {
		return nil, err
	}

	// Log audit event
	if s.AuditService != nil {
		s.AuditService.LogUserAction(ctx, user.ID, "email_verified", "user", map[string]interface{}{
			"email": user.Email,
		}, "", "")
	}

	return &AuthResponse{
		User:         s.toUserResponse(user),
		Token:        accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// VerifyEmail with email and OTP parameters (kept for backward compatibility)
func (s *Service) VerifyEmailWithOTP(ctx context.Context, email, otp string) (*AuthResponse, error) {
	// Get user by email
	user, err := s.Repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Validate OTP
	otpRecord, err := s.Repo.GetValidOTP(ctx, email, otp, "email_verification")
	if err != nil {
		return nil, err
	}

	// Mark OTP as used
	otpRecord.Used = true
	if err := s.Repo.UpdateOTP(ctx, otpRecord); err != nil {
		return nil, err
	}

	// Mark user as verified and set status to active
	user.IsVerified = true
	user.Status = "active"
	if err := s.Repo.Update(ctx, user); err != nil {
		return nil, err
	}

	// Generate JWT tokens
	accessToken, err := s.JWTService.GenerateToken(user)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.JWTService.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	// Store refresh token
	if err := s.Repo.StoreRefreshToken(ctx, user.ID, refreshToken); err != nil {
		return nil, err
	}

	// Log audit event
	if s.AuditService != nil {
		s.AuditService.LogUserAction(ctx, user.ID, "email_verified", "user", map[string]interface{}{
			"email": user.Email,
		}, "", "")
	}

	return &AuthResponse{
		User:         s.toUserResponse(user),
		Token:        accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// RefreshToken method
func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (*RefreshTokenResponse, error) {
	// Validate refresh token
	tokenRecord, err := s.Repo.GetRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, err
	}

	// Get user
	user, err := s.Repo.GetByID(ctx, tokenRecord.UserID)
	if err != nil {
		return nil, err
	}

	// Generate new tokens
	newToken, err := s.JWTService.GenerateToken(user)
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := s.JWTService.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	// Delete old refresh token
	s.Repo.DeleteRefreshToken(ctx, refreshToken)

	// Save new refresh token
	refreshTokenRecord := &RefreshToken{
		UserID:    user.ID,
		Token:     newRefreshToken,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	s.Repo.CreateRefreshToken(ctx, refreshTokenRecord)

	return &RefreshTokenResponse{
		Token:        newToken,
		RefreshToken: newRefreshToken,
	}, nil
}

// Logout method
func (s *Service) Logout(ctx context.Context, userID uuid.UUID) error {
	return s.Repo.DeleteUserRefreshTokens(ctx, userID)
}

// ForgotPassword method
func (s *Service) RequestPasswordReset(ctx context.Context, email string) error {
	user, err := s.Repo.GetByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	// Generate OTP
	otp, err := s.generateOTP()
	if err != nil {
		return err
	}

	// Save OTP
	otpRecord := &OTP{
		UserID:    user.ID,
		Email:     email,
		Code:      otp,
		Type:      "password_reset",
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}

	if err := s.Repo.SaveOTP(ctx, otpRecord); err != nil {
		return err
	}

	// TODO: Send email using Resend
	fmt.Printf("Password reset OTP for %s: %s\n", email, otp)

	return nil
}

// ResetPassword method
func (s *Service) ResetPassword(ctx context.Context, email, otp, newPassword string) error {
	// Validate OTP
	otpRecord, err := s.Repo.GetValidOTP(ctx, email, otp, "password_reset")
	if err != nil {
		return err
	}

	// Get user (use normalized email for lookup to avoid case issues)
	normalizedEmail := strings.TrimSpace(strings.ToLower(email))
	user, err := s.Repo.GetByEmail(ctx, normalizedEmail)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return err
	}

	if err := s.Repo.UpdateUserPassword(ctx, user.ID, string(hashedPassword)); err != nil {
		return err
	}

	// Clear force reset flag if set
	if user.ForceReset {
		_ = s.Repo.ForcePasswordReset(ctx, user.ID, false)
	}

	// Mark OTP as used
	otpRecord.Used = true
	return s.Repo.UpdateOTP(ctx, otpRecord)
}

// GetUserByID method
func (s *Service) GetUserByID(ctx context.Context, userID uuid.UUID) (*UserResponse, error) {
	user, err := s.Repo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	userResponse := s.toUserResponse(user)
	return &userResponse, nil
}

// GetUsers method
func (s *Service) GetUsers(ctx context.Context, page, limit int, search, role, status string) ([]*UserResponse, int64, error) {
	offset := (page - 1) * limit
	users, total, err := s.Repo.GetUsers(ctx, offset, limit, search, role, status)
	if err != nil {
		return nil, 0, err
	}

	userResponses := make([]*UserResponse, len(users))
	for i, user := range users {
		userResponse := s.toUserResponse(user)
		userResponses[i] = &userResponse
	}

	return userResponses, total, nil
}

// UpdateUserStatus method
func (s *Service) UpdateUserStatus(ctx context.Context, userID uuid.UUID, status string) error {
	return s.Repo.UpdateUserStatus(ctx, userID, status)
}

// Update toUserResponse to include new fields
func (s *Service) toUserResponse(user *User) UserResponse {
	// Get permissions if not set
	permissions := user.Permissions
	if len(permissions) == 0 {
		permissions = GetUserPermissions(user.Role)
	}

	return UserResponse{
		ID:          user.ID.String(),
		Name:        user.Name,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Email:       user.Email,
		Phone:       user.Phone,
		Avatar:      user.Avatar,
		Role:        user.Role,
		Permissions: permissions,
		Status:      user.Status,
		CreatedAt:   user.CreatedAt.Format(time.RFC3339),
	}
}



// Add these methods to the Service struct

// CreateUser creates a new user (admin only)
func (s *Service) CreateUser(ctx context.Context, req CreateUserRequest) (*UserResponse, error) {
	// Check if user already exists
	existingUser, _ := s.Repo.GetByEmail(ctx, req.Email)
	if existingUser != nil {
		return nil, fmt.Errorf("email already exists")
	}

	// Create user
	user := &User{
		Name:       req.Name,
		Email:      req.Email,
		Password:   req.Password,
		Phone:      *req.Phone,
		Role:       req.Role,
		Status:     req.Status,
		IsVerified: true, // Admin created users are auto-verified
		ForceReset: true, // Force password reset on first login
	}

	if err := s.Repo.Create(ctx, user); err != nil {
		return nil, err
	}

	// Audit log
	s.AuditService.LogUserAction(ctx, user.ID, "user_created_by_admin", "admin", map[string]interface{}{
		"email": user.Email,
		"role":  user.Role,
	}, "", "")

	userResponse := s.toUserResponse(user)
	return &userResponse, nil
}

// UpdateUser updates user details (admin only)
func (s *Service) UpdateUser(ctx context.Context, userID uuid.UUID, req UpdateUserRequest) (*UserResponse, error) {
	user, err := s.Repo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Check for phone number uniqueness if phone is being updated
	if req.Phone != nil {
		normalizedPhone := strings.TrimSpace(*req.Phone)
		if normalizedPhone != user.Phone {
			existingUser, err := s.Repo.GetByPhone(ctx, normalizedPhone)
			if err == nil && existingUser.ID != userID {
				return nil, fmt.Errorf("phone number already exists")
			}
			user.Phone = normalizedPhone
		}
	}

	// Check for email uniqueness if email is being updated
	if req.Email != nil {
		normalizedEmail := strings.ToLower(strings.TrimSpace(*req.Email))
		if normalizedEmail != user.Email {
			existingUser, err := s.Repo.GetByEmail(ctx, normalizedEmail)
			if err == nil && existingUser.ID != userID {
				return nil, fmt.Errorf("email already exists")
			}
			user.Email = normalizedEmail
		}
	}

	// Update fields if provided
	if req.Name != nil {
		user.Name = *req.Name
	}
	if req.Email != nil {
		user.Email = *req.Email
	}
	if req.Phone != nil {
		user.Phone = *req.Phone
	}
	if req.Role != nil {
		user.Role = *req.Role
	}
	if req.Status != nil {
		user.Status = *req.Status
	}

	if err := s.Repo.Update(ctx, user); err != nil {
		return nil, err
	}

	// Audit log
	s.AuditService.LogUserAction(ctx, user.ID, "user_updated_by_admin", "admin", map[string]interface{}{
		"user_id": userID,
	}, "", "")

	userResponse := s.toUserResponse(user)
	return &userResponse, nil
}

// DeleteUser soft deletes a user (admin only)
func (s *Service) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	user, err := s.Repo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if err := s.Repo.DeleteUser(ctx, userID); err != nil {
		return err
	}

	// Audit log
	s.AuditService.LogUserAction(ctx, userID, "user_deleted_by_admin", "admin", map[string]interface{}{
		"email": user.Email,
	}, "", "")

	return nil
}

// UpdateUserPermissions updates user role/permissions (admin only)
func (s *Service) UpdateUserPermissions(ctx context.Context, userID uuid.UUID, permissions []string) error {
	// In this system, permissions are role-based, so we update the role
	// You could extend this to have a more granular permission system
	var role string
	if contains(permissions, "*") {
		role = "superadmin"
	} else if contains(permissions, "admin:manage:users") {
		role = "admin"
	} else {
		role = "customer"
	}

	if err := s.Repo.UpdateUserRole(ctx, userID, role); err != nil {
		return err
	}

	// Audit log
	s.AuditService.LogUserAction(ctx, userID, "user_permissions_updated", "admin", map[string]interface{}{
		"new_role":    role,
		"permissions": permissions,
	}, "", "")

	return nil
}

// ForcePasswordReset forces user to reset password on next login
func (s *Service) ForcePasswordReset(ctx context.Context, userID uuid.UUID) error {
	if err := s.Repo.ForcePasswordReset(ctx, userID, true); err != nil {
		return err
	}

	// Audit log
	s.AuditService.LogUserAction(ctx, userID, "password_reset_forced", "admin", map[string]interface{}{
		"user_id": userID,
	}, "", "")

	return nil
}

// ChangePassword allows authenticated users to change their password
func (s *Service) ChangePassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error {
	user, err := s.Repo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	// Verify current password
	if !user.CheckPassword(currentPassword) {
		return fmt.Errorf("current password is incorrect")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return err
	}

	if err := s.Repo.UpdateUserPassword(ctx, userID, string(hashedPassword)); err != nil {
		return err
	}

	// Clear force reset flag if set
	if user.ForceReset {
		s.Repo.ForcePasswordReset(ctx, userID, false)
	}

	// Audit log
	s.AuditService.LogUserAction(ctx, userID, "password_changed", "user", map[string]interface{}{
		"user_id": userID,
	}, "", "")

	return nil
}

// ResendOTP resends OTP for email verification
func (s *Service) ResendOTP(ctx context.Context, email string) error {
	// Get user by email
	user, err := s.Repo.GetByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	// Check if email is already verified
	if user.IsVerified {
		return fmt.Errorf("email already verified")
	}

	// Generate new OTP
	otpCode, err := s.generateOTP()
	if err != nil {
		return fmt.Errorf("failed to generate OTP: %w", err)
	}

	// Save OTP to database
	otp := &OTP{
		UserID:    user.ID,
		Email:     user.Email,
		Code:      otpCode,
		Type:      "email_verification",
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}

	if err := s.Repo.SaveOTP(ctx, otp); err != nil {
		return fmt.Errorf("failed to save OTP: %w", err)
	}

	// Send OTP via email
	if err := s.EmailService.SendOTPEmail(ctx, user.Email, otpCode, "email_verification"); err != nil {
		return fmt.Errorf("failed to send OTP email: %w", err)
	}

	// Audit log
	s.AuditService.LogUserAction(ctx, user.ID, "otp_resent", "user", map[string]interface{}{
		"email": email,
		"type":  "email_verification",
	}, "", "")

	return nil
}

// Helper function
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
