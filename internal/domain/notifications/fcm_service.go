package notifications

import (
	"context"
	"fmt"
	"log"
	"time"

	"errandShop/internal/services/firebase"
	"github.com/google/uuid"
)

type FCMService interface {
	SendToSingleUser(userID uuid.UUID, userType, title, body string, data map[string]interface{}, imageURL string) (*FCMMessage, error)
	SendToMultipleUsers(userIDs []uuid.UUID, userType, title, body string, data map[string]interface{}, imageURL string) (*FCMMessage, error)
	BroadcastToAllUsers(title, body string, data map[string]interface{}, imageURL string) (*FCMMessage, error)
	RegisterToken(userID uuid.UUID, userType, token, platform, deviceID string) error
	UnregisterToken(token string) error
	GetMessages(page, limit int) ([]FCMMessage, int64, error)
	GetMessagesByUser(userID uuid.UUID, page, limit int) ([]FCMMessage, int64, error)
	GetStats() (map[string]interface{}, error)
	TestMessage(userID uuid.UUID, userType string) error
}

type fcmService struct {
	fcmTokenRepo     FCMTokenRepository
	fcmMessageRepo   FCMMessageRepository
	fcmRecipientRepo FCMMessageRecipientRepository
	firebaseService  *firebase.FCMService
}

func NewFCMService(
	fcmTokenRepo FCMTokenRepository,
	fcmMessageRepo FCMMessageRepository,
	fcmRecipientRepo FCMMessageRecipientRepository,
) FCMService {
	// Initialize Firebase FCM service
	firebaseService, err := firebase.NewFCMService()
	if err != nil {
		log.Printf("Warning: Firebase FCM service not initialized: %v", err)
		log.Println("FCM messages will be logged only. Set FIREBASE_CREDENTIALS_PATH or FIREBASE_CREDENTIALS_JSON to enable FCM.")
	}

	return &fcmService{
		fcmTokenRepo:     fcmTokenRepo,
		fcmMessageRepo:   fcmMessageRepo,
		fcmRecipientRepo: fcmRecipientRepo,
		firebaseService:  firebaseService,
	}
}

func (s *fcmService) SendToSingleUser(userID uuid.UUID, userType, title, body string, data map[string]interface{}, imageURL string) (*FCMMessage, error) {
	// Get user's active tokens
	tokens, err := s.fcmTokenRepo.GetActiveTokens(userID, userType)
	if err != nil {
		return nil, fmt.Errorf("failed to get user tokens: %v", err)
	}

	if len(tokens) == 0 {
		return nil, fmt.Errorf("no active tokens found for user %s", userID)
	}

	// Create FCM message record
	message := &FCMMessage{
		Title:       title,
		Body:        body,
		Data:        data,
		ImageURL:    imageURL,
		MessageType: "single",
		SentBy:      uuid.MustParse("00000000-0000-0000-0000-000000000001"), // TODO: Get from context/auth
		SentAt:      time.Now(),
	}

	err = s.fcmMessageRepo.Create(message)
	if err != nil {
		return nil, fmt.Errorf("failed to create message record: %v", err)
	}

	// Create recipients and send messages
	var recipients []FCMMessageRecipient
	for _, token := range tokens {
		recipient := FCMMessageRecipient{
			MessageID: message.ID,
			UserID:    userID,
			UserType:  userType,
			TokenID:   token.ID,
			Status:    "pending",
		}
		recipients = append(recipients, recipient)
	}

	err = s.fcmRecipientRepo.CreateBatch(recipients)
	if err != nil {
		return nil, fmt.Errorf("failed to create recipients: %v", err)
	}

	// Send actual FCM messages
	s.sendFCMMessages(tokens, title, body, data, imageURL, recipients)

	return message, nil
}

func (s *fcmService) SendToMultipleUsers(userIDs []uuid.UUID, userType, title, body string, data map[string]interface{}, imageURL string) (*FCMMessage, error) {
	// Create FCM message record
	message := &FCMMessage{
		Title:       title,
		Body:        body,
		Data:        data,
		ImageURL:    imageURL,
		MessageType: "multiple",
		SentBy:      uuid.MustParse("00000000-0000-0000-0000-000000000001"), // TODO: Get from context/auth
		SentAt:      time.Now(),
	}

	err := s.fcmMessageRepo.Create(message)
	if err != nil {
		return nil, fmt.Errorf("failed to create message record: %v", err)
	}

	// Get tokens for all users and create recipients
	var allTokens []FCMToken
	var recipients []FCMMessageRecipient

	for _, userID := range userIDs {
		tokens, err := s.fcmTokenRepo.GetActiveTokens(userID, userType)
		if err != nil {
			log.Printf("Failed to get tokens for user %s: %v", userID, err)
			continue
		}

		for _, token := range tokens {
			allTokens = append(allTokens, token)
			recipient := FCMMessageRecipient{
				MessageID: message.ID,
				UserID:    userID,
				UserType:  userType,
				TokenID:   token.ID,
				Status:    "pending",
			}
			recipients = append(recipients, recipient)
		}
	}

	if len(recipients) == 0 {
		return nil, fmt.Errorf("no active tokens found for any users")
	}

	err = s.fcmRecipientRepo.CreateBatch(recipients)
	if err != nil {
		return nil, fmt.Errorf("failed to create recipients: %v", err)
	}

	// Send actual FCM messages
	s.sendFCMMessages(allTokens, title, body, data, imageURL, recipients)

	return message, nil
}

func (s *fcmService) BroadcastToAllUsers(title, body string, data map[string]interface{}, imageURL string) (*FCMMessage, error) {
	// Get all active tokens
	tokens, err := s.fcmTokenRepo.GetAllActiveTokens()
	if err != nil {
		return nil, fmt.Errorf("failed to get all tokens: %v", err)
	}

	if len(tokens) == 0 {
		return nil, fmt.Errorf("no active tokens found")
	}

	// Create FCM message record
	message := &FCMMessage{
		Title:       title,
		Body:        body,
		Data:        data,
		ImageURL:    imageURL,
		MessageType: "broadcast",
		SentBy:      uuid.MustParse("00000000-0000-0000-0000-000000000001"), // TODO: Get from context/auth
		SentAt:      time.Now(),
	}

	err = s.fcmMessageRepo.Create(message)
	if err != nil {
		return nil, fmt.Errorf("failed to create message record: %v", err)
	}

	// Create recipients
	var recipients []FCMMessageRecipient
	for _, token := range tokens {
		recipient := FCMMessageRecipient{
			MessageID: message.ID,
			UserID:    token.UserID,
			UserType:  token.UserType,
			TokenID:   token.ID,
			Status:    "pending",
		}
		recipients = append(recipients, recipient)
	}

	err = s.fcmRecipientRepo.CreateBatch(recipients)
	if err != nil {
		return nil, fmt.Errorf("failed to create recipients: %v", err)
	}

	// Send actual FCM messages
	s.sendFCMMessages(tokens, title, body, data, imageURL, recipients)

	return message, nil
}

func (s *fcmService) RegisterToken(userID uuid.UUID, userType, token, platform, deviceID string) error {
	// Check if token already exists
	existingToken, err := s.fcmTokenRepo.GetByToken(token)
	if err == nil {
		// Token exists, update it
		existingToken.UserID = userID
		existingToken.UserType = userType
		existingToken.Platform = DevicePlatform(platform)
		existingToken.DeviceID = deviceID
		existingToken.IsActive = true
		existingToken.UpdatedAt = time.Now()
		return s.fcmTokenRepo.Create(existingToken) // This will update due to unique constraint
	}

	// Create new token
	newToken := &FCMToken{
		UserID:    userID,
		UserType:  userType,
		Token:     token,
		Platform:  DevicePlatform(platform),
		DeviceID:  deviceID,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return s.fcmTokenRepo.Create(newToken)
}

func (s *fcmService) UnregisterToken(token string) error {
	return s.fcmTokenRepo.DeleteByToken(token)
}

func (s *fcmService) GetMessages(page, limit int) ([]FCMMessage, int64, error) {
	return s.fcmMessageRepo.GetMessages(page, limit)
}

func (s *fcmService) GetMessagesByUser(userID uuid.UUID, page, limit int) ([]FCMMessage, int64, error) {
	return s.fcmMessageRepo.GetMessagesByUser(userID, page, limit)
}

func (s *fcmService) GetStats() (map[string]interface{}, error) {
	recipientStats, err := s.fcmRecipientRepo.GetStats()
	if err != nil {
		return nil, err
	}

	// Get total messages count
	messages, _, err := s.fcmMessageRepo.GetMessages(1, 1)
	if err != nil {
		return nil, err
	}

	// Get total tokens count
	allTokens, err := s.fcmTokenRepo.GetAllActiveTokens()
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"totalMessages":    len(messages),
		"totalActiveTokens": len(allTokens),
		"deliveryStats":    recipientStats,
	}

	return stats, nil
}

func (s *fcmService) TestMessage(userID uuid.UUID, userType string) error {
	title := "Test Notification"
	body := "This is a test message from the FCM service"
	data := map[string]interface{}{
		"type": "test",
		"timestamp": time.Now().Unix(),
	}

	_, err := s.SendToSingleUser(userID, userType, title, body, data, "")
	return err
}

// Helper method to send actual FCM messages
func (s *fcmService) sendFCMMessages(tokens []FCMToken, title, body string, data map[string]interface{}, imageURL string, recipients []FCMMessageRecipient) {
	if s.firebaseService == nil {
		// Log messages if Firebase is not configured
		for i, token := range tokens {
			log.Printf("FCM not configured - Logging message to token %s (platform: %s): %s - %s",
				token.Token[:10]+"...", token.Platform, title, body)
			
			// Update recipient status
			if i < len(recipients) {
				s.fcmRecipientRepo.UpdateStatus(recipients[i].ID, "sent", "")
			}
		}
		log.Printf("Message data: %+v", data)
		return
	}

	// Convert data to string map for FCM
	fcmData := make(map[string]string)
	for k, v := range data {
		if str, ok := v.(string); ok {
			fcmData[k] = str
		} else {
			fcmData[k] = fmt.Sprintf("%v", v)
		}
	}

	// Send to each token
	for i, token := range tokens {
		fcmMsg := &firebase.FCMMessage{
			Token:    token.Token,
			Title:    title,
			Body:     body,
			Data:     fcmData,
			ImageURL: imageURL,
		}

		_, err := s.firebaseService.SendMessage(context.Background(), fcmMsg)
		if err != nil {
			log.Printf("Failed to send FCM to token %s: %v", token.Token[:10]+"...", err)
			if i < len(recipients) {
				s.fcmRecipientRepo.UpdateStatus(recipients[i].ID, "failed", err.Error())
			}
		} else {
			log.Printf("Successfully sent FCM to token %s", token.Token[:10]+"...")
			if i < len(recipients) {
				s.fcmRecipientRepo.UpdateStatus(recipients[i].ID, "sent", "")
			}
		}
	}
}