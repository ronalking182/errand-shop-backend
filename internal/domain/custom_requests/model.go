package custom_requests

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"errandShop/internal/domain/products"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CustomRequest represents a user request for items not in the catalog
type CustomRequest struct {
	ID                 uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID             uuid.UUID      `gorm:"type:uuid;not null" json:"userId"`
	DeliveryAddressID  *uuid.UUID     `gorm:"type:uuid" json:"deliveryAddressId"`
	Status             RequestStatus  `gorm:"type:varchar(50);not null;default:'submitted'" json:"status"`
	Priority           RequestPriority `gorm:"type:varchar(20);default:'MEDIUM'" json:"priority"`
	AllowSubstitutions bool           `gorm:"default:true" json:"allowSubstitutions"`
	Notes              string         `gorm:"type:text" json:"notes"`
	AssigneeID         *uuid.UUID     `gorm:"type:uuid" json:"assigneeId"`
	SubmittedAt        time.Time      `gorm:"default:now()" json:"submittedAt"`
	UpdatedAt          time.Time      `gorm:"default:now()" json:"updatedAt"`
	ExpiresAt          *time.Time     `json:"expiresAt"`

	// Relationships
	Items    []RequestItem           `gorm:"foreignKey:CustomRequestID;constraint:OnDelete:CASCADE" json:"items"`
	Quotes   []Quote                 `gorm:"foreignKey:CustomRequestID;constraint:OnDelete:CASCADE" json:"quotes"`
	Messages []CustomRequestMessage `gorm:"foreignKey:CustomRequestID;constraint:OnDelete:CASCADE" json:"messages"`
}

// RequestItem represents an individual item within a custom request
type RequestItem struct {
	ID               uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CustomRequestID  uuid.UUID  `gorm:"type:uuid;not null" json:"customRequestId"`
	Name             string     `gorm:"type:varchar(255);not null" json:"name"`
	Description      string     `gorm:"type:text" json:"description"`
	Quantity         float64    `gorm:"not null" json:"quantity"`
	Unit             string     `gorm:"type:varchar(50)" json:"unit"`
	PreferredBrand   string     `gorm:"type:varchar(255)" json:"preferredBrand"`
	EstimatedPrice   *int64     `gorm:"column:estimated_price" json:"estimatedPrice"` // in kobo
	QuotedPrice      *int64     `gorm:"column:quoted_price" json:"quotedPrice"`    // in kobo
	AdminNotes       string     `gorm:"column:admin_notes;type:text" json:"adminNotes"`
	Images           products.StringSlice   `gorm:"type:jsonb;default:'[]'" json:"images"`

	// Relationships
	CustomRequest CustomRequest `gorm:"foreignKey:CustomRequestID" json:"-"`
}

// Quote represents an admin quote for a custom request
type Quote struct {
	ID               uuid.UUID   `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CustomRequestID  uuid.UUID   `gorm:"type:uuid;not null" json:"customRequestId"`
	ItemsSubtotal    int64       `gorm:"not null" json:"itemsSubtotal"` // in kobo
	Fees             QuoteFees   `gorm:"type:jsonb;not null" json:"fees"`
	FeesTotal        int64       `gorm:"not null" json:"feesTotal"` // in kobo
	GrandTotal       int64       `gorm:"not null" json:"grandTotal"` // in kobo
	Status           QuoteStatus `gorm:"type:varchar(20);not null;default:'DRAFT'" json:"status"`
	ValidUntil       *time.Time  `json:"validUntil"`
	SentAt           *time.Time  `json:"sentAt"`
	AcceptedAt       *time.Time  `json:"acceptedAt"`
	CreatedAt        time.Time   `gorm:"default:now()" json:"createdAt"`
	UpdatedAt        time.Time   `gorm:"default:now()" json:"updatedAt"`

	// Relationships
	CustomRequest CustomRequest `gorm:"foreignKey:CustomRequestID" json:"-"`
	Items         []QuoteItem   `gorm:"foreignKey:QuoteID;constraint:OnDelete:CASCADE" json:"items"`
}

// QuoteItem represents an individual item within a quote
type QuoteItem struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	QuoteID       uuid.UUID `gorm:"type:uuid;not null" json:"quoteId"`
	RequestItemID uuid.UUID `gorm:"type:uuid;not null" json:"requestItemId"`
	QuotedPrice   int64     `gorm:"not null" json:"quotedPrice"` // in kobo
	AdminNotes    string    `gorm:"type:text" json:"adminNotes"`
	CreatedAt     time.Time `gorm:"default:now()" json:"createdAt"`
	UpdatedAt     time.Time `gorm:"default:now()" json:"updatedAt"`

	// Relationships
	Quote       Quote       `gorm:"foreignKey:QuoteID" json:"-"`
	RequestItem RequestItem `gorm:"foreignKey:RequestItemID" json:"-"`
}

// CustomRequestMessage represents communication between users and admins
type CustomRequestMessage struct {
	ID               uuid.UUID   `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CustomRequestID  uuid.UUID   `gorm:"type:uuid;not null" json:"customRequestId"`
	SenderType       SenderType  `gorm:"type:varchar(20);not null" json:"senderType"`
	SenderID         uuid.UUID   `gorm:"type:uuid;not null" json:"senderId"`
	Message          string      `gorm:"type:text;not null" json:"message"`
	CreatedAt        time.Time   `gorm:"default:now()" json:"createdAt"`

	// Relationships
	CustomRequest CustomRequest `gorm:"foreignKey:CustomRequestID" json:"-"`
}

// QuoteFees represents the fees structure in a quote
type QuoteFees struct {
	Delivery  int64 `json:"delivery"`  // in kobo
	Service   int64 `json:"service"`   // in kobo
	Packaging int64 `json:"packaging"` // in kobo
}

// Value implements the driver.Valuer interface for database storage
func (qf QuoteFees) Value() (driver.Value, error) {
	return json.Marshal(qf)
}

// Scan implements the sql.Scanner interface for database retrieval
func (qf *QuoteFees) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, qf)
	case string:
		return json.Unmarshal([]byte(v), qf)
	default:
		return errors.New("cannot scan QuoteFees from non-string/[]byte value")
	}
}

// RequestStatus represents the status of a custom request
type RequestStatus string

const (
	RequestSubmitted        RequestStatus = "submitted"
	RequestUnderReview      RequestStatus = "under_review"
	RequestQuoteSent        RequestStatus = "quote_sent"
	RequestQuoteReady       RequestStatus = "quote_ready"
	RequestNeedsInfo        RequestStatus = "needs_info"
	RequestCustomerAccepted RequestStatus = "customer_accepted"
	RequestCustomerDeclined RequestStatus = "customer_declined"
	RequestApproved         RequestStatus = "approved"
	RequestInCart           RequestStatus = "in_cart"
	RequestCancelled        RequestStatus = "cancelled"
)

// RequestPriority represents the priority level of a custom request
type RequestPriority string

const (
	PriorityLow    RequestPriority = "LOW"
	PriorityMedium RequestPriority = "MEDIUM"
	PriorityHigh   RequestPriority = "HIGH"
	PriorityUrgent RequestPriority = "URGENT"
)

// QuoteStatus represents the status of a quote
type QuoteStatus string

const (
	QuoteDraft    QuoteStatus = "DRAFT"
	QuoteSent     QuoteStatus = "SENT"
	QuoteAccepted QuoteStatus = "ACCEPTED"
	QuoteDeclined QuoteStatus = "DECLINED"
)

// SenderType represents who sent a message
type SenderType string

const (
	SenderUser  SenderType = "user"
	SenderAdmin SenderType = "admin"
)

// GORM Hooks
func (cr *CustomRequest) BeforeUpdate(tx *gorm.DB) error {
	cr.UpdatedAt = time.Now()
	return nil
}

func (q *Quote) BeforeUpdate(tx *gorm.DB) error {
	q.UpdatedAt = time.Now()
	return nil
}

// Helper methods
func (cr *CustomRequest) IsExpired() bool {
	if cr.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*cr.ExpiresAt)
}

func (q *Quote) IsValid() bool {
	if q.ValidUntil == nil {
		return true
	}
	return time.Now().Before(*q.ValidUntil)
}

func (q *Quote) IsExpired() bool {
	return !q.IsValid()
}

// CalculateTotal calculates the grand total from items subtotal and fees
func (q *Quote) CalculateTotal() {
	q.FeesTotal = q.Fees.Delivery + q.Fees.Service + q.Fees.Packaging
	q.GrandTotal = q.ItemsSubtotal + q.FeesTotal
}

// GetActiveQuote returns the most recent active quote for a custom request
// Includes both sent and accepted quotes to handle the full quote lifecycle
func (cr *CustomRequest) GetActiveQuote() *Quote {
	for i := len(cr.Quotes) - 1; i >= 0; i-- {
		quote := &cr.Quotes[i]
		// Include both SENT and ACCEPTED quotes to fix amount disappearing issue
		if (quote.Status == QuoteSent || quote.Status == QuoteAccepted) && quote.IsValid() {
			return quote
		}
	}
	return nil
}

// CanBeAccepted checks if a custom request can be accepted by the user
func (cr *CustomRequest) CanBeAccepted() bool {
	return cr.Status == RequestQuoteSent && !cr.IsExpired()
}

// CanBeModified checks if a custom request can be modified
func (cr *CustomRequest) CanBeModified() bool {
	return cr.Status == RequestSubmitted || cr.Status == RequestNeedsInfo
}

// GetLatestMessage returns the most recent message
func (cr *CustomRequest) GetLatestMessage() *CustomRequestMessage {
	if len(cr.Messages) == 0 {
		return nil
	}
	return &cr.Messages[len(cr.Messages)-1]
}

// TableName specifies the table name for GORM
func (CustomRequest) TableName() string {
	return "custom_requests"
}

func (RequestItem) TableName() string {
	return "request_items"
}

func (Quote) TableName() string {
	return "quotes"
}

func (CustomRequestMessage) TableName() string {
	return "custom_request_messages"
}