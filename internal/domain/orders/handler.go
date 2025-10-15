package orders

import (
    "errors"
    "fmt"
    "log"
    "strings"

    "github.com/go-playground/validator/v10"
    "github.com/gofiber/fiber/v2"
    "github.com/google/uuid"
    "gorm.io/gorm"
)

type Handler struct {
	svc    *Service
	logger *log.Logger
}

func NewHandler(svc *Service) *Handler {
	return &Handler{
		svc:    svc,
		logger: log.Default(),
	}
}

var validate = validator.New()

// Helper methods
func (h *Handler) errorResponse(c *fiber.Ctx, statusCode int, message string, err error) error {
	if err != nil {
		h.logger.Printf("Error: %v", err)
	}
	return c.Status(statusCode).JSON(fiber.Map{
		"error":   true,
		"message": message,
	})
}

func (h *Handler) successResponse(c *fiber.Ctx, data interface{}, message string) error {
	response := fiber.Map{
		"error":   false,
		"message": message,
	}
	if data != nil {
		response["data"] = data
	}
	return c.JSON(response)
}

func (h *Handler) getUserID(c *fiber.Ctx) (uuid.UUID, error) {
	userID := c.Locals("userID")
	if userID == nil {
		return uuid.Nil, errors.New("user not authenticated")
	}
	
	// Handle different types from JWT claims
	switch v := userID.(type) {
	case uuid.UUID:
		// Direct UUID from JWT claims
		return v, nil
	case string:
		// Try to parse as UUID
		return uuid.Parse(v)
	case uint:
		// Convert uint to string and parse as UUID
		return uuid.Parse(fmt.Sprintf("%d", v))
	case float64:
		// Convert float64 to string and parse as UUID
		return uuid.Parse(fmt.Sprintf("%.0f", v))
	default:
		return uuid.Nil, errors.New("invalid user ID format")
	}
}

// List orders for authenticated user
func (h *Handler) List(c *fiber.Ctx) error {
	userID, err := h.getUserID(c)
	if err != nil {
		return h.errorResponse(c, fiber.StatusUnauthorized, "Authentication required", err)
	}

	var query ListQuery
	if err := c.QueryParser(&query); err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "Invalid query parameters", err)
	}

	// Set default values if not provided
	if query.Page == 0 {
		query.Page = 1
	}
	if query.Limit == 0 {
		query.Limit = 10
	}

	if err := validate.Struct(&query); err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "Validation failed", err)
	}

	result, err := h.svc.List(c.Context(), userID, query)
	if err != nil {
		return h.errorResponse(c, fiber.StatusInternalServerError, "Failed to fetch orders", err)
	}

	return h.successResponse(c, result, "Orders retrieved successfully")
}

func (h *Handler) Get(c *fiber.Ctx) error {
	userID, err := h.getUserID(c)
	if err != nil {
		return h.errorResponse(c, fiber.StatusUnauthorized, "Authentication required", err)
	}

	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "Invalid order ID", err)
	}

	order, err := h.svc.Get(c.Context(), id, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return h.errorResponse(c, fiber.StatusNotFound, "Order not found", err)
		}
		return h.errorResponse(c, fiber.StatusInternalServerError, "Failed to fetch order", err)
	}

	return h.successResponse(c, order, "Order retrieved successfully")
}

func (h *Handler) Create(c *fiber.Ctx) error {
	userID, err := h.getUserID(c)
	if err != nil {
		return h.errorResponse(c, fiber.StatusUnauthorized, "Authentication required", err)
	}

	var req CreateOrderRequest
	if err := c.BodyParser(&req); err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "Invalid request body", err)
	}

	if err := validate.Struct(&req); err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "Validation failed", err)
	}

	order, err := h.svc.CreateWithPayment(c.Context(), userID, req)
	if err != nil {
		if err.Error() == "insufficient stock" {
			return h.errorResponse(c, fiber.StatusBadRequest, "Insufficient stock for one or more items", err)
		}
		// Map expired custom request error to 400 to support user-facing popup
		if strings.Contains(err.Error(), "custom request") && strings.Contains(err.Error(), "has expired") {
			return h.errorResponse(c, fiber.StatusBadRequest, "Custom request has expired. Please create a new request.", err)
		}
		return h.errorResponse(c, fiber.StatusInternalServerError, "Failed to create order", err)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"error":   false,
		"message": "Order created successfully",
		"data":    order,
	})
}

// CreateFromCart creates an order from the user's cart
func (h *Handler) CreateFromCart(c *fiber.Ctx) error {
	userID, err := h.getUserID(c)
	if err != nil {
		return h.errorResponse(c, fiber.StatusUnauthorized, "Authentication required", err)
	}

	var req CreateOrderFromCartRequest
	if err := c.BodyParser(&req); err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "Invalid request body", err)
	}

	if err := validate.Struct(&req); err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "Validation failed", err)
	}



	order, err := h.svc.CreateFromCart(c.Context(), userID, req)
	if err != nil {
		if err.Error() == "insufficient stock" {
			return h.errorResponse(c, fiber.StatusBadRequest, "Insufficient stock for one or more items", err)
		}
		return h.errorResponse(c, fiber.StatusInternalServerError, "Failed to create order from cart", err)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"error":   false,
		"message": "Order created successfully from cart",
		"data":    order,
	})
}

func (h *Handler) UpdateStatus(c *fiber.Ctx) error {
	userID, err := h.getUserID(c)
	if err != nil {
		return h.errorResponse(c, fiber.StatusUnauthorized, "Authentication required", err)
	}

	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "Invalid order ID", err)
	}

	var req UpdateOrderStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "Invalid request body", err)
	}

	if err := validate.Struct(&req); err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "Validation failed", err)
	}

	if err := h.svc.UpdateStatus(c.Context(), id, userID, req.Status); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return h.errorResponse(c, fiber.StatusNotFound, "Order not found", err)
		}
		if err.Error() == "customers can only cancel orders" ||
			err.Error() == "can only cancel pending orders" {
			return h.errorResponse(c, fiber.StatusBadRequest, err.Error(), err)
		}
		return h.errorResponse(c, fiber.StatusInternalServerError, "Failed to update order status", err)
	}

	return h.successResponse(c, nil, "Order status updated successfully")
}

// CancelOrder cancels an order
func (h *Handler) CancelOrder(c *fiber.Ctx) error {
	userID, err := h.getUserID(c)
	if err != nil {
		return h.errorResponse(c, fiber.StatusUnauthorized, "Authentication required", err)
	}

	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "Invalid order ID", err)
	}

	var req CancelOrderRequest
	if err := c.BodyParser(&req); err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "Invalid request body", err)
	}

	if err := validate.Struct(&req); err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "Validation failed", err)
	}

	err = h.svc.CancelOrder(c.Context(), id, userID, req.Reason)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return h.errorResponse(c, fiber.StatusNotFound, "Order not found", err)
		}
		return h.errorResponse(c, fiber.StatusInternalServerError, "Failed to cancel order", err)
	}

	return h.successResponse(c, nil, "Order cancelled successfully")
}

// Admin endpoints
func (h *Handler) AdminList(c *fiber.Ctx) error {
	var query AdminListQuery
	if err := c.QueryParser(&query); err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "Invalid query parameters", err)
	}

	// Set default values for pagination if not provided
	if query.Page == 0 {
		query.Page = 1
	}
	if query.Limit == 0 {
		query.Limit = 20
	}

	if err := validate.Struct(&query); err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "Validation failed", err)
	}

	result, err := h.svc.AdminList(c.Context(), query)
	if err != nil {
		return h.errorResponse(c, fiber.StatusInternalServerError, "Failed to fetch orders", err)
	}

	return h.successResponse(c, result, "Orders retrieved successfully")
}

func (h *Handler) AdminGet(c *fiber.Ctx) error {
    idStr := c.Params("id")
    // If the dynamic :id route captured the "stats" path segment, delegate to stats handler
    if idStr == "stats" {
        return h.GetStats(c)
    }
    id, err := uuid.Parse(idStr)
    if err != nil {
        return h.errorResponse(c, fiber.StatusBadRequest, "Invalid order ID", err)
    }

	order, err := h.svc.AdminGet(c.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return h.errorResponse(c, fiber.StatusNotFound, "Order not found", err)
		}
		return h.errorResponse(c, fiber.StatusInternalServerError, "Failed to fetch order", err)
	}

	return h.successResponse(c, order, "Order retrieved successfully")
}

func (h *Handler) AdminUpdateStatus(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "Invalid order ID", err)
	}

	var req UpdateOrderStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "Invalid request body", err)
	}

	if err := validate.Struct(&req); err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "Validation failed", err)
	}

	if err := h.svc.AdminUpdateStatus(c.Context(), id, req.Status); err != nil {
		return h.errorResponse(c, fiber.StatusInternalServerError, "Failed to update order status", err)
	}

	return h.successResponse(c, nil, "Order status updated successfully")
}

func (h *Handler) AdminUpdatePaymentStatus(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "Invalid order ID", err)
	}

	var req UpdatePaymentStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "Invalid request body", err)
	}

	if err := validate.Struct(&req); err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "Validation failed", err)
	}

	if err := h.svc.AdminUpdatePaymentStatus(c.Context(), id, req.PaymentStatus); err != nil {
		return h.errorResponse(c, fiber.StatusInternalServerError, "Failed to update payment status", err)
	}

	return h.successResponse(c, nil, "Payment status updated successfully")
}



func (h *Handler) AdminCancelOrder(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "Invalid order ID", err)
	}

	var req CancelOrderRequest
	if err := c.BodyParser(&req); err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "Invalid request body", err)
	}

	if err := validate.Struct(&req); err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "Validation failed", err)
	}

	err = h.svc.AdminCancelOrder(c.Context(), id, req.Reason)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return h.errorResponse(c, fiber.StatusNotFound, "Order not found", err)
		}
		return h.errorResponse(c, fiber.StatusInternalServerError, "Failed to cancel order", err)
	}

	return h.successResponse(c, nil, "Order cancelled successfully")
}

func (h *Handler) GetStats(c *fiber.Ctx) error {
	stats, err := h.svc.GetStats(c.Context())
	if err != nil {
		return h.errorResponse(c, fiber.StatusInternalServerError, "Failed to fetch order statistics", err)
	}

	return h.successResponse(c, stats, "Order statistics retrieved successfully")
}
