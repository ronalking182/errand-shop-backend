package payments

import (
	"time"

	"gorm.io/gorm"
)

// PaymentStatus represents the status of a payment
type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusCompleted PaymentStatus = "completed"
	PaymentStatusFailed    PaymentStatus = "failed"
	PaymentStatusCancelled PaymentStatus = "cancelled"
	PaymentStatusRefunded  PaymentStatus = "refunded"
)

// OrderStatus represents the status of an order
type OrderStatus string

const (
	OrderStatusPending       OrderStatus = "pending"
	OrderStatusConfirmed     OrderStatus = "confirmed"
	OrderStatusPreparing     OrderStatus = "preparing"
	OrderStatusOutForDelivery OrderStatus = "out_for_delivery"
	OrderStatusDelivered     OrderStatus = "delivered"
	OrderStatusCancelled     OrderStatus = "cancelled"
)

// PaymentMethod represents supported payment methods
type PaymentMethod string

const (
	PaymentMethodCard     PaymentMethod = "card"
	PaymentMethodBank     PaymentMethod = "bank_transfer"
	PaymentMethodPaystack PaymentMethod = "paystack"
)

// Payment represents a payment transaction
type Payment struct {
	ID               string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	OrderID          string         `json:"order_id" gorm:"type:uuid;not null;index"`
	CustomerID       uint           `json:"customer_id" gorm:"not null;index"`
	AmountKobo       int64          `json:"amount_kobo" gorm:"not null"` // Amount in kobo (smallest currency unit)
	Currency         string         `json:"currency" gorm:"not null;default:'NGN'"`
	PaymentMethod    PaymentMethod  `json:"payment_method" gorm:"not null"`
	Status           PaymentStatus  `json:"status" gorm:"not null;default:'pending'"`
	TransactionRef   string         `json:"transaction_ref" gorm:"unique;not null"`
	ProviderRef      string         `json:"provider_ref" gorm:"index"`          // Reference from payment provider
	ProviderResponse string         `json:"provider_response" gorm:"type:text"` // JSON response from provider
	FailureReason    string         `json:"failure_reason"`
	ProcessedAt      *time.Time     `json:"processed_at"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `json:"-" gorm:"index"`
}

// PaymentRefund represents a refund transaction
type PaymentRefund struct {
	ID               string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	PaymentID        string         `json:"payment_id" gorm:"type:uuid;not null;index"`
	AmountKobo       int64          `json:"amount_kobo" gorm:"not null"`
	Reason           string         `json:"reason" gorm:"not null"`
	Status           PaymentStatus  `json:"status" gorm:"not null;default:'pending'"`
	RefundRef        string         `json:"refund_ref" gorm:"unique;not null"`
	ProviderRef      string         `json:"provider_ref" gorm:"index"`
	ProviderResponse string         `json:"provider_response" gorm:"type:text"`
	ProcessedAt      *time.Time     `json:"processed_at"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Payment Payment `json:"payment" gorm:"foreignKey:PaymentID"`
}

// PaymentWebhook represents webhook events from payment providers
type PaymentWebhook struct {
	ID          string     `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Provider    string     `json:"provider" gorm:"not null;index"`
	EventType   string     `json:"event_type" gorm:"not null"`
	EventData   string     `json:"event_data" gorm:"type:text;not null"`
	Signature   string     `json:"signature"`
	Processed   bool       `json:"processed" gorm:"default:false;index"`
	ProcessedAt *time.Time `json:"processed_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

// Order represents a payment order for tracking
type Order struct {
	ID            string        `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Reference     string        `json:"reference" gorm:"unique;not null;index"`
	CustomerEmail string        `json:"customer_email" gorm:"not null"`
	AmountKobo    int64         `json:"amount_kobo" gorm:"not null"`
	Currency      string        `json:"currency" gorm:"not null;default:'NGN'"`
	Status        PaymentStatus `json:"status" gorm:"not null;default:'pending'"`
	ItemsSubtotal int64         `json:"items_subtotal" gorm:"not null"`
	TotalAmount   int64         `json:"total_amount" gorm:"not null"`
	Metadata      string        `json:"metadata" gorm:"type:jsonb"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}
