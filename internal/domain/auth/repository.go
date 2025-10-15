package auth

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// User operations - Fixed method names
func (r *Repository) Create(ctx context.Context, user *User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *Repository) GetByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

func (r *Repository) GetByPhone(ctx context.Context, phone string) (*User, error) {
	var user User
	err := r.db.WithContext(ctx).Where("phone = ?", phone).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	var user User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

func (r *Repository) Update(ctx context.Context, user *User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *Repository) GetUsers(ctx context.Context, offset, limit int, search, role, status string) ([]*User, int64, error) {
	query := r.db.WithContext(ctx).Model(&User{})

	if search != "" {
		query = query.Where("name ILIKE ? OR email ILIKE ?", "%"+search+"%", "%"+search+"%")
	}
	if role != "" {
		query = query.Where("role = ?", role)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var users []*User
	err := query.Offset(offset).Limit(limit).Find(&users).Error
	return users, total, err
}

func (r *Repository) UpdateUserStatus(ctx context.Context, userID uuid.UUID, status string) error {
	return r.db.WithContext(ctx).Model(&User{}).Where("id = ?", userID).Update("status", status).Error
}

// OTP operations - Fixed method names
func (r *Repository) SaveOTP(ctx context.Context, otp *OTP) error {
	return r.db.WithContext(ctx).Create(otp).Error
}

func (r *Repository) GetOTPByUserIDAndCode(ctx context.Context, userID uuid.UUID, code string) (*OTP, error) {
	var otp OTP
	err := r.db.WithContext(ctx).Where(
		"user_id = ? AND code = ? AND expires_at > ? AND used = false",
		userID, code, time.Now(),
	).First(&otp).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid or expired OTP")
		}
		return nil, err
	}
	return &otp, nil
}

func (r *Repository) GetValidOTP(ctx context.Context, email, otpCode, purpose string) (*OTP, error) {
	var otp OTP
	err := r.db.WithContext(ctx).Where(
		"email = ? AND code = ? AND type = ? AND expires_at > ? AND used = false",
		email, otpCode, purpose, time.Now(),
	).First(&otp).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid or expired OTP")
		}
		return nil, err
	}
	return &otp, nil
}

func (r *Repository) UpdateOTP(ctx context.Context, otp *OTP) error {
	return r.db.WithContext(ctx).Save(otp).Error
}

func (r *Repository) GetValidOTPByCode(ctx context.Context, code, purpose string) (*OTP, error) {
	var otp OTP
	err := r.db.WithContext(ctx).Where("code = ? AND type = ? AND used = ? AND expires_at > ?", code, purpose, false, time.Now()).First(&otp).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid or expired OTP")
		}
		return nil, err
	}
	return &otp, nil
}

// Refresh token operations
func (r *Repository) CreateRefreshToken(ctx context.Context, token *RefreshToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

func (r *Repository) GetRefreshToken(ctx context.Context, token string) (*RefreshToken, error) {
	var refreshToken RefreshToken
	err := r.db.WithContext(ctx).Where("token = ? AND expires_at > ?", token, time.Now()).First(&refreshToken).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("refresh token not found or expired")
		}
		return nil, err
	}
	return &refreshToken, nil
}

func (r *Repository) DeleteRefreshToken(ctx context.Context, token string) error {
	return r.db.WithContext(ctx).Where("token = ?", token).Delete(&RefreshToken{}).Error
}

func (r *Repository) DeleteUserRefreshTokens(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&RefreshToken{}).Error
}

func (r *Repository) StoreRefreshToken(ctx context.Context, userID uuid.UUID, token string) error {
	refreshToken := &RefreshToken{
		UserID:    userID,
		Token:     token,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // 7 days
	}
	return r.db.WithContext(ctx).Create(refreshToken).Error
}

// Address operations
func (r *Repository) CreateAddress(ctx context.Context, address *Address) error {
	return r.db.WithContext(ctx).Create(address).Error
}

func (r *Repository) GetUserAddresses(ctx context.Context, userID uuid.UUID) ([]*Address, error) {
	var addresses []*Address
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&addresses).Error
	return addresses, err
}

// Add these methods to the Repository struct

// DeleteUser soft deletes a user
func (r *Repository) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&User{}, userID).Error
}

// UpdateUserPermissions updates user role (permissions are role-based)
func (r *Repository) UpdateUserRole(ctx context.Context, userID uuid.UUID, role string) error {
	return r.db.WithContext(ctx).Model(&User{}).Where("id = ?", userID).Update("role", role).Error
}

// ForcePasswordReset sets force reset flag
func (r *Repository) ForcePasswordReset(ctx context.Context, userID uuid.UUID, forceReset bool) error {
	return r.db.WithContext(ctx).Model(&User{}).Where("id = ?", userID).Update("force_reset", forceReset).Error
}

// UpdateUserPassword updates user password
func (r *Repository) UpdateUserPassword(ctx context.Context, userID uuid.UUID, hashedPassword string) error {
	return r.db.WithContext(ctx).Model(&User{}).Where("id = ?", userID).Update("password", hashedPassword).Error
}
