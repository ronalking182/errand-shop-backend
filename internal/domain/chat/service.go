package chat

import (
	"fmt"
	"log"
	"math"
	"time"

	"errandShop/internal/domain/notifications"
	"github.com/google/uuid"
)

// ChatService interface
type ChatService interface {
	// Chat Room operations
	CreateChatRoom(req *CreateChatRoomRequest, senderType SenderType, senderID uint) (*ChatRoomResponse, error)
	GetChatRooms(userID uint, userType string, page, limit int) (*ChatRoomListResponse, error)
	GetChatRoom(roomID uint) (*ChatRoomResponse, error)
	UpdateChatRoom(roomID uint, req *UpdateChatRoomRequest) (*ChatRoomResponse, error)
	DeleteChatRoom(roomID uint) error
	AssignAdminToRoom(roomID, adminID uint) error
	UnassignAdminFromRoom(roomID uint) error

	// Chat Message operations
	SendMessage(message *ChatMessage) error
	SendMessageWithRequest(req *SendMessageRequest, senderType SenderType, senderID uint) (*ChatMessageResponse, error)
	GetMessages(roomID uint, page, limit int) (*ChatMessageListResponse, error)
	MarkMessagesAsRead(roomID, userID uint, userType SenderType) error
	DeleteMessage(messageID uint) error

	// Statistics and utilities
	GetChatStats() (*ChatStatsResponse, error)
	SendTypingIndicator(roomID, userID uint, userType SenderType, isTyping bool) error

	// WebSocket integration
	SetHub(hub *Hub)
}

// chatService implementation
type chatService struct {
	roomRepo        ChatRoomRepository
	messageRepo     ChatMessageRepository
	notificationSvc notifications.NotificationService
	hub             *Hub
}

// NewChatService creates a new chat service
func NewChatService(
	roomRepo ChatRoomRepository,
	messageRepo ChatMessageRepository,
	notificationSvc notifications.NotificationService,
) ChatService {
	return &chatService{
		roomRepo:        roomRepo,
		messageRepo:     messageRepo,
		notificationSvc: notificationSvc,
	}
}

// SetHub sets the WebSocket hub for real-time messaging
func (s *chatService) SetHub(hub *Hub) {
	s.hub = hub
}



// Chat Room operations
func (s *chatService) CreateChatRoom(req *CreateChatRoomRequest, senderType SenderType, senderID uint) (*ChatRoomResponse, error) {
	// Create the chat room
	room := &ChatRoom{
		CustomerID: req.CustomerID,
		Subject:    req.Subject,
		Priority:   req.Priority,
		Status:     ChatStatusActive,
	}

	// If admin is creating the room, assign them immediately
	if senderType == SenderTypeAdmin {
		room.AdminID = &senderID
	}

	if err := s.roomRepo.Create(room); err != nil {
		return nil, fmt.Errorf("failed to create chat room: %w", err)
	}

	// Send the initial message
	initialMessage := &ChatMessage{
		RoomID:      room.ID,
		SenderID:    senderID,
		SenderType:  senderType,
		Message:     req.Message,
		MessageType: MessageTypeText,
	}

	if err := s.messageRepo.Create(initialMessage); err != nil {
		return nil, fmt.Errorf("failed to create initial message: %w", err)
	}

	// Send notification to admins if customer created the room
	if senderType == SenderTypeCustomer {
		go s.notifyNewChatRoom(room, initialMessage)
	}

	return s.toChatRoomResponse(room, initialMessage, 1), nil
}

func (s *chatService) GetChatRooms(userID uint, userType string, page, limit int) (*ChatRoomListResponse, error) {
	var rooms []ChatRoom
	var total int64
	var err error

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	switch userType {
	case "customer":
		rooms, total, err = s.roomRepo.GetByCustomerID(userID, page, limit)
	case "admin", "superadmin":
		rooms, total, err = s.roomRepo.GetAll(page, limit)
	default:
		return nil, fmt.Errorf("invalid user type: %s", userType)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get chat rooms: %w", err)
	}

	// Convert to response format
	roomResponses := make([]ChatRoomResponse, len(rooms))
	for i, room := range rooms {
		var lastMessage *ChatMessage
		if len(room.Messages) > 0 {
			lastMessage = &room.Messages[0]
		}

		// Get unread count for this user
		unreadCount, _ := s.messageRepo.GetUnreadCount(room.ID, userID, SenderType(userType))

		roomResponses[i] = *s.toChatRoomResponse(&room, lastMessage, unreadCount)
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return &ChatRoomListResponse{
		Rooms:      roomResponses,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

func (s *chatService) GetChatRoom(roomID uint) (*ChatRoomResponse, error) {
	room, err := s.roomRepo.GetByID(roomID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat room: %w", err)
	}

	var lastMessage *ChatMessage
	if len(room.Messages) > 0 {
		lastMessage = &room.Messages[0]
	}

	return s.toChatRoomResponse(room, lastMessage, 0), nil
}

func (s *chatService) UpdateChatRoom(roomID uint, req *UpdateChatRoomRequest) (*ChatRoomResponse, error) {
	updates := make(map[string]interface{})

	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.Priority != nil {
		updates["priority"] = *req.Priority
	}
	if req.AdminID != nil {
		updates["admin_id"] = *req.AdminID
	}

	if len(updates) == 0 {
		return nil, fmt.Errorf("no updates provided")
	}

	updates["updated_at"] = time.Now()

	if err := s.roomRepo.Update(roomID, updates); err != nil {
		return nil, fmt.Errorf("failed to update chat room: %w", err)
	}

	return s.GetChatRoom(roomID)
}

func (s *chatService) DeleteChatRoom(roomID uint) error {
	return s.roomRepo.Delete(roomID)
}

func (s *chatService) AssignAdminToRoom(roomID, adminID uint) error {
	if err := s.roomRepo.AssignAdmin(roomID, adminID); err != nil {
		return fmt.Errorf("failed to assign admin to room: %w", err)
	}

	// Send system message about admin joining
	systemMessage := &ChatMessage{
		RoomID:      roomID,
		SenderID:    adminID,
		SenderType:  SenderTypeSystem,
		Message:     "Admin has joined the conversation",
		MessageType: MessageTypeText,
	}

	return s.messageRepo.Create(systemMessage)
}

func (s *chatService) UnassignAdminFromRoom(roomID uint) error {
	return s.roomRepo.UnassignAdmin(roomID)
}

// Chat Message operations
// SendMessage saves a message and broadcasts it via WebSocket (used by WebSocket handler)
func (s *chatService) SendMessage(message *ChatMessage) error {
	// Save message to database
	if err := s.messageRepo.Create(message); err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}

	// Update room's updated_at timestamp
	s.roomRepo.Update(message.RoomID, map[string]interface{}{
		"updated_at": time.Now(),
	})

	// Broadcast via WebSocket if hub is available
	if s.hub != nil {
		roomUUID, err := uuid.Parse(fmt.Sprintf("%08d-0000-0000-0000-000000000000", message.RoomID))
		if err == nil {
			senderUUID, _ := uuid.Parse(fmt.Sprintf("%08d-0000-0000-0000-000000000000", message.SenderID))
			wsMsg := WebSocketMessage{
				Type:      "new_message",
				RoomID:    &roomUUID,
				SenderID:  senderUUID,
				Message:   message.Message,
				Timestamp: time.Now().Unix(),
				Data: map[string]interface{}{
					"message_id":   message.ID,
					"message_type": message.MessageType,
					"sender_type":  message.SenderType,
					"attachments":  message.Attachments,
					"created_at":   message.CreatedAt,
				},
			}
			s.hub.BroadcastToRoom(roomUUID, wsMsg)
		}
	}

	return nil
}

// SendMessageWithRequest handles REST API message sending
func (s *chatService) SendMessageWithRequest(req *SendMessageRequest, senderType SenderType, senderID uint) (*ChatMessageResponse, error) {
	// Verify room exists
	room, err := s.roomRepo.GetByID(req.RoomID)
	if err != nil {
		return nil, fmt.Errorf("chat room not found: %w", err)
	}

	// Create message
	message := &ChatMessage{
		RoomID:      req.RoomID,
		SenderID:    senderID,
		SenderType:  senderType,
		Message:     req.Message,
		MessageType: req.MessageType,
		Attachments: req.Attachments,
	}

	if message.MessageType == "" {
		message.MessageType = MessageTypeText
	}

	// Use the new SendMessage method for consistency
	if err := s.SendMessage(message); err != nil {
		return nil, err
	}

	// Send push notification to the other party
	go s.sendMessageNotification(room, message)

	return s.toChatMessageResponse(message), nil
}

func (s *chatService) GetMessages(roomID uint, page, limit int) (*ChatMessageListResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}

	messages, total, err := s.messageRepo.GetByRoomID(roomID, page, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	// Convert to response format
	messageResponses := make([]ChatMessageResponse, len(messages))
	for i, message := range messages {
		messageResponses[i] = *s.toChatMessageResponse(&message)
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return &ChatMessageListResponse{
		Messages:   messageResponses,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

func (s *chatService) MarkMessagesAsRead(roomID, userID uint, userType SenderType) error {
	return s.messageRepo.MarkRoomMessagesAsRead(roomID, userID, userType)
}

func (s *chatService) DeleteMessage(messageID uint) error {
	return s.messageRepo.Delete(messageID)
}

// Statistics and utilities
func (s *chatService) GetChatStats() (*ChatStatsResponse, error) {
	return s.roomRepo.GetStats()
}

func (s *chatService) SendTypingIndicator(roomID, userID uint, userType SenderType, isTyping bool) error {
	// This would typically be handled by WebSocket connections
	// For now, we'll just log it
	log.Printf("User %d (%s) typing in room %d: %v", userID, userType, roomID, isTyping)
	return nil
}

// Helper methods
func (s *chatService) toChatRoomResponse(room *ChatRoom, lastMessage *ChatMessage, unreadCount int64) *ChatRoomResponse {
	response := &ChatRoomResponse{
		ID:          room.ID,
		CustomerID:  room.CustomerID,
		AdminID:     room.AdminID,
		Status:      room.Status,
		Subject:     room.Subject,
		Priority:    room.Priority,
		UnreadCount: unreadCount,
		CreatedAt:   room.CreatedAt,
		UpdatedAt:   room.UpdatedAt,
	}

	if lastMessage != nil {
		response.LastMessage = s.toChatMessageResponse(lastMessage)
	}

	return response
}

func (s *chatService) toChatMessageResponse(message *ChatMessage) *ChatMessageResponse {
	return &ChatMessageResponse{
		ID:          message.ID,
		RoomID:      message.RoomID,
		SenderID:    message.SenderID,
		SenderType:  message.SenderType,
		Message:     message.Message,
		MessageType: message.MessageType,
		Attachments: message.Attachments,
		IsRead:      message.IsRead,
		ReadAt:      message.ReadAt,
		CreatedAt:   message.CreatedAt,
		UpdatedAt:   message.UpdatedAt,
	}
}

// Notification helpers
func (s *chatService) notifyNewChatRoom(room *ChatRoom, message *ChatMessage) {
	// Create notification for admins
	notificationReq := &notifications.CreateNotificationRequest{
		RecipientID:   uuid.MustParse("00000000-0000-0000-0000-000000000001"), // System admin UUID
		RecipientType: notifications.RecipientAdmin,
		Type:          notifications.TypeOrderUpdate, // Using existing notification type
		Title:         "New Chat Request",
		Body:          fmt.Sprintf("New chat from customer: %s", room.Subject),
		Data: map[string]interface{}{
			"room_id":     room.ID,
			"customer_id": room.CustomerID,
			"priority":    room.Priority,
		},
	}

	if _, err := s.notificationSvc.CreateNotification(notificationReq); err != nil {
		log.Printf("Failed to create chat notification: %v", err)
	}
}

func (s *chatService) sendMessageNotification(room *ChatRoom, message *ChatMessage) {
	var recipientUUID uuid.UUID
	var recipientType notifications.NotificationRecipient
	var title, body string

	// Determine recipient based on sender
	if message.SenderType == SenderTypeCustomer {
		// Notify admin - use system admin UUID for now
		recipientUUID = uuid.MustParse("00000000-0000-0000-0000-000000000001")
		recipientType = notifications.RecipientAdmin
		title = "New Message"
		body = fmt.Sprintf("Customer: %s", message.Message)
	} else {
		// Notify customer - use system customer UUID for now
		// TODO: Implement proper customer UUID lookup
		recipientUUID = uuid.MustParse("00000000-0000-0000-0000-000000000002")
		recipientType = notifications.RecipientCustomer
		title = "Admin Reply"
		body = fmt.Sprintf("Admin: %s", message.Message)
	}

	// Create notification
	notificationReq := &notifications.CreateNotificationRequest{
		RecipientID:   recipientUUID,
		RecipientType: recipientType,
		Type:          notifications.TypeOrderUpdate, // Using existing notification type
		Title:         title,
		Body:          body,
		Data: map[string]interface{}{
			"room_id":    room.ID,
			"message_id": message.ID,
		},
	}

	if _, err := s.notificationSvc.CreateNotification(notificationReq); err != nil {
		log.Printf("Failed to create message notification: %v", err)
	}

	// Send push notification via notification service
	go s.sendFCMNotification(recipientUUID, recipientType, title, body, map[string]string{
		"room_id":    fmt.Sprintf("%d", room.ID),
		"message_id": fmt.Sprintf("%d", message.ID),
		"type":       "chat_message",
	})
}

func (s *chatService) sendFCMNotification(userID uuid.UUID, userType notifications.NotificationRecipient, title, body string, data map[string]string) {
	// Use the existing notification service to send push notifications
	notificationData := make(map[string]interface{})
	for k, v := range data {
		notificationData[k] = v
	}

	// Convert recipient type to string
	var userTypeStr string
	switch userType {
	case notifications.RecipientAdmin:
		userTypeStr = "admin"
	case notifications.RecipientCustomer:
		userTypeStr = "customer"
	default:
		userTypeStr = "customer"
	}

	// Send push notification via the notification service
	req := &notifications.SendPushNotificationRequest{
		UserID:   userID,
		UserType: userTypeStr,
		Title:    title,
		Body:     body,
		Data:     notificationData,
	}

	err := s.notificationSvc.SendPushNotification(req)
	if err != nil {
		log.Printf("Failed to send push notification: %v", err)
	}
}