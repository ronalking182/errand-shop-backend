package coupons

import (
	"time"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Coupon struct {
	ID                   uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Code                 string         `gorm:"uniqueIndex;size:50;not null" json:"code"`
	Type                 CouponType     `gorm:"type:varchar(20);not null" json:"type"`
	Value                float64        `gorm:"type:decimal(10,2);not null" json:"value"`
	Description          string         `gorm:"type:text" json:"description"`
	MaxUsage             *int           `json:"maxUsage"`
	UsageCount           int            `gorm:"default:0" json:"usageCount"`
	ExpiryDate           *time.Time     `json:"expiryDate"`
	IsActive             bool           `gorm:"default:true" json:"isActive"`
	CreatedBy            string         `gorm:"size:20;not null" json:"createdBy"` // 'owner', 'system'
	CreatedByUserID      *uuid.UUID     `gorm:"type:uuid" json:"createdByUserId"`
	LinkedOrderID        *uuid.UUID     `gorm:"type:uuid" json:"linkedOrderId"`
	LinkedUserID         *uuid.UUID     `gorm:"type:uuid" json:"linkedUserId"`
	MinimumOrderAmount   float64        `gorm:"type:decimal(10,2);default:0" json:"minimumOrderAmount"`
	CreatedAt            time.Time      `json:"createdAt"`
	UpdatedAt            time.Time      `json:"updatedAt"`
	DeletedAt            gorm.DeletedAt `gorm:"index" json:"-"`
}

type CouponUsage struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CouponID       uuid.UUID `gorm:"type:uuid;not null" json:"couponId"`
	UserID         uuid.UUID `gorm:"type:uuid;not null" json:"userId"`
	OrderID        uuid.UUID `gorm:"type:uuid;not null" json:"orderId"`
	DiscountAmount float64   `gorm:"type:decimal(10,2)" json:"discountAmount"`
	UsedAt         time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"usedAt"`
	Coupon         Coupon    `gorm:"foreignKey:CouponID" json:"coupon,omitempty"`
}

type UserRefundCredit struct {
	ID                  uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID              uuid.UUID  `gorm:"type:uuid;not null" json:"userId"`
	OriginalOrderID     uuid.UUID  `gorm:"type:uuid;not null" json:"originalOrderId"`
	RefundAmount        float64    `gorm:"type:decimal(10,2)" json:"refundAmount"`
	ConvertedToCoupon   bool       `gorm:"default:false" json:"convertedToCoupon"`
	CouponID            *uuid.UUID `gorm:"type:uuid" json:"couponId"`
	CreatedAt           time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"createdAt"`
	Coupon              *Coupon    `gorm:"foreignKey:CouponID" json:"coupon,omitempty"`
}

type CouponType string

const (
	CouponPercentage CouponType = "percentage"
	CouponFixed      CouponType = "fixed"
)

type CreatedBy string

const (
	CreatedByOwner  CreatedBy = "owner"
	CreatedBySystem CreatedBy = "system"
)
