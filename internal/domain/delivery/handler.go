package delivery

import (
	"errandShop/internal/presenter"
	"errandShop/internal/validation"
	"strconv"
	"time"

	"errandShop/internal/core/match"
	"errandShop/internal/core/types"
	"errandShop/internal/repos"
	"github.com/gofiber/fiber/v2"
)

// DeliveryHandler handles delivery HTTP requests
type DeliveryHandler struct {
	service     DeliveryService
	matcher     *match.Matcher
	addressRepo repos.AddressRepo
}

// NewDeliveryHandler creates a new delivery handler
func NewDeliveryHandler(service DeliveryService) *DeliveryHandler {
	return &DeliveryHandler{service: service}
}

// NewDeliveryHandlerWithCosting creates a new delivery handler with costing functionality
func NewDeliveryHandlerWithCosting(service DeliveryService, matcher *match.Matcher, addressRepo repos.AddressRepo) *DeliveryHandler {
	return &DeliveryHandler{
		service:     service,
		matcher:     matcher,
		addressRepo: addressRepo,
	}
}

// Customer Endpoints

// CreateDelivery creates a new delivery
func (h *DeliveryHandler) CreateDelivery(c *fiber.Ctx) error {
	var req CreateDeliveryRequest
	if err := c.BodyParser(&req); err != nil {
		return presenter.BadRequest(c, "Invalid request body")
	}

	if err := validation.ValidateStruct(&req); err != nil {
		return presenter.BadRequest(c, err.Error())
	}

	delivery, err := h.service.CreateDelivery(&req)
	if err != nil {
		return presenter.InternalServerError(c, err.Error())
	}

	return presenter.Success(c, "Delivery created successfully", delivery)
}

// GetDelivery gets delivery by ID
func (h *DeliveryHandler) GetDelivery(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return presenter.BadRequest(c, "Invalid delivery ID")
	}

	delivery, err := h.service.GetDelivery(uint(id))
	if err != nil {
		return presenter.NotFound(c, "Delivery not found")
	}

	return presenter.Success(c, "Delivery retrieved successfully", delivery)
}

// GetDeliveryByTrackingNumber gets delivery by tracking number
func (h *DeliveryHandler) GetDeliveryByTrackingNumber(c *fiber.Ctx) error {
	trackingNumber := c.Params("tracking_number")
	if trackingNumber == "" {
		return presenter.BadRequest(c, "Tracking number is required")
	}

	delivery, err := h.service.GetDeliveryByTrackingNumber(trackingNumber)
	if err != nil {
		return presenter.NotFound(c, "Delivery not found")
	}

	return presenter.Success(c, "Delivery retrieved successfully", delivery)
}

// GetDeliveryByOrderID gets delivery by order ID
func (h *DeliveryHandler) GetDeliveryByOrderID(c *fiber.Ctx) error {
	orderID := c.Params("order_id")
	if orderID == "" {
		return presenter.BadRequest(c, "Invalid order ID")
	}

	delivery, err := h.service.GetDeliveryByOrderID(orderID)
	if err != nil {
		return presenter.NotFound(c, "Delivery not found")
	}

	return presenter.Success(c, "Delivery retrieved successfully", delivery)
}

// GetDeliveryQuote gets delivery quote
func (h *DeliveryHandler) GetDeliveryQuote(c *fiber.Ctx) error {
	var req DeliveryQuoteRequest
	if err := c.BodyParser(&req); err != nil {
		return presenter.BadRequest(c, "Invalid request body")
	}

	if err := validation.ValidateStruct(&req); err != nil {
		return presenter.BadRequest(c, err.Error())
	}

	quote, err := h.service.GetDeliveryQuote(&req)
	if err != nil {
		return presenter.InternalServerError(c, "Failed to calculate quote")
	}

	return presenter.Success(c, "Quote calculated successfully", quote)
}

// GetTrackingUpdates gets tracking updates for a delivery
func (h *DeliveryHandler) GetTrackingUpdates(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return presenter.BadRequest(c, "Invalid delivery ID")
	}

	updates, err := h.service.GetTrackingUpdates(uint(id))
	if err != nil {
		return presenter.InternalServerError(c, "Failed to get tracking updates")
	}

	return presenter.Success(c, "Tracking updates retrieved successfully", updates)
}

// Driver Endpoints

// CreateDriver creates a new driver
func (h *DeliveryHandler) CreateDriver(c *fiber.Ctx) error {
	var req CreateDriverRequest
	if err := c.BodyParser(&req); err != nil {
		return presenter.BadRequest(c, "Invalid request body")
	}

	if err := validation.ValidateStruct(&req); err != nil {
		return presenter.BadRequest(c, err.Error())
	}

	driver, err := h.service.CreateDriver(&req)
	if err != nil {
		return presenter.InternalServerError(c, "Failed to create driver")
	}

	return presenter.Success(c, "Driver created successfully", driver)
}

// GetDriver gets a driver by ID
func (h *DeliveryHandler) GetDriver(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return presenter.BadRequest(c, "Invalid driver ID")
	}

	driver, err := h.service.GetDriver(uint(id))
	if err != nil {
		return presenter.NotFound(c, "Driver not found")
	}

	return presenter.Success(c, "Driver retrieved successfully", driver)
}

// GetDriverByUserID gets driver by user ID
func (h *DeliveryHandler) GetDriverByUserID(c *fiber.Ctx) error {
	userID, err := strconv.ParseUint(c.Params("user_id"), 10, 32)
	if err != nil {
		return presenter.BadRequest(c, "Invalid user ID")
	}

	driver, err := h.service.GetDriverByUserID(uint(userID))
	if err != nil {
		return presenter.NotFound(c, "Driver not found")
	}

	return presenter.Success(c, "Driver retrieved successfully", driver)
}

// UpdateDriverLocation updates driver location
func (h *DeliveryHandler) UpdateDriverLocation(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return presenter.BadRequest(c, "Invalid driver ID")
	}

	var req UpdateDriverLocationRequest
	if err := c.BodyParser(&req); err != nil {
		return presenter.BadRequest(c, "Invalid request body")
	}

	if err := validation.ValidateStruct(&req); err != nil {
		return presenter.BadRequest(c, err.Error())
	}

	err = h.service.UpdateDriverLocation(uint(id), &req)
	if err != nil {
		return presenter.InternalServerError(c, "Failed to update driver location")
	}

	return presenter.Success(c, "Driver location updated successfully", nil)
}

// ToggleDriverAvailability toggles driver availability
func (h *DeliveryHandler) ToggleDriverAvailability(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return presenter.BadRequest(c, "Invalid driver ID")
	}

	var req struct {
		Available bool `json:"available"`
	}
	if err := c.BodyParser(&req); err != nil {
		return presenter.BadRequest(c, "Invalid request body")
	}

	err = h.service.ToggleDriverAvailability(uint(id), req.Available)
	if err != nil {
		return presenter.InternalServerError(c, "Failed to update driver availability")
	}

	return presenter.Success(c, "Driver availability updated successfully", nil)
}

// GetDriverDeliveries gets deliveries for a driver
func (h *DeliveryHandler) GetDriverDeliveries(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return presenter.BadRequest(c, "Invalid driver ID")
	}

	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	deliveries, total, err := h.service.GetDeliveriesByDriver(uint(id), limit, offset)
	if err != nil {
		return presenter.InternalServerError(c, "Failed to get driver deliveries")
	}

	response := map[string]interface{}{
		"deliveries": deliveries,
		"total":      total,
		"limit":      limit,
		"offset":     offset,
	}

	return presenter.Success(c, "Driver deliveries retrieved successfully", response)
}

// UpdateDeliveryStatus updates delivery status (Admin only)
func (h *DeliveryHandler) UpdateDeliveryStatus(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return presenter.BadRequest(c, "Invalid delivery ID")
	}

	var req UpdateDeliveryStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return presenter.BadRequest(c, "Invalid request body")
	}

	if err := validation.ValidateStruct(&req); err != nil {
		return presenter.BadRequest(c, err.Error())
	}

	delivery, err := h.service.UpdateDeliveryStatus(uint(id), &req)
	if err != nil {
		return presenter.InternalServerError(c, err.Error())
	}

	return presenter.Success(c, "Delivery status updated successfully", delivery)
}

// AssignLogisticsProvider assigns a logistics provider to a delivery
// TODO: Implement logistics provider assignment functionality
func (h *DeliveryHandler) AssignLogisticsProvider(c *fiber.Ctx) error {
	return presenter.NotImplemented(c, "Logistics provider assignment not implemented")
}

// GetLogisticsProviders returns available logistics providers
func (h *DeliveryHandler) GetLogisticsProviders(c *fiber.Ctx) error {
	providers := []map[string]interface{}{
		{"id": "dhl", "name": "DHL Express", "description": "International and domestic express delivery"},
		{"id": "fedex", "name": "FedEx", "description": "Global courier delivery services"},
		{"id": "ups", "name": "UPS", "description": "United Parcel Service"},
		{"id": "gig", "name": "GIG Logistics", "description": "Nigerian logistics company"},
		{"id": "kwik", "name": "Kwik Delivery", "description": "On-demand delivery service"},
		{"id": "sendbox", "name": "Sendbox", "description": "E-commerce logistics platform"},
	}

	return presenter.Success(c, "Logistics providers retrieved successfully", providers)
}

// Admin Endpoints

// ListDeliveries lists all deliveries (admin)
func (h *DeliveryHandler) ListDeliveries(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	status := c.Query("status")

	var statusFilter *DeliveryStatus
	if status != "" {
		statusValue := DeliveryStatus(status)
		statusFilter = &statusValue
	}

	deliveries, total, err := h.service.ListDeliveries(limit, offset, statusFilter)
	if err != nil {
		return presenter.InternalServerError(c, err.Error())
	}

	response := map[string]interface{}{
		"deliveries": deliveries,
		"total":      total,
		"limit":      limit,
		"offset":     offset,
	}

	return presenter.Success(c, "Deliveries retrieved successfully", response)
}

// ListDrivers lists all drivers (admin)
func (h *DeliveryHandler) ListDrivers(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	isActiveStr := c.Query("is_active")

	var isActive *bool
	if isActiveStr != "" {
		value := isActiveStr == "true"
		isActive = &value
	}

	drivers, total, err := h.service.ListDrivers(limit, offset, isActive)
	if err != nil {
		return presenter.InternalServerError(c, err.Error())
	}

	response := map[string]interface{}{
		"drivers": drivers,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	}

	return presenter.Success(c, "Drivers retrieved successfully", response)
}

// AssignDriver assigns driver to delivery (admin)
func (h *DeliveryHandler) AssignDriver(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return presenter.BadRequest(c, "Invalid delivery ID")
	}

	var req AssignDriverRequest
	if err := c.BodyParser(&req); err != nil {
		return presenter.BadRequest(c, "Invalid request body")
	}

	if err := validation.ValidateStruct(&req); err != nil {
		return presenter.BadRequest(c, err.Error())
	}

	delivery, err := h.service.AssignDriver(uint(id), req.DriverID)
	if err != nil {
		return presenter.InternalServerError(c, err.Error())
	}

	return presenter.Success(c, "Driver assigned successfully", delivery)
}

// CancelDelivery cancels a delivery (admin)
func (h *DeliveryHandler) CancelDelivery(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return presenter.BadRequest(c, "Invalid delivery ID")
	}

	var req struct {
		Reason string `json:"reason" validate:"required,min=5,max=500"`
	}
	if err := c.BodyParser(&req); err != nil {
		return presenter.BadRequest(c, "Invalid request body")
	}

	if err := validation.ValidateStruct(&req); err != nil {
		return presenter.BadRequest(c, err.Error())
	}

	delivery, err := h.service.CancelDelivery(uint(id), req.Reason)
	if err != nil {
		return presenter.InternalServerError(c, err.Error())
	}

	return presenter.Success(c, "Delivery cancelled successfully", delivery)
}

// GetAvailableDrivers gets available drivers (admin)
func (h *DeliveryHandler) GetAvailableDrivers(c *fiber.Ctx) error {
	vehicleTypeStr := c.Query("vehicle_type")
	latStr := c.Query("latitude")
	lngStr := c.Query("longitude")
	radiusStr := c.Query("radius", "10")

	var vehicleType *VehicleType
	if vehicleTypeStr != "" {
		vt := VehicleType(vehicleTypeStr)
		vehicleType = &vt
	}

	var lat, lng *float64
	if latStr != "" && lngStr != "" {
		if latVal, err := strconv.ParseFloat(latStr, 64); err == nil {
			lat = &latVal
		}
		if lngVal, err := strconv.ParseFloat(lngStr, 64); err == nil {
			lng = &lngVal
		}
	}

	radius, _ := strconv.ParseFloat(radiusStr, 64)

	drivers, err := h.service.GetAvailableDrivers(vehicleType, lat, lng, radius)
	if err != nil {
		return presenter.InternalServerError(c, err.Error())
	}

	return presenter.Success(c, "Available drivers retrieved successfully", drivers)
}

// GetDeliveryStats gets delivery statistics (admin)
func (h *DeliveryHandler) GetDeliveryStats(c *fiber.Ctx) error {
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	var startDate, endDate *time.Time
	if startDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = &parsed
		}
	}
	if endDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = &parsed
		}
	}

	stats, err := h.service.GetDeliveryStats(startDate, endDate)
	if err != nil {
		return presenter.InternalServerError(c, err.Error())
	}

	return presenter.Success(c, "Delivery statistics retrieved successfully", stats)
}

// GetDriverStats gets driver statistics (admin)
func (h *DeliveryHandler) GetDriverStats(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return presenter.BadRequest(c, "Invalid driver ID")
	}

	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	var startDate, endDate *time.Time
	if startDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = &parsed
		}
	}
	if endDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = &parsed
		}
	}

	stats, err := h.service.GetDriverStats(uint(id), startDate, endDate)
	if err != nil {
		return presenter.InternalServerError(c, err.Error())
	}

	return presenter.Success(c, "Driver statistics retrieved successfully", stats)
}

// EstimateDelivery handles POST /api/v1/delivery/estimate
func (h *DeliveryHandler) EstimateDelivery(c *fiber.Ctx) error {
	// Check if costing functionality is available
	if h.matcher == nil || h.addressRepo == nil {
		return presenter.BadRequest(c, "Delivery costing functionality not available")
	}

	// Parse request
	var req types.DeliveryEstimateRequest
	if err := c.BodyParser(&req); err != nil {
		return presenter.BadRequest(c, "Invalid request body")
	}

	// Validate required fields
	if req.AddressID == "" {
		return presenter.BadRequest(c, "addressId is required")
	}

	// Get user ID from JWT context if available, otherwise from request
	userID := c.Locals("userID")
	var userIDStr string
	if userID != nil {
		userIDStr = userID.(string)
	} else if req.UserID != "" {
		userIDStr = req.UserID
	} else {
		return presenter.BadRequest(c, "userID is required when not authenticated")
	}

	// Fetch address by ID
	address, err := h.addressRepo.GetByID(userIDStr, req.AddressID)
	if err != nil {
		return presenter.NotFound(c, "Address not found or does not belong to user")
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
	// Check if costing functionality is available
	if h.matcher == nil || h.addressRepo == nil {
		return presenter.BadRequest(c, "Delivery costing functionality not available")
	}

	// Parse request
	var req types.OrderConfirmRequest
	if err := c.BodyParser(&req); err != nil {
		return presenter.BadRequest(c, "Invalid request body")
	}

	// Validate required fields
	if req.AddressID == "" {
		return presenter.BadRequest(c, "addressId is required")
	}

	if req.ClientPrice <= 0 {
		return presenter.BadRequest(c, "clientPrice must be greater than 0")
	}

	// Get user ID from context
	userID := c.Locals("userID")
	if userID == nil {
		userID = "U-TEST" // Default for testing
	}

	// Fetch address by ID
	address, err := h.addressRepo.GetByID(userID.(string), req.AddressID)
	if err != nil {
		return presenter.NotFound(c, "Address not found or does not belong to user")
	}

	// Recompute delivery price using the same algorithm
	matchResult, noMatchResult := h.matcher.MatchAddress(address.Text)

	if matchResult == nil {
		// Address cannot be matched to any zone
		return c.Status(fiber.StatusUnprocessableEntity).JSON(map[string]interface{}{
			"error":       "no_delivery_zone",
			"message":     "Address cannot be matched to any delivery zone",
			"suggestions": noMatchResult.Suggestions,
		})
	}

	// Check if client price matches computed price
	if req.ClientPrice != matchResult.Price {
		return presenter.Conflict(c, "Client price does not match computed delivery price")
	}

	// Price matches - order can be confirmed
	return presenter.Success(c, "Order delivery price confirmed", map[string]interface{}{
		"addressId":      req.AddressID,
		"confirmedPrice": matchResult.Price,
		"zoneId":         matchResult.ZoneID,
		"zoneName":       matchResult.ZoneName,
		"matchedBy":      matchResult.MatchedBy,
		"confidence":     matchResult.Confidence,
	})
}

// GetDeliveryByID gets delivery by ID
func (h *DeliveryHandler) GetDeliveryByID(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return presenter.BadRequest(c, "Invalid delivery ID")
	}

	delivery, err := h.service.GetDeliveryByID(uint(id))
	if err != nil {
		return presenter.NotFound(c, "Delivery not found")
	}

	return presenter.Success(c, "Delivery retrieved successfully", delivery)
}
