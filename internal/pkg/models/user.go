package models

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	FirstName   string         `json:"first_name" gorm:"not null"`
	LastName    string         `json:"last_name" gorm:"not null"`
	Name        string         `json:"name" gorm:"not null"`
	Email       string         `json:"email" gorm:"not null"`
	Password    string         `json:"-" gorm:"not null"`
	Phone       string         `json:"phone" gorm:"not null"`
	Avatar      *string        `json:"avatar"`
	Role        string         `json:"role" gorm:"default:'customer'"`
	Permissions []string       `json:"permissions" gorm:"type:text[]"`
	Status      string         `json:"status" gorm:"default:'active'"`
	IsVerified  bool           `json:"is_verified" gorm:"default:false"`
	ForceReset  bool           `json:"force_reset" gorm:"default:false"`
	LastLoginAt *time.Time     `json:"last_login_at"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	RefreshTokens []RefreshToken `json:"-" gorm:"foreignKey:UserID"`
	OTPs          []OTP         `json:"-" gorm:"foreignKey:UserID"`
}

// BeforeCreate hook to hash password
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		u.Password = string(hashedPassword)
	}
	return nil
}

// CheckPassword verifies the password
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

// HasPermission checks if user has a specific permission
func (u *User) HasPermission(permission string) bool {
	for _, p := range u.Permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// Address represents a user address
type Address struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID    uuid.UUID      `json:"user_id" gorm:"type:uuid;not null"`
	Street    string         `json:"street" gorm:"not null"`
	City      string         `json:"city" gorm:"not null"`
	State     string         `json:"state" gorm:"not null"`
	ZipCode   string         `json:"zip_code" gorm:"not null"`
	Country   string         `json:"country" gorm:"not null"`
	IsDefault bool           `json:"is_default" gorm:"default:false"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	User User `json:"-" gorm:"foreignKey:UserID"`
}

// RefreshToken represents a refresh token
type RefreshToken struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID    uuid.UUID      `json:"user_id" gorm:"type:uuid;not null"`
	Token     string         `json:"token" gorm:"unique;not null"`
	ExpiresAt time.Time      `json:"expires_at" gorm:"not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	User User `json:"-" gorm:"foreignKey:UserID"`
}

// OTP represents an OTP token
type OTP struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID    uuid.UUID      `json:"user_id" gorm:"type:uuid;not null"`
	Email     string         `json:"email" gorm:"not null"`
	Code      string         `json:"code" gorm:"not null"`
	Type      string         `json:"type" gorm:"not null"` // email_verification, password_reset
	ExpiresAt time.Time      `json:"expires_at" gorm:"not null"`
	Used      bool           `json:"used" gorm:"default:false"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	User User `json:"-" gorm:"foreignKey:UserID"`
}