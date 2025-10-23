package products

import (
	"time"
	"github.com/google/uuid"
)

type ListQuery struct {
	Q     string
	Page  int
	Limit int
}

// Product DTOs
type CreateProductRequest struct {
	Name              string   `json:"name" validate:"required,min=2,max=255"`
	Description       string   `json:"description" validate:"omitempty,max=2000"`
	CostPrice         float64  `json:"costPrice" validate:"required,min=0"`
	SellingPrice      float64  `json:"sellingPrice" validate:"required,min=0"`
	StockQuantity     int      `json:"stockQuantity" validate:"required,min=0"`
	ImageURL          string   `json:"imageUrl" validate:"omitempty,url"`
	ImagePublicID     string   `json:"imagePublicId" validate:"omitempty"`
	Category          string   `json:"category" validate:"required,min=2,max=100"`
	Tags              StringSlice `json:"tags" validate:"omitempty,dive,min=1,max=50"`
	LowStockThreshold int      `json:"lowStockThreshold" validate:"min=0"`
}

type UpdateProductRequest struct {
	Name              *string   `json:"name" validate:"omitempty,min=2,max=255"`
	Description       *string   `json:"description" validate:"omitempty,max=2000"`
	CostPrice         *float64  `json:"costPrice" validate:"omitempty,min=0"`
	SellingPrice      *float64  `json:"sellingPrice" validate:"omitempty,min=0"`
	StockQuantity     *int      `json:"stockQuantity" validate:"omitempty,min=0"`
	ImageURL          *string   `json:"imageUrl" validate:"omitempty,url"`
	ImagePublicID     *string   `json:"imagePublicId" validate:"omitempty"`
	Category          *string   `json:"category" validate:"omitempty,min=2,max=100"`
	Tags              *StringSlice `json:"tags" validate:"omitempty,dive,min=1,max=50"`
	LowStockThreshold *int      `json:"lowStockThreshold" validate:"omitempty,min=0"`
	IsActive          *bool     `json:"isActive" validate:"omitempty"`
}

type ProductResponse struct {
	ID                uuid.UUID `json:"id"`
	Name              string    `json:"name"`
	SKU               string    `json:"sku"`
	Slug              string    `json:"slug"`
	Description       string    `json:"description"`
	CostPrice         float64   `json:"costPrice"`
	SellingPrice      float64   `json:"sellingPrice"`
	Profit            float64   `json:"profit"`
	StockQuantity     int       `json:"stockQuantity"`
	ImageURL          string    `json:"imageUrl"`
	ImagePublicID     string    `json:"imagePublicId"`
	Category          string    `json:"category"`
	Tags              StringSlice  `json:"tags"`
	LowStockThreshold int       `json:"lowStockThreshold"`
	IsLowStock        bool      `json:"isLowStock"`
	IsActive          bool      `json:"isActive"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

// Stock Management DTOs
type StockUpdateRequest struct {
	Quantity   int    `json:"quantity" validate:"required"`
	ChangeType string `json:"changeType" validate:"required,oneof=ADD REMOVE ADJUST"`
	Reason     string `json:"reason" validate:"omitempty,max=500"`
}

type StockUpdateResponse struct {
	PreviousQuantity int `json:"previousQuantity"`
	NewQuantity      int `json:"newQuantity"`
	Change           int `json:"change"`
}

// Category DTOs
type CreateCategoryRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Description string `json:"description" validate:"omitempty,max=500"`
}

type UpdateCategoryRequest struct {
	Name        *string `json:"name" validate:"omitempty,min=2,max=100"`
	Description *string `json:"description" validate:"omitempty,max=500"`
	IsActive    *bool   `json:"isActive" validate:"omitempty"`
}

type CategoryResponse struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	ProductCount int       `json:"productCount"`
	IsActive     bool      `json:"isActive"`
	CreatedAt    time.Time `json:"createdAt"`
}

// Image Upload DTOs
type ImageUploadResponse struct {
	ImageURL    string `json:"imageUrl"`
	PublicID    string `json:"publicId"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	Size        int64  `json:"size"`
	Format      string `json:"format"`
}

// Bulk Operations DTOs
type BulkUpdateStockRequest struct {
	Updates []BulkStockUpdate `json:"updates" validate:"required,dive"`
	Reason  string            `json:"reason" validate:"omitempty,max=500"`
}

type BulkStockUpdate struct {
	ProductID  uuid.UUID `json:"productId" validate:"required"`
	Quantity   int       `json:"quantity" validate:"required"`
	ChangeType string    `json:"changeType" validate:"required,oneof=ADD REMOVE ADJUST"`
}

// Query DTOs
type AdminListQuery struct {
	ListQuery
	Category   string `query:"category"`
	LowStock   bool   `query:"low_stock"`
	OutOfStock bool   `query:"out_of_stock"`
	Search     string `query:"search"`
	SortBy     string `query:"sort_by" validate:"omitempty,oneof=name price stock created_at"`
	SortOrder  string `query:"sort_order" validate:"omitempty,oneof=asc desc"`
}

// Analytics DTOs
type ProductAnalytics struct {
	TotalProducts    int     `json:"totalProducts"`
	LowStockCount    int     `json:"lowStockCount"`
	OutOfStockCount  int     `json:"outOfStockCount"`
	TotalValue       float64 `json:"totalValue"`
	AveragePrice     float64 `json:"averagePrice"`
	TopCategories    []CategoryStats `json:"topCategories"`
}

type CategoryStats struct {
	Category     string `json:"category"`
	ProductCount int    `json:"productCount"`
	TotalValue   float64 `json:"totalValue"`
}

// List Result DTOs
type ListResult struct {
	Data []ProductResponse `json:"data"`
	Meta PageMeta         `json:"meta"`
}

type PageMeta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
}
