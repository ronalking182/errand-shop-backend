package notifications

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type FCMHandler struct {
	fcmService FCMService
}

func NewFCMHandler(fcmService FCMService) *FCMHandler {
	return &FCMHandler{
		fcmService: fcmService,
	}
}

// DTOs for FCM endpoints
type SendSingleRequest struct {
	UserID   uuid.UUID              `json:"userId" validate:"required"`
	UserType string                 `json:"userType" validate:"required"`
	Title    string                 `json:"title" validate:"required"`
	Body     string                 `json:"body" validate:"required"`
	Data     map[string]interface{} `json:"data,omitempty"`
	ImageURL string                 `json:"imageUrl,omitempty"`
}

type SendMultipleRequest struct {
	UserIDs  []uuid.UUID            `json:"userIds" validate:"required,min=1"`
	UserType string                 `json:"userType" validate:"required"`
	Title    string                 `json:"title" validate:"required"`
	Body     string                 `json:"body" validate:"required"`
	Data     map[string]interface{} `json:"data,omitempty"`
	ImageURL string                 `json:"imageUrl,omitempty"`
}

type BroadcastRequest struct {
	Title    string                 `json:"title" validate:"required"`
	Body     string                 `json:"body" validate:"required"`
	Data     map[string]interface{} `json:"data,omitempty"`
	ImageURL string                 `json:"imageUrl,omitempty"`
}

type RegisterTokenRequest struct {
	UserID   uuid.UUID `json:"userId" validate:"required"`
	UserType string   `json:"userType" validate:"required"`
	Token    string `json:"token" validate:"required"`
	Platform string `json:"platform" validate:"required,oneof=ios android web"`
	DeviceID string `json:"deviceId,omitempty"`
}

type UnregisterTokenRequest struct {
	Token string `json:"token" validate:"required"`
}

type TestMessageRequest struct {
	UserID   uuid.UUID `json:"userId" validate:"required"`
	UserType string   `json:"userType" validate:"required"`
}

// POST /api/fcm/send - Send to single user
func (h *FCMHandler) SendToSingleUser(c *fiber.Ctx) error {
	var req SendSingleRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	message, err := h.fcmService.SendToSingleUser(
		req.UserID,
		req.UserType,
		req.Title,
		req.Body,
		req.Data,
		req.ImageURL,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Message sent successfully",
		"data":    message,
	})
}

// POST /api/fcm/send-multiple - Send to multiple users
func (h *FCMHandler) SendToMultipleUsers(c *fiber.Ctx) error {
	var req SendMultipleRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	message, err := h.fcmService.SendToMultipleUsers(
		req.UserIDs,
		req.UserType,
		req.Title,
		req.Body,
		req.Data,
		req.ImageURL,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Messages sent successfully",
		"data":    message,
	})
}

// POST /api/fcm/broadcast - Send to all users
func (h *FCMHandler) BroadcastToAllUsers(c *fiber.Ctx) error {
	var req BroadcastRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	message, err := h.fcmService.BroadcastToAllUsers(
		req.Title,
		req.Body,
		req.Data,
		req.ImageURL,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Broadcast sent successfully",
		"data":    message,
	})
}

// POST /api/fcm/register-token - Register device token
func (h *FCMHandler) RegisterToken(c *fiber.Ctx) error {
	var req RegisterTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	err := h.fcmService.RegisterToken(
		req.UserID,
		req.UserType,
		req.Token,
		req.Platform,
		req.DeviceID,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Token registered successfully",
	})
}

// DELETE /api/fcm/unregister-token - Remove token
func (h *FCMHandler) UnregisterToken(c *fiber.Ctx) error {
	var req UnregisterTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	err := h.fcmService.UnregisterToken(req.Token)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Token unregistered successfully",
	})
}

// GET /api/fcm/messages - Message history
func (h *FCMHandler) GetMessages(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	messages, total, err := h.fcmService.GetMessages(page, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"messages": messages,
			"total":    total,
			"page":     page,
			"limit":    limit,
			"pages":    (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// GET /api/fcm/stats - Analytics data
func (h *FCMHandler) GetStats(c *fiber.Ctx) error {
	stats, err := h.fcmService.GetStats()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    stats,
	})
}

// POST /api/fcm/test - Test message delivery
func (h *FCMHandler) TestMessage(c *fiber.Ctx) error {
	var req TestMessageRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	err := h.fcmService.TestMessage(req.UserID, req.UserType)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Test message sent successfully",
	})
}