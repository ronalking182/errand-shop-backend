package handlers

import (
	"encoding/json"
	"log"
	"os"

	"errandShop/internal/core/match"
	"errandShop/internal/core/types"
	"errandShop/internal/repos"
	"github.com/gofiber/fiber/v2"
)

// DeliveryHandler handles delivery-related endpoints
type DeliveryHandler struct {
	matcher     *match.Matcher
	addressRepo repos.AddressRepo
}

// NewDeliveryHandler creates a new delivery handler
func NewDeliveryHandler(zonesFilePath string, addressRepo repos.AddressRepo) (*DeliveryHandler, error) {
	// Load delivery zones from JSON file
	zones, err := loadDeliveryZones(zonesFilePath)
	if err != nil {
		return nil, err
	}
	
	matcher := match.NewMatcher(zones)
	
	return &DeliveryHandler{
		matcher:     matcher,
		addressRepo: addressRepo,
	}, nil
}

// EstimateDelivery handles POST /api/v1/delivery/estimate
func (h *DeliveryHandler) EstimateDelivery(c *fiber.Ctx) error {
	// Parse request
	var req types.DeliveryEstimateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}
	
	// Validate required fields
	if req.AddressID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{
			Error:   "validation_error",
			Message: "addressId is required",
		})
	}
	
	// Get user ID from context (dummy middleware sets this)
	userID := c.Locals("userID")
	if userID == nil {
		userID = "U-TEST" // Default for testing
	}
	
	// Fetch address by ID
	address, err := h.addressRepo.GetByID(userID.(string), req.AddressID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(types.ErrorResponse{
			Error:   "address_not_found",
			Message: "Address not found or does not belong to user",
		})
	}
	
	// Match address to delivery zone
	matchResult, noMatchResult := h.matcher.MatchAddress(address.Text)
	
	if matchResult != nil {
		// Success - return match result
		return c.Status(fiber.StatusOK).JSON(matchResult)
	}
	
	// No match found - return suggestions
	return c.Status(fiber.StatusUnprocessableEntity).JSON(noMatchResult)
}

// ConfirmOrder handles POST /api/v1/orders/confirm
func (h *DeliveryHandler) ConfirmOrder(c *fiber.Ctx) error {
	// Parse request
	var req types.OrderConfirmRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}
	
	// Validate required fields
	if req.AddressID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{
			Error:   "validation_error",
			Message: "addressId is required",
		})
	}
	
	if req.ClientPrice <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{
			Error:   "validation_error",
			Message: "clientPrice must be greater than 0",
		})
	}
	
	// Get user ID from context
	userID := c.Locals("userID")
	if userID == nil {
		userID = "U-TEST" // Default for testing
	}
	
	// Fetch address by ID
	address, err := h.addressRepo.GetByID(userID.(string), req.AddressID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(types.ErrorResponse{
			Error:   "address_not_found",
			Message: "Address not found or does not belong to user",
		})
	}
	
	// Recompute delivery price using the same algorithm
	matchResult, noMatchResult := h.matcher.MatchAddress(address.Text)
	
	if matchResult == nil {
		// Address cannot be matched to any zone
		return c.Status(fiber.StatusUnprocessableEntity).JSON(map[string]interface{}{
			"error":   "no_delivery_zone",
			"message": "Address cannot be matched to any delivery zone",
			"suggestions": noMatchResult.Suggestions,
		})
	}
	
	// Check if client price matches computed price
	if req.ClientPrice != matchResult.Price {
		return c.Status(fiber.StatusConflict).JSON(types.ErrorResponse{
			Error:   "price_mismatch",
			Message: "Client price does not match computed delivery price",
		})
	}
	
	// Price matches - order can be confirmed
	return c.Status(fiber.StatusOK).JSON(map[string]interface{}{
		"success":        true,
		"message":        "Order delivery price confirmed",
		"addressId":      req.AddressID,
		"confirmedPrice": matchResult.Price,
		"zoneId":         matchResult.ZoneID,
		"zoneName":       matchResult.ZoneName,
		"matchedBy":      matchResult.MatchedBy,
		"confidence":     matchResult.Confidence,
	})
}

// GetMatcher returns the matcher instance
func (h *DeliveryHandler) GetMatcher() *match.Matcher {
	return h.matcher
}

// loadDeliveryZones loads delivery zones from JSON file
func loadDeliveryZones(filePath string) ([]types.DeliveryZone, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("Error reading delivery zones file: %v", err)
		return nil, err
	}
	
	var zones []types.DeliveryZone
	if err := json.Unmarshal(data, &zones); err != nil {
		log.Printf("Error parsing delivery zones JSON: %v", err)
		return nil, err
	}
	
	log.Printf("Loaded %d delivery zones", len(zones))
	return zones, nil
}