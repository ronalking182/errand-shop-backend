package payments

import (
	"errandShop/internal/presenter"
	"errandShop/internal/validation"
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// Helper function to convert UUID to uint for database compatibility
func (h *Handler) uuidToUint(userUUID uuid.UUID) uint {
	// Convert UUID to string and then hash to uint
	uuidStr := userUUID.String()
	hash := uint(0)
	for _, char := range uuidStr {
		hash = hash*31 + uint(char)
	}
	return hash % 1000000 // Limit to 6 digits
}

// Customer endpoints

// InitializePayment godoc
// @Summary Initialize a new payment
// @Description Create a new payment for an order
// @Tags payments
// @Accept json
// @Produce json
// @Param request body CreatePaymentRequest true "Payment initialization request"
// @Success 201 {object} presenter.Response{data=PaymentInitResponse}
// @Failure 400 {object} presenter.Response
// @Failure 401 {object} presenter.Response
// @Failure 500 {object} presenter.Response
// @Router /payments/initialize [post]
// @Security BearerAuth
func (h *Handler) InitializePayment(c *fiber.Ctx) error {
	// Extract customer ID from context or params
	userIDRaw := c.Locals("userID")
	if userIDRaw == nil {
		return presenter.Err(c, fiber.StatusUnauthorized, "User not authenticated")
	}

	// Handle different types from JWT claims
	var userIDUUID uuid.UUID
	switch v := userIDRaw.(type) {
	case uuid.UUID:
		userIDUUID = v
	case string:
		parsedUUID, err := uuid.Parse(v)
		if err != nil {
			return presenter.Err(c, fiber.StatusBadRequest, "Invalid user ID format")
		}
		userIDUUID = parsedUUID
	default:
		return presenter.Err(c, fiber.StatusInternalServerError, "Invalid user ID type")
	}

	customerID := h.uuidToUint(userIDUUID)

	// Parse CreatePaymentRequest format
	var req CreatePaymentRequest
	if err := c.BodyParser(&req); err != nil {
		return presenter.BadRequest(c, "Invalid request body")
	}

	if err := req.Validate(); err != nil {
		return presenter.BadRequest(c, err.Error())
	}

	resp, err := h.service.InitializePayment(req, customerID)
	if err != nil {
		return presenter.InternalServerError(c, err.Error())
	}
	return presenter.Created(c, resp)
}

// ProcessPayment godoc
// @Summary Process a payment
// @Description Update payment status after processing
// @Tags payments
// @Accept json
// @Produce json
// @Param request body ProcessPaymentRequest true "Payment processing request"
// @Success 200 {object} presenter.Response
// @Failure 400 {object} presenter.Response
// @Failure 404 {object} presenter.Response
// @Failure 500 {object} presenter.Response
// @Router /payments/process [post]
func (h *Handler) getCurrentUserID(c *fiber.Ctx) (uint, error) {
	userIDRaw := c.Locals("userID")
	if userIDRaw == nil {
		return 0, errors.New("user not authenticated")
	}

	// Handle different types from JWT claims
	switch v := userIDRaw.(type) {
	case uuid.UUID:
		return h.uuidToUint(v), nil
	case string:
		parsedUUID, err := uuid.Parse(v)
		if err != nil {
			return 0, errors.New("invalid user ID format")
		}
		return h.uuidToUint(parsedUUID), nil
	default:
		return 0, errors.New("invalid user ID type")
	}
}

func (h *Handler) ProcessPayment(c *fiber.Ctx) error {
	var req ProcessPaymentRequest
	if err := c.BodyParser(&req); err != nil {
		return presenter.BadRequest(c, "Invalid request body")
	}

	if err := validation.ValidateStruct(&req); err != nil {
		return presenter.BadRequest(c, "Validation failed")
	}

	payment := h.service.ProcessPayment(req)
	if payment == nil {
		return presenter.InternalServerError(c, "Failed to process payment")
	}

	return presenter.Success(c, "Payment processed successfully", payment)
}

func (h *Handler) GetPayment(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return presenter.BadRequest(c, "Invalid payment ID")
	}

	payment, err := h.service.GetPayment(id)
	if err != nil {
		if err.Error() == "payment not found" {
			return presenter.NotFound(c, "Payment not found")
		}
		return presenter.InternalServerError(c, "Failed to get payment")
	}

	return presenter.Success(c, "Payment retrieved successfully", payment)
}

func (h *Handler) GetPaymentByTransactionRef(c *fiber.Ctx) error {
	transactionRef := c.Params("ref")
	if transactionRef == "" {
		return presenter.BadRequest(c, "Transaction reference is required")
	}

	payment, err := h.service.GetPaymentByTransactionRef(transactionRef)
	if err != nil {
		if err.Error() == "payment not found" {
			return presenter.NotFound(c, "Payment not found")
		}
		return presenter.InternalServerError(c, "Failed to get payment")
	}

	return presenter.Success(c, "Payment retrieved successfully", payment)
}

func (h *Handler) InitiateRefund(c *fiber.Ctx) error {
	var req RefundPaymentRequest
	if err := c.BodyParser(&req); err != nil {
		return presenter.BadRequest(c, "Invalid request body")
	}

	if err := validation.ValidateStruct(&req); err != nil {
		return presenter.BadRequest(c, "Validation failed")
	}

	// Fix: Pass value instead of pointer if service expects value
	refund, err := h.service.InitiateRefund(req)
	if err != nil {
		return presenter.InternalServerError(c, "Failed to initiate refund")
	}

	return presenter.Success(c, "Refund initiated successfully", refund)
}

// Fix other presenter calls to use correct signatures

// 	if err := validation.ValidateStruct(&req); err != nil {
// 		return presenter.BadRequest(c, "Validation failed") // Remove extra parameters
// 	}
//
// 	customerID, err := h.getCurrentUserID(c)
// 	if err != nil {
// 		return presenter.BadRequest(c, "User not authenticated")
// 	}
//
// 	payment, err := h.service.ProcessPayment(req, customerID)
// 	if err != nil {
// 		fmt.Printf("[DEBUG] Process payment error: %v\n", err)
// 		return presenter.InternalServerError(c, "Failed to process payment")
// 	}
//
// 	fmt.Printf("[DEBUG] Payment processed: %+v\n", payment)
// 	return presenter.Success(c, "Payment processed successfully", payment)
// }

// GetRefund godoc
// @Summary Get refund by ID
// @Description Retrieve refund details by ID
// @Tags payments
// @Produce json
// @Param id path int true "Refund ID"
// @Success 200 {object} presenter.Response{data=RefundResponse}
// @Failure 400 {object} presenter.Response
// @Failure 401 {object} presenter.Response
// @Failure 404 {object} presenter.Response
// @Failure 500 {object} presenter.Response
// @Router /payments/refund/{id} [get]
// @Security BearerAuth
func (h *Handler) GetRefund(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return presenter.BadRequest(c, "Invalid refund ID")
	}

	refund, err := h.service.GetRefund(id)
	if err != nil {
		if err.Error() == "refund not found" {
			return presenter.NotFound(c, "Refund not found")
		}
		return presenter.InternalServerError(c, "Failed to get refund")
	}

	return presenter.Success(c, "Refund retrieved successfully", refund)
}

// GetPaymentRefunds godoc
// @Summary Get payment refunds
// @Description Retrieve all refunds for a specific payment
// @Tags payments
// @Produce json
// @Param payment_id path int true "Payment ID"
// @Success 200 {object} presenter.Response{data=[]RefundResponse}
// @Failure 400 {object} presenter.Response
// @Failure 401 {object} presenter.Response
// @Failure 500 {object} presenter.Response
// @Router /payments/{payment_id}/refunds [get]
// @Security BearerAuth
func (h *Handler) GetPaymentRefunds(c *fiber.Ctx) error {
	paymentID := c.Params("payment_id")
	if paymentID == "" {
		return presenter.BadRequest(c, "Invalid payment ID")
	}

	refunds, err := h.service.GetPaymentRefunds(paymentID)
	if err != nil {
		return presenter.InternalServerError(c, "Failed to get payment refunds")
	}

	return presenter.Success(c, "Payment refunds retrieved successfully", refunds)
}

// Webhook endpoint

// ProcessWebhook godoc
// @Summary Process payment webhook
// @Description Handle webhook events from payment providers
// @Tags payments
// @Accept json
// @Produce json
// @Param request body WebhookEventRequest true "Webhook event"
// @Success 200 {object} presenter.Response
// @Failure 400 {object} presenter.Response
// @Failure 500 {object} presenter.Response
// @Router /payments/webhook [post]
func (h *Handler) ProcessWebhook(c *fiber.Ctx) error {
	var req WebhookEventRequest
	if err := c.BodyParser(&req); err != nil {
		return presenter.BadRequest(c, "Invalid request body")
	}

	if err := validation.ValidateStruct(&req); err != nil {
		return presenter.BadRequest(c, "Validation failed")
	}

	if err := h.service.ProcessWebhook(req); err != nil {
		return presenter.InternalServerError(c, "Failed to process webhook")
	}

	return presenter.Success(c, "Webhook processed successfully", nil)
}

// Admin endpoints

// GetAllPayments godoc
// @Summary Get all payments (Admin)
// @Description Retrieve all payments with pagination
// @Tags admin,payments
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param status query string false "Filter by status"
// @Success 200 {object} presenter.Response{data=[]PaymentResponse}
// @Failure 401 {object} presenter.Response
// @Failure 403 {object} presenter.Response
// @Failure 500 {object} presenter.Response
// @Router /admin/payments [get]
// @Security BearerAuth
func (h *Handler) GetAllPayments(c *fiber.Ctx) error {
	// TODO: Implement pagination and filtering
	// This is a placeholder implementation
	return presenter.Success(c, "All payments retrieved successfully", []PaymentResponse{})
}

// ProcessRefund godoc
// @Summary Process a refund (Admin)
// @Description Update refund status after processing
// @Tags admin,payments
// @Accept json
// @Produce json
// @Param id path int true "Refund ID"
// @Param status query string true "New status" Enums(completed,failed)
// @Param provider_ref query string false "Provider reference"
// @Success 200 {object} presenter.Response
// @Failure 400 {object} presenter.Response
// @Failure 401 {object} presenter.Response
// @Failure 403 {object} presenter.Response
// @Failure 404 {object} presenter.Response
// @Failure 500 {object} presenter.Response
// @Router /admin/payments/refund/{id}/process [post]
// @Security BearerAuth
func (h *Handler) ProcessRefund(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return presenter.BadRequest(c, "Invalid refund ID")
	}

	status := c.Query("status")
	if status == "" {
		return presenter.BadRequest(c, "Status is required")
	}

	providerRef := c.Query("provider_ref")

	if err := h.service.ProcessRefund(id, PaymentStatus(status), providerRef); err != nil {
		if err.Error() == "refund not found" {
			return presenter.NotFound(c, "Refund not found")
		}
		return presenter.InternalServerError(c, "Failed to process refund")
	}

	return presenter.Success(c, "Refund processed successfully", nil)
}

// GetPaymentStats godoc
// @Summary Get payment statistics (Admin)
// @Description Retrieve payment analytics and statistics
// @Tags admin,payments
// @Produce json
// @Param customer_id query int false "Filter by customer ID"
// @Success 200 {object} presenter.Response{data=map[string]interface{}}
// @Failure 401 {object} presenter.Response
// @Failure 403 {object} presenter.Response
// @Failure 500 {object} presenter.Response
// @Router /admin/payments/stats [get]
// @Security BearerAuth
func (h *Handler) GetPaymentStats(c *fiber.Ctx) error {
	// Parse optional customer_id filter
	var customerID *uint
	if cidStr := c.Query("customer_id"); cidStr != "" {
		if cid, err := strconv.ParseUint(cidStr, 10, 32); err == nil {
			cidUint := uint(cid)
			customerID = &cidUint
		}
	}

	stats, err := h.service.GetPaymentStats(customerID)
	if err != nil {
		return presenter.InternalServerError(c, "Failed to get payment stats")
	}

	return presenter.Success(c, "Payment statistics retrieved successfully", stats)
}

// Paystack endpoints

// InitializePaystackPayment godoc
// @Summary Initialize a Paystack payment
// @Description Initialize a payment transaction with Paystack
// @Tags paystack
// @Accept json
// @Produce json
// @Param request body PaystackInitializeRequest true "Paystack initialization request"
// @Success 200 {object} presenter.Response{data=PaystackInitializeResponse}
// @Failure 400 {object} presenter.Response
// @Failure 500 {object} presenter.Response
// @Router /payments/paystack/initialize [post]
func (h *Handler) InitializePaystackPayment(c *fiber.Ctx) error {
	var req PaystackInitializeRequest
	if err := c.BodyParser(&req); err != nil {
		return presenter.BadRequest(c, "Invalid request body")
	}

	if err := validation.ValidateStruct(&req); err != nil {
		return presenter.BadRequest(c, "Validation failed")
	}

	resp, err := h.service.InitializePaystackPayment(req.Email, req.Amount, req.Metadata)
	if err != nil {
		return presenter.InternalServerError(c, "Failed to initialize payment")
	}

	return presenter.Success(c, "Payment initialized successfully", resp)
}

// VerifyPaystackPayment godoc
// @Summary Verify a Paystack payment
// @Description Verify a payment transaction with Paystack
// @Tags paystack
// @Produce json
// @Param reference path string true "Payment reference"
// @Success 200 {object} presenter.Response{data=PaystackVerifyResponse}
// @Failure 400 {object} presenter.Response
// @Failure 404 {object} presenter.Response
// @Failure 500 {object} presenter.Response
// @Router /payments/paystack/verify/{reference} [get]
func (h *Handler) VerifyPaystackPayment(c *fiber.Ctx) error {
	reference := c.Params("reference")
	if reference == "" {
		return presenter.BadRequest(c, "Reference is required")
	}

	resp, err := h.service.VerifyPaystackPayment(reference)
	if err != nil {
		if err.Error() == "order not found" {
			return presenter.NotFound(c, "Payment not found")
		}
		return presenter.InternalServerError(c, "Failed to verify payment")
	}

	return presenter.Success(c, "Payment verified successfully", resp)
}

// PaystackWebhook godoc
// @Summary Handle Paystack webhook
// @Description Process webhook events from Paystack
// @Tags paystack
// @Accept json
// @Produce json
// @Success 200 {object} presenter.Response
// @Failure 400 {object} presenter.Response
// @Failure 500 {object} presenter.Response
// @Router /webhooks/paystack [post]
func (h *Handler) PaystackWebhook(c *fiber.Ctx) error {
	// Get signature from header
	signature := c.Get("X-Paystack-Signature")
	if signature == "" {
		return presenter.BadRequest(c, "Missing signature header")
	}

	// Get raw body
	body := c.Body()
	if len(body) == 0 {
		return presenter.BadRequest(c, "Empty request body")
	}

	// Process webhook
	if err := h.service.ProcessPaystackWebhook(signature, body); err != nil {
		if err.Error() == "invalid webhook signature" {
			return presenter.BadRequest(c, "Invalid signature")
		}
		return presenter.InternalServerError(c, "Failed to process webhook")
	}

	return presenter.Success(c, "Webhook processed successfully", nil)
}
