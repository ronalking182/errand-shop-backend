package chat

import (
	"time"

	"gorm.io/gorm"
)

// ChatRoom represents a conversation between admin and customer
type ChatRoom struct {
	ID         uint           `json:"id" gorm:"primaryKey"`
	CustomerID uint           `json:"customer_id" gorm:"not null;index"`
	AdminID    *uint          `json:"admin_id" gorm:"index"` // Nullable, assigned when admin joins
	Status     ChatStatus     `json:"status" gorm:"type:varchar(20);default:'active'"`
	Subject    string         `json:"subject" gorm:"type:varchar(255)"`
	Priority   ChatPriority   `json:"priority" gorm:"type:varchar(20);default:'normal'"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Messages []ChatMessage `json:"messages,omitempty" gorm:"foreignKey:RoomID"`
}

// ChatMessage represents individual messages in a chat room
type ChatMessage struct {
	ID         uint           `json:"id" gorm:"primaryKey"`
	RoomID     uint           `json:"room_id" gorm:"not null;index"`
	SenderID   uint           `json:"sender_id" gorm:"not null"`
	SenderType SenderType     `json:"sender_type" gorm:"type:varchar(20);not null"`
	Message    string         `json:"message" gorm:"type:text;not null"`
	MessageType MessageType   `json:"message_type" gorm:"type:varchar(20);default:'text'"`
	Attachments []string      `json:"attachments" gorm:"type:json"`
	IsRead     bool           `json:"is_read" gorm:"default:false"`
	ReadAt     *time.Time     `json:"read_at"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Room ChatRoom `json:"room,omitempty" gorm:"foreignKey:RoomID"`
}

// ChatStatus represents the status of a chat room
type ChatStatus string

const (
	ChatStatusActive   ChatStatus = "active"
	ChatStatusClosed   ChatStatus = "closed"
	ChatStatusArchived ChatStatus = "archived"
)

// ChatPriority represents the priority level of a chat
type ChatPriority string

const (
	ChatPriorityLow    ChatPriority = "low"
	ChatPriorityNormal ChatPriority = "normal"
	ChatPriorityHigh   ChatPriority = "high"
	ChatPriorityUrgent ChatPriority = "urgent"
)

// SenderType represents who sent the message
type SenderType string

const (
	SenderTypeCustomer SenderType = "customer"
	SenderTypeAdmin    SenderType = "admin"
	SenderTypeSystem   SenderType = "system"
)

// MessageType represents the type of message
type MessageType string

const (
	MessageTypeText  MessageType = "text"
	MessageTypeImage MessageType = "image"
	MessageTypeFile  MessageType = "file"
	MessageTypeAudio MessageType = "audio"
	MessageTypeVideo MessageType = "video"
)

// TableName sets the table name for ChatRoom
func (ChatRoom) TableName() string {
	return "chat_rooms"
}

// TableName sets the table name for ChatMessage
func (ChatMessage) TableName() string {
	return "chat_messages"
}