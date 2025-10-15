package users

import "time"

// Repository interface for user data access
type Repository interface {
	Create(user *User) error
	GetByID(id uint) (*User, error)
	GetByEmail(email string) (*User, error)
	Update(user *User) error
	Delete(id uint) error
	List(offset, limit int) ([]User, int64, error)
	UpdatePermissions(userID uint, permissions []string) error
}

type User struct {
	ID          uint       `json:"id" gorm:"primaryKey"`
	Name        string     `json:"name" gorm:"not null"`
	Email       string     `json:"email" gorm:"unique;not null"`
	Phone       *string    `json:"phone" gorm:"unique"`
	Gender      *string    `json:"gender"`
	Password    string     `json:"-" gorm:"not null"`
	Avatar      *string    `json:"avatar"`
	Role        string     `json:"role" gorm:"default:'customer'"`
	Permissions []string   `json:"permissions" gorm:"type:text;serializer:json"`
	Status      string     `json:"status" gorm:"default:'active'"`
	IsVerified  bool       `json:"isVerified" gorm:"default:false"`
	ForceReset  bool       `json:"forceReset" gorm:"default:false"`
	LastLoginAt *time.Time `json:"lastLoginAt"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}
