package payments

import (
	"errors"
	"time"
)

// CreatePaymentRequest represents a payment creation request
type CreatePaymentRequest struct {
	OrderID       string        `json:"order_id" validate:"required"`
	PaymentMethod PaymentMethod `json:"payment_method" validate:"required,oneof=card bank_transfer paystack"`
	ReturnURL     string        `json:"return_url,omitempty"`
	CancelURL     string        `json:"cancel_url,omitempty"`
}

// ProcessPaymentRequest represents a payment processing request
type ProcessPaymentRequest struct {
	TransactionRef string `json:"transaction_ref" validate:"required"`
	ProviderRef    string `json:"provider_ref,omitempty"`
	Status         string `json:"status" validate:"required,oneof=completed failed cancelled"`
}

// RefundPaymentRequest represents a refund request
type RefundPaymentRequest struct {
	PaymentID  string `json:"payment_id" validate:"required"`
	AmountKobo int64  `json:"amount_kobo" validate:"required,min=1"`
	Reason     string `json:"reason" validate:"required,min=3,max=500"`
}

// PaymentResponse represents a payment response
type PaymentResponse struct {
	ID             string        `json:"id"`
	OrderID        string        `json:"order_id"`
	CustomerID     uint          `json:"customer_id"`
	AmountKobo     int64         `json:"amount_kobo"`
	AmountNaira    float64       `json:"amount_naira"`
	Currency       string        `json:"currency"`
	PaymentMethod  PaymentMethod `json:"payment_method"`
	Status         PaymentStatus `json:"status"`
	TransactionRef string        `json:"transaction_ref"`
	ProviderRef    string        `json:"provider_ref,omitempty"`
	FailureReason  string        `json:"failure_reason,omitempty"`
	ProcessedAt    *time.Time    `json:"processed_at"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
}

// PaymentInitResponse represents the response after payment initialization
type PaymentInitResponse struct {
	PaymentID      string `json:"payment_id"`
	TransactionRef string `json:"transaction_ref"`
	PaymentURL     string `json:"payment_url,omitempty"`
	Status         string `json:"status"`
	Message        string `json:"message"`
}

// RefundResponse represents a refund response
type RefundResponse struct {
	ID          string        `json:"id"`
	PaymentID   string        `json:"payment_id"`
	AmountKobo  int64         `json:"amount_kobo"`
	AmountNaira float64       `json:"amount_naira"`
	Reason      string        `json:"reason"`
	Status      PaymentStatus `json:"status"`
	RefundRef   string        `json:"refund_ref"`
	ProviderRef string        `json:"provider_ref,omitempty"`
	ProcessedAt *time.Time    `json:"processed_at"`
	CreatedAt   time.Time     `json:"created_at"`
}

// WebhookEventRequest represents a webhook event
type WebhookEventRequest struct {
	Provider  string `json:"provider" validate:"required"`
	EventType string `json:"event_type" validate:"required"`
	EventData string `json:"event_data" validate:"required"`
	Signature string `json:"signature,omitempty"`
}

func (r *CreatePaymentRequest) Validate() error {
	if r.OrderID == "" {
		return errors.New("order ID is required")
	}
	if r.PaymentMethod == "" {
		return errors.New("payment method is required")
	}
	return nil
}

func (r *ProcessPaymentRequest) Validate() error {
	if r.TransactionRef == "" {
		return errors.New("transaction reference is required")
	}
	if r.Status == "" {
		return errors.New("status is required")
	}
	return nil
}

func (r *RefundPaymentRequest) Validate() error {
	if r.PaymentID == "" {
		return errors.New("payment ID is required")
	}
	if r.AmountKobo <= 0 {
		return errors.New("amount must be greater than 0")
	}
	if r.Reason == "" {
		return errors.New("reason is required")
	}
	return nil
}

func (r *WebhookEventRequest) Validate() error {
	if r.Provider == "" {
		return errors.New("provider is required")
	}
	if r.EventType == "" {
		return errors.New("event type is required")
	}
	if r.EventData == "" {
		return errors.New("event_data is required")
	}
	return nil
}

// Paystack-specific DTOs

// PaystackInitializeRequest represents a Paystack payment initialization request
type PaystackInitializeRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Amount    int64  `json:"amount" validate:"required,min=1"`
	Reference string `json:"reference" validate:"required"`
	CallbackURL string `json:"callback_url,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// PaystackInitializeResponse represents Paystack initialization response
type PaystackInitializeResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		AuthorizationURL string `json:"authorization_url"`
		AccessCode       string `json:"access_code"`
		Reference        string `json:"reference"`
	} `json:"data"`
}

// PaystackVerifyResponse represents Paystack verification response
type PaystackVerifyResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		ID              int64  `json:"id"`
		Domain          string `json:"domain"`
		Status          string `json:"status"`
		Reference       string `json:"reference"`
		Amount          int64  `json:"amount"`
		GatewayResponse string `json:"gateway_response"`
		PaidAt          string `json:"paid_at"`
		CreatedAt       string `json:"created_at"`
		Channel         string `json:"channel"`
		Currency        string `json:"currency"`
		Customer        struct {
			ID           int64  `json:"id"`
			FirstName    string `json:"first_name"`
			LastName     string `json:"last_name"`
			Email        string `json:"email"`
			CustomerCode string `json:"customer_code"`
			Phone        string `json:"phone"`
		} `json:"customer"`
	} `json:"data"`
}
