package orders

import (
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type CartHandler struct {
	service *Service
}

func NewCartHandler(service *Service) *CartHandler {
	return &CartHandler{service: service}
}

// GetCart godoc
// @Summary Get user's cart
// @Description Get the current user's shopping cart
// @Tags Cart
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} CartResponse
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/cart [get]
func (h *CartHandler) GetCart(c *fiber.Ctx) error {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	cart, err := h.service.GetCart(c.Context(), userID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get cart",
		})
	}

	return c.JSON(cart)
}

// AddToCart godoc
// @Summary Add item to cart
// @Description Add a product to the user's shopping cart
// @Tags Cart
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body AddToCartRequest true "Add to cart request"
// @Success 200 {object} CartResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/cart/items [post]
func (h *CartHandler) AddToCart(c *fiber.Ctx) error {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	var req AddToCartRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate request
	if req.ProductID == uuid.Nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Product ID is required",
		})
	}
	if req.Quantity <= 0 || req.Quantity > 100 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Quantity must be between 1 and 100",
		})
	}

	cart, err := h.service.AddToCart(c.Context(), userID, req)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to add item to cart",
		})
	}

	return c.JSON(cart)
}

// UpdateCartItem godoc
// @Summary Update cart item quantity
// @Description Update the quantity of an item in the cart
// @Tags Cart
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param itemId path string true "Cart Item ID"
// @Param request body UpdateCartItemRequest true "Update cart item request"
// @Success 200 {object} CartResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/cart/items/{itemId} [put]
func (h *CartHandler) UpdateCartItem(c *fiber.Ctx) error {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	itemIDStr := c.Params("itemId")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid item ID",
		})
	}

	var req UpdateCartItemRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate request
	if req.Quantity <= 0 || req.Quantity > 100 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Quantity must be between 1 and 100",
		})
	}

	cart, err := h.service.UpdateCartItem(c.Context(), userID, itemID, req)
	if err != nil {
		if err.Error() == "cart item not found" {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"error": "Cart item not found",
			})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update cart item",
		})
	}

	return c.JSON(cart)
}

// RemoveFromCart godoc
// @Summary Remove item from cart
// @Description Remove an item from the user's shopping cart
// @Tags Cart
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param itemId path string true "Cart Item ID"
// @Success 200 {object} CartResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/cart/items/{itemId} [delete]
func (h *CartHandler) RemoveFromCart(c *fiber.Ctx) error {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	itemIDStr := c.Params("itemId")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid item ID",
		})
	}

	cart, err := h.service.RemoveFromCart(c.Context(), userID, itemID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to remove item from cart",
		})
	}

	return c.JSON(cart)
}

// ClearCart godoc
// @Summary Clear cart
// @Description Remove all items from the user's shopping cart
// @Tags Cart
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/cart [delete]
func (h *CartHandler) ClearCart(c *fiber.Ctx) error {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	err = h.service.ClearCart(c.Context(), userID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to clear cart",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Cart cleared successfully",
	})
}

// ValidateCoupon godoc
// @Summary Validate coupon for cart
// @Description Validate a coupon code against the current cart total
// @Tags Cart
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body ValidateCouponRequest true "Validate coupon request"
// @Success 200 {object} ValidateCouponResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/cart/validate-coupon [post]
func (h *CartHandler) ValidateCoupon(c *fiber.Ctx) error {
	// TODO: Add user authentication when coupon validation is implemented
	// _, err := getUserIDFromContext(c)
	// if err != nil {
	//	return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
	//		"error": "Unauthorized",
	//	})
	// }

	var req ValidateCouponRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate request
	if req.CouponCode == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Coupon code is required",
		})
	}
	if req.OrderSubtotal <= 0 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Order subtotal must be greater than 0",
		})
	}

	// TODO: Implement coupon validation through coupon service
	// For now, return a placeholder response
	response := ValidateCouponResponse{
		Valid:         false,
		DiscountAmount: 0,
		DiscountAmountNaira: 0,
		Message:       "Coupon validation not implemented yet",
	}

	return c.JSON(response)
}

// Helper function to get user ID from context
func getUserIDFromContext(c *fiber.Ctx) (uuid.UUID, error) {
	userIDRaw := c.Locals("userID")
	if userIDRaw == nil {
		return uuid.Nil, fiber.NewError(fiber.StatusUnauthorized, "User not authenticated")
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
		return uuid.Nil, fiber.NewError(fiber.StatusUnauthorized, "Invalid user ID format")
	}
}