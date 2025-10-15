package notifications

import (
	"errandShop/internal/presenter"
	"errandShop/internal/validation"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type NotificationHandler struct {
	service NotificationService
}

func NewNotificationHandler(service NotificationService) *NotificationHandler {
	return &NotificationHandler{service: service}
}

// GET /api/v1/notifications
func (h *NotificationHandler) GetNotifications(c *fiber.Ctx) error {
	userID := c.Locals("userID")
	if userID == nil {
		return presenter.ErrorResponse(c, 401, "User not authenticated")
	}

	// Handle different types from JWT claims
	var userIDUUID uuid.UUID
	switch v := userID.(type) {
	case uuid.UUID:
		userIDUUID = v
	case string:
		parsedUUID, err := uuid.Parse(v)
		if err != nil {
			return presenter.ErrorResponse(c, 400, "Invalid user ID format")
		}
		userIDUUID = parsedUUID
	default:
		return presenter.ErrorResponse(c, 500, "Invalid user ID type")
	}

	role := c.Locals("role")
	if role == nil {
		return presenter.ErrorResponse(c, 401, "User role not found")
	}

	roleStr, ok := role.(string)
	if !ok {
		return presenter.ErrorResponse(c, 500, "Invalid role type")
	}

	var recipientType NotificationRecipient
	switch roleStr {
	case "admin", "superadmin":
		recipientType = RecipientAdmin
	case "customer":
		recipientType = RecipientCustomer
	default:
		recipientType = RecipientCustomer
	}

	notifications, err := h.service.GetNotifications(userIDUUID, recipientType, 0, 10)
	if err != nil {
		return presenter.ErrorResponse(c, 500, err.Error())
	}

	return c.Status(200).JSON(fiber.Map{
		"status":  "success",
		"message": "Notifications retrieved successfully",
		"data":    notifications,
	})
}

// PUT /api/v1/notifications/:id/read
func (h *NotificationHandler) MarkAsRead(c *fiber.Ctx) error {
	notificationID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return presenter.ErrorResponse(c, 400, "Invalid notification ID")
	}

	if err := h.service.MarkAsRead(uint(notificationID)); err != nil {
		return presenter.ErrorResponse(c, 500, err.Error())
	}
	return presenter.SuccessResponse(c, "Notification marked as read", nil)
}

// PUT /api/v1/notifications/read-all
func (h *NotificationHandler) MarkAllAsRead(c *fiber.Ctx) error {
	userIDRaw := c.Locals("userID")
	if userIDRaw == nil {
		return presenter.ErrorResponse(c, 401, "User not authenticated")
	}

	// Handle different types from JWT claims
	var userID uuid.UUID
	switch v := userIDRaw.(type) {
	case uuid.UUID:
		userID = v
	case string:
		parsedUUID, err := uuid.Parse(v)
		if err != nil {
			return presenter.ErrorResponse(c, 400, "Invalid user ID format")
		}
		userID = parsedUUID
	default:
		return presenter.ErrorResponse(c, 500, "Invalid user ID type")
	}

	role := c.Locals("role")
	if role == nil {
		return presenter.ErrorResponse(c, 401, "User role not found")
	}

	roleStr, ok := role.(string)
	if !ok {
		return presenter.ErrorResponse(c, 500, "Invalid role type")
	}

	var recipientType NotificationRecipient
	switch roleStr {
	case "admin", "superadmin":
		recipientType = RecipientAdmin
	case "customer":
		recipientType = RecipientCustomer
	default:
		recipientType = RecipientCustomer
	}

	if err := h.service.MarkAllAsRead(userID, recipientType); err != nil {
		return presenter.ErrorResponse(c, 500, err.Error())
	}

	return c.Status(200).JSON(fiber.Map{
		"status":  "success",
		"message": "All notifications marked as read",
		"data":    nil,
	})
}

// POST /api/v1/notifications/push-token
func (h *NotificationHandler) RegisterPushToken(c *fiber.Ctx) error {
	userIDRaw := c.Locals("userID")
	if userIDRaw == nil {
		return presenter.ErrorResponse(c, 401, "User not authenticated")
	}

	// Handle different types from JWT claims
	var userID uuid.UUID
	switch v := userIDRaw.(type) {
	case uuid.UUID:
		userID = v
	case string:
		parsedUUID, err := uuid.Parse(v)
		if err != nil {
			return presenter.ErrorResponse(c, 400, "Invalid user ID format")
		}
		userID = parsedUUID
	default:
		return presenter.ErrorResponse(c, 500, "Invalid user ID type")
	}

	var req RegisterPushTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return presenter.ErrorResponse(c, 400, "Invalid request body")
	}

	if err := validation.ValidateStruct(&req); err != nil {
		return presenter.ErrorResponse(c, 400, err.Error())
	}

	response, err := h.service.RegisterPushToken(userID, "customer", &req)
	if err != nil {
		return presenter.ErrorResponse(c, 500, err.Error())
	}

	return c.Status(201).JSON(fiber.Map{
		"status":  "success",
		"message": "Push token registered successfully",
		"data":    response,
	})
}

// POST /api/v1/notifications
func (h *NotificationHandler) CreateNotification(c *fiber.Ctx) error {
	var req CreateNotificationRequest
	if err := c.BodyParser(&req); err != nil {
		return presenter.ErrorResponse(c, 400, "Invalid request body")
	}

	if err := validation.ValidateStruct(&req); err != nil {
		return presenter.ErrorResponse(c, 400, err.Error())
	}

	notification, err := h.service.CreateNotification(&req)
	if err != nil {
		return presenter.ErrorResponse(c, 500, err.Error())
	}

	return c.Status(201).JSON(fiber.Map{
		"status":  "success",
		"message": "Notification created successfully",
		"data":    notification,
	})
}

// Add these methods to NotificationHandler

func (h *NotificationHandler) SendBroadcastNotification(c *fiber.Ctx) error {
	var req BroadcastNotificationRequest
	if err := c.BodyParser(&req); err != nil {
		return presenter.ErrorResponse(c, 400, "Invalid request body")
	}

	if err := validation.ValidateStruct(&req); err != nil {
		return presenter.ErrorResponse(c, 400, err.Error())
	}

	if err := h.service.SendBroadcastNotification(&req); err != nil {
		return presenter.ErrorResponse(c, 500, err.Error())
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Broadcast notification sent successfully",
	})
}

func (h *NotificationHandler) GetNotificationTemplates(c *fiber.Ctx) error {
	templates, err := h.service.GetTemplates()
	if err != nil {
		return presenter.ErrorResponse(c, 500, err.Error())
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   templates,
	})
}

func (h *NotificationHandler) CreateNotificationTemplate(c *fiber.Ctx) error {
	var req CreateTemplateRequest
	if err := c.BodyParser(&req); err != nil {
		return presenter.ErrorResponse(c, 400, "Invalid request body")
	}

	if err := validation.ValidateStruct(&req); err != nil {
		return presenter.ErrorResponse(c, 400, err.Error())
	}

	template, err := h.service.CreateTemplate(&req)
	if err != nil {
		return presenter.ErrorResponse(c, 500, err.Error())
	}

	return c.Status(201).JSON(fiber.Map{
		"status": "success",
		"data":   template,
	})
}

func (h *NotificationHandler) UpdateNotificationTemplate(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return presenter.ErrorResponse(c, 400, "Invalid template ID")
	}

	var req CreateTemplateRequest
	if err := c.BodyParser(&req); err != nil {
		return presenter.ErrorResponse(c, 400, "Invalid request body")
	}

	if err := validation.ValidateStruct(&req); err != nil {
		return presenter.ErrorResponse(c, 400, err.Error())
	}

	template, err := h.service.UpdateTemplate(uint(id), &req)
	if err != nil {
		return presenter.ErrorResponse(c, 500, err.Error())
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   template,
	})
}

func (h *NotificationHandler) DeleteNotificationTemplate(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return presenter.ErrorResponse(c, 400, "Invalid template ID")
	}

	if err := h.service.DeleteTemplate(uint(id)); err != nil {
		return presenter.ErrorResponse(c, 500, err.Error())
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Template deleted successfully",
	})
}

func (h *NotificationHandler) GetNotificationStats(c *fiber.Ctx) error {
	// Implement notification statistics logic
	return c.JSON(fiber.Map{
		"status": "success",
		"data": fiber.Map{
			"total_sent":    0,
			"total_read":    0,
			"total_pending": 0,
		},
	})
}
