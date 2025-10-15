package chat

import (
	"errandShop/internal/presenter"
	"errandShop/internal/validation"
	"errors"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// ChatHandler handles HTTP requests for chat functionality
type ChatHandler struct {
	service ChatService
}

// NewChatHandler creates a new chat handler
func NewChatHandler(service ChatService) *ChatHandler {
	return &ChatHandler{service: service}
}

// Chat Room Handlers

// POST /api/v1/chat/rooms
func (h *ChatHandler) CreateChatRoom(c *fiber.Ctx) error {
	userUUID, err := h.getUserIDFromContext(c)
	if err != nil {
		return presenter.ErrorResponse(c, 401, err.Error())
	}
	
	// Convert UUID to uint for chat service compatibility
	userID := h.uuidToUint(userUUID)
	// Get role from context with type checking
	roleRaw := c.Locals("role")
	if roleRaw == nil {
		return presenter.ErrorResponse(c, 401, "User role not found")
	}
	role, ok := roleRaw.(string)
	if !ok {
		return presenter.ErrorResponse(c, 500, "Invalid role type")
	}

	var req CreateChatRoomRequest
	if err := c.BodyParser(&req); err != nil {
		return presenter.ErrorResponse(c, 400, "Invalid request body")
	}

	if err := validation.ValidateStruct(&req); err != nil {
		return presenter.ErrorResponse(c, 400, err.Error())
	}

	// Determine sender type
	var senderType SenderType
	switch role {
	case "admin", "superadmin":
		senderType = SenderTypeAdmin
	case "customer":
		senderType = SenderTypeCustomer
		// Customers can only create rooms for themselves
		req.CustomerID = userID
	default:
		return presenter.ErrorResponse(c, 403, "Invalid user role")
	}

	room, err := h.service.CreateChatRoom(&req, senderType, userID)
	if err != nil {
		return presenter.ErrorResponse(c, 500, err.Error())
	}

	return c.Status(201).JSON(fiber.Map{
		"status":  "success",
		"message": "Chat room created successfully",
		"data":    room,
	})
}

// Helper function to get user ID from context
func (h *ChatHandler) getUserIDFromContext(c *fiber.Ctx) (uuid.UUID, error) {
	userIDRaw := c.Locals("userID")
	if userIDRaw == nil {
		return uuid.Nil, errors.New("user not authenticated")
	}

	// Handle different types from JWT claims
	switch v := userIDRaw.(type) {
	case uuid.UUID:
		return v, nil
	case string:
		return uuid.Parse(v)
	case uint:
		return uuid.Parse(fmt.Sprintf("%d", v))
	case float64:
		return uuid.Parse(fmt.Sprintf("%.0f", v))
	default:
		return uuid.Nil, errors.New("invalid user ID format")
	}
}

// Helper function to convert UUID to uint for chat service compatibility
func (h *ChatHandler) uuidToUint(id uuid.UUID) uint {
	// Convert first 4 bytes of UUID to uint32, then to uint
	return uint(uint32(id[0])<<24 | uint32(id[1])<<16 | uint32(id[2])<<8 | uint32(id[3]))
}

// GET /api/v1/chat/rooms
func (h *ChatHandler) GetChatRooms(c *fiber.Ctx) error {
	userUUID, err := h.getUserIDFromContext(c)
	if err != nil {
		return presenter.ErrorResponse(c, 401, err.Error())
	}
	
	// Convert UUID to uint for chat service compatibility
	userID := h.uuidToUint(userUUID)
	
	// Get role from context with type checking
	roleRaw := c.Locals("role")
	if roleRaw == nil {
		return presenter.ErrorResponse(c, 401, "User role not found")
	}
	role, ok := roleRaw.(string)
	if !ok {
		return presenter.ErrorResponse(c, 500, "Invalid role type")
	}

	// Parse query parameters
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	rooms, err := h.service.GetChatRooms(userID, role, page, limit)
	if err != nil {
		return presenter.ErrorResponse(c, 500, err.Error())
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   rooms,
	})
}

// GET /api/v1/chat/rooms/:id
func (h *ChatHandler) GetChatRoom(c *fiber.Ctx) error {
	roomID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return presenter.ErrorResponse(c, 400, "Invalid room ID")
	}

	room, err := h.service.GetChatRoom(uint(roomID))
	if err != nil {
		return presenter.ErrorResponse(c, 404, "Chat room not found")
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   room,
	})
}

// PUT /api/v1/chat/rooms/:id
func (h *ChatHandler) UpdateChatRoom(c *fiber.Ctx) error {
	roomID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return presenter.ErrorResponse(c, 400, "Invalid room ID")
	}

	var req UpdateChatRoomRequest
	if err := c.BodyParser(&req); err != nil {
		return presenter.ErrorResponse(c, 400, "Invalid request body")
	}

	if err := validation.ValidateStruct(&req); err != nil {
		return presenter.ErrorResponse(c, 400, err.Error())
	}

	room, err := h.service.UpdateChatRoom(uint(roomID), &req)
	if err != nil {
		return presenter.ErrorResponse(c, 500, err.Error())
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Chat room updated successfully",
		"data":    room,
	})
}

// DELETE /api/v1/chat/rooms/:id
func (h *ChatHandler) DeleteChatRoom(c *fiber.Ctx) error {
	roomID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return presenter.ErrorResponse(c, 400, "Invalid room ID")
	}

	if err := h.service.DeleteChatRoom(uint(roomID)); err != nil {
		return presenter.ErrorResponse(c, 500, err.Error())
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Chat room deleted successfully",
	})
}

// POST /api/v1/chat/rooms/:id/assign
func (h *ChatHandler) AssignAdminToRoom(c *fiber.Ctx) error {
	roomID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return presenter.ErrorResponse(c, 400, "Invalid room ID")
	}

	adminUUID, err := h.getUserIDFromContext(c)
	if err != nil {
		return presenter.ErrorResponse(c, 401, err.Error())
	}
	
	// Convert UUID to uint for chat service compatibility
	adminID := h.uuidToUint(adminUUID)
	
	// Get role from context with type checking
	roleRaw := c.Locals("role")
	if roleRaw == nil {
		return presenter.ErrorResponse(c, 401, "User role not found")
	}
	role, ok := roleRaw.(string)
	if !ok {
		return presenter.ErrorResponse(c, 500, "Invalid role type")
	}

	// Only admins can assign themselves to rooms
	if role != "admin" && role != "superadmin" {
		return presenter.ErrorResponse(c, 403, "Only admins can assign themselves to rooms")
	}

	if err := h.service.AssignAdminToRoom(uint(roomID), adminID); err != nil {
		return presenter.ErrorResponse(c, 500, err.Error())
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Admin assigned to room successfully",
	})
}

// POST /api/v1/chat/rooms/:id/unassign
func (h *ChatHandler) UnassignAdminFromRoom(c *fiber.Ctx) error {
	roomID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return presenter.ErrorResponse(c, 400, "Invalid room ID")
	}

	if err := h.service.UnassignAdminFromRoom(uint(roomID)); err != nil {
		return presenter.ErrorResponse(c, 500, err.Error())
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Admin unassigned from room successfully",
	})
}

// Chat Message Handlers

// POST /api/v1/chat/messages
func (h *ChatHandler) SendMessage(c *fiber.Ctx) error {
	userUUID, err := h.getUserIDFromContext(c)
	if err != nil {
		return presenter.ErrorResponse(c, 401, err.Error())
	}
	
	// Convert UUID to uint for chat service compatibility
	userID := h.uuidToUint(userUUID)
	
	// Get role from context with type checking
	roleRaw := c.Locals("role")
	if roleRaw == nil {
		return presenter.ErrorResponse(c, 401, "User role not found")
	}
	role, ok := roleRaw.(string)
	if !ok {
		return presenter.ErrorResponse(c, 500, "Invalid role type")
	}

	var req SendMessageRequest
	if err := c.BodyParser(&req); err != nil {
		return presenter.ErrorResponse(c, 400, "Invalid request body")
	}

	if err := validation.ValidateStruct(&req); err != nil {
		return presenter.ErrorResponse(c, 400, err.Error())
	}

	// Determine sender type
	var senderType SenderType
	switch role {
	case "admin", "superadmin":
		senderType = SenderTypeAdmin
	case "customer":
		senderType = SenderTypeCustomer
	default:
		return presenter.ErrorResponse(c, 403, "Invalid user role")
	}

	message, err := h.service.SendMessageWithRequest(&req, senderType, userID)
	if err != nil {
		return presenter.ErrorResponse(c, 500, err.Error())
	}

	return c.Status(201).JSON(fiber.Map{
		"status":  "success",
		"message": "Message sent successfully",
		"data":    message,
	})
}

// GET /api/v1/chat/rooms/:id/messages
func (h *ChatHandler) GetMessages(c *fiber.Ctx) error {
	roomID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return presenter.ErrorResponse(c, 400, "Invalid room ID")
	}

	// Parse query parameters
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "50"))

	messages, err := h.service.GetMessages(uint(roomID), page, limit)
	if err != nil {
		return presenter.ErrorResponse(c, 500, err.Error())
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   messages,
	})
}

// PUT /api/v1/chat/rooms/:id/read
func (h *ChatHandler) MarkMessagesAsRead(c *fiber.Ctx) error {
	roomID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return presenter.ErrorResponse(c, 400, "Invalid room ID")
	}

	userUUID, err := h.getUserIDFromContext(c)
	if err != nil {
		return presenter.ErrorResponse(c, 401, err.Error())
	}
	
	// Convert UUID to uint for chat service compatibility
	userID := h.uuidToUint(userUUID)
	
	// Get role from context with type checking
	roleRaw := c.Locals("role")
	if roleRaw == nil {
		return presenter.ErrorResponse(c, 401, "User role not found")
	}
	role, ok := roleRaw.(string)
	if !ok {
		return presenter.ErrorResponse(c, 500, "Invalid role type")
	}

	// Determine user type
	var userType SenderType
	switch role {
	case "admin", "superadmin":
		userType = SenderTypeAdmin
	case "customer":
		userType = SenderTypeCustomer
	default:
		return presenter.ErrorResponse(c, 403, "Invalid user role")
	}

	if err := h.service.MarkMessagesAsRead(uint(roomID), userID, userType); err != nil {
		return presenter.ErrorResponse(c, 500, err.Error())
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Messages marked as read",
	})
}

// DELETE /api/v1/chat/messages/:id
func (h *ChatHandler) DeleteMessage(c *fiber.Ctx) error {
	messageID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return presenter.ErrorResponse(c, 400, "Invalid message ID")
	}

	if err := h.service.DeleteMessage(uint(messageID)); err != nil {
		return presenter.ErrorResponse(c, 500, err.Error())
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Message deleted successfully",
	})
}

// Utility Handlers

// GET /api/v1/chat/stats
func (h *ChatHandler) GetChatStats(c *fiber.Ctx) error {
	stats, err := h.service.GetChatStats()
	if err != nil {
		return presenter.ErrorResponse(c, 500, err.Error())
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   stats,
	})
}

// POST /api/v1/chat/typing
func (h *ChatHandler) SendTypingIndicator(c *fiber.Ctx) error {
	userUUID, err := h.getUserIDFromContext(c)
	if err != nil {
		return presenter.ErrorResponse(c, 401, err.Error())
	}
	
	// Convert UUID to uint for chat service compatibility
	userID := h.uuidToUint(userUUID)
	
	// Get role from context with type checking
	roleRaw := c.Locals("role")
	if roleRaw == nil {
		return presenter.ErrorResponse(c, 401, "User role not found")
	}
	role, ok := roleRaw.(string)
	if !ok {
		return presenter.ErrorResponse(c, 500, "Invalid role type")
	}

	var req TypingIndicatorRequest
	if err := c.BodyParser(&req); err != nil {
		return presenter.ErrorResponse(c, 400, "Invalid request body")
	}

	if err := validation.ValidateStruct(&req); err != nil {
		return presenter.ErrorResponse(c, 400, err.Error())
	}

	// Determine user type
	var userType SenderType
	switch role {
	case "admin", "superadmin":
		userType = SenderTypeAdmin
	case "customer":
		userType = SenderTypeCustomer
	default:
		return presenter.ErrorResponse(c, 403, "Invalid user role")
	}

	if err := h.service.SendTypingIndicator(req.RoomID, userID, userType, req.IsTyping); err != nil {
		return presenter.ErrorResponse(c, 500, err.Error())
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Typing indicator sent",
	})
}