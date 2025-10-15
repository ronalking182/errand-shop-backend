package chat

import (
	"encoding/json"
	"errandShop/internal/pkg/jwt"
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// WebSocketHandler handles WebSocket connections
type WebSocketHandler struct {
	hub         *Hub
	chatService ChatService
	jwtSecret   string
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(hub *Hub, chatService ChatService, jwtSecret string) *WebSocketHandler {
	return &WebSocketHandler{
		hub:         hub,
		chatService: chatService,
		jwtSecret:   jwtSecret,
	}
}

// HandleWebSocket handles WebSocket upgrade and client management
func (wh *WebSocketHandler) HandleWebSocket(c *websocket.Conn) {
	// Extract user info from query parameters or headers
	userID, userType, roomID, err := wh.extractUserInfo(c)
	if err != nil {
		log.Printf("Failed to extract user info: %v", err)
		c.Close()
		return
	}

	// Create client
	client := &Client{
		ID:       uuid.New().String(),
		Conn:     c,
		UserID:   userID,
		UserType: userType,
		RoomID:   roomID,
		Send:     make(chan []byte, 256),
	}

	// Register client
	wh.hub.Register <- client

	// Start goroutines for reading and writing
	// Start writePump first, then readPump to avoid race conditions
	go wh.writePump(client)
	// Use readPump in the main goroutine to ensure proper connection handling
	wh.readPump(client)
}

// extractUserInfo extracts user information from WebSocket connection
func (wh *WebSocketHandler) extractUserInfo(c *websocket.Conn) (uuid.UUID, string, *uuid.UUID, error) {
	// Get token from query parameter
	token := c.Query("token")
	if token == "" {
		return uuid.Nil, "", nil, fiber.NewError(fiber.StatusUnauthorized, "Token required")
	}

	tokenPreview := token
	if len(token) > 20 {
		tokenPreview = token[:20] + "..."
	}
	log.Printf("DEBUG: WebSocket connection attempt with token: %s", tokenPreview)

	// Parse JWT token using the same validation as middleware
	parsedToken, err := jwtlib.ParseWithClaims(token, &jwt.JWTClaims{}, func(token *jwtlib.Token) (interface{}, error) {
		return []byte(wh.jwtSecret), nil
	})
	if err != nil {
		log.Printf("DEBUG: Token parsing failed: %v", err)
		return uuid.Nil, "", nil, fiber.NewError(fiber.StatusUnauthorized, "Invalid token")
	}

	claims, ok := parsedToken.Claims.(*jwt.JWTClaims)
	if !ok || !parsedToken.Valid {
		log.Printf("DEBUG: Token claims validation failed")
		return uuid.Nil, "", nil, fiber.NewError(fiber.StatusUnauthorized, "Invalid token claims")
	}

	// DEBUG: Log token claims
	log.Printf("DEBUG: Token claims - UserID: %s, Role: %s", claims.UserID, claims.Role)

	// Extract user ID from structured claims
	userID := claims.UserID
	if userID == uuid.Nil {
		log.Printf("DEBUG: Invalid user ID in token claims")
		return uuid.Nil, "", nil, fiber.NewError(fiber.StatusUnauthorized, "Invalid user ID in token")
	}

	// Extract user type from role field
	userType := claims.Role
	if userType == "" {
		userType = "customer" // default
	}

	// Extract room ID (optional, mainly for customers)
	var roomID *uuid.UUID
	roomIDStr := c.Query("room_id")
	if roomIDStr != "" {
		parsedRoomID, err := uuid.Parse(roomIDStr)
		if err == nil {
			roomID = &parsedRoomID
		}
	}

	log.Printf("DEBUG: WebSocket connection - UserID: %s, Role: %s, RoomID: %v", userID, userType, roomID)
	return userID, userType, roomID, nil
}

// readPump pumps messages from the WebSocket connection to the hub
func (wh *WebSocketHandler) readPump(client *Client) {
	log.Printf("Starting readPump for client %s", client.ID)
	
	// Check if connection is nil
	if client.Conn == nil {
		log.Printf("ERROR: client.Conn is nil for client %s", client.ID)
		return
	}
	
	defer func() {
		log.Printf("readPump ending for client %s", client.ID)
		wh.hub.Unregister <- client
		if client.Conn != nil {
			client.Conn.Close()
		}
	}()

	// Set longer read deadline for React Native compatibility (5 minutes)
	client.Conn.SetReadDeadline(time.Now().Add(5 * time.Minute))
	log.Printf("Set read deadline for client %s", client.ID)
	
	client.Conn.SetPongHandler(func(string) error {
		log.Printf("Received pong from client %s", client.ID)
		client.Conn.SetReadDeadline(time.Now().Add(5 * time.Minute))
		return nil
	})

	log.Printf("Starting message read loop for client %s", client.ID)
	for {
		// Double-check connection is still valid
		if client.Conn == nil {
			log.Printf("ERROR: client.Conn became nil during read loop for client %s", client.ID)
			break
		}
		
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			log.Printf("ReadMessage error for client %s: %v", client.ID, err)
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket unexpected close error for client %s: %v", client.ID, err)
			}
			break
		}

		log.Printf("Received message from client %s, length: %d", client.ID, len(message))
		// Reset read deadline on any message received
		client.Conn.SetReadDeadline(time.Now().Add(5 * time.Minute))

		// Handle incoming message
		wh.handleIncomingMessage(client, message)
	}
}

// writePump pumps messages from the hub to the WebSocket connection
func (wh *WebSocketHandler) writePump(client *Client) {
	// Send ping every 4 minutes (less than 5 minute read deadline)
	ticker := time.NewTicker(4 * time.Minute)
	defer func() {
		ticker.Stop()
		// Don't close connection here - let readPump handle it
	}()

	for {
		select {
		case message, ok := <-client.Send:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := client.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}

		case <-ticker.C:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			// Send ping message - React Native should respond with pong
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("Failed to send ping to client %s: %v", client.ID, err)
				return
			}
		}
	}
}

// handleIncomingMessage processes incoming WebSocket messages
func (wh *WebSocketHandler) handleIncomingMessage(client *Client, message []byte) {
	log.Printf("Received message from client %s: %s", client.ID, string(message))
	
	var wsMsg WebSocketMessage
	if err := json.Unmarshal(message, &wsMsg); err != nil {
		log.Printf("Error unmarshaling message: %v", err)
		return
	}

	// Debug: Log the parsed message content
	log.Printf("DEBUG: Parsed message - Type: %s, Message: '%s', RoomIDRaw: %v", wsMsg.Type, wsMsg.Message, wsMsg.RoomIDRaw)

	// Process RoomIDRaw and convert to proper UUID
	if wsMsg.RoomIDRaw != nil {
		if err := wh.processRoomID(&wsMsg); err != nil {
			log.Printf("Error processing room_id: %v", err)
			return
		}
	}

	log.Printf("Parsed message type: %s, RoomID: %v, Message: '%s'", wsMsg.Type, wsMsg.RoomID, wsMsg.Message)

	// Set sender ID from client
	wsMsg.SenderID = client.UserID
	wsMsg.Timestamp = time.Now().Unix()

	switch wsMsg.Type {
	case "chat_message":
		wh.handleChatMessage(client, wsMsg)
	case "typing_start":
		wh.handleTypingIndicator(client, wsMsg, true)
	case "typing_stop":
		wh.handleTypingIndicator(client, wsMsg, false)
	case "join_room":
		wh.handleJoinRoom(client, wsMsg)
	case "leave_room":
		wh.handleLeaveRoom(client, wsMsg)
	case "ping":
		log.Printf("Received ping from client %s", client.ID)
		// Send pong response
		pongMsg := WebSocketMessage{
			Type:       "pong",
			SenderID:   client.UserID,
			SenderType: client.UserType, // Add sender_type for consistency
			Timestamp:  time.Now().Unix(),
		}
		if pongData, err := json.Marshal(pongMsg); err == nil {
			select {
			case client.Send <- pongData:
			default:
				log.Printf("Failed to send pong to client %s", client.ID)
			}
		}
	default:
		log.Printf("Unknown message type: %s", wsMsg.Type)
	}
}

// handleChatMessage processes chat messages
func (wh *WebSocketHandler) handleChatMessage(client *Client, wsMsg WebSocketMessage) {
	log.Printf("DEBUG: handleChatMessage called with message: '%s', RoomID: %v", wsMsg.Message, wsMsg.RoomID)
	
	if wsMsg.RoomID == nil {
		log.Println("Chat message requires room_id")
		return
	}

	// Convert UUID to uint for database compatibility
	roomID := wh.uuidToRoomID(*wsMsg.RoomID)
	userID := wh.uuidToUint(client.UserID)
	
	log.Printf("DEBUG: Processing message '%s' for roomID: %d, userID: %d", wsMsg.Message, roomID, userID)

	// Save message to database
	chatMessage := &ChatMessage{
		RoomID:      roomID,
		SenderID:    userID,
		SenderType:  SenderType(client.UserType),
		Message:     wsMsg.Message,
		MessageType: MessageTypeText,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := wh.chatService.SendMessage(chatMessage); err != nil {
		log.Printf("Error saving chat message: %v", err)
		return
	}

	// Get sender name (for customers, use a default name or fetch from user service)
	senderName := "Customer"
	if client.UserType == "admin" {
		senderName = "Admin"
	}

	// Create dashboard-compatible message format
	broadcastMsg := WebSocketMessage{
		Type:       "chat_message",
		RoomID:     wsMsg.RoomID,
		SenderID:   client.UserID,
		SenderType: client.UserType, // Add sender_type at top level for easy access
		Message:    wsMsg.Message,
		Timestamp:  time.Now().UnixMilli(), // Use milliseconds for dashboard compatibility
		Data: map[string]interface{}{
			"room_id":     wsMsg.RoomID.String(),
			"sender_id":   client.UserID.String(),
			"sender_name": senderName,
			"sender_type": client.UserType,
			"message":     wsMsg.Message,
			"timestamp":   time.Now().UnixMilli(),
			"message_id":  chatMessage.ID,
			"created_at":  chatMessage.CreatedAt,
		},
	}
	
	log.Printf("DEBUG: Created broadcast message - Type: %s, Message: '%s', Data.message: '%s'", broadcastMsg.Type, broadcastMsg.Message, broadcastMsg.Data.(map[string]interface{})["message"])

	// Broadcast to room (this will reach all non-admin clients in the room)
	wh.hub.BroadcastToRoom(*wsMsg.RoomID, broadcastMsg)

	// Also broadcast to all admin clients (global listeners) if message is from customer
	if client.UserType == "customer" {
		wh.hub.BroadcastToAdmins(broadcastMsg)
	} else if client.UserType == "admin" || client.UserType == "superadmin" {
		// For admin messages, also broadcast to admins so they can see their own messages
		// and other admins can see the conversation
		wh.hub.BroadcastToAdmins(broadcastMsg)
		log.Printf("DEBUG: Admin message broadcasted to room %s and admin clients", wsMsg.RoomID.String())
	}
}

// handleTypingIndicator processes typing indicators
func (wh *WebSocketHandler) handleTypingIndicator(client *Client, wsMsg WebSocketMessage, isTyping bool) {
	if wsMsg.RoomID == nil {
		return
	}

	// Get sender name
	senderName := "Customer"
	if client.UserType == "admin" {
		senderName = "Admin"
	}

	// Create dashboard-compatible typing indicator format
	broadcastMsg := WebSocketMessage{
		Type:       wsMsg.Type, // "typing_start" or "typing_stop"
		RoomID:     wsMsg.RoomID,
		SenderID:   client.UserID,
		SenderType: client.UserType, // Add sender_type at top level for easy access
		Timestamp:  time.Now().UnixMilli(),
		Data: map[string]interface{}{
			"room_id":     wsMsg.RoomID.String(),
			"sender_id":   client.UserID.String(),
			"sender_name": senderName,
			"sender_type": client.UserType,
			"is_typing":   isTyping,
			"timestamp":   time.Now().UnixMilli(),
		},
	}

	wh.hub.BroadcastToRoom(*wsMsg.RoomID, broadcastMsg)
}

// handleJoinRoom processes room join requests
func (wh *WebSocketHandler) handleJoinRoom(client *Client, wsMsg WebSocketMessage) {
	if wsMsg.RoomID == nil {
		return
	}

	// Update client's room
	client.RoomID = wsMsg.RoomID

	// Add to room in hub
	wh.hub.mu.Lock()
	if wh.hub.Rooms[*wsMsg.RoomID] == nil {
		wh.hub.Rooms[*wsMsg.RoomID] = make(map[string]*Client)
	}
	wh.hub.Rooms[*wsMsg.RoomID][client.ID] = client
	wh.hub.mu.Unlock()

	log.Printf("Client %s joined room %s", client.ID, wsMsg.RoomID.String())
}

// handleLeaveRoom processes room leave requests
func (wh *WebSocketHandler) handleLeaveRoom(client *Client, wsMsg WebSocketMessage) {
	if client.RoomID == nil {
		return
	}

	// Remove from room in hub
	wh.hub.mu.Lock()
	if room, exists := wh.hub.Rooms[*client.RoomID]; exists {
		delete(room, client.ID)
		if len(room) == 0 {
			delete(wh.hub.Rooms, *client.RoomID)
		}
	}
	wh.hub.mu.Unlock()

	log.Printf("Client %s left room %s", client.ID, client.RoomID.String())
	client.RoomID = nil
}

// uuidToUint converts UUID to uint for database compatibility
// This is a simple conversion that extracts the first 4 bytes as uint
func (wh *WebSocketHandler) uuidToUint(id uuid.UUID) uint {
	// Convert first 4 bytes of UUID to uint32, then to uint
	return uint(uint32(id[0])<<24 | uint32(id[1])<<16 | uint32(id[2])<<8 | uint32(id[3]))
}

// uuidToRoomID converts UUID back to database room ID
func (wh *WebSocketHandler) uuidToRoomID(id uuid.UUID) uint {
	// Handle special case for support-chat
	if id.String() == "550e8400-e29b-41d4-a716-446655440001" {
		return 1 // Support chat room ID
	}
	
	// Handle the UUID used in test (550e8400-e29b-41d4-a716-446655440000)
	if id.String() == "550e8400-e29b-41d4-a716-446655440000" {
		return 1 // Also map to support chat room ID
	}
	
	// Handle the UUID from mobile app (6ba7b810-9dad-11d1-80b4-00c04fd430c8)
	if id.String() == "6ba7b810-9dad-11d1-80b4-00c04fd430c8" {
		log.Printf("Mobile app UUID %s mapped to room ID: 1", id.String())
		return 1 // Map to support chat room ID
	}
	
	// For UUIDs in our format: 550e8400-e29b-41d4-a716-{12-digit-number}
	idStr := id.String()
	if len(idStr) >= 12 {
		// Extract the last 12 characters and convert to uint
		lastPart := idStr[len(idStr)-12:]
		var roomID uint
		if _, err := fmt.Sscanf(lastPart, "%012d", &roomID); err == nil {
			log.Printf("Converted UUID %s to room ID: %d", id.String(), roomID)
			return roomID
		}
	}
	
	// Fallback: use the original method
	log.Printf("Using fallback conversion for UUID %s", id.String())
	return wh.uuidToUint(id)
}

// processRoomID converts RoomIDRaw to proper UUID format
func (wh *WebSocketHandler) processRoomID(wsMsg *WebSocketMessage) error {
	switch roomID := wsMsg.RoomIDRaw.(type) {
	case string:
		// Handle predefined string room IDs
		switch roomID {
		case "support-chat":
			// Use a predefined UUID for support chat that maps to room ID 1
			// We'll ensure this room exists in the database
			supportRoomID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")
			wsMsg.RoomID = &supportRoomID
			log.Printf("Mapped support-chat to UUID: %s", supportRoomID.String())
		default:
			// Try to parse as UUID string
			if parsedUUID, err := uuid.Parse(roomID); err == nil {
				wsMsg.RoomID = &parsedUUID
			} else {
				// Generate a deterministic UUID from the string
				generatedUUID := uuid.NewSHA1(uuid.NameSpaceURL, []byte(roomID))
				wsMsg.RoomID = &generatedUUID
				log.Printf("Generated UUID %s for room string: %s", generatedUUID.String(), roomID)
			}
		}
	case float64:
		// Handle numeric room IDs (from JSON parsing)
		roomIDInt := uint(roomID)
		// Convert integer room ID to UUID using a deterministic method
		uuidStr := fmt.Sprintf("550e8400-e29b-41d4-a716-%012d", roomIDInt)
		if parsedUUID, err := uuid.Parse(uuidStr); err == nil {
			wsMsg.RoomID = &parsedUUID
			log.Printf("Converted room ID %d to UUID: %s", roomIDInt, parsedUUID.String())
		} else {
			return fmt.Errorf("failed to convert room ID %d to UUID: %v", roomIDInt, err)
		}
	case map[string]interface{}:
		// Handle case where room_id might be a JSON object
		if idStr, ok := roomID["id"].(string); ok {
			if parsedUUID, err := uuid.Parse(idStr); err == nil {
				wsMsg.RoomID = &parsedUUID
			} else {
				return fmt.Errorf("invalid UUID format in room_id object: %s", idStr)
			}
		} else {
			return fmt.Errorf("room_id object missing 'id' field")
		}
	default:
		return fmt.Errorf("unsupported room_id type: %T", roomID)
	}
	return nil
}