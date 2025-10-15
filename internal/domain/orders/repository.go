package orders

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Customer methods
func (r *Repository) List(ctx context.Context, userID uuid.UUID, query ListQuery) ([]Order, int64, error) {
	var orders []Order
	var total int64

	db := r.db.WithContext(ctx).Model(&Order{}).Where("customer_id = ?", userID)

	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.Limit
	err := db.Preload("Items.Product").Order("created_at DESC").Offset(offset).Limit(query.Limit).Find(&orders).Error
	return orders, total, err
}

func (r *Repository) Get(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*Order, error) {
	var order Order
	err := r.db.WithContext(ctx).Preload("Items.Product").Preload("StatusHistory").Where("id = ? AND customer_id = ?", id, userID).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *Repository) Create(ctx context.Context, order *Order) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(order).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *Repository) UpdateStatus(ctx context.Context, id uuid.UUID, userID uuid.UUID, status OrderStatus, reason string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Update order status
		if err := tx.Model(&Order{}).Where("id = ? AND customer_id = ?", id, userID).Updates(map[string]interface{}{
			"status": status,
		}).Error; err != nil {
			return err
		}

		// Add status history entry
		statusHistory := OrderStatusHistory{
			OrderID:   id,
			ToStatus:  status,
			Note:      reason,
		}
		return tx.Create(&statusHistory).Error
	})
}

// CancelOrder cancels an order
func (r *Repository) CancelOrder(ctx context.Context, id uuid.UUID, userID uuid.UUID, reason string) error {
	return r.UpdateStatus(ctx, id, userID, OrderStatusCancelled, reason)
}

// CheckIdempotency checks if an order with the given idempotency key exists
func (r *Repository) CheckIdempotency(ctx context.Context, userID uuid.UUID, idempotencyKey string) (*Order, error) {
	var order Order
	err := r.db.WithContext(ctx).Where("customer_id = ? AND idempotency_key = ?", userID, idempotencyKey).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// Admin methods
func (r *Repository) AdminList(ctx context.Context, query AdminListQuery) ([]Order, int64, error) {
	var orders []Order
	var total int64

	db := r.db.WithContext(ctx).Model(&Order{})

	// Apply filters
	if query.UserID != nil {
		db = db.Where("customer_id = ?", *query.UserID)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if query.PaymentStatus != "" {
		db = db.Where("payment_status = ?", query.PaymentStatus)
	}
	if query.DateFrom != nil {
		db = db.Where("created_at >= ?", *query.DateFrom)
	}
	if query.DateTo != nil {
		db = db.Where("created_at <= ?", *query.DateTo)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	sortBy := "created_at"
	if query.SortBy != "" {
		sortBy = query.SortBy
	}
	sortOrder := "desc"
	if query.SortOrder != "" {
		sortOrder = query.SortOrder
	}

	offset := (query.Page - 1) * query.Limit
	err := db.Preload("Items.Product").Order(sortBy + " " + sortOrder).Offset(offset).Limit(query.Limit).Find(&orders).Error
	return orders, total, err
}

func (r *Repository) AdminGet(ctx context.Context, id uuid.UUID) (*Order, error) {
	var order Order
	err := r.db.WithContext(ctx).Preload("Items.Product").Preload("StatusHistory").Where("id = ?", id).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *Repository) AdminUpdateStatus(ctx context.Context, id uuid.UUID, status OrderStatus) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Update order status
		if err := tx.Model(&Order{}).Where("id = ?", id).Updates(map[string]interface{}{
			"status": status,
		}).Error; err != nil {
			return err
		}

		//// Add status history entry
		statusHistory := &OrderStatusHistory{
			OrderID:   id,
			ToStatus:  status,
			Note:      "Updated by admin",
		}
		return tx.Create(statusHistory).Error
	})
}

func (r *Repository) AdminUpdatePaymentStatus(ctx context.Context, id uuid.UUID, paymentStatus PaymentStatus) error {
	return r.db.WithContext(ctx).Model(&Order{}).Where("id = ?", id).Updates(map[string]interface{}{
		"payment_status": paymentStatus,
	}).Error
}

func (r *Repository) AdminCancelOrder(ctx context.Context, id uuid.UUID, reason string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Update order status to cancelled
		if err := tx.Model(&Order{}).Where("id = ?", id).Updates(map[string]interface{}{
			"status": OrderStatusCancelled,
			"cancellation_reason": reason,
			"cancelled_at": time.Now(),
		}).Error; err != nil {
			return err
		}

		// Add status history entry
		statusHistory := &OrderStatusHistory{
			OrderID:   id,
			ToStatus:  OrderStatusCancelled,
			Note:      fmt.Sprintf("Cancelled by admin: %s", reason),
		}
		return tx.Create(statusHistory).Error
	})
}

func (r *Repository) GetStats(ctx context.Context, query OrderStatsQuery) (*OrderStats, error) {
	stats := &OrderStats{}

	db := r.db.WithContext(ctx).Model(&Order{})

	// Apply date filters if provided
	if query.DateFrom != nil {
		db = db.Where("created_at >= ?", *query.DateFrom)
	}
	if query.DateTo != nil {
		db = db.Where("created_at <= ?", *query.DateTo)
	}
	if query.UserID != nil {
		db = db.Where("customer_id = ?", *query.UserID)
	}

	// Total orders
	db.Count(&stats.TotalOrders)

	// Orders by status
	db.Where("status = ?", OrderStatusPending).Count(&stats.PendingOrders)
	db.Where("status = ?", OrderStatusConfirmed).Count(&stats.ConfirmedOrders)
	db.Where("status = ?", OrderStatusPreparing).Count(&stats.PreparingOrders)
	db.Where("status = ?", OrderStatusOutForDelivery).Count(&stats.ShippedOrders)
	db.Where("status = ?", OrderStatusDelivered).Count(&stats.DeliveredOrders)
	db.Where("status = ?", OrderStatusCancelled).Count(&stats.CancelledOrders)
	// RefundedOrders count removed as status doesn't exist in DB constraint
	stats.RefundedOrders = 0

	// Revenue calculations
	type RevenueResult struct {
		TotalRevenue   float64
		AverageOrder  float64
	}
	var revenueResult RevenueResult
	db.Select("COALESCE(SUM(total_amount), 0) as total_revenue, COALESCE(AVG(total_amount), 0) as average_order").Where("payment_status = ?", PaymentStatusPaid).Scan(&revenueResult)
	stats.TotalRevenue = int64(revenueResult.TotalRevenue)
	stats.AverageOrderValue = int64(revenueResult.AverageOrder)
	stats.RevenueNaira = revenueResult.TotalRevenue
	stats.AverageOrderValueNaira = revenueResult.AverageOrder

	return stats, nil
}

// Cart methods
func (r *Repository) GetCart(ctx context.Context, userID uuid.UUID) (*Cart, error) {
	var cart Cart
	err := r.db.WithContext(ctx).Preload("Items.Product").Where("customer_id = ?", userID).First(&cart).Error
	if err != nil {
		return nil, err
	}
	return &cart, nil
}

func (r *Repository) CreateCart(ctx context.Context, cart *Cart) error {
	return r.db.WithContext(ctx).Create(cart).Error
}

func (r *Repository) UpdateCart(ctx context.Context, cart *Cart) error {
	return r.db.WithContext(ctx).Save(cart).Error
}

func (r *Repository) ClearCart(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("customer_id = ?", userID).Delete(&CartItem{}).Error
}
