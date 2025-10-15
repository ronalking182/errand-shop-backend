package notifications

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type NotificationRecipient string
type NotificationType string
type NotificationStatus string
type DevicePlatform string

const (
	// Recipients
	RecipientCustomer NotificationRecipient = "customer"
	RecipientAdmin    NotificationRecipient = "admin"
	RecipientDriver   NotificationRecipient = "driver"

	// Types
	TypeOrderUpdate    NotificationType = "order_update"
	TypeDeliveryUpdate NotificationType = "delivery_update"
	TypePaymentUpdate  NotificationType = "payment_update"
	TypePromotion      NotificationType = "promotion"
	TypeSystem         NotificationType = "system"

	// Status
	StatusPending NotificationStatus = "pending"
	StatusSent    NotificationStatus = "sent"
	StatusRead    NotificationStatus = "read"
	StatusFailed  NotificationStatus = "failed"

	// Platforms
	PlatformIOS     DevicePlatform = "ios"
	PlatformAndroid DevicePlatform = "android"
	PlatformWeb     DevicePlatform = "web"
)

type Notification struct {
	ID            uint                  `gorm:"primaryKey" json:"id"`
	RecipientID   uuid.UUID             `gorm:"type:uuid;not null" json:"recipientId"`
	RecipientType NotificationRecipient `gorm:"type:varchar(20);not null" json:"recipientType"`
	Type          NotificationType      `gorm:"type:varchar(30);not null" json:"type"`
	Title         string                `gorm:"size:200;not null" json:"title"`
	Body          string                `gorm:"type:text;not null" json:"body"`
	Data          JSONMap               `gorm:"type:jsonb" json:"data,omitempty"`
	Status        NotificationStatus    `gorm:"type:varchar(20);default:'pending'" json:"status"`
	ReadAt        *time.Time            `json:"readAt,omitempty"`
	SentAt        *time.Time            `json:"sentAt,omitempty"`
	CreatedAt     time.Time             `json:"createdAt"`
	UpdatedAt     time.Time             `json:"updatedAt"`
	DeletedAt     gorm.DeletedAt        `gorm:"index" json:"-"`
}

type NotificationTemplate struct {
	ID        uint             `gorm:"primaryKey" json:"id"`
	Type      NotificationType `gorm:"type:varchar(30);uniqueIndex;not null" json:"type"`
	Title     string           `gorm:"size:200;not null" json:"title"`
	Body      string           `gorm:"type:text;not null" json:"body"`
	IsActive  bool             `gorm:"default:true" json:"isActive"`
	CreatedAt time.Time        `json:"createdAt"`
	UpdatedAt time.Time        `json:"updatedAt"`
}

type PushToken struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	UserID     uuid.UUID      `gorm:"type:uuid;not null" json:"userId"`
	UserType   string         `gorm:"type:varchar(20);not null" json:"userType"`
	Token      string         `gorm:"size:500;not null" json:"token"`
	Platform   DevicePlatform `gorm:"type:varchar(20);not null" json:"platform"`
	DeviceID   string         `gorm:"size:100" json:"deviceId"`
	IsActive   bool           `gorm:"default:true" json:"isActive"`
	LastUsedAt time.Time      `json:"lastUsedAt"`
	CreatedAt  time.Time      `json:"createdAt"`
	UpdatedAt  time.Time      `json:"updatedAt"`
}

// FCM-specific models for dashboard requirements
type FCMToken struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"userId"`
	UserType  string         `gorm:"type:varchar(20);not null" json:"userType"`
	Token     string         `gorm:"size:500;not null;uniqueIndex" json:"token"`
	Platform  DevicePlatform `gorm:"type:varchar(20);not null" json:"platform"`
	DeviceID  string         `gorm:"size:100" json:"deviceId"`
	IsActive  bool           `gorm:"default:true" json:"isActive"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
}

// JSONMap is a custom type for handling JSON data in database
type JSONMap map[string]interface{}

// Scan implements the sql.Scanner interface for database retrieval
func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, j)
	case string:
		return json.Unmarshal([]byte(v), j)
	default:
		return fmt.Errorf("cannot scan %T into JSONMap", value)
	}
}

// Value implements the driver.Valuer interface for database storage
func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

type FCMMessage struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Title       string    `gorm:"size:200;not null" json:"title"`
	Body        string    `gorm:"type:text;not null" json:"body"`
	Data        JSONMap   `gorm:"type:jsonb" json:"data,omitempty"`
	ImageURL    string    `gorm:"size:500" json:"imageUrl,omitempty"`
	MessageType string    `gorm:"type:varchar(20);not null" json:"messageType"` // single, multiple, broadcast
	SentBy      uuid.UUID `gorm:"type:uuid;not null" json:"sentBy"`
	SentAt      time.Time `json:"sentAt"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	
	// Relationships
	Recipients []FCMMessageRecipient `gorm:"foreignKey:MessageID" json:"recipients,omitempty"`
}

type FCMMessageRecipient struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	MessageID   uint      `gorm:"not null;index" json:"messageId"`
	UserID      uuid.UUID `gorm:"type:uuid;not null;index" json:"userId"`
	UserType    string    `gorm:"type:varchar(20);not null" json:"userType"`
	TokenID     uint      `gorm:"not null;index" json:"tokenId"`
	Status      string    `gorm:"type:varchar(20);default:'pending'" json:"status"` // pending, sent, failed, delivered
	Error       string    `gorm:"type:text" json:"error,omitempty"`
	DeliveredAt *time.Time `json:"deliveredAt,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	
	// Relationships
	Message FCMMessage `gorm:"foreignKey:MessageID" json:"message,omitempty"`
	Token   FCMToken   `gorm:"foreignKey:TokenID" json:"token,omitempty"`
}

// Table names
func (FCMToken) TableName() string {
	return "fcm_tokens"
}

func (FCMMessage) TableName() string {
	return "fcm_messages"
}

func (FCMMessageRecipient) TableName() string {
	return "fcm_message_recipients"
}
