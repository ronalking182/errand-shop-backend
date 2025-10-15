package customers

import (
	"time"
	"github.com/google/uuid"
)

type Customer struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	UserID      uuid.UUID      `json:"user_id" gorm:"type:uuid;not null;index"`
	FirstName   string         `json:"first_name" gorm:"size:100;not null"`
	LastName    string         `json:"last_name" gorm:"size:100;not null"`
	Phone       string         `json:"phone" gorm:"size:20;index"`
	DateOfBirth *time.Time     `json:"date_of_birth"`
	Gender      string         `json:"gender" gorm:"size:10"`
	Avatar      string         `json:"avatar" gorm:"size:500"`
	Status      CustomerStatus `json:"status" gorm:"default:active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`

	// Relationships
	Addresses []Address `json:"addresses" gorm:"foreignKey:CustomerID"`
}

type Address struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	CustomerID uint      `json:"customer_id" gorm:"not null;index"`
	UserID     uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	Label      string    `json:"label" gorm:"size:100;not null"`
	Type       string    `json:"type" gorm:"size:20;not null"` // home, work, other
	Street     string    `json:"street" gorm:"size:255;not null"`
	City       string    `json:"city" gorm:"size:100;not null"`
	State      string    `json:"state" gorm:"size:100;not null"`
	Country    string    `json:"country" gorm:"size:100;not null"`
	PostalCode string    `json:"postal_code" gorm:"size:20"`
	ZipCode    string    `json:"zip_code" gorm:"size:20;not null"`
	IsDefault  bool      `json:"is_default" gorm:"default:false"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}

type CustomerStatus string

const (
	CustomerStatusActive    CustomerStatus = "active"
	CustomerStatusInactive  CustomerStatus = "inactive"
	CustomerStatusSuspended CustomerStatus = "suspended"
)
