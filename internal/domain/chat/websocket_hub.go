package chat

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
)

// Client represents a WebSocket client
type Client struct {
	ID       string
	Conn     *websocket.Conn
	UserID   uuid.UUID
	UserType string // "customer" or "admin"
	RoomID   *uuid.UUID // nil for admins (they can join multiple rooms)
	Send     chan []byte
}

// Hub maintains the set of active clients and broadcasts messages to the clients
type Hub struct {
	Clients    map[string]*Client
	Broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client
	Rooms      map[uuid.UUID]map[string]*Client // roomID -> clientID -> client
	mu         sync.RWMutex
}

// WebSocketMessage represents a message sent through WebSocket
type WebSocketMessage struct {
	Type       string      `json:"type"`
	RoomID     *uuid.UUID  `json:"-"` // Will be set after custom unmarshaling
	RoomIDRaw  interface{} `json:"room_id,omitempty"` // Can be string or UUID
	SenderID   uuid.UUID   `json:"sender_id"`
	SenderType string      `json:"sender_type,omitempty"` // "customer", "admin", or "superadmin"
	Message    string      `json:"message,omitempty"`
	Data       interface{} `json:"data,omitempty"`
	Timestamp  int64       `json:"timestamp"`
}

// NewHub creates a new WebSocket hub
func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[string]*Client),
		Broadcast:  make(chan []byte, 1000),    // Buffer for broadcast messages
		Register:   make(chan *Client, 100),    // Buffer for client registrations
		Unregister: make(chan *Client, 100),    // Buffer for client unregistrations
		Rooms:      make(map[uuid.UUID]map[string]*Client),
	}
}

// Run starts the hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.registerClient(client)

		case client := <-h.Unregister:
			h.unregisterClient(client)

		case message := <-h.Broadcast:
			h.broadcastMessage(message)
		}
	}
}

// registerClient registers a new client
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.Clients[client.ID] = client

	// If client has a room ID, add them to that room
	if client.RoomID != nil {
		if h.Rooms[*client.RoomID] == nil {
			h.Rooms[*client.RoomID] = make(map[string]*Client)
		}
		h.Rooms[*client.RoomID][client.ID] = client
	}

	log.Printf("Client %s connected (UserID: %s, UserType: %s, RoomID: %v)", 
		client.ID, client.UserID, client.UserType, client.RoomID)
	
	// DEBUG: Admin client registration tracking
	if client.UserType == "admin" || client.UserType == "superadmin" {
		adminCount := 0
		for _, c := range h.Clients {
			if (c.UserType == "admin" || c.UserType == "superadmin") && c.RoomID == nil {
				adminCount++
			}
		}
		log.Printf("DEBUG: Admin client registered: %s (Total global admin clients: %d)", client.ID, adminCount)
	}
}

// unregisterClient unregisters a client
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.Clients[client.ID]; ok {
		delete(h.Clients, client.ID)
		close(client.Send)

		// Remove from room if applicable
		if client.RoomID != nil {
			if room, exists := h.Rooms[*client.RoomID]; exists {
				delete(room, client.ID)
				// Clean up empty rooms
				if len(room) == 0 {
					delete(h.Rooms, *client.RoomID)
				}
			}
		}

		log.Printf("Client %s disconnected", client.ID)
	}
}

// broadcastMessage broadcasts a message to relevant clients
func (h *Hub) broadcastMessage(message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var wsMsg WebSocketMessage
	if err := json.Unmarshal(message, &wsMsg); err != nil {
		log.Printf("Error unmarshaling WebSocket message: %v", err)
		return
	}

	// If message has a room ID, broadcast to that room only
	if wsMsg.RoomID != nil {
		log.Printf("DEBUG: broadcastMessage - Broadcasting room message to room %s", wsMsg.RoomID.String())
		if room, exists := h.Rooms[*wsMsg.RoomID]; exists {
			for _, client := range room {
				// Send to ALL clients in the room (customers and admins)
				// Admin clients will also receive via BroadcastToAdmins, but room clients should get room messages
				log.Printf("DEBUG: broadcastMessage - Sending to room client %s (UserType: %s)", client.ID, client.UserType)
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.Clients, client.ID)
					delete(room, client.ID)
				}
			}
		}
	} else {
		log.Printf("DEBUG: broadcastMessage - Broadcasting global message to all clients")
		// Broadcast to all clients (admin notifications, etc.)
		for clientID, client := range h.Clients {
			log.Printf("DEBUG: broadcastMessage - Sending global message to client %s (UserType: %s)", client.ID, client.UserType)
			select {
			case client.Send <- message:
			default:
				close(client.Send)
				delete(h.Clients, clientID)
			}
		}
	}
}

// BroadcastToRoom broadcasts a message to a specific room
func (h *Hub) BroadcastToRoom(roomID uuid.UUID, message WebSocketMessage) {
	message.RoomID = &roomID
	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	select {
	case h.Broadcast <- data:
	default:
		log.Println("Broadcast channel is full")
	}
}

// BroadcastToAdmins broadcasts a message to all admin clients (global listeners)
func (h *Hub) BroadcastToAdmins(message WebSocketMessage) {
	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling admin message: %v", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	// DEBUG: Detailed admin client analysis
	totalClients := len(h.Clients)
	adminCount := 0
	globalAdminCount := 0
	
	log.Printf("DEBUG: BroadcastToAdmins called - Total clients: %d", totalClients)
	
	for clientID, client := range h.Clients {
		log.Printf("DEBUG: Checking client %s - UserType: %s, RoomID: %v", clientID, client.UserType, client.RoomID)
		
		if client.UserType == "admin" || client.UserType == "superadmin" {
			adminCount++
			if client.RoomID == nil {
				globalAdminCount++
				log.Printf("DEBUG: Broadcasting to admin client: %s (UserID: %s)", client.ID, client.UserID.String())
				select {
				case client.Send <- data:
					log.Printf("SUCCESS: Message sent to admin %s", client.ID)
				default:
					log.Printf("ERROR: Failed to send to admin client %s - channel blocked", client.ID)
				}
			} else {
				log.Printf("DEBUG: Skipping admin client %s - has RoomID: %v (not global listener)", client.ID, client.RoomID)
			}
		}
	}
	
	log.Printf("DEBUG: BroadcastToAdmins completed - Total admins: %d, Global admins: %d, Messages sent: %d", adminCount, globalAdminCount, globalAdminCount)
}

// BroadcastToUser broadcasts a message to a specific user (all their connections)
func (h *Hub) BroadcastToUser(userID uuid.UUID, message WebSocketMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	for _, client := range h.Clients {
		if client.UserID == userID {
			select {
			case client.Send <- data:
			default:
				close(client.Send)
				delete(h.Clients, client.ID)
			}
		}
	}
}

// GetRoomClients returns all clients in a specific room
func (h *Hub) GetRoomClients(roomID uuid.UUID) []*Client {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var clients []*Client
	if room, exists := h.Rooms[roomID]; exists {
		for _, client := range room {
			clients = append(clients, client)
		}
	}
	return clients
}

// GetClientCount returns the total number of connected clients
func (h *Hub) GetClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.Clients)
}

// GetRoomCount returns the total number of active rooms
func (h *Hub) GetRoomCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.Rooms)
}