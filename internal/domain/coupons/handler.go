package coupons

import (
	"errandShop/internal/presenter"
	"errandShop/internal/validation"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// Admin Coupon Management Endpoints

// CreateCoupon creates a new coupon
// POST /api/v1/admin/coupons
func (h *Handler) CreateCoupon(c *fiber.Ctx) error {
	var req CreateCouponRequest
	if err := c.BodyParser(&req); err != nil {
		return presenter.BadRequest(c, "Invalid request body")
	}

	if err := validation.ValidateStruct(&req); err != nil {
		return presenter.BadRequest(c, err.Error())
	}

	// Get the user ID from context (set by JWT middleware)
	userIDInterface := c.Locals("userID")
	if userIDInterface == nil {
		return presenter.Unauthorized(c, "User not authenticated")
	}

	userID, ok := userIDInterface.(uuid.UUID)
	if !ok {
		return presenter.Unauthorized(c, "Invalid user ID")
	}

	// Use UUID directly as pointer
	var createdByUserID *uuid.UUID
	if userID != uuid.Nil {
		createdByUserID = &userID
	}

	coupon, err := h.service.CreateCoupon(req, createdByUserID)
	if err != nil {
		return presenter.InternalServerError(c, "Failed to create coupon")
	}

	return presenter.Success(c, "Coupon created successfully", coupon)
}

// GetCoupon gets a single coupon by ID
// GET /api/v1/admin/coupons/:id
func (h *Handler) GetCoupon(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return presenter.BadRequest(c, "Invalid coupon ID")
	}

	coupon, err := h.service.GetCoupon(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return presenter.NotFound(c, "Coupon not found")
		}
		return presenter.InternalServerError(c, "Failed to get coupon")
	}

	return presenter.Success(c, "Coupon retrieved successfully", coupon)
}

// UpdateCoupon updates an existing coupon
// PUT /api/v1/admin/coupons/:id
func (h *Handler) UpdateCoupon(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return presenter.BadRequest(c, "Invalid coupon ID")
	}

	var req UpdateCouponRequest
	if err := c.BodyParser(&req); err != nil {
		return presenter.BadRequest(c, "Invalid request body")
	}

	if err := validation.ValidateStruct(&req); err != nil {
		return presenter.BadRequest(c, "Validation failed")
	}

	coupon, err := h.service.UpdateCoupon(id, req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return presenter.NotFound(c, "Coupon not found")
		}
		return presenter.InternalServerError(c, "Failed to update coupon")
	}

	return presenter.Success(c, "Coupon updated successfully", coupon)
}

// DeleteCoupon deletes a coupon
// DELETE /api/v1/admin/coupons/:id
func (h *Handler) DeleteCoupon(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return presenter.BadRequest(c, "Invalid coupon ID")
	}

	err = h.service.DeleteCoupon(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return presenter.NotFound(c, "Coupon not found")
		}
		return presenter.InternalServerError(c, "Failed to delete coupon")
	}

	return presenter.Success(c, "Coupon deleted successfully", nil)
}

// ListCoupons lists all coupons with pagination and filters
// GET /api/v1/admin/coupons
func (h *Handler) ListCoupons(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	filters := make(map[string]interface{})

	// Apply filters
	if isActive := c.Query("is_active"); isActive != "" {
		if active, err := strconv.ParseBool(isActive); err == nil {
			filters["is_active"] = active
		}
	}

	if couponType := c.Query("type"); couponType != "" {
		filters["type"] = couponType
	}

	if createdBy := c.Query("created_by"); createdBy != "" {
		filters["created_by"] = createdBy
	}

	if search := c.Query("search"); search != "" {
		filters["search"] = search
	}

	if linkedUserID := c.Query("linked_user_id"); linkedUserID != "" {
		if userID, err := uuid.Parse(linkedUserID); err == nil {
			filters["linked_user_id"] = userID
		}
	}

	coupons, err := h.service.ListCoupons(page, limit, filters)
	if err != nil {
		return presenter.InternalServerError(c, "Failed to list coupons")
	}

	return presenter.Success(c, "Coupons retrieved successfully", coupons)
}

// ToggleCouponActive toggles the active status of a coupon
// POST /api/v1/admin/coupons/:id/toggle
func (h *Handler) ToggleCouponActive(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return presenter.BadRequest(c, "Invalid coupon ID")
	}

	coupon, err := h.service.ToggleCouponActive(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return presenter.NotFound(c, "Coupon not found")
		}
		return presenter.InternalServerError(c, "Failed to toggle coupon status")
	}

	return presenter.Success(c, "Coupon status toggled successfully", coupon)
}

// User Coupon Operations

// GetAvailableCoupons gets available coupons for a user
// GET /api/v1/user/coupons/available
func (h *Handler) GetAvailableCoupons(c *fiber.Ctx) error {
	userIDInterface := c.Locals("userID")
	if userIDInterface == nil {
		return presenter.Unauthorized(c, "User not authenticated")
	}

	userID, ok := userIDInterface.(uuid.UUID)
	if !ok {
		return presenter.Unauthorized(c, "Invalid user ID")
	}

	if userID == uuid.Nil {
		return presenter.BadRequest(c, "Invalid user ID")
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	coupons, err := h.service.GetAvailableCoupons(userID, page, limit)
	if err != nil {
		return presenter.InternalServerError(c, "Failed to get available coupons")
	}

	return presenter.Success(c, "Available coupons retrieved successfully", coupons)
}

// ValidateCoupon validates a coupon code
// POST /api/v1/user/coupons/validate
func (h *Handler) ValidateCoupon(c *fiber.Ctx) error {
	userIDInterface := c.Locals("userID")
	if userIDInterface == nil {
		return presenter.Unauthorized(c, "User not authenticated")
	}

	userID, ok := userIDInterface.(uuid.UUID)
	if !ok {
		return presenter.Unauthorized(c, "Invalid user ID")
	}

	if userID == uuid.Nil {
		return presenter.BadRequest(c, "Invalid user ID")
	}

	var req ValidateCouponRequest
	if err := c.BodyParser(&req); err != nil {
		return presenter.BadRequest(c, "Invalid request body")
	}

	req.UserID = userID // Set user ID from context

	if err := validation.ValidateStruct(&req); err != nil {
		return presenter.BadRequest(c, "Validation failed")
	}

	validation, err := h.service.ValidateCoupon(req)
	if err != nil {
		return presenter.InternalServerError(c, "Failed to validate coupon")
	}

	return presenter.Success(c, "Coupon validation completed", validation)
}

// ApplyCoupon applies a coupon to an order
// POST /api/v1/user/coupons/apply
func (h *Handler) ApplyCoupon(c *fiber.Ctx) error {
	userIDInterface := c.Locals("userID")
	if userIDInterface == nil {
		return presenter.Unauthorized(c, "User not authenticated")
	}

	userID, ok := userIDInterface.(uuid.UUID)
	if !ok {
		return presenter.Unauthorized(c, "Invalid user ID")
	}

	if userID == uuid.Nil {
		return presenter.BadRequest(c, "Invalid user ID")
	}

	var req ApplyCouponRequest
	if err := c.BodyParser(&req); err != nil {
		return presenter.BadRequest(c, "Invalid request body")
	}

	req.UserID = userID // Set user ID from context

	if err := validation.ValidateStruct(&req); err != nil {
		return presenter.BadRequest(c, "Validation failed")
	}

	result, err := h.service.ApplyCoupon(req)
	if err != nil {
		return presenter.InternalServerError(c, "Failed to apply coupon")
	}

	if !result.Valid {
		return presenter.BadRequest(c, result.Message)
	}

	return presenter.Success(c, "Coupon applied successfully", result)
}

// Refund Credits Operations

// GetUserRefundCredits gets user's refund credits
// GET /api/v1/user/refund-credits
func (h *Handler) GetUserRefundCredits(c *fiber.Ctx) error {
	userIDInterface := c.Locals("userID")
	if userIDInterface == nil {
		return presenter.Unauthorized(c, "User not authenticated")
	}

	userID, ok := userIDInterface.(uuid.UUID)
	if !ok {
		return presenter.Unauthorized(c, "Invalid user ID")
	}

	if userID == uuid.Nil {
		return presenter.BadRequest(c, "Invalid user ID")
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	credits, err := h.service.GetUserRefundCredits(userID, page, limit)
	if err != nil {
		return presenter.InternalServerError(c, "Failed to get refund credits")
	}

	return presenter.Success(c, "Refund credits retrieved successfully", credits)
}

// ConvertRefundToCredit converts refund credit to coupon
// POST /api/v1/user/refund-credits/convert
func (h *Handler) ConvertRefundToCredit(c *fiber.Ctx) error {
	userIDInterface := c.Locals("userID")
	if userIDInterface == nil {
		return presenter.Unauthorized(c, "User not authenticated")
	}

	userID, ok := userIDInterface.(uuid.UUID)
	if !ok {
		return presenter.Unauthorized(c, "Invalid user ID")
	}

	if userID == uuid.Nil {
		return presenter.BadRequest(c, "Invalid user ID")
	}

	var req ConvertRefundRequest
	if err := c.BodyParser(&req); err != nil {
		return presenter.BadRequest(c, "Invalid request body")
	}

	if err := validation.ValidateStruct(&req); err != nil {
		return presenter.BadRequest(c, "Validation failed")
	}

	coupon, err := h.service.ConvertRefundToCredit(req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return presenter.NotFound(c, "Refund credit not found")
		}
		if strings.Contains(err.Error(), "already converted") {
			return presenter.BadRequest(c, "Refund credit already converted")
		}
		return presenter.InternalServerError(c, "Failed to convert refund credit")
	}

	return presenter.Success(c, "Refund credit converted to coupon successfully", coupon)
}

// System Operations (Admin only)

// CreateRefundCredit creates a refund credit for a user
// POST /api/v1/system/refund-credits/create
func (h *Handler) CreateRefundCredit(c *fiber.Ctx) error {
	var req CreateRefundCreditRequest
	if err := c.BodyParser(&req); err != nil {
		return presenter.BadRequest(c, "Invalid request body")
	}

	if err := validation.ValidateStruct(&req); err != nil {
		return presenter.BadRequest(c, "Validation failed")
	}

	credit, err := h.service.CreateRefundCredit(req)
	if err != nil {
		return presenter.InternalServerError(c, "Failed to create refund credit")
	}

	return presenter.Success(c, "Refund credit created successfully", credit)
}

// GenerateRefundCoupon generates a refund coupon for an order
// POST /api/v1/system/coupons/generate-refund
func (h *Handler) GenerateRefundCoupon(c *fiber.Ctx) error {
	type GenerateRefundCouponRequest struct {
		OrderID      uuid.UUID `json:"orderId" validate:"required"`
		UserID       uuid.UUID `json:"userId" validate:"required"`
		RefundAmount float64   `json:"refundAmount" validate:"required,gt=0"`
	}

	var req GenerateRefundCouponRequest
	if err := c.BodyParser(&req); err != nil {
		return presenter.BadRequest(c, "Invalid request body")
	}

	if err := validation.ValidateStruct(&req); err != nil {
		return presenter.BadRequest(c, "Validation failed")
	}

	coupon, err := h.service.GenerateRefundCoupon(req.OrderID, req.UserID, req.RefundAmount)
	if err != nil {
		return presenter.InternalServerError(c, "Failed to generate refund coupon")
	}

	return presenter.Success(c, "Refund coupon generated successfully", coupon)
}

// AutoGenerateUserCoupon generates a user-specific coupon
// POST /api/v1/system/coupons/auto-generate
func (h *Handler) AutoGenerateUserCoupon(c *fiber.Ctx) error {
	type AutoGenerateUserCouponRequest struct {
		UserID      uuid.UUID  `json:"userId" validate:"required"`
		Type        CouponType `json:"type" validate:"required,oneof=percentage fixed"`
		Value       float64    `json:"value" validate:"required,gt=0"`
		Description string     `json:"description" validate:"required"`
	}

	var req AutoGenerateUserCouponRequest
	if err := c.BodyParser(&req); err != nil {
		return presenter.BadRequest(c, "Invalid request body")
	}

	if err := validation.ValidateStruct(&req); err != nil {
		return presenter.BadRequest(c, "Validation failed")
	}

	coupon, err := h.service.AutoGenerateUserCoupon(req.UserID, req.Type, req.Value, req.Description)
	if err != nil {
		return presenter.InternalServerError(c, "Failed to generate user coupon")
	}

	return presenter.Success(c, "User coupon generated successfully", coupon)
}

// MobileAutoGenerateCoupon allows mobile users to generate their own coupons
// POST /api/v1/user/coupons/generate
func (h *Handler) MobileAutoGenerateCoupon(c *fiber.Ctx) error {
	var req MobileAutoGenerateCouponRequest
	if err := c.BodyParser(&req); err != nil {
		return presenter.BadRequest(c, "Invalid request body")
	}

	if err := validation.ValidateStruct(&req); err != nil {
		return presenter.BadRequest(c, err.Error())
	}

	// Get the user ID from context (set by JWT middleware)
	userIDInterface := c.Locals("userID")
	if userIDInterface == nil {
		return presenter.Unauthorized(c, "User not authenticated")
	}

	userID, ok := userIDInterface.(uuid.UUID)
	if !ok {
		return presenter.Unauthorized(c, "Invalid user ID")
	}

	// Check if userID is valid
	if userID == uuid.Nil {
		return presenter.Unauthorized(c, "Invalid user ID")
	}

	coupon, err := h.service.MobileAutoGenerateCoupon(userID, req)
	if err != nil {
		if strings.Contains(err.Error(), "daily coupon generation limit") {
			return presenter.BadRequest(c, err.Error())
		}
		if strings.Contains(err.Error(), "cannot exceed") {
			return presenter.BadRequest(c, err.Error())
		}
		return presenter.InternalServerError(c, "Failed to generate coupon")
	}

	return presenter.Success(c, "Coupon generated successfully", coupon)
}

// Analytics

// GetCouponStats gets coupon statistics
// GET /api/v1/admin/coupons/stats
func (h *Handler) GetCouponStats(c *fiber.Ctx) error {
	stats, err := h.service.GetCouponStats()
	if err != nil {
		return presenter.InternalServerError(c, "Failed to get coupon statistics")
	}

	return presenter.Success(c, "Coupon statistics retrieved successfully", stats)
}

// GetPublicCoupons gets publicly available coupons (no auth required)
func (h *Handler) GetPublicCoupons(c *fiber.Ctx) error {
	coupons, err := h.service.GetPublicCoupons()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    coupons,
	})
}

// ValidatePublicCoupon validates a coupon without requiring authentication
// POST /public/coupons/validate
func (h *Handler) ValidatePublicCoupon(c *fiber.Ctx) error {
	var req struct {
		Code        string  `json:"code" validate:"required"`
		OrderAmount float64 `json:"order_amount" validate:"required,min=0"`
	}

	if err := c.BodyParser(&req); err != nil {
		return presenter.BadRequest(c, "Invalid request body")
	}

	if err := validation.ValidateStruct(&req); err != nil {
		return presenter.BadRequest(c, "Validation failed")
	}

	// Create validation request without user ID for public validation
	validationReq := ValidateCouponRequest{
		Code:        req.Code,
		OrderAmount: req.OrderAmount,
		// UserID is zero value (will be handled in service layer)
	}

	validation, err := h.service.ValidateCoupon(validationReq)
	if err != nil {
		return presenter.InternalServerError(c, "Failed to validate coupon")
	}

	return presenter.Success(c, "Coupon validation completed", validation)
}