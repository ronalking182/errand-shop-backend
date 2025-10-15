package customers

import (
	"time"
	"github.com/google/uuid"
)

type CreateCustomerRequest struct {
	UserID      uuid.UUID  `json:"user_id" validate:"required"`
	FirstName   string     `json:"first_name" validate:"required,min=2,max=100"`
	LastName    string     `json:"last_name" validate:"required,min=2,max=100"`
	Phone       string     `json:"phone" validate:"omitempty,min=10,max=20"`
	DateOfBirth *time.Time `json:"date_of_birth" validate:"omitempty"`
	Gender      string     `json:"gender" validate:"omitempty,oneof=male female other"`
}

type UpdateCustomerRequest struct {
	FirstName   *string    `json:"first_name" validate:"omitempty,min=2,max=100"`
	LastName    *string    `json:"last_name" validate:"omitempty,min=2,max=100"`
	Phone       *string    `json:"phone" validate:"omitempty,min=10,max=20"`
	DateOfBirth *time.Time `json:"date_of_birth" validate:"omitempty"`
	Gender      *string    `json:"gender" validate:"omitempty,oneof=male female other"`
	Avatar      *string    `json:"avatar" validate:"omitempty,max=500"`
}

type CreateAddressRequest struct {
	Label      string `json:"label" validate:"required,max=100"`
	Type       string `json:"type" validate:"required,oneof=home work other"`
	Street     string `json:"street" validate:"required,max=255"`
	City       string `json:"city" validate:"required,max=100"`
	State      string `json:"state" validate:"required,max=100"`
	Country    string `json:"country" validate:"required,max=100"`
	PostalCode string `json:"postal_code" validate:"omitempty,max=20"`
	IsDefault  bool   `json:"is_default"`
}

type UpdateAddressRequest struct {
	Street     *string `json:"street,omitempty" validate:"omitempty,min=5,max=200"`
	City       *string `json:"city,omitempty" validate:"omitempty,min=2,max=100"`
	State      *string `json:"state,omitempty" validate:"omitempty,min=2,max=100"`
	Country    *string `json:"country,omitempty" validate:"omitempty,min=2,max=100"`
	PostalCode *string `json:"postal_code,omitempty" validate:"omitempty,min=3,max=20"` // Changed from ZipCode
	IsDefault  *bool   `json:"is_default,omitempty"`
}

type CustomerResponse struct {
	ID          uint              `json:"id"`
	UserID      uuid.UUID         `json:"user_id"`
	FirstName   string            `json:"first_name"`
	LastName    string            `json:"last_name"`
	Phone       string            `json:"phone"`
	DateOfBirth *time.Time        `json:"date_of_birth"`
	Gender      string            `json:"gender"`
	Avatar      string            `json:"avatar"`
	Status      CustomerStatus    `json:"status"`
	Addresses   []AddressResponse `json:"addresses"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

type AddressResponse struct {
	ID         uint      `json:"id"`
	UserID     uuid.UUID `json:"user_id"`
	Label      string    `json:"label"`
	Type       string    `json:"type"`
	Street     string    `json:"street"`
	City       string    `json:"city"`
	State      string    `json:"state"`
	Country    string    `json:"country"`
	PostalCode string    `json:"postal_code"`
	IsDefault  bool      `json:"is_default"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
