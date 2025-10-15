package coupons

import (
	"time"
	"github.com/google/uuid"
)

// Request DTOs
type CreateCouponRequest struct {
	Code               string     `json:"code" validate:"required,min=3,max=50"`
	Type               CouponType `json:"type" validate:"required,oneof=percentage fixed"`
	Value              float64    `json:"value" validate:"required,gt=0"`
	Description        string     `json:"description"`
	MaxUsage           *int       `json:"maxUsage" validate:"omitempty,gt=0"`
	ExpiryDate         *time.Time `json:"expiryDate"`
	IsActive           *bool      `json:"isActive"`
	LinkedUserID       *uuid.UUID `json:"linkedUserId"`
	MinimumOrderAmount float64    `json:"minimumOrderAmount" validate:"gte=0"`
}

type UpdateCouponRequest struct {
	Description        *string    `json:"description"`
	MaxUsage           *int       `json:"maxUsage" validate:"omitempty,gt=0"`
	ExpiryDate         *time.Time `json:"expiryDate"`
	IsActive           *bool      `json:"isActive"`
	MinimumOrderAmount *float64   `json:"minimumOrderAmount" validate:"omitempty,gte=0"`
}

type ValidateCouponRequest struct {
	Code        string    `json:"code" validate:"required"`
	UserID      uuid.UUID `json:"userId" validate:"required"`
	OrderAmount float64   `json:"orderAmount" validate:"required,gt=0"`
}

type ApplyCouponRequest struct {
	Code        string    `json:"code" validate:"required"`
	UserID      uuid.UUID `json:"userId" validate:"required"`
	OrderID     uuid.UUID `json:"orderId" validate:"required"`
	OrderAmount float64   `json:"orderAmount" validate:"required,gt=0"`
}

type ConvertRefundRequest struct {
	RefundCreditID uuid.UUID `json:"refundCreditId" validate:"required"`
	CouponCode     *string   `json:"couponCode"`
}

// Mobile App Auto-Generation Request
type MobileAutoGenerateCouponRequest struct {
	Type        CouponType `json:"type" validate:"required,oneof=percentage fixed"`
	Value       float64    `json:"value" validate:"required,gt=0,lte=50"` // Max 50% or $50
	Description string     `json:"description" validate:"required,min=10,max=200"`
	ExpiryDays  *int       `json:"expiryDays" validate:"omitempty,min=1,max=365"` // Optional, max 1 year
}

type CreateRefundCreditRequest struct {
	OrderID      uuid.UUID `json:"orderId" validate:"required"`
	UserID       uuid.UUID `json:"userId" validate:"required"`
	RefundAmount float64   `json:"refundAmount" validate:"required,gt=0"`
	Reason       string    `json:"reason" validate:"required"`
}

// Response DTOs
type CouponResponse struct {
	ID                 uuid.UUID  `json:"id"`
	Code               string     `json:"code"`
	Type               CouponType `json:"type"`
	Value              float64    `json:"value"`
	Description        string     `json:"description"`
	MaxUsage           *int       `json:"maxUsage"`
	UsageCount         int        `json:"usageCount"`
	ExpiryDate         *time.Time `json:"expiryDate"`
	IsActive           bool       `json:"isActive"`
	CreatedBy          string     `json:"createdBy"`
	LinkedOrderID      *uuid.UUID `json:"linkedOrderId"`
	LinkedUserID       *uuid.UUID `json:"linkedUserId"`
	MinimumOrderAmount float64    `json:"minimumOrderAmount"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt"`
}

type CouponValidationResponse struct {
	Valid          bool    `json:"valid"`
	DiscountAmount float64 `json:"discountAmount"`
	Message        string  `json:"message"`
	Coupon         *CouponResponse `json:"coupon,omitempty"`
}

type UserRefundCreditResponse struct {
	ID                uuid.UUID        `json:"id"`
	UserID            uuid.UUID        `json:"userId"`
	OriginalOrderID   uuid.UUID        `json:"originalOrderId"`
	RefundAmount      float64          `json:"refundAmount"`
	ConvertedToCoupon bool             `json:"convertedToCoupon"`
	CouponID          *uuid.UUID       `json:"couponId"`
	CreatedAt         time.Time        `json:"createdAt"`
	Coupon            *CouponResponse  `json:"coupon,omitempty"`
}

type CouponStatsResponse struct {
	TotalCoupons           int     `json:"totalCoupons"`
	ActiveCoupons          int     `json:"activeCoupons"`
	TotalUsage             int     `json:"totalUsage"`
	SystemCoupons          int     `json:"systemCoupons"`
	TotalDiscountGiven     float64 `json:"totalDiscountGiven"`
	TopPerformingCoupons   []CouponPerformance `json:"topPerformingCoupons"`
}

type CouponPerformance struct {
	Coupon         CouponResponse `json:"coupon"`
	UsageCount     int            `json:"usageCount"`
	TotalDiscount  float64        `json:"totalDiscount"`
}

// Pagination
type CouponListResponse struct {
	Coupons []CouponResponse `json:"coupons"`
	Total   int64            `json:"total"`
	Page    int              `json:"page"`
	Limit   int              `json:"limit"`
}

type RefundCreditListResponse struct {
	Credits []UserRefundCreditResponse `json:"credits"`
	Total   int64                      `json:"total"`
	Page    int                        `json:"page"`
	Limit   int                        `json:"limit"`
}