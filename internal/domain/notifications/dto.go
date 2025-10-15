package notifications

import (
	"errors"
	"time"
	"github.com/google/uuid"
)

// Request DTOs
type CreateNotificationRequest struct {
	RecipientID   uuid.UUID              `json:"recipientId" validate:"required"`
	RecipientType NotificationRecipient  `json:"recipientType" validate:"required,oneof=customer admin driver"`
	Type          NotificationType       `json:"type" validate:"required"`
	Title         string                 `json:"title" validate:"required,max=200"`
	Body          string                 `json:"body" validate:"required"`
	Data          map[string]interface{} `json:"data,omitempty"`
}

type SendPushNotificationRequest struct {
	UserID   uuid.UUID              `json:"userId" validate:"required"`
	UserType string                 `json:"userType" validate:"required"`
	Title    string                 `json:"title" validate:"required,max=200"`
	Body     string                 `json:"body" validate:"required"`
	Data     map[string]interface{} `json:"data,omitempty"`
}

type BroadcastNotificationRequest struct {
	Title    string                 `json:"title" validate:"required,max=200"`
	Body     string                 `json:"body" validate:"required"`
	Data     map[string]interface{} `json:"data,omitempty"`
}

type RegisterPushTokenRequest struct {
	Token      string `json:"token" validate:"required"`
	DeviceType string `json:"device_type" validate:"required,oneof=ios android web"`
	DeviceID   string `json:"device_id,omitempty"`
}

func (r *RegisterPushTokenRequest) Validate() error {
	if r.Token == "" {
		return errors.New("token is required")
	}
	if r.DeviceType == "" {
		return errors.New("device_type is required")
	}
	if r.DeviceType != "ios" && r.DeviceType != "android" && r.DeviceType != "web" {
		return errors.New("device_type must be ios, android, or web")
	}
	return nil
}

type UpdateNotificationStatusRequest struct {
	Status NotificationStatus `json:"status" validate:"required,oneof=read"`
}

type CreateTemplateRequest struct {
	Type     NotificationType `json:"type" validate:"required"`
	Title    string           `json:"title" validate:"required,max=200"`
	Body     string           `json:"body" validate:"required"`
	IsActive bool             `json:"isActive"`
}

// Response DTOs
type NotificationResponse struct {
	ID            uint                   `json:"id"`
	RecipientID   uuid.UUID              `json:"recipientId"`
	RecipientType NotificationRecipient  `json:"recipientType"`
	Type          NotificationType       `json:"type"`
	Title         string                 `json:"title"`
	Body          string                 `json:"body"`
	Data          map[string]interface{} `json:"data,omitempty"`
	Status        NotificationStatus     `json:"status"`
	ReadAt        *time.Time             `json:"readAt,omitempty"`
	SentAt        *time.Time             `json:"sentAt,omitempty"`
	CreatedAt     time.Time              `json:"createdAt"`
}

type NotificationListResponse struct {
	Notifications []NotificationResponse `json:"notifications"`
	Total         int64                  `json:"total"`
	Page          int                    `json:"page"`
	Limit         int                    `json:"limit"`
	UnreadCount   int64                  `json:"unreadCount"`
}

type TemplateResponse struct {
	ID        uint             `json:"id"`
	Type      NotificationType `json:"type"`
	Title     string           `json:"title"`
	Body      string           `json:"body"`
	IsActive  bool             `json:"isActive"`
	CreatedAt time.Time        `json:"createdAt"`
	UpdatedAt time.Time        `json:"updatedAt"`
}

type PushTokenResponse struct {
	ID         uint           `json:"id"`
	Token      string         `json:"token"`
	Platform   DevicePlatform `json:"platform"`
	DeviceID   string         `json:"deviceId,omitempty"`
	IsActive   bool           `json:"isActive"`
	LastUsedAt time.Time      `json:"lastUsedAt"`
	CreatedAt  time.Time      `json:"createdAt"`
}
