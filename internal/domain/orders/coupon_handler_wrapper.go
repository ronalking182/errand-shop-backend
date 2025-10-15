package orders

import (
	"errandShop/internal/domain/coupons"
	"errandShop/internal/presenter"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// CouponHandler wraps the coupon service for order-related coupon operations
type CouponHandler struct {
	couponService coupons.Service
}

// NewCouponHandler creates a new coupon handler wrapper
func NewCouponHandler(couponService coupons.Service) *CouponHandler {
	return &CouponHandler{
		couponService: couponService,
	}
}

// ValidateCoupon validates a coupon code for order usage
func (h *CouponHandler) ValidateCoupon(c *fiber.Ctx) error {
	userIDRaw := c.Locals("userID")
	if userIDRaw == nil {
		return presenter.Unauthorized(c, "User not authenticated")
	}

	// Handle UUID from JWT claims
	userID, ok := userIDRaw.(uuid.UUID)
	if !ok {
		return presenter.Unauthorized(c, "Invalid user ID format")
	}

	if userID == uuid.Nil {
		return presenter.BadRequest(c, "Invalid user ID")
	}

	var req coupons.ValidateCouponRequest
	if err := c.BodyParser(&req); err != nil {
		return presenter.BadRequest(c, "Invalid request body")
	}

	// Set user ID from context
	req.UserID = userID

	validation, err := h.couponService.ValidateCoupon(req)
	if err != nil {
		return presenter.InternalServerError(c, "Failed to validate coupon")
	}

	return presenter.Success(c, "Coupon validation completed", validation)
}