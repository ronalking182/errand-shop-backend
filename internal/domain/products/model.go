package products

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// StringSlice is a custom type for handling []string with JSONB
type StringSlice []string

// Value implements the driver.Valuer interface for database storage
func (s StringSlice) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	return json.Marshal(s)
}

// Scan implements the sql.Scanner interface for database retrieval
func (s *StringSlice) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, s)
	case string:
		return json.Unmarshal([]byte(v), s)
	default:
		return errors.New("cannot scan into StringSlice")
	}
}

// Product represents a product in the system
type Product struct {
	ID                uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name              string    `gorm:"size:255;not null" json:"name"`
	SKU               string    `gorm:"size:50;uniqueIndex;not null" json:"sku"`
	Slug              string    `gorm:"uniqueIndex;size:200" json:"slug"`
	Description       string    `gorm:"type:text" json:"description"`
	CostPrice         float64   `gorm:"type:decimal(10,2);not null" json:"costPrice"`
	SellingPrice      float64   `gorm:"type:decimal(10,2);not null" json:"sellingPrice"`
	StockQuantity     int       `gorm:"not null;default:0" json:"stockQuantity"`
	LowStockThreshold int       `gorm:"not null;default:10" json:"lowStockThreshold"`
	ImageURL          string    `gorm:"size:500" json:"imageUrl"`
	ImagePublicID     string    `gorm:"size:255" json:"imagePublicId"`
	Category          string    `gorm:"size:120" json:"category"`
	Tags              StringSlice  `gorm:"type:jsonb" json:"tags"`
	IsActive          bool      `gorm:"default:true" json:"isActive"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

// Category represents a product category
type Category struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name        string    `gorm:"size:100;not null;uniqueIndex" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	IsActive    bool      `gorm:"default:true" json:"isActive"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// StockHistory tracks stock changes
type StockHistory struct {
	ID               uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ProductID        uuid.UUID `gorm:"type:uuid;not null;index" json:"productId"`
	ChangeType       string    `gorm:"size:20;not null" json:"changeType"` // ADD, REMOVE, ADJUST, SALE
	QuantityChange   int       `gorm:"not null" json:"quantityChange"`
	PreviousQuantity int       `gorm:"not null" json:"previousQuantity"`
	NewQuantity      int       `gorm:"not null" json:"newQuantity"`
	Reason           string    `gorm:"type:text" json:"reason"`
	CreatedAt        time.Time `json:"createdAt"`
	CreatedBy        uuid.UUID `gorm:"type:uuid" json:"createdBy"`
	
	// Relationships
	Product Product `gorm:"foreignKey:ProductID" json:"-"`
}

// TableName sets the table name for Product
func (Product) TableName() string {
	return "products"
}

// TableName sets the table name for Category
func (Category) TableName() string {
	return "categories"
}

// TableName sets the table name for StockHistory
func (StockHistory) TableName() string {
	return "stock_history"
}

// BeforeCreate generates SKU for new products
func (p *Product) BeforeCreate(tx *gorm.DB) error {
	if p.SKU == "" {
		p.SKU = generateSKU()
	}
	return nil
}

// Profit calculates the profit margin
func (p *Product) Profit() float64 {
	return p.SellingPrice - p.CostPrice
}

// IsLowStock checks if product is below threshold
func (p *Product) IsLowStock() bool {
	return p.StockQuantity <= p.LowStockThreshold
}

// generateSKU creates a unique SKU
func generateSKU() string {
	return "PRD-" + uuid.New().String()[:8]
}
