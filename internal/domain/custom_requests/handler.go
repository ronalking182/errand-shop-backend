package custom_requests

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// User endpoints

// CreateCustomRequest creates a new custom request
// @Summary Create custom request
// @Description Create a new custom request for items not in catalog
// @Tags custom-requests
// @Accept json
// @Produce json
// @Param request body CreateCustomRequestReq true "Custom request data"
// @Success 201 {object} CustomRequestRes
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/custom-requests [post]
func (h *Handler) CreateCustomRequest(c *fiber.Ctx) error {
	// Get user ID from context (set by auth middleware)
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	var req CreateCustomRequestReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate request
	if len(req.Items) == 0 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "At least one item is required",
		})
	}

	result, err := h.service.CreateCustomRequest(userID, req)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(http.StatusCreated).JSON(result)
}

// GetCustomRequest gets a custom request by ID
// @Summary Get custom request
// @Description Get a custom request by ID (user can only access their own)
// @Tags custom-requests
// @Produce json
// @Param id path string true "Custom Request ID"
// @Success 200 {object} CustomRequestRes
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/custom-requests/{id} [get]
func (h *Handler) GetCustomRequest(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	requestID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request ID",
		})
	}

	result, err := h.service.GetCustomRequest(userID, requestID)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(result)
}

// UpdateCustomRequest updates a custom request
// @Summary Update custom request
// @Description Update a custom request (only allowed in certain statuses)
// @Tags custom-requests
// @Accept json
// @Produce json
// @Param id path string true "Custom Request ID"
// @Param request body UpdateCustomRequestReq true "Update data"
// @Success 200 {object} CustomRequestRes
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/custom-requests/{id} [put]
func (h *Handler) UpdateCustomRequest(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	requestID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request ID",
		})
	}

	var req UpdateCustomRequestReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	result, err := h.service.UpdateCustomRequest(userID, requestID, req)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(result)
}

// DeleteCustomRequest deletes a custom request
// @Summary Delete custom request
// @Description Delete a custom request (only allowed in certain statuses)
// @Tags custom-requests
// @Param id path string true "Custom Request ID"
// @Success 204
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/custom-requests/{id} [delete]
func (h *Handler) DeleteCustomRequest(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	requestID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request ID",
		})
	}

	if err := h.service.DeleteCustomRequest(userID, requestID); err != nil {
		return h.handleError(c, err)
	}

	return c.SendStatus(http.StatusNoContent)
}

// CancelCustomRequest cancels a custom request
// @Summary Cancel custom request
// @Description Cancel a custom request and set status to cancelled
// @Tags custom-requests
// @Accept json
// @Produce json
// @Param id path string true "Custom Request ID"
// @Param request body object{reason=string} true "Cancel request body"
// @Success 200 {object} CustomRequestRes
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/custom-requests/{id}/cancel [post]
func (h *Handler) CancelCustomRequest(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	requestID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request ID",
		})
	}

	var req struct {
		Reason string `json:"reason" validate:"required,min=3,max=500"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Reason == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Reason is required",
		})
	}

	result, err := h.service.CancelCustomRequest(userID, requestID, req.Reason)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Custom request cancelled successfully",
		"data":    result,
	})
}

// ListUserCustomRequests lists user's custom requests
// @Summary List user custom requests
// @Description List custom requests for the authenticated user
// @Tags custom-requests
// @Produce json
// @Param status query string false "Filter by status"
// @Param priority query string false "Filter by priority"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param sort_by query string false "Sort field" default("submitted_at")
// @Param sort_order query string false "Sort order (asc/desc)" default("desc")
// @Success 200 {object} CustomRequestListRes
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/custom-requests [get]
func (h *Handler) ListUserCustomRequests(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	query := h.parseListQuery(c)
	result, err := h.service.ListUserCustomRequests(userID, query)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(result)
}

// AcceptQuote accepts a quote for a custom request
// @Summary Accept quote
// @Description Accept a quote for a custom request
// @Tags custom-requests
// @Accept json
// @Produce json
// @Param request body AcceptQuoteReq true "Accept quote data"
// @Success 200 {object} CustomRequestRes
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/custom-requests/accept-quote [post]
func (h *Handler) AcceptQuote(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	var req AcceptQuoteReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	result, err := h.service.AcceptQuote(userID, req)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(result)
}

// AcceptQuoteByRequestID accepts the active quote for a custom request by request ID
// @Summary Accept quote by request ID
// @Description Accept the active quote for a custom request using the request ID
// @Tags custom-requests
// @Accept json
// @Produce json
// @Param id path string true "Custom Request ID"
// @Success 200 {object} CustomRequestRes
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/custom-requests/{id}/accept [post]
func (h *Handler) AcceptQuoteByRequestID(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	requestID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request ID",
		})
	}

	result, err := h.service.AcceptQuoteByRequestID(userID, requestID)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(result)
}

// SendMessage sends a message for a custom request
// @Summary Send message
// @Description Send a message for a custom request
// @Tags custom-requests
// @Accept json
// @Produce json
// @Param id path string true "Custom Request ID"
// @Param request body SendMessageReq true "Message data"
// @Success 201 {object} CustomRequestMsgRes
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/custom-requests/{id}/messages [post]
func (h *Handler) SendMessage(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	requestID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request ID",
		})
	}

	var req SendMessageReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	result, err := h.service.SendMessage(userID, requestID, req)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(http.StatusCreated).JSON(result)
}

// Admin endpoints

// GetCustomRequestAdmin gets a custom request by ID (admin access)
// @Summary Get custom request (admin)
// @Description Get a custom request by ID with admin access
// @Tags admin,custom-requests
// @Produce json
// @Param id path string true "Custom Request ID"
// @Success 200 {object} CustomRequestRes
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/admin/custom-requests/{id} [get]
func (h *Handler) GetCustomRequestAdmin(c *fiber.Ctx) error {
	requestID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request ID",
		})
	}

	result, err := h.service.GetCustomRequestAdmin(requestID)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(result)
}

// ListCustomRequestsAdmin lists all custom requests (admin access)
// @Summary List custom requests (admin)
// @Description List all custom requests with admin access
// @Tags admin,custom-requests
// @Produce json
// @Param status query string false "Filter by status"
// @Param priority query string false "Filter by priority"
// @Param assignee_id query string false "Filter by assignee ID"
// @Param user_id query string false "Filter by user ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param sort_by query string false "Sort field" default("submitted_at")
// @Param sort_order query string false "Sort order (asc/desc)" default("desc")
// @Success 200 {object} CustomRequestListRes
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/admin/custom-requests [get]
func (h *Handler) ListCustomRequestsAdmin(c *fiber.Ctx) error {
	query := h.parseListQuery(c)

	// Parse additional admin filters
	if assigneeIDStr := c.Query("assignee_id"); assigneeIDStr != "" {
		if assigneeID, err := uuid.Parse(assigneeIDStr); err == nil {
			query.AssigneeID = &assigneeID
		}
	}
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if userID, err := uuid.Parse(userIDStr); err == nil {
			query.UserID = &userID
		}
	}

	result, err := h.service.ListCustomRequestsAdmin(query)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(result)
}

// UpdateCustomRequestStatus updates the status of a custom request
// @Summary Update custom request status
// @Description Update the status of a custom request (admin only)
// @Tags admin,custom-requests
// @Accept json
// @Produce json
// @Param id path string true "Custom Request ID"
// @Param request body UpdateRequestStatusReq true "Status update data"
// @Success 200 {object} CustomRequestRes
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/admin/custom-requests/{id}/status [put]
func (h *Handler) UpdateCustomRequestStatus(c *fiber.Ctx) error {
	requestID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request ID",
		})
	}

	var req UpdateRequestStatusReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	result, err := h.service.UpdateCustomRequestStatus(requestID, req)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(result)
}

// AssignCustomRequest assigns a custom request to an admin
// @Summary Assign custom request
// @Description Assign a custom request to an admin
// @Tags admin,custom-requests
// @Param id path string true "Custom Request ID"
// @Param assignee_id path string true "Assignee ID"
// @Success 200 {object} CustomRequestRes
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/admin/custom-requests/{id}/assign/{assignee_id} [put]
func (h *Handler) AssignCustomRequest(c *fiber.Ctx) error {
	requestID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request ID",
		})
	}

	assigneeID, err := uuid.Parse(c.Params("assignee_id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid assignee ID",
		})
	}

	result, err := h.service.AssignCustomRequest(requestID, assigneeID)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(result)
}

// CreateQuote creates a quote for a custom request
// @Summary Create quote
// @Description Create a quote for a custom request
// @Tags admin,custom-requests
// @Accept json
// @Produce json
// @Param request body CreateQuoteReq true "Quote data"
// @Success 201 {object} QuoteRes
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/admin/custom-requests/quotes [post]
func (h *Handler) CreateQuote(c *fiber.Ctx) error {
	adminID, err := h.getUserID(c)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	var req CreateQuoteReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate request items for null UUIDs
	for i, item := range req.Items {
		if item.RequestItemID == uuid.Nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("Invalid requestItemId at index %d: cannot be null UUID", i),
			})
		}
	}

	result, err := h.service.CreateQuote(adminID, req)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(http.StatusCreated).JSON(result)
}

// UpdateQuote updates a quote
// @Summary Update quote
// @Description Update a quote for a custom request
// @Tags admin,custom-requests
// @Accept json
// @Produce json
// @Param id path string true "Quote ID"
// @Param request body CreateQuoteReq true "Quote data"
// @Success 200 {object} QuoteRes
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/admin/custom-requests/quotes/{id} [put]
func (h *Handler) UpdateQuote(c *fiber.Ctx) error {
	quoteID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid quote ID",
		})
	}

	var req CreateQuoteReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	result, err := h.service.UpdateQuote(quoteID, req)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(result)
}

// SendQuote sends a quote to the customer
// @Summary Send quote
// @Description Send a quote to the customer
// @Tags admin,custom-requests
// @Param id path string true "Quote ID"
// @Success 200 {object} QuoteRes
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/admin/custom-requests/quotes/{id}/send [put]
func (h *Handler) SendQuote(c *fiber.Ctx) error {
	quoteID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid quote ID",
		})
	}

	result, err := h.service.SendQuote(quoteID)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(result)
}

// SendMessageAdmin sends a message as admin
// @Summary Send message (admin)
// @Description Send a message for a custom request as admin
// @Tags admin,custom-requests
// @Accept json
// @Produce json
// @Param id path string true "Custom Request ID"
// @Param request body SendMessageReq true "Message data"
// @Success 201 {object} CustomRequestMsgRes
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/admin/custom-requests/{id}/messages [post]
func (h *Handler) SendMessageAdmin(c *fiber.Ctx) error {
	adminID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	requestID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request ID",
		})
	}

	var req SendMessageReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	result, err := h.service.SendMessageAdmin(adminID, requestID, req)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(http.StatusCreated).JSON(result)
}

// BulkUpdateStatus updates the status of multiple custom requests
// @Summary Bulk update custom request status
// @Description Update the status of multiple custom requests (admin only)
// @Tags admin,custom-requests
// @Accept json
// @Produce json
// @Param request body BulkUpdateStatusReq true "Bulk status update data"
// @Success 200 {object} BulkUpdateStatusRes
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/admin/custom-requests/bulk/status [put]
func (h *Handler) BulkUpdateStatus(c *fiber.Ctx) error {
	var req BulkUpdateStatusReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	result, err := h.service.BulkUpdateStatus(req)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(result)
}

// BulkAssign assigns multiple custom requests to an admin
// @Summary Bulk assign custom requests
// @Description Assign multiple custom requests to an admin
// @Tags admin,custom-requests
// @Accept json
// @Produce json
// @Param request body BulkAssignReq true "Bulk assign data"
// @Success 200 {object} BulkAssignRes
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/admin/custom-requests/bulk/assign [put]
func (h *Handler) BulkAssign(c *fiber.Ctx) error {
	var req BulkAssignReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	result, err := h.service.BulkAssign(req)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(result)
}

// PermanentlyDeleteCustomRequest permanently deletes a cancelled custom request
// @Summary Permanently delete custom request
// @Description Permanently delete a cancelled custom request so it no longer appears in the app
// @Tags custom-requests
// @Param id path string true "Custom Request ID"
// @Success 204
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/custom-requests/{id}/permanent-delete [delete]
func (h *Handler) PermanentlyDeleteCustomRequest(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	requestID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request ID",
		})
	}

	if err := h.service.PermanentlyDeleteCustomRequest(userID, requestID); err != nil {
		return h.handleError(c, err)
	}

	return c.SendStatus(http.StatusNoContent)
}

// CancelCustomRequestAdmin allows admins to cancel any custom request regardless of status
// @Summary Cancel custom request (Admin)
// @Description Admin endpoint to cancel any custom request regardless of status
// @Tags admin,custom-requests
// @Accept json
// @Produce json
// @Param id path string true "Custom Request ID"
// @Param request body map[string]string false "Cancellation reason"
// @Success 200 {object} CustomRequestRes
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/admin/custom-requests/{id}/cancel [post]
func (h *Handler) CancelCustomRequestAdmin(c *fiber.Ctx) error {
	adminID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	requestID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request ID",
		})
	}

	// Parse optional reason from request body
	var reqBody map[string]string
	reason := "Cancelled by admin"
	if err := c.BodyParser(&reqBody); err == nil {
		if r, exists := reqBody["reason"]; exists && r != "" {
			reason = r
		}
	}

	result, err := h.service.CancelCustomRequestAdmin(adminID, requestID, reason)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Custom request cancelled successfully",
		"data":    result,
	})
}

// PermanentlyDeleteCustomRequestAdmin allows admins to permanently delete any custom request regardless of status
// @Summary Permanently delete custom request (Admin)
// @Description Admin endpoint to permanently delete any custom request regardless of status
// @Tags admin,custom-requests
// @Param id path string true "Custom Request ID"
// @Success 204
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/admin/custom-requests/{id}/permanent-delete [delete]
func (h *Handler) PermanentlyDeleteCustomRequestAdmin(c *fiber.Ctx) error {
	adminID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	requestID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request ID",
		})
	}

	if err := h.service.PermanentlyDeleteCustomRequestAdmin(adminID, requestID); err != nil {
		return h.handleError(c, err)
	}

	return c.SendStatus(http.StatusNoContent)
}

// GetCustomRequestStats gets custom request statistics
// @Summary Get custom request statistics
// @Description Get statistics for custom requests
// @Tags admin,custom-requests
// @Produce json
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Success 200 {object} CustomRequestStatsRes
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/admin/custom-requests/stats [get]
func (h *Handler) GetCustomRequestStats(c *fiber.Ctx) error {
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if startDateStr == "" && endDateStr == "" {
		result, err := h.service.GetCustomRequestStats()
		if err != nil {
			return h.handleError(c, err)
		}
		return c.JSON(result)
	}

	var startDate, endDate time.Time
	var err error

	if startDateStr != "" {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid start date format (use YYYY-MM-DD)",
			})
		}
	}

	if endDateStr != "" {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid end date format (use YYYY-MM-DD)",
			})
		}
		// Set end date to end of day
		endDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	}

	result, err := h.service.GetCustomRequestStatsByDateRange(startDate, endDate)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(result)
}

// Helper methods

func (h *Handler) parseListQuery(c *fiber.Ctx) CustomRequestListQuery {
	query := CustomRequestListQuery{}

	// Parse status
	if statusStr := c.Query("status"); statusStr != "" {
		status := RequestStatus(statusStr)
		query.Status = &status
	}

	// Parse priority
	if priorityStr := c.Query("priority"); priorityStr != "" {
		priority := RequestPriority(priorityStr)
		query.Priority = &priority
	}

	// Parse pagination
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			query.Page = page
		}
	}
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 100 {
			query.Limit = limit
		}
	}

	// Parse sorting
	if sortBy := c.Query("sort_by"); sortBy != "" {
		query.SortBy = sortBy
	}
	if sortOrder := c.Query("sort_order"); sortOrder != "" {
		query.SortOrder = sortOrder
	}

	// Set defaults
	query.SetDefaults()

	return query
}

// Helper function to get user ID from context
func (h *Handler) getUserID(c *fiber.Ctx) (uuid.UUID, error) {
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

func (h *Handler) handleError(c *fiber.Ctx, err error) error {
	switch err {
	case ErrCustomRequestNotFound:
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Custom request not found",
		})
	case ErrQuoteNotFound:
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Quote not found",
		})
	case ErrUnauthorizedAccess:
		return c.Status(http.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied",
		})
	case ErrInvalidStatusTransition:
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid status transition",
		})
	case ErrCannotModifyRequest:
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot modify request in current status",
		})
	case ErrQuoteExpired:
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Quote has expired",
		})
	case ErrQuoteNotActive:
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Quote is not active",
		})
	case ErrDuplicateActiveQuote:
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Custom request already has an active quote",
		})
	case ErrEmptyRequestItems:
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "At least one item is required",
		})
	case ErrInvalidPriority:
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid priority level",
		})
	case ErrInvalidQuoteStatus:
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid quote status",
		})
	default:
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}
}