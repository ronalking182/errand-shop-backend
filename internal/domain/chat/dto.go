package chat

import "time"

// Chat Room DTOs
type CreateChatRoomRequest struct {
	CustomerID uint         `json:"customer_id" validate:"required"`
	Subject    string       `json:"subject" validate:"required,min=3,max=255"`
	Priority   ChatPriority `json:"priority" validate:"omitempty,oneof=low normal high urgent"`
	Message    string       `json:"message" validate:"required,min=1,max=1000"`
}

type ChatRoomResponse struct {
	ID               uint              `json:"id"`
	CustomerID       uint              `json:"customer_id"`
	AdminID          *uint             `json:"admin_id"`
	Status           ChatStatus        `json:"status"`
	Subject          string            `json:"subject"`
	Priority         ChatPriority      `json:"priority"`
	LastMessage      *ChatMessageResponse `json:"last_message,omitempty"`
	UnreadCount      int64             `json:"unread_count"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
}

type ChatRoomListResponse struct {
	Rooms      []ChatRoomResponse `json:"rooms"`
	Total      int64              `json:"total"`
	Page       int                `json:"page"`
	Limit      int                `json:"limit"`
	TotalPages int                `json:"total_pages"`
}

type UpdateChatRoomRequest struct {
	Status   *ChatStatus   `json:"status,omitempty" validate:"omitempty,oneof=active closed archived"`
	Priority *ChatPriority `json:"priority,omitempty" validate:"omitempty,oneof=low normal high urgent"`
	AdminID  *uint         `json:"admin_id,omitempty"`
}

// Chat Message DTOs
type SendMessageRequest struct {
	RoomID      uint        `json:"room_id" validate:"required"`
	Message     string      `json:"message" validate:"required,min=1,max=1000"`
	MessageType MessageType `json:"message_type" validate:"omitempty,oneof=text image file audio video"`
	Attachments []string    `json:"attachments,omitempty"`
}

type ChatMessageResponse struct {
	ID          uint        `json:"id"`
	RoomID      uint        `json:"room_id"`
	SenderID    uint        `json:"sender_id"`
	SenderType  SenderType  `json:"sender_type"`
	Message     string      `json:"message"`
	MessageType MessageType `json:"message_type"`
	Attachments []string    `json:"attachments"`
	IsRead      bool        `json:"is_read"`
	ReadAt      *time.Time  `json:"read_at"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

type ChatMessageListResponse struct {
	Messages   []ChatMessageResponse `json:"messages"`
	Total      int64                 `json:"total"`
	Page       int                   `json:"page"`
	Limit      int                   `json:"limit"`
	TotalPages int                   `json:"total_pages"`
}

// Chat Statistics DTOs
type ChatStatsResponse struct {
	TotalRooms    int64 `json:"total_rooms"`
	ActiveRooms   int64 `json:"active_rooms"`
	ClosedRooms   int64 `json:"closed_rooms"`
	UnreadMessages int64 `json:"unread_messages"`
	TodayMessages int64 `json:"today_messages"`
}

// Real-time chat DTOs
type ChatEventType string

const (
	EventNewMessage    ChatEventType = "new_message"
	EventMessageRead   ChatEventType = "message_read"
	EventRoomUpdated   ChatEventType = "room_updated"
	EventAdminJoined   ChatEventType = "admin_joined"
	EventAdminLeft     ChatEventType = "admin_left"
	EventTypingStart   ChatEventType = "typing_start"
	EventTypingStop    ChatEventType = "typing_stop"
)

type ChatEvent struct {
	Type      ChatEventType `json:"type"`
	RoomID    uint          `json:"room_id"`
	UserID    uint          `json:"user_id"`
	UserType  SenderType    `json:"user_type"`
	Data      interface{}   `json:"data"`
	Timestamp time.Time     `json:"timestamp"`
}

type TypingIndicatorRequest struct {
	RoomID   uint `json:"room_id" validate:"required"`
	IsTyping bool `json:"is_typing"`
}