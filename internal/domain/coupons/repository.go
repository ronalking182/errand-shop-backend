package coupons

import (
	"fmt"
	"time"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	// Coupon CRUD
	Create(coupon *Coupon) error
	GetByID(id uuid.UUID) (*Coupon, error)
	GetByCode(code string) (*Coupon, error)
	Update(coupon *Coupon) error
	Delete(id uuid.UUID) error
	List(page, limit int, filters map[string]interface{}) ([]Coupon, int64, error)
	ToggleActive(id uuid.UUID) error
	
	// Coupon Usage
	CreateUsage(usage *CouponUsage) error
	GetUsageByUserAndCoupon(userID, couponID uuid.UUID) (*CouponUsage, error)
	GetUsageCountByCoupon(couponID uuid.UUID) (int64, error)
	GetUsageCountByUserAndCoupon(userID, couponID uuid.UUID) (int64, error)
	
	// Refund Credits
	CreateRefundCredit(credit *UserRefundCredit) error
	GetRefundCreditByID(id uuid.UUID) (*UserRefundCredit, error)
	GetRefundCreditsByUser(userID uuid.UUID, page, limit int) ([]UserRefundCredit, int64, error)
	UpdateRefundCredit(credit *UserRefundCredit) error
	
	// Analytics
	GetCouponStats() (*CouponStatsResponse, error)
	GetTopPerformingCoupons(limit int) ([]CouponPerformance, error)
	
	// Public operations
	GetPublicCoupons() ([]Coupon, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// Coupon CRUD Operations
func (r *repository) Create(coupon *Coupon) error {
	return r.db.Create(coupon).Error
}

func (r *repository) GetByID(id uuid.UUID) (*Coupon, error) {
	var coupon Coupon
	err := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&coupon).Error
	if err != nil {
		return nil, err
	}
	return &coupon, nil
}

func (r *repository) GetByCode(code string) (*Coupon, error) {
	var coupon Coupon
	err := r.db.Where("code = ? AND deleted_at IS NULL", code).First(&coupon).Error
	if err != nil {
		return nil, err
	}
	return &coupon, nil
}

func (r *repository) Update(coupon *Coupon) error {
	return r.db.Save(coupon).Error
}

func (r *repository) Delete(id uuid.UUID) error {
	return r.db.Where("id = ?", id).Delete(&Coupon{}).Error
}

func (r *repository) List(page, limit int, filters map[string]interface{}) ([]Coupon, int64, error) {
	var coupons []Coupon
	var total int64
	
	query := r.db.Model(&Coupon{}).Where("deleted_at IS NULL")
	
	// Apply filters
	for key, value := range filters {
		switch key {
		case "is_active":
			query = query.Where("is_active = ?", value)
		case "type":
			query = query.Where("type = ?", value)
		case "created_by":
			query = query.Where("created_by = ?", value)
		case "linked_user_id":
			query = query.Where("linked_user_id = ?", value)
		case "search":
			query = query.Where("code ILIKE ? OR description ILIKE ?", 
				fmt.Sprintf("%%%s%%", value), fmt.Sprintf("%%%s%%", value))
		}
	}
	
	// Count total
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	
	// Get paginated results
	offset := (page - 1) * limit
	err = query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&coupons).Error
	if err != nil {
		return nil, 0, err
	}
	
	return coupons, total, nil
}

func (r *repository) ToggleActive(id uuid.UUID) error {
	return r.db.Model(&Coupon{}).Where("id = ?", id).Update("is_active", gorm.Expr("NOT is_active")).Error
}

// Coupon Usage Operations
func (r *repository) CreateUsage(usage *CouponUsage) error {
	return r.db.Create(usage).Error
}

func (r *repository) GetUsageByUserAndCoupon(userID, couponID uuid.UUID) (*CouponUsage, error) {
	var usage CouponUsage
	err := r.db.Where("user_id = ? AND coupon_id = ?", userID, couponID).First(&usage).Error
	if err != nil {
		return nil, err
	}
	return &usage, nil
}

func (r *repository) GetUsageCountByCoupon(couponID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&CouponUsage{}).Where("coupon_id = ?", couponID).Count(&count).Error
	return count, err
}

func (r *repository) GetUsageCountByUserAndCoupon(userID, couponID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&CouponUsage{}).Where("user_id = ? AND coupon_id = ?", userID, couponID).Count(&count).Error
	return count, err
}

// Refund Credits Operations
func (r *repository) CreateRefundCredit(credit *UserRefundCredit) error {
	return r.db.Create(credit).Error
}

func (r *repository) GetRefundCreditByID(id uuid.UUID) (*UserRefundCredit, error) {
	var credit UserRefundCredit
	err := r.db.Preload("Coupon").Where("id = ?", id).First(&credit).Error
	if err != nil {
		return nil, err
	}
	return &credit, nil
}

func (r *repository) GetRefundCreditsByUser(userID uuid.UUID, page, limit int) ([]UserRefundCredit, int64, error) {
	var credits []UserRefundCredit
	var total int64
	
	// Count total
	err := r.db.Model(&UserRefundCredit{}).Where("user_id = ?", userID).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	
	// Get paginated results
	offset := (page - 1) * limit
	err = r.db.Preload("Coupon").Where("user_id = ?", userID).
		Order("created_at DESC").Offset(offset).Limit(limit).Find(&credits).Error
	if err != nil {
		return nil, 0, err
	}
	
	return credits, total, nil
}

func (r *repository) UpdateRefundCredit(credit *UserRefundCredit) error {
	return r.db.Save(credit).Error
}

// Analytics Operations
func (r *repository) GetCouponStats() (*CouponStatsResponse, error) {
	var stats CouponStatsResponse
	var err error
	
	// Total coupons
	var totalCount int64
	err = r.db.Model(&Coupon{}).Where("deleted_at IS NULL").Count(&totalCount).Error
	if err != nil {
		return nil, err
	}
	stats.TotalCoupons = int(totalCount)
	
	// Active coupons
	var activeCoupons int64
	err = r.db.Model(&Coupon{}).Where("deleted_at IS NULL AND is_active = true").Count(&activeCoupons).Error
	if err != nil {
		return nil, err
	}
	stats.ActiveCoupons = int(activeCoupons)
	
	// System coupons
	var systemCoupons int64
	err = r.db.Model(&Coupon{}).Where("deleted_at IS NULL AND created_by = 'system'").Count(&systemCoupons).Error
	if err != nil {
		return nil, err
	}
	stats.SystemCoupons = int(systemCoupons)
	
	// Total usage and discount
	var result struct {
		TotalUsage    int64   `json:"total_usage"`
		TotalDiscount float64 `json:"total_discount"`
	}
	
	err = r.db.Model(&CouponUsage{}).
		Select("COUNT(*) as total_usage, COALESCE(SUM(discount_amount), 0) as total_discount").
		Scan(&result).Error
	if err != nil {
		return nil, err
	}
	
	stats.TotalUsage = int(result.TotalUsage)
	stats.TotalDiscountGiven = result.TotalDiscount
	
	// Get top performing coupons
	topCoupons, err := r.GetTopPerformingCoupons(5)
	if err != nil {
		return nil, err
	}
	stats.TopPerformingCoupons = topCoupons
	
	return &stats, nil
}

func (r *repository) GetTopPerformingCoupons(limit int) ([]CouponPerformance, error) {
	var performances []CouponPerformance
	
	type result struct {
		CouponID      uuid.UUID `json:"coupon_id"`
		UsageCount    int64     `json:"usage_count"`
		TotalDiscount float64   `json:"total_discount"`
	}
	
	var results []result
	err := r.db.Model(&CouponUsage{}).
		Select("coupon_id, COUNT(*) as usage_count, COALESCE(SUM(discount_amount), 0) as total_discount").
		Group("coupon_id").
		Order("usage_count DESC").
		Limit(limit).
		Scan(&results).Error
	if err != nil {
		return nil, err
	}
	
	for _, result := range results {
		coupon, err := r.GetByID(result.CouponID)
		if err != nil {
			continue
		}
		
		performance := CouponPerformance{
			Coupon:        *r.toCouponResponse(coupon),
			UsageCount:    int(result.UsageCount),
			TotalDiscount: result.TotalDiscount,
		}
		performances = append(performances, performance)
	}
	
	return performances, nil
}

// GetPublicCoupons gets publicly available coupons
func (r *repository) GetPublicCoupons() ([]Coupon, error) {
	var coupons []Coupon
	err := r.db.Where("is_active = ? AND expiry_date > ?", true, time.Now()).Find(&coupons).Error
	return coupons, err
}

// Helper methods
func (r *repository) toCouponResponse(coupon *Coupon) *CouponResponse {
	return &CouponResponse{
		ID:                 coupon.ID,
		Code:               coupon.Code,
		Type:               coupon.Type,
		Value:              coupon.Value,
		Description:        coupon.Description,
		MaxUsage:           coupon.MaxUsage,
		UsageCount:         coupon.UsageCount,
		ExpiryDate:         coupon.ExpiryDate,
		IsActive:           coupon.IsActive,
		CreatedBy:          coupon.CreatedBy,
		LinkedOrderID:      coupon.LinkedOrderID,
		LinkedUserID:       coupon.LinkedUserID,
		MinimumOrderAmount: coupon.MinimumOrderAmount,
		CreatedAt:          coupon.CreatedAt,
		UpdatedAt:          coupon.UpdatedAt,
	}
}