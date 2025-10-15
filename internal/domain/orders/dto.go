package orders

import (
	"time"
	"github.com/google/uuid"
)

// Cart DTOs
type AddToCartRequest struct {
	ProductID uuid.UUID `json:"productId" validate:"required"`
	Quantity  int       `json:"quantity" validate:"required,min=1,max=100"`
}

type UpdateCartItemRequest struct {
	Quantity int `json:"quantity" validate:"required,min=1,max=100"`
}

type CartResponse struct {
	ID        uuid.UUID          `json:"id"`
	UserID    uuid.UUID          `json:"userId"`
	Items     []CartItemResponse `json:"items"`
	TotalItems int               `json:"totalItems"`
	TotalKobo  int64             `json:"totalKobo"`
	TotalNaira float64           `json:"totalNaira"`
	CreatedAt time.Time          `json:"createdAt"`
	UpdatedAt time.Time          `json:"updatedAt"`
}

type CartItemResponse struct {
	ID            uuid.UUID    `json:"id"`
	ProductID     uuid.UUID    `json:"productId"`
	Quantity      int          `json:"quantity"`
	PriceKobo     int64        `json:"priceKobo"`
	PriceNaira    float64      `json:"priceNaira"`
	SubtotalKobo  int64        `json:"subtotalKobo"`
	SubtotalNaira float64      `json:"subtotalNaira"`
	Product       *ProductInfo `json:"product,omitempty"`
	CreatedAt     time.Time    `json:"createdAt"`
	UpdatedAt     time.Time    `json:"updatedAt"`
}

// Order Request DTOs
type CreateOrderRequest struct {
	DeliveryAddressID *string                   `json:"delivery_address_id"`
	DeliveryMode      string                    `json:"delivery_mode"`
	PaymentMethod     string                    `json:"payment_method"`
	Items             []CreateOrderItemRequest  `json:"items" validate:"dive"`
	CustomRequests    []CreateOrderCustomRequest `json:"custom_requests,omitempty"`
	CouponCode        *string                   `json:"couponCode"`
	Notes             string                    `json:"notes"`
	IdempotencyKey    string                    `json:"IdempotencyKey" validate:"required"`
}

type CreateOrderCustomRequest struct {
	CustomRequestID uuid.UUID `json:"CustomRequestID" validate:"required"`
	Title           string    `json:"title,omitempty"`
	Quantity        int       `json:"quantity,omitempty"`
	Price           int64     `json:"price,omitempty"`
}

type CreateOrderItemRequest struct {
	ProductID uuid.UUID `json:"ProductID" validate:"required"`
	Quantity  int       `json:"quantity" validate:"required,min=1,max=100"`
	Name      string    `json:"name,omitempty"`
	SKU       string    `json:"sku,omitempty"`
}

type CreateOrderFromCartRequest struct {
	DeliveryAddressID *string `json:"delivery_address_id"`
	DeliveryMode      string  `json:"delivery_mode"`
	PaymentMethod     string  `json:"payment_method"`
	CouponCode        *string `json:"couponCode"`
	Notes             string  `json:"notes"`
	IdempotencyKey    string  `json:"IdempotencyKey" validate:"required"`
}

type UpdateOrderStatusRequest struct {
	Status OrderStatus `json:"status" validate:"required,oneof=pending confirmed preparing ready shipped out_for_delivery delivered cancelled refunded"`
	Notes  string      `json:"notes"`
}

type UpdatePaymentStatusRequest struct {
	PaymentStatus PaymentStatus `json:"paymentStatus" validate:"required,oneof=unpaid pending paid partially_refunded refunded failed expired"`
}

type CancelOrderRequest struct {
	Reason string `json:"reason" validate:"required,min=3,max=500"`
}

// Query DTOs
type ListQuery struct {
	Page   int         `query:"page" validate:"omitempty,min=1"`
	Limit  int         `query:"limit" validate:"omitempty,min=1,max=100"`
	Status OrderStatus `query:"status" validate:"omitempty,oneof=pending confirmed preparing ready shipped out_for_delivery delivered cancelled refunded"`
}

type AdminListQuery struct {
	ListQuery
	UserID        *uuid.UUID    `query:"user_id"`
	PaymentStatus PaymentStatus `query:"payment_status" validate:"omitempty,oneof=unpaid pending paid partially_refunded refunded failed expired"`
	SortBy        string        `query:"sort_by" validate:"omitempty,oneof=created_at total_amount status"`
	SortOrder     string        `query:"sort_order" validate:"omitempty,oneof=asc desc"`
	DateFrom      *time.Time    `query:"date_from"`
	DateTo        *time.Time    `query:"date_to"`
}

type OrderStatsQuery struct {
	DateFrom *time.Time `query:"date_from"`
	DateTo   *time.Time `query:"date_to"`
	UserID   *uuid.UUID `query:"user_id"`
}

// Response DTOs
type OrderResponse struct {
	ID                uuid.UUID               `json:"id"`
	CustomerID        uint                    `json:"customerId"`
	Customer          *CustomerInfo           `json:"customer,omitempty"`
	DeliveryAddressID *uint                   `json:"deliveryAddressId"`
	DeliveryAddress   *AddressInfo            `json:"deliveryAddress,omitempty"`
	Status            OrderStatus             `json:"status"`
	PaymentStatus     PaymentStatus           `json:"paymentStatus"`
	IdempotencyKey    string                  `json:"idempotencyKey"`
	CouponCode        *string                 `json:"couponCode"`
	CouponDiscount    int64                   `json:"couponDiscount"`
	CouponDiscountNaira float64               `json:"couponDiscountNaira"`
	ItemsSubtotal     int64                   `json:"itemsSubtotal"`
	ItemsSubtotalNaira float64                `json:"itemsSubtotalNaira"`
	DeliveryFee       int64                   `json:"deliveryFee"`
	DeliveryFeeNaira  float64                 `json:"deliveryFeeNaira"`
	ServiceFee        int64                   `json:"serviceFee"`
	ServiceFeeNaira   float64                 `json:"serviceFeeNaira"`
	TotalAmount       int64                   `json:"totalAmount"`
	TotalAmountNaira  float64                 `json:"totalAmountNaira"`
	CustomRequests    []uuid.UUID             `json:"customRequests"`
	CustomRequestDetails []CustomRequestInfo  `json:"customRequestDetails"`
	Notes             string                  `json:"notes"` 
	EstimatedDelivery *time.Time              `json:"estimatedDelivery"`
	DeliveredAt       *time.Time              `json:"deliveredAt"`
	CancelledAt       *time.Time              `json:"cancelledAt"`
	CancellationReason string                 `json:"cancellationReason"`
	Items             []OrderItemResponse     `json:"items"`
	StatusHistory     []OrderStatusHistoryResponse `json:"statusHistory,omitempty"`
	CreatedAt         time.Time               `json:"createdAt"`
	UpdatedAt         time.Time               `json:"updatedAt"`
}

type OrderItemResponse struct {
	ID           uuid.UUID    `json:"id"`
	ProductID    uuid.UUID    `json:"productId"`
	Name         string       `json:"name"`
	SKU          string       `json:"sku"`
	Quantity     int          `json:"quantity"`
	UnitPrice    int64        `json:"unitPrice"`
	UnitPriceNaira float64    `json:"unitPriceNaira"`
	TotalPrice   int64        `json:"totalPrice"`
	TotalPriceNaira float64   `json:"totalPriceNaira"`
	Product      *ProductInfo `json:"product,omitempty"`
	CreatedAt    time.Time    `json:"createdAt"`
	UpdatedAt    time.Time    `json:"updatedAt"`
}

type OrderStatusHistoryResponse struct {
	ID        uuid.UUID   `json:"id"`
	Status    OrderStatus `json:"status"`
	Notes     string      `json:"notes"`
	ChangedBy *uuid.UUID  `json:"changedBy"`
	CreatedAt time.Time   `json:"createdAt"`
}

type ProductInfo struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Slug     string    `json:"slug"`
	ImageURL string    `json:"imageUrl"`
	Price    int64     `json:"price"`
	PriceNaira float64 `json:"priceNaira"`
}

type OrderStats struct {
	TotalOrders       int64   `json:"totalOrders"`
	PendingOrders     int64   `json:"pendingOrders"`
	ConfirmedOrders   int64   `json:"confirmedOrders"`
	PreparingOrders   int64   `json:"preparingOrders"`
	ShippedOrders     int64   `json:"shippedOrders"`
	DeliveredOrders   int64   `json:"deliveredOrders"`
	CancelledOrders   int64   `json:"cancelledOrders"`
	RefundedOrders    int64   `json:"refundedOrders"`
	TotalRevenue      int64   `json:"totalRevenue"`
	RevenueNaira      float64 `json:"revenueNaira"`
	AverageOrderValue int64   `json:"averageOrderValue"`
	AverageOrderValueNaira float64 `json:"averageOrderValueNaira"`
}

// Validation DTOs
type ValidateCouponRequest struct {
	CouponCode    string `json:"couponCode" validate:"required"`
	OrderSubtotal int64  `json:"orderSubtotal" validate:"required,min=1"`
}

type ValidateCouponResponse struct {
	Valid         bool    `json:"valid"`
	DiscountAmount int64  `json:"discountAmount"`
	DiscountAmountNaira float64 `json:"discountAmountNaira"`
	Message       string  `json:"message"`
}

// Additional Response DTOs
type CouponInfo struct {
	ID   uuid.UUID `json:"id"`
	Code string    `json:"code"`
	Name string    `json:"name"`
}

type CustomerInfo struct {
	ID        uint   `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
}

type AddressInfo struct {
	ID         uint   `json:"id"`
	Label      string `json:"label"`
	Street     string `json:"street"`
	City       string `json:"city"`
	State      string `json:"state"`
	Country    string `json:"country"`
	PostalCode string `json:"postalCode"`
}

type CustomRequestInfo struct {
	ID                 uuid.UUID                    `json:"id"`
	Status             string                       `json:"status"`
	Priority           string                       `json:"priority"`
	AllowSubstitutions bool                         `json:"allowSubstitutions"`
	Notes              string                       `json:"notes"`
	SubmittedAt        time.Time                    `json:"submittedAt"`
	Items              []CustomRequestItemInfo      `json:"items"`
	ActiveQuote        *CustomRequestQuoteInfo      `json:"activeQuote,omitempty"`
}

type CustomRequestItemInfo struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	Quantity       float64   `json:"quantity"`
	Unit           string    `json:"unit"`
	PreferredBrand string    `json:"preferredBrand"`
	QuotedPrice    *int64    `json:"quotedPrice"`
	QuotedPriceNaira *float64 `json:"quotedPriceNaira"`
	AdminNotes     string    `json:"adminNotes"`
	Images         []string  `json:"images"`
}

type CustomRequestQuoteInfo struct {
	ID            uuid.UUID `json:"id"`
	ItemsSubtotal int64     `json:"itemsSubtotal"`
	ItemsSubtotalNaira float64 `json:"itemsSubtotalNaira"`
	GrandTotal    int64     `json:"grandTotal"`
	GrandTotalNaira float64 `json:"grandTotalNaira"`
	Status        string    `json:"status"`
	ValidUntil    *time.Time `json:"validUntil"`
	AcceptedAt    *time.Time `json:"acceptedAt"`
}

// CreateOrderResponse represents the response after creating an order with payment initialization
type CreateOrderResponse struct {
	OrderID     uuid.UUID `json:"order_id"`
	OrderNumber string    `json:"order_number"`
	Payment     PaymentInfo `json:"payment"`
}

// PaymentInfo represents payment initialization data
type PaymentInfo struct {
	Provider   string `json:"provider"`
	Reference  string `json:"reference"`
	PaymentURL string `json:"payment_url"`
}

// Helper function to convert kobo to naira (now returns same value in kobo format)
func KoboToNaira(kobo int64) float64 {
	return float64(kobo)
}
