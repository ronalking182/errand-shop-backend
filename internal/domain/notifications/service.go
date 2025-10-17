package notifications

import (
	"context"
	"errandShop/internal/services/firebase"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
)

type NotificationService interface {
	// Core notification methods
	CreateNotification(req *CreateNotificationRequest) (*NotificationResponse, error)
	GetNotifications(recipientID uuid.UUID, recipientType NotificationRecipient, page, limit int) (*NotificationListResponse, error)
	MarkAsRead(id uint) error
	MarkAllAsRead(recipientID uuid.UUID, recipientType NotificationRecipient) error
	DeleteNotification(id uint) error

	// Push notification methods
	SendPushNotification(req *SendPushNotificationRequest) error
	SendBroadcastNotification(req *BroadcastNotificationRequest) error
	RegisterPushToken(userID uuid.UUID, userType string, req *RegisterPushTokenRequest) (*PushTokenResponse, error)
	SendDeliveryUpdate(deliveryID uint, status string, customerID uuid.UUID) error

	// Template methods
	CreateTemplate(req *CreateTemplateRequest) (*TemplateResponse, error)
	GetTemplates() ([]TemplateResponse, error)
	UpdateTemplate(id uint, req *CreateTemplateRequest) (*TemplateResponse, error)
	DeleteTemplate(id uint) error
}

type notificationService struct {
	notificationRepo NotificationRepository
	templateRepo     TemplateRepository
	pushTokenRepo    PushTokenRepository
	fcmService       *firebase.FCMService
}

func NewNotificationService(
	notificationRepo NotificationRepository,
	templateRepo TemplateRepository,
	pushTokenRepo PushTokenRepository,
) NotificationService {
	// Initialize FCM service (optional - will log if not configured)
	fcmService, err := firebase.NewFCMService()
	if err != nil {
		log.Printf("Warning: FCM service not initialized: %v", err)
		log.Println("Push notifications will be logged only. Set FIREBASE_CREDENTIALS_PATH or FIREBASE_CREDENTIALS_JSON to enable FCM.")
	}

	return &notificationService{
		notificationRepo: notificationRepo,
		templateRepo:     templateRepo,
		pushTokenRepo:    pushTokenRepo,
		fcmService:       fcmService,
	}
}

// Core notification methods
func (s *notificationService) CreateNotification(req *CreateNotificationRequest) (*NotificationResponse, error) {
	notification := &Notification{
		RecipientID:   req.RecipientID,
		RecipientType: req.RecipientType,
		Type:          req.Type,
		Title:         req.Title,
		Body:          req.Body,
		Data:          req.Data,
		Status:        StatusPending,
	}

	if err := s.notificationRepo.Create(notification); err != nil {
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	// Send push notification asynchronously
	go s.sendPushToUser(req.RecipientID, string(req.RecipientType), req.Title, req.Body, req.Data)

	return s.toNotificationResponse(notification), nil
}

func (s *notificationService) GetNotifications(recipientID uuid.UUID, recipientType NotificationRecipient, page, limit int) (*NotificationListResponse, error) {
	notifications, total, err := s.notificationRepo.GetByRecipient(recipientID, recipientType, page, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get notifications: %w", err)
	}

	unreadCount, err := s.notificationRepo.GetUnreadCount(recipientID, recipientType)
	if err != nil {
		log.Printf("Failed to get unread count: %v", err)
		unreadCount = 0
	}

	responses := make([]NotificationResponse, len(notifications))
	for i, notification := range notifications {
		responses[i] = *s.toNotificationResponse(&notification)
	}

	return &NotificationListResponse{
		Notifications: responses,
		Total:         total,
		Page:          page,
		Limit:         limit,
		UnreadCount:   unreadCount,
	}, nil
}

func (s *notificationService) MarkAsRead(id uint) error {
	return s.notificationRepo.MarkAsRead(id)
}

func (s *notificationService) MarkAllAsRead(recipientID uuid.UUID, recipientType NotificationRecipient) error {
	return s.notificationRepo.MarkAllAsRead(recipientID, recipientType)
}

func (s *notificationService) DeleteNotification(id uint) error {
	return s.notificationRepo.Delete(id)
}

// Push notification methods
func (s *notificationService) SendPushNotification(req *SendPushNotificationRequest) error {
	return s.sendPushToUser(req.UserID, req.UserType, req.Title, req.Body, req.Data)
}

func (s *notificationService) SendBroadcastNotification(req *BroadcastNotificationRequest) error {
	log.Printf("[BROADCAST] Sending broadcast notification: %s - %s", req.Title, req.Body)
	
	// Get all active push tokens
	tokens, err := s.pushTokenRepo.GetActiveTokens(uuid.Nil, "")
	if err != nil {
		// Try a different approach - get all tokens by querying without user filter
		log.Printf("[BROADCAST] Failed to get active tokens, trying alternative approach: %v", err)
		// For now, we'll use FCM topic-based broadcasting
		if s.fcmService != nil {
			ctx := context.Background()
			// Convert data map from interface{} to string for FCM
			fcmData := make(map[string]string)
			for k, v := range req.Data {
				if str, ok := v.(string); ok {
					fcmData[k] = str
				} else {
					fcmData[k] = fmt.Sprintf("%v", v)
				}
			}
			
			// Use topic-based broadcasting to "all_users" topic
			_, err := s.fcmService.SendToTopic(ctx, "all_users", req.Title, req.Body, fcmData, "")
			if err != nil {
				log.Printf("[BROADCAST] FCM topic broadcast failed: %v", err)
				return fmt.Errorf("failed to send broadcast notification: %w", err)
			}
			log.Printf("[BROADCAST] Successfully sent broadcast notification via FCM topic")
			return nil
		}
		
		// Fallback: log the broadcast (FCM not configured)
		log.Printf("[BROADCAST] FCM not configured, broadcast logged only: Title=%s, Body=%s", req.Title, req.Body)
		return nil
	}
	
	if len(tokens) == 0 {
		log.Printf("[BROADCAST] No active push tokens found")
		return nil
	}
	
	// Send push notifications via FCM if available
	if s.fcmService != nil {
		// Convert data map from interface{} to string for FCM
		fcmData := make(map[string]string)
		for k, v := range req.Data {
			if str, ok := v.(string); ok {
				fcmData[k] = str
			} else {
				fcmData[k] = fmt.Sprintf("%v", v)
			}
		}
		
		// Extract token strings for multicast
		tokenStrings := make([]string, len(tokens))
		for i, token := range tokens {
			tokenStrings[i] = token.Token
		}
		
		ctx := context.Background()
		_, err := s.fcmService.SendMulticast(ctx, tokenStrings, req.Title, req.Body, fcmData, "")
		if err != nil {
			log.Printf("[BROADCAST] FCM multicast failed: %v", err)
			return fmt.Errorf("failed to send broadcast notification: %w", err)
		}
		log.Printf("[BROADCAST] Successfully sent broadcast notification to %d tokens via FCM multicast", len(tokens))
		return nil
	}
	
	// Fallback: log the broadcast (FCM not configured)
	log.Printf("[BROADCAST] FCM not configured - Logging broadcast notification to %d tokens: %s - %s", len(tokens), req.Title, req.Body)
	return nil
}

func (s *notificationService) RegisterPushToken(userID uuid.UUID, userType string, req *RegisterPushTokenRequest) (*PushTokenResponse, error) {
	// Check for existing tokens and deactivate duplicates
	existingTokens, err := s.pushTokenRepo.GetByUserID(userID, userType)
	if err == nil {
		for _, token := range existingTokens {
			// Fix: Use req.DeviceID instead of req.Device
			if token.DeviceID == req.DeviceID && req.DeviceID != "" {
				s.pushTokenRepo.Deactivate(token.ID)
			}
		}
	}

	var platform DevicePlatform
	switch req.DeviceType {
	case "ios":
		platform = PlatformIOS
	case "android":
		platform = PlatformAndroid
	case "web":
		platform = PlatformWeb
	default:
		return nil, fmt.Errorf("invalid device type: %s", req.DeviceType)
	}

	pushToken := &PushToken{
		UserID:     userID,
		UserType:   userType,
		Token:      req.Token,
		Platform:   platform,
		DeviceID:   req.DeviceID,
		IsActive:   true,
		LastUsedAt: time.Now(),
	}

	if err := s.pushTokenRepo.Create(pushToken); err != nil {
		return nil, fmt.Errorf("failed to register push token: %w", err)
	}

	return &PushTokenResponse{
		ID:         pushToken.ID,
		Token:      pushToken.Token,
		Platform:   pushToken.Platform,
		DeviceID:   pushToken.DeviceID,
		IsActive:   pushToken.IsActive,
		LastUsedAt: pushToken.LastUsedAt,
		CreatedAt:  pushToken.CreatedAt,
	}, nil
}

func (s *notificationService) SendDeliveryUpdate(deliveryID uint, status string, customerID uuid.UUID) error {
	// Get template for delivery updates
	template, err := s.templateRepo.GetByType(TypeDeliveryUpdate)
	if err != nil {
		// Use default template if not found
		template = &NotificationTemplate{
			Title: "Delivery Update",
			Body:  "Your delivery status has been updated to: {{.status}}",
		}
	}

	// Replace placeholders
	title := strings.ReplaceAll(template.Title, "{{.status}}", status)
	body := strings.ReplaceAll(template.Body, "{{.status}}", status)

	// Create notification
	notificationReq := &CreateNotificationRequest{
		RecipientID:   customerID,
		RecipientType: RecipientCustomer,
		Type:          TypeDeliveryUpdate,
		Title:         title,
		Body:          body,
		Data: map[string]interface{}{
			"deliveryId": deliveryID,
			"status":     status,
		},
	}

	_, err = s.CreateNotification(notificationReq)
	return err
}

// Template methods
func (s *notificationService) CreateTemplate(req *CreateTemplateRequest) (*TemplateResponse, error) {
	template := &NotificationTemplate{
		Type:     req.Type,
		Title:    req.Title,
		Body:     req.Body,
		IsActive: req.IsActive,
	}

	if err := s.templateRepo.Create(template); err != nil {
		return nil, fmt.Errorf("failed to create template: %w", err)
	}

	return &TemplateResponse{
		ID:        template.ID,
		Type:      template.Type,
		Title:     template.Title,
		Body:      template.Body,
		IsActive:  template.IsActive,
		CreatedAt: template.CreatedAt,
		UpdatedAt: template.UpdatedAt,
	}, nil
}

func (s *notificationService) GetTemplates() ([]TemplateResponse, error) {
	templates, err := s.templateRepo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get templates: %w", err)
	}

	responses := make([]TemplateResponse, len(templates))
	for i, template := range templates {
		responses[i] = TemplateResponse{
			ID:        template.ID,
			Type:      template.Type,
			Title:     template.Title,
			Body:      template.Body,
			IsActive:  template.IsActive,
			CreatedAt: template.CreatedAt,
			UpdatedAt: template.UpdatedAt,
		}
	}

	return responses, nil
}

func (s *notificationService) UpdateTemplate(id uint, req *CreateTemplateRequest) (*TemplateResponse, error) {
	template := &NotificationTemplate{
		Type:     req.Type,
		Title:    req.Title,
		Body:     req.Body,
		IsActive: req.IsActive,
	}

	if err := s.templateRepo.Update(id, template); err != nil {
		return nil, fmt.Errorf("failed to update template: %w", err)
	}

	updatedTemplate, err := s.templateRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated template: %w", err)
	}

	return &TemplateResponse{
		ID:        updatedTemplate.ID,
		Type:      updatedTemplate.Type,
		Title:     updatedTemplate.Title,
		Body:      updatedTemplate.Body,
		IsActive:  updatedTemplate.IsActive,
		CreatedAt: updatedTemplate.CreatedAt,
		UpdatedAt: updatedTemplate.UpdatedAt,
	}, nil
}

func (s *notificationService) DeleteTemplate(id uint) error {
	return s.templateRepo.Delete(id)
}

// Helper methods
func (s *notificationService) sendPushToUser(userID uuid.UUID, userType, title, body string, data map[string]interface{}) error {
	tokens, err := s.pushTokenRepo.GetActiveTokens(userID, userType)
	if err != nil {
		log.Printf("Failed to get push tokens for user %s: %v", userID, err)
		return err
	}

	if len(tokens) == 0 {
		log.Printf("No active push tokens found for user %s", userID)
		return nil
	}

	// Send push notifications via FCM if available
	if s.fcmService != nil {
		// Convert data map from interface{} to string for FCM
		fcmData := make(map[string]string)
		for k, v := range data {
			if str, ok := v.(string); ok {
				fcmData[k] = str
			} else {
				fcmData[k] = fmt.Sprintf("%v", v)
			}
		}

		for _, token := range tokens {
			fcmMsg := &firebase.FCMMessage{
				Token: token.Token,
				Title: title,
				Body:  body,
				Data:  fcmData,
			}

			_, err := s.fcmService.SendMessage(context.Background(), fcmMsg)
			if err != nil {
				log.Printf("Failed to send FCM to token %s: %v", token.Token, err)
			// TODO: Handle invalid tokens (remove from database)
		} else {
			log.Printf("Successfully sent FCM to user %s via token %s", userID, token.Token[:10]+"...")
		}

			// Update last used timestamp
			s.pushTokenRepo.UpdateLastUsed(token.ID)
		}
	} else {
		// Fallback: log the notification
		for _, token := range tokens {
			log.Printf("FCM not configured - Logging push notification to token %s (platform: %s): %s - %s",
				token.Token[:10]+"...", token.Platform, title, body)

			// Update last used timestamp
			s.pushTokenRepo.UpdateLastUsed(token.ID)
		}
		log.Printf("Notification data: %+v", data)
		log.Printf("Active tokens: %d", len(tokens))
	}

	return nil
}

func (s *notificationService) toNotificationResponse(notification *Notification) *NotificationResponse {
	return &NotificationResponse{
		ID:            notification.ID,
		RecipientID:   notification.RecipientID,
		RecipientType: notification.RecipientType,
		Type:          notification.Type,
		Title:         notification.Title,
		Body:          notification.Body,
		Data:          notification.Data,
		Status:        notification.Status,
		ReadAt:        notification.ReadAt,
		SentAt:        notification.SentAt,
		CreatedAt:     notification.CreatedAt,
	}
}
