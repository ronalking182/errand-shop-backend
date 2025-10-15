package payments

import (
	"errors"

	"gorm.io/gorm"
)

type Repository interface {
	// Payment operations
	CreatePayment(payment *Payment) error
	GetPaymentByID(id string) (*Payment, error)
	GetPaymentByTransactionRef(ref string) (*Payment, error)
	GetPaymentsByOrderID(orderID string) ([]Payment, error)
	GetPaymentsByCustomerID(customerID uint) ([]Payment, error)
	UpdatePayment(payment *Payment) error
	UpdatePaymentStatus(id string, status PaymentStatus, providerRef, providerResponse string) error

	// Order operations
	CreateOrder(order *Order) error
	GetOrderByID(id string) (*Order, error)
	GetOrderByReference(reference string) (*Order, error)
	UpdateOrderStatus(reference string, status OrderStatus) error

	// Refund operations
	CreateRefund(refund *PaymentRefund) error
	GetRefundByID(id string) (*PaymentRefund, error)
	GetRefundsByPaymentID(paymentID string) ([]PaymentRefund, error)
	UpdateRefund(refund *PaymentRefund) error

	// Webhook operations
	CreateWebhook(webhook *PaymentWebhook) error
	GetUnprocessedWebhooks() ([]PaymentWebhook, error)
	MarkWebhookProcessed(id string) error

	// Analytics
	GetPaymentStats(customerID *uint) (map[string]interface{}, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// Payment operations
func (r *repository) CreatePayment(payment *Payment) error {
	return r.db.Create(payment).Error
}

func (r *repository) GetPaymentByID(id string) (*Payment, error) {
	var payment Payment
	err := r.db.First(&payment, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("payment not found")
		}
		return nil, err
	}
	return &payment, nil
}

func (r *repository) GetPaymentByTransactionRef(ref string) (*Payment, error) {
	var payment Payment
	err := r.db.Where("transaction_ref = ?", ref).First(&payment).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("payment not found")
		}
		return nil, err
	}
	return &payment, nil
}

func (r *repository) GetPaymentsByOrderID(orderID string) ([]Payment, error) {
	var payments []Payment
	err := r.db.Where("order_id = ?", orderID).Order("created_at DESC").Find(&payments).Error
	return payments, err
}

func (r *repository) GetPaymentsByCustomerID(customerID uint) ([]Payment, error) {
	var payments []Payment
	err := r.db.Where("customer_id = ?", customerID).Order("created_at DESC").Find(&payments).Error
	return payments, err
}

func (r *repository) UpdatePayment(payment *Payment) error {
	return r.db.Save(payment).Error
}

func (r *repository) UpdatePaymentStatus(id string, status PaymentStatus, providerRef, providerResponse string) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if providerRef != "" {
		updates["provider_ref"] = providerRef
	}
	if providerResponse != "" {
		updates["provider_response"] = providerResponse
	}
	if status == PaymentStatusCompleted || status == PaymentStatusFailed {
		updates["processed_at"] = "NOW()"
	}

	return r.db.Model(&Payment{}).Where("id = ?", id).Updates(updates).Error
}

// Refund operations
func (r *repository) CreateRefund(refund *PaymentRefund) error {
	return r.db.Create(refund).Error
}

func (r *repository) GetRefundByID(id string) (*PaymentRefund, error) {
	var refund PaymentRefund
	err := r.db.Preload("Payment").First(&refund, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("refund not found")
		}
		return nil, err
	}
	return &refund, nil
}

func (r *repository) GetRefundsByPaymentID(paymentID string) ([]PaymentRefund, error) {
	var refunds []PaymentRefund
	err := r.db.Where("payment_id = ?", paymentID).Order("created_at DESC").Find(&refunds).Error
	return refunds, err
}

func (r *repository) UpdateRefund(refund *PaymentRefund) error {
	return r.db.Save(refund).Error
}

// Webhook operations
func (r *repository) CreateWebhook(webhook *PaymentWebhook) error {
	return r.db.Create(webhook).Error
}

func (r *repository) GetUnprocessedWebhooks() ([]PaymentWebhook, error) {
	var webhooks []PaymentWebhook
	err := r.db.Where("processed = ?", false).Order("created_at ASC").Find(&webhooks).Error
	return webhooks, err
}

func (r *repository) MarkWebhookProcessed(id string) error {
	return r.db.Model(&PaymentWebhook{}).Where("id = ?", id).Updates(map[string]interface{}{
		"processed":    true,
		"processed_at": "NOW()",
	}).Error
}

// Analytics
func (r *repository) GetPaymentStats(customerID *uint) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	query := r.db.Model(&Payment{})
	if customerID != nil {
		query = query.Where("customer_id = ?", *customerID)
	}

	// Total payments
	var totalCount int64
	var totalAmount int64
	query.Count(&totalCount)
	query.Select("COALESCE(SUM(amount_kobo), 0)").Scan(&totalAmount)

	// Successful payments
	var successfulCount int64
	var successfulAmount int64
	successQuery := query.Where("status = ?", PaymentStatusCompleted)
	successQuery.Count(&successfulCount)
	successQuery.Select("COALESCE(SUM(amount_kobo), 0)").Scan(&successfulAmount)

	stats["total_payments"] = totalCount
	stats["total_amount_kobo"] = totalAmount
	stats["total_amount_naira"] = float64(totalAmount)
	stats["successful_payments"] = successfulCount
	stats["successful_amount_kobo"] = successfulAmount
	stats["successful_amount_naira"] = float64(successfulAmount)

	if totalCount > 0 {
		stats["success_rate"] = float64(successfulCount) / float64(totalCount) * 100
	} else {
		stats["success_rate"] = 0.0
	}

	return stats, nil
}

// Order operations
func (r *repository) CreateOrder(order *Order) error {
	return r.db.Create(order).Error
}

func (r *repository) GetOrderByID(id string) (*Order, error) {
	var order Order
	err := r.db.First(&order, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("order not found")
		}
		return nil, err
	}
	return &order, nil
}

func (r *repository) GetOrderByReference(reference string) (*Order, error) {
	var order Order
	err := r.db.First(&order, "reference = ?", reference).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("order not found")
		}
		return nil, err
	}
	return &order, nil
}

func (r *repository) UpdateOrderStatus(reference string, status OrderStatus) error {
	return r.db.Model(&Order{}).Where("reference = ?", reference).Update("status", status).Error
}
