package payments

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// PaymentStatus represents the status of a payment in orders
type OrderPaymentStatus string

const (
	OrderPaymentStatusPaid OrderPaymentStatus = "paid"
)

// OrderServiceInterface defines the interface for orders service operations
type OrderServiceInterface interface {
	AdminUpdatePaymentStatus(ctx context.Context, id uuid.UUID, paymentStatus interface{}) error
}

type Service interface {
	// Payment operations
	InitializePayment(req CreatePaymentRequest, customerID uint) (*PaymentInitResponse, error)
	ProcessPayment(req ProcessPaymentRequest) error
	GetPayment(id string) (*PaymentResponse, error)
	GetPaymentByTransactionRef(ref string) (*PaymentResponse, error)
	GetCustomerPayments(customerID uint) ([]PaymentResponse, error)
	GetOrderPayments(orderID string) ([]PaymentResponse, error)

	// Paystack operations
	InitializePaystackPayment(email string, amount int64, metadata map[string]interface{}) (*PaystackInitializeResponse, error)
	VerifyPaystackPayment(reference string) (*PaystackVerifyResponse, error)
	ProcessPaystackWebhook(signature string, payload []byte) error

	// Refund operations
	InitiateRefund(req RefundPaymentRequest) (*RefundResponse, error)
	ProcessRefund(refundID string, status PaymentStatus, providerRef string) error
	GetRefund(id string) (*RefundResponse, error)
	GetPaymentRefunds(paymentID string) ([]RefundResponse, error)

	// Webhook operations
	ProcessWebhook(req WebhookEventRequest) error

	// Analytics
	GetPaymentStats(customerID *uint) (map[string]interface{}, error)
}

type service struct {
	repo           Repository
	paystackClient *PaystackClient
	orderService   OrderServiceInterface
}

func NewService(repo Repository, paystackClient *PaystackClient, orderService OrderServiceInterface) Service {
	return &service{
		repo:           repo,
		paystackClient: paystackClient,
		orderService:   orderService,
	}
}

// Payment operations
func (s *service) InitializePayment(req CreatePaymentRequest, customerID uint) (*PaymentInitResponse, error) {
	// Generate unique transaction reference
	transactionRef, err := s.generateTransactionRef()
	if err != nil {
		return nil, fmt.Errorf("failed to generate transaction reference: %w", err)
	}

	// TODO: Get order details and amount from orders service
	// For now, using placeholder amount
	amountKobo := int64(100000) // ₦1000.00

	payment := &Payment{
		OrderID:        req.OrderID,
		CustomerID:     customerID,
		AmountKobo:     amountKobo,
		Currency:       "NGN",
		PaymentMethod:  req.PaymentMethod,
		Status:         PaymentStatusPending,
		TransactionRef: transactionRef,
	}

	if err := s.repo.CreatePayment(payment); err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	// Generate payment URL based on payment method
	paymentURL := s.generatePaymentURL(payment, req.ReturnURL, req.CancelURL)

	return &PaymentInitResponse{
		PaymentID:      payment.ID,
		TransactionRef: transactionRef,
		PaymentURL:     paymentURL,
		Status:         string(PaymentStatusPending),
		Message:        "Payment initialized successfully",
	}, nil
}

func (s *service) ProcessPayment(req ProcessPaymentRequest) error {
	payment, err := s.repo.GetPaymentByTransactionRef(req.TransactionRef)
	if err != nil {
		return fmt.Errorf("payment not found: %w", err)
	}

	if payment.Status != PaymentStatusPending {
		return errors.New("payment is not in pending status")
	}

	status := PaymentStatus(req.Status)
	if err := s.repo.UpdatePaymentStatus(payment.ID, status, req.ProviderRef, ""); err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	// Update order payment status when payment is successful
	if status == PaymentStatusCompleted {
		// Parse order ID from string to UUID
		orderID, err := uuid.Parse(payment.OrderID)
		if err != nil {
			fmt.Printf("Warning: Failed to parse order ID %s: %v\n", payment.OrderID, err)
		} else {
			if err := s.orderService.AdminUpdatePaymentStatus(context.Background(), orderID, OrderPaymentStatusPaid); err != nil {
				// Log error but don't fail the payment processing
				fmt.Printf("Warning: Failed to update order payment status: %v\n", err)
			}
		}
	}

	// TODO: Send notification to customer

	return nil
}

func (s *service) GetPayment(id string) (*PaymentResponse, error) {
	payment, err := s.repo.GetPaymentByID(id)
	if err != nil {
		return nil, err
	}
	return s.toPaymentResponse(payment), nil
}

func (s *service) GetPaymentByTransactionRef(ref string) (*PaymentResponse, error) {
	payment, err := s.repo.GetPaymentByTransactionRef(ref)
	if err != nil {
		return nil, err
	}
	return s.toPaymentResponse(payment), nil
}

func (s *service) GetCustomerPayments(customerID uint) ([]PaymentResponse, error) {
	payments, err := s.repo.GetPaymentsByCustomerID(customerID)
	if err != nil {
		return nil, err
	}

	responses := make([]PaymentResponse, len(payments))
	for i, payment := range payments {
		responses[i] = *s.toPaymentResponse(&payment)
	}
	return responses, nil
}

func (s *service) GetOrderPayments(orderID string) ([]PaymentResponse, error) {
	payments, err := s.repo.GetPaymentsByOrderID(orderID)
	if err != nil {
		return nil, err
	}

	responses := make([]PaymentResponse, len(payments))
	for i, payment := range payments {
		responses[i] = *s.toPaymentResponse(&payment)
	}
	return responses, nil
}

// Refund operations
func (s *service) InitiateRefund(req RefundPaymentRequest) (*RefundResponse, error) {
	payment, err := s.repo.GetPaymentByID(req.PaymentID)
	if err != nil {
		return nil, fmt.Errorf("payment not found: %w", err)
	}

	if payment.Status != PaymentStatusCompleted {
		return nil, errors.New("can only refund completed payments")
	}

	if req.AmountKobo > payment.AmountKobo {
		return nil, errors.New("refund amount cannot exceed payment amount")
	}

	// Generate unique refund reference
	refundRef, err := s.generateRefundRef()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refund reference: %w", err)
	}

	refund := &PaymentRefund{
		PaymentID:  req.PaymentID,
		AmountKobo: req.AmountKobo,
		Reason:     req.Reason,
		Status:     PaymentStatusPending,
		RefundRef:  refundRef,
	}

	if err := s.repo.CreateRefund(refund); err != nil {
		return nil, fmt.Errorf("failed to create refund: %w", err)
	}

	return s.toRefundResponse(refund), nil
}

func (s *service) ProcessRefund(refundID string, status PaymentStatus, providerRef string) error {
	refund, err := s.repo.GetRefundByID(refundID)
	if err != nil {
		return fmt.Errorf("refund not found: %w", err)
	}

	if refund.Status != PaymentStatusPending {
		return errors.New("refund is not in pending status")
	}

	refund.Status = status
	refund.ProviderRef = providerRef
	if status == PaymentStatusCompleted || status == PaymentStatusFailed {
		now := time.Now()
		refund.ProcessedAt = &now
	}

	if err := s.repo.UpdateRefund(refund); err != nil {
		return fmt.Errorf("failed to update refund: %w", err)
	}

	return nil
}

func (s *service) GetRefund(id string) (*RefundResponse, error) {
	refund, err := s.repo.GetRefundByID(id)
	if err != nil {
		return nil, err
	}
	return s.toRefundResponse(refund), nil
}

func (s *service) GetPaymentRefunds(paymentID string) ([]RefundResponse, error) {
	refunds, err := s.repo.GetRefundsByPaymentID(paymentID)
	if err != nil {
		return nil, err
	}

	responses := make([]RefundResponse, len(refunds))
	for i, refund := range refunds {
		responses[i] = *s.toRefundResponse(&refund)
	}
	return responses, nil
}

// Webhook operations
func (s *service) ProcessWebhook(req WebhookEventRequest) error {
	webhook := &PaymentWebhook{
		Provider:  req.Provider,
		EventType: req.EventType,
		EventData: req.EventData,
		Signature: req.Signature,
		Processed: false,
	}

	if err := s.repo.CreateWebhook(webhook); err != nil {
		return fmt.Errorf("failed to create webhook: %w", err)
	}

	// TODO: Process webhook based on provider and event type
	// TODO: Update payment status accordingly

	return nil
}

// Analytics
func (s *service) GetPaymentStats(customerID *uint) (map[string]interface{}, error) {
	return s.repo.GetPaymentStats(customerID)
}

// Helper methods
func (s *service) generateTransactionRef() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "TXN_" + hex.EncodeToString(bytes), nil
}

func (s *service) generateRefundRef() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "REF_" + hex.EncodeToString(bytes), nil
}

func (s *service) generatePaymentURL(payment *Payment, returnURL, cancelURL string) string {
	// TODO: Implement actual payment gateway integration
	// This is a placeholder implementation
	switch payment.PaymentMethod {
	case PaymentMethodCard:
		return fmt.Sprintf("https://payment.example.com/pay/%s", payment.TransactionRef)
	default:
		return fmt.Sprintf("https://payment.example.com/pay/%s", payment.TransactionRef)
	}
}

func (s *service) toPaymentResponse(payment *Payment) *PaymentResponse {
    return &PaymentResponse{
        ID:             payment.ID,
        OrderID:        payment.OrderID,
        CustomerID:     payment.CustomerID,
        AmountKobo:     payment.AmountKobo,
        AmountNaira:    float64(payment.AmountKobo) / 100,
        Currency:       payment.Currency,
        PaymentMethod:  payment.PaymentMethod,
        Status:         payment.Status,
        TransactionRef: payment.TransactionRef,
        ProviderRef:    payment.ProviderRef,
        FailureReason:  payment.FailureReason,
        ProcessedAt:    payment.ProcessedAt,
        CreatedAt:      payment.CreatedAt,
        UpdatedAt:      payment.UpdatedAt,
    }
}

func (s *service) toRefundResponse(refund *PaymentRefund) *RefundResponse {
	return &RefundResponse{
		ID:          refund.ID,
		PaymentID:   refund.PaymentID,
		AmountKobo:  refund.AmountKobo,
		AmountNaira: float64(refund.AmountKobo) / 100,
		Reason:      refund.Reason,
		Status:      refund.Status,
		RefundRef:   refund.RefundRef,
		ProviderRef: refund.ProviderRef,
		ProcessedAt: refund.ProcessedAt,
		CreatedAt:   refund.CreatedAt,
	}
}

// Paystack operations
func (s *service) InitializePaystackPayment(email string, amount int64, metadata map[string]interface{}) (*PaystackInitializeResponse, error) {
	// Generate unique reference
	reference, err := s.generateTransactionRef()
	if err != nil {
		return nil, fmt.Errorf("failed to generate reference: %w", err)
	}

	// Convert metadata to JSON string
	metadataJSON := "{}"
	if metadata != nil && len(metadata) > 0 {
		if jsonBytes, err := json.Marshal(metadata); err == nil {
			metadataJSON = string(jsonBytes)
		}
	}

	// Create order record
	order := &Order{
		Reference:     reference,
		CustomerEmail: email,
		AmountKobo:    amount,
		Currency:      "NGN",
		Status:        PaymentStatusPending,
		ItemsSubtotal: amount, // Set items subtotal to match amount
		TotalAmount:   amount, // Set total amount to match amount
		Metadata:      metadataJSON,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.repo.CreateOrder(order); err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Initialize payment with Paystack
	return s.paystackClient.InitializeTransaction(email, amount, reference, metadata)
}

func (s *service) VerifyPaystackPayment(reference string) (*PaystackVerifyResponse, error) {
	// Get order by reference
	order, err := s.repo.GetOrderByReference(reference)
	if err != nil {
		return nil, fmt.Errorf("order not found: %w", err)
	}

	// Verify payment with Paystack
	verifyResp, err := s.paystackClient.VerifyTransaction(reference)
	if err != nil {
		return nil, fmt.Errorf("failed to verify payment: %w", err)
	}

	// Update order status based on verification result
	if verifyResp.Status && verifyResp.Data.Status == "success" {
		// Verify amount matches
		if verifyResp.Data.Amount == order.AmountKobo {
			if err := s.repo.UpdateOrderStatus(reference, OrderStatusConfirmed); err != nil {
				return nil, fmt.Errorf("failed to update order status: %w", err)
			}
		} else {
			return nil, errors.New("amount mismatch")
		}
	} else {
		if err := s.repo.UpdateOrderStatus(reference, OrderStatusCancelled); err != nil {
			return nil, fmt.Errorf("failed to update order status: %w", err)
		}
	}

	return verifyResp, nil
}

func (s *service) ProcessPaystackWebhook(signature string, payload []byte) error {
	// Validate webhook signature
	if !s.paystackClient.ValidateWebhookSignature(payload, signature) {
		return errors.New("invalid webhook signature")
	}

	// Parse webhook event
	event, err := s.paystackClient.ParseWebhookEvent(payload)
	if err != nil {
		return fmt.Errorf("failed to parse webhook event: %w", err)
	}

	// Process charge.success event
	if event.Event == "charge.success" {
		reference := event.Data.Reference
		amount := event.Data.Amount

		// Get order by reference
		order, err := s.repo.GetOrderByReference(reference)
		if err != nil {
			return fmt.Errorf("order not found: %w", err)
		}

		// Verify amount matches and order is still pending
		if order.AmountKobo == amount && order.Status == PaymentStatusPending {
			// Update order status in payments domain
			if err := s.repo.UpdateOrderStatus(reference, OrderStatusConfirmed); err != nil {
				return fmt.Errorf("failed to update order status in payments: %w", err)
			}

			// Also update payment status in orders domain
			if s.orderService != nil {
				ctx := context.Background()
				orderUUID, err := uuid.Parse(reference)
				if err == nil {
					if err := s.orderService.AdminUpdatePaymentStatus(ctx, orderUUID, OrderPaymentStatusPaid); err != nil {
						// Log error but don't fail the webhook processing
						return fmt.Errorf("failed to update payment status in orders domain: %w", err)
					}
				}
			}
		}
	}

	return nil
}
