package orders

import (
	"database/sql/driver"
	"encoding/json"
	"errandShop/internal/domain/products"
	"time"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Cart represents a user's shopping cart
type Cart struct {
	ID        uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex" json:"userId"`
	Items     []CartItem `gorm:"foreignKey:CartID;constraint:OnDelete:CASCADE" json:"items"`
	CreatedAt time.Time  `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`
}

// CartItem represents an item in a user's cart
type CartItem struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CartID    uuid.UUID `gorm:"type:uuid;not null" json:"cartId"`
	ProductID uuid.UUID `gorm:"type:uuid;not null" json:"productId"`
	Quantity  int       `gorm:"not null;check:quantity > 0" json:"quantity"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`

	// Relationships
	Cart    Cart             `gorm:"foreignKey:CartID" json:"-"`
	Product products.Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

// Order represents a customer order
type Order struct {
	ID                uuid.UUID     `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CustomerID        uuid.UUID     `gorm:"type:uuid;not null;column:customer_id" json:"customerId"`
	DeliveryAddressID *uint         `gorm:"column:delivery_address_id" json:"deliveryAddressId"`
	Status            OrderStatus   `gorm:"type:varchar(50);not null;default:'pending'" json:"status"`
	PaymentStatus     PaymentStatus `gorm:"type:varchar(50);not null;default:'unpaid'" json:"paymentStatus"`
	IdempotencyKey    string        `gorm:"type:varchar(255);uniqueIndex" json:"idempotencyKey"`
	CouponCode        *string       `gorm:"type:varchar(100)" json:"couponCode"`
	CouponDiscount    int64         `gorm:"default:0" json:"couponDiscount"` // in kobo
	ItemsSubtotal     int64         `gorm:"not null" json:"itemsSubtotal"`   // in kobo
	DeliveryFee       int64         `gorm:"default:0" json:"deliveryFee"`    // in kobo
	ServiceFee        int64         `gorm:"default:0" json:"serviceFee"`     // in kobo
	TotalAmount       int64         `gorm:"not null" json:"totalAmount"`     // in kobo
	CustomRequests    UUIDSlice     `gorm:"type:jsonb;default:'[]'" json:"customRequests"` // Custom request IDs
	Notes             string        `gorm:"type:text" json:"notes"`
	EstimatedDelivery *time.Time    `json:"estimatedDelivery"`
	DeliveredAt       *time.Time    `json:"deliveredAt"`
	CancelledAt       *time.Time    `json:"cancelledAt"`
	CancellationReason string       `gorm:"type:text" json:"cancellationReason"`
	Items             []OrderItem   `gorm:"foreignKey:OrderID;constraint:OnDelete:CASCADE" json:"items"`
	StatusHistory     []OrderStatusHistory `gorm:"foreignKey:OrderID;constraint:OnDelete:CASCADE" json:"statusHistory,omitempty"`
	CreatedAt         time.Time     `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt         time.Time     `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"` 
}

// OrderItem represents an item within an order
type OrderItem struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	OrderID   uuid.UUID `gorm:"type:uuid;not null;column:order_id" json:"orderId"`
	ProductID uuid.UUID `gorm:"type:uuid;not null;column:product_id" json:"productId"`
	Name      string    `gorm:"type:varchar(255);not null" json:"name"`
	SKU       string    `gorm:"type:varchar(100)" json:"sku"`
	Source    string    `gorm:"type:varchar(50);default:'catalog'" json:"source"`
	Quantity  int       `gorm:"not null;check:quantity > 0" json:"quantity"`
	UnitPrice int64     `gorm:"not null" json:"unitPrice"` // Price per unit in kobo at time of order
	TotalPrice int64    `gorm:"not null" json:"totalPrice"` // Total price for this item in kobo
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`

	// Relationships
	Order   Order            `gorm:"foreignKey:OrderID" json:"-"`
	Product products.Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

// OrderStatusHistory tracks status changes for orders
type OrderStatusHistory struct {
	ID         uuid.UUID   `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	OrderID    uuid.UUID   `json:"orderId" gorm:"type:uuid;not null;index;column:order_id"`
	FromStatus *OrderStatus `json:"fromStatus" gorm:"size:50;column:from_status"`
	ToStatus   OrderStatus `json:"toStatus" gorm:"size:50;not null;column:to_status"`
	ByAdminID  *uuid.UUID  `json:"byAdminId" gorm:"type:uuid;column:by_admin_id"`
	Note       string      `json:"note" gorm:"type:text;column:note"`
	CreatedAt  time.Time   `json:"createdAt" gorm:"autoCreateTime;column:created_at"`

	// Relationships
	Order Order `gorm:"foreignKey:OrderID" json:"-"`
}

// TableName specifies the table name for OrderStatusHistory
func (OrderStatusHistory) TableName() string {
	return "order_status_history"
}

type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "pending"
	OrderStatusConfirmed  OrderStatus = "confirmed"
	OrderStatusPreparing  OrderStatus = "preparing"
	OrderStatusOutForDelivery OrderStatus = "out_for_delivery"
	OrderStatusDelivered  OrderStatus = "delivered"
	OrderStatusCancelled  OrderStatus = "cancelled"
)

type PaymentStatus string

const (
	PaymentStatusUnpaid     PaymentStatus = "unpaid"
	PaymentStatusPending    PaymentStatus = "pending"
	PaymentStatusPaid       PaymentStatus = "paid"
	PaymentStatusPartiallyRefunded PaymentStatus = "partially_refunded"
	PaymentStatusRefunded   PaymentStatus = "refunded"
	PaymentStatusFailed     PaymentStatus = "failed"
	PaymentStatusExpired    PaymentStatus = "expired"
)

// Helper methods for Order
func (o *Order) CanBeCancelled() bool {
	return o.Status == OrderStatusPending || o.Status == OrderStatusConfirmed
}

func (o *Order) IsDelivered() bool {
	return o.Status == OrderStatusDelivered
}

func (o *Order) IsCancelled() bool {
	return o.Status == OrderStatusCancelled
}

func (o *Order) CalculateTotal() int64 {
	return o.ItemsSubtotal + o.DeliveryFee + o.ServiceFee - o.CouponDiscount
}

// BeforeCreate GORM hook
func (o *Order) BeforeCreate(tx *gorm.DB) error {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	o.TotalAmount = o.CalculateTotal()
	
	// Set estimated delivery to 2 hours from now if not already set
	if o.EstimatedDelivery == nil {
		estimatedTime := time.Now().Add(2 * time.Hour)
		o.EstimatedDelivery = &estimatedTime
	}
	
	return nil
}

// BeforeUpdate GORM hook
func (o *Order) BeforeUpdate(tx *gorm.DB) error {
	o.TotalAmount = o.CalculateTotal()
	return nil
}

// Helper methods for Cart
func (c *Cart) GetTotalItems() int {
	total := 0
	for _, item := range c.Items {
		total += item.Quantity
	}
	return total
}

func (c *Cart) IsEmpty() bool {
	return len(c.Items) == 0
}

// BeforeCreate GORM hook for Cart
func (c *Cart) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

// UUIDSlice is a custom type for handling UUID slices in JSONB
type UUIDSlice []uuid.UUID

// Value implements the driver.Valuer interface for database storage
func (us UUIDSlice) Value() (driver.Value, error) {
	return json.Marshal(us)
}

// Scan implements the sql.Scanner interface for database retrieval
func (us *UUIDSlice) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, us)
	case string:
		return json.Unmarshal([]byte(v), us)
	default:
		return nil
	}
}
