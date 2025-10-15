package customers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"errandShop/internal/presenter"
	"errandShop/internal/validation"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// Helper methods
func (h *Handler) getCurrentUserID(c *fiber.Ctx) (uuid.UUID, error) {
	userID := c.Locals("userID")
	if userID == nil {
		return uuid.Nil, errors.New("user not authenticated")
	}

	switch v := userID.(type) {
	case uuid.UUID:
		return v, nil
	case string:
		return uuid.Parse(v)
	case uint:
		return uuid.Parse(fmt.Sprintf("%d", v))
	case float64:
		return uuid.Parse(fmt.Sprintf("%.0f", v))
	default:
		return uuid.Nil, errors.New("invalid user ID type")
	}
}

func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func getBoolValue(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

// Customer handlers
func (h *Handler) CreateCustomer(c *fiber.Ctx) error {
	var req CreateCustomerRequest
	if err := c.BodyParser(&req); err != nil {
		return presenter.BadRequest(c, "Invalid request body")
	}

	// Get user ID from JWT token
	userID, err := h.getCurrentUserID(c)
	if err != nil {
		return presenter.Unauthorized(c, "User not authenticated")
	}
	req.UserID = userID

	if err := validation.ValidateStruct(&req); err != nil {
		return presenter.ValidationErrorResponse(c, err)
	}

	customerResult, err := h.service.CreateCustomer(&req)
	if err != nil {
		if err.Error() == "customer profile already exists for this user" {
			return presenter.Conflict(c, err.Error())
		}
		return presenter.InternalServerError(c, "Failed to create customer")
	}

	// Cast the interface{} result back to *CustomerResponse
	customer, ok := customerResult.(*CustomerResponse)
	if !ok {
		return presenter.InternalServerError(c, "Invalid response type")
	}

	fmt.Printf("[DEBUG] Customer created: %+v\n", customer)
	return presenter.Created(c, customer)
}

func (h *Handler) GetCustomerProfile(c *fiber.Ctx) error {
	userID, err := h.getCurrentUserID(c)
	if err != nil {
		return presenter.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
	}

	customer, err := h.service.GetCustomerByUserID(userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return presenter.ErrorResponse(c, http.StatusNotFound, "Customer profile not found")
		}
		return presenter.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve customer profile")
	}

	fmt.Printf("[DEBUG] Retrieved customer profile for user %d: %+v\n", userID, customer)
	return presenter.SuccessResponse(c, "Customer profile retrieved successfully", customer)
}

func (h *Handler) CreateAddress(c *fiber.Ctx) error {
	userID, err := h.getCurrentUserID(c)
	if err != nil {
		return presenter.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
	}

	// Get customer by user ID first
	customer, err := h.service.GetCustomerByUserID(userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return presenter.ErrorResponse(c, http.StatusNotFound, "Customer profile not found")
		}
		return presenter.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve customer profile")
	}

	var req CreateAddressRequest
	if err := c.BodyParser(&req); err != nil {
		return presenter.BadRequest(c, "Invalid request body")
	}

	if err := validation.ValidateStruct(&req); err != nil {
		return presenter.BadRequest(c, "Validation failed")
	}

	address, err := h.service.CreateAddress(customer.ID, &req)
	if err != nil {
		return presenter.InternalServerError(c, "Failed to create address")
	}

	fmt.Printf("[DEBUG] Address created for customer %d: %+v\n", customer.ID, address)
	return presenter.Created(c, address)
}

func (h *Handler) GetCustomerAddresses(c *fiber.Ctx) error {
	userID, err := h.getCurrentUserID(c)
	if err != nil {
		return presenter.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
	}

	// Get customer by user ID first
	customer, err := h.service.GetCustomerByUserID(userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return presenter.ErrorResponse(c, http.StatusNotFound, "Customer profile not found")
		}
		return presenter.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve customer profile")
	}

	addresses, err := h.service.GetCustomerAddresses(customer.ID)
	if err != nil {
		return presenter.InternalServerError(c, "Failed to get addresses")
	}

	fmt.Printf("[DEBUG] Retrieved %d addresses for customer %d\n", len(addresses), customer.ID)
	return presenter.SuccessResponse(c, "Addresses retrieved successfully", addresses)
}

// Fix UpdateAddress - handle PostalCode field issue
func (h *Handler) UpdateAddress(c *fiber.Ctx) error {
	userID, err := h.getCurrentUserID(c)
	if err != nil {
		return presenter.BadRequest(c, "Unauthorized")
	}

	addressID, err := strconv.ParseUint(c.Params("addressId"), 10, 32)
	if err != nil {
		return presenter.BadRequest(c, "Invalid address ID")
	}

	var req UpdateAddressRequest
	if err := c.BodyParser(&req); err != nil {
		return presenter.BadRequest(c, "Invalid request body")
	}

	if err := validation.ValidateStruct(&req); err != nil {
		return presenter.BadRequest(c, err.Error())
	}

	// Get customer by user ID first
	customer, err := h.service.GetCustomerByUserID(userID)
	if err != nil {
		return presenter.NotFound(c, "Customer not found")
	}

	// Convert UpdateAddressRequest to CreateAddressRequest for service call
	createReq := &CreateAddressRequest{
		Street:     getStringValue(req.Street),
		City:       getStringValue(req.City),
		State:      getStringValue(req.State),
		Country:    getStringValue(req.Country),
		PostalCode: getStringValue(req.PostalCode),
		IsDefault:  getBoolValue(req.IsDefault),
	}

	address, err := h.service.UpdateAddress(customer.ID, uint(addressID), createReq)
	if err != nil {
		return presenter.InternalServerError(c, "Failed to update address")
	}

	return presenter.Success(c, "Address updated successfully", address)
}

// Fix ListCustomers - handle service signature
func (h *Handler) ListCustomers(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	// Fix: Match service signature (returns 2 values, not 3)
	customers, err := h.service.ListCustomers(page, limit)
	if err != nil {
		return presenter.InternalServerError(c, "Failed to list customers")
	}

	return presenter.OK(c, customers, nil)
}

// Fix GetCustomerByID
func (h *Handler) GetCustomerByID(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return presenter.BadRequest(c, "Invalid customer ID")
	}

	customer, err := h.service.GetCustomerByID(uint(id))
	if err != nil {
		return presenter.NotFound(c, "Customer not found")
	}

	response := CustomerResponse{
		ID:        customer.ID,
		FirstName: customer.FirstName,
		LastName:  customer.LastName,
		Phone:     customer.Phone,
	}

	// Fix: data first, meta second
	return presenter.OK(c, response, nil)
}

// Fix UpdateCustomerProfile
func (h *Handler) UpdateCustomerProfile(c *fiber.Ctx) error {
	userID, err := h.getCurrentUserID(c)
	if err != nil {
		return presenter.BadRequest(c, "Unauthorized")
	}

	var req UpdateCustomerRequest
	if err := c.BodyParser(&req); err != nil {
		return presenter.BadRequest(c, "Invalid request body")
	}

	if err := validation.ValidateStruct(&req); err != nil {
		return presenter.BadRequest(c, err.Error())
	}

	// Get existing customer
	customer, err := h.service.GetCustomerByUserID(userID)
	if err != nil {
		return presenter.NotFound(c, "Customer not found")
	}

	// Actually update the customer profile
	updatedCustomer, err := h.service.UpdateCustomer(customer.ID, &req)
	if err != nil {
		return presenter.InternalServerError(c, "Failed to update customer profile")
	}

	fmt.Printf("[DEBUG] Customer profile updated for user %d: %+v\n", userID, updatedCustomer)
	return presenter.OK(c, updatedCustomer, nil)
}

// Fix DeleteAddress
func (h *Handler) DeleteAddress(c *fiber.Ctx) error {
	userID, err := h.getCurrentUserID(c)
	if err != nil {
		return presenter.BadRequest(c, "Unauthorized")
	}

	addressID, err := strconv.ParseUint(c.Params("addressId"), 10, 32)
	if err != nil {
		return presenter.BadRequest(c, "Invalid address ID")
	}

	// Get customer by user ID first
	customer, err := h.service.GetCustomerByUserID(userID)
	if err != nil {
		return presenter.NotFound(c, "Customer not found")
	}

	err = h.service.DeleteAddress(customer.ID, uint(addressID))
	if err != nil {
		return presenter.InternalServerError(c, "Failed to delete address")
	}

	return presenter.Success(c, "Address deleted successfully", nil)
}

func (h *Handler) SetDefaultAddress(c *fiber.Ctx) error {
	addressID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return presenter.BadRequest(c, "Invalid address ID")
	}

	userID, err := h.getCurrentUserID(c)
	if err != nil {
		return presenter.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
	}

	// Get customer by user ID first
	customer, err := h.service.GetCustomerByUserID(userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return presenter.ErrorResponse(c, http.StatusNotFound, "Customer profile not found")
		}
		return presenter.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve customer profile")
	}

	if err := h.service.SetDefaultAddress(customer.ID, uint(addressID)); err != nil {
		return presenter.InternalServerError(c, "Failed to set default address")
	}

	fmt.Printf("[DEBUG] Set address %d as default for customer %d\n", addressID, customer.ID)
	return presenter.SuccessResponse(c, "Default address set successfully", nil)
}

// Admin methods
func (h *Handler) CreateCustomerAddress(c *fiber.Ctx) error {
	customerID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return presenter.ErrorResponse(c, http.StatusBadRequest, "Invalid customer ID")
	}

	var req CreateAddressRequest
	if err := c.BodyParser(&req); err != nil {
		return presenter.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}

	if err := validation.ValidateStruct(&req); err != nil {
		return presenter.ErrorResponse(c, http.StatusBadRequest, "Validation failed")
	}

	address, err := h.service.CreateAddress(uint(customerID), &req)
	if err != nil {
		return presenter.ErrorResponse(c, http.StatusInternalServerError, "Failed to create address")
	}

	return presenter.SuccessResponse(c, "Address created successfully", address)
}

func (h *Handler) GetCustomerAddressesAdmin(c *fiber.Ctx) error {
	customerID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return presenter.ErrorResponse(c, http.StatusBadRequest, "Invalid customer ID")
	}

	addresses, err := h.service.GetCustomerAddresses(uint(customerID))
	if err != nil {
		return presenter.ErrorResponse(c, http.StatusInternalServerError, "Failed to get addresses")
	}

	return presenter.SuccessResponse(c, "Addresses retrieved successfully", addresses)
}

// REMOVE ALL DUPLICATE METHODS - keep only the above definitions
