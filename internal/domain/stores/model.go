package stores

import "time"

// Store represents a physical store location
type Store struct {
	ID           uint                   `gorm:"primaryKey" json:"id"`
	Name         string                 `gorm:"size:200;not null" json:"name"`
	Description  string                 `gorm:"type:text" json:"description"`
	Address      string                 `gorm:"not null" json:"address"`
	City         string                 `gorm:"size:100" json:"city"`
	State        string                 `gorm:"size:100" json:"state"`
	ZipCode      string                 `gorm:"size:20" json:"zipCode"`
	Country      string                 `gorm:"size:100;default:'USA'" json:"country"`
	Lat          *float64               `json:"lat"`
	Lng          *float64               `json:"lng"`
	Phone        string                 `gorm:"size:20" json:"phone"`
	Email        string                 `gorm:"size:255" json:"email"`
	Website      string                 `gorm:"size:255" json:"website"`
	IsActive     bool                   `gorm:"default:true" json:"isActive"`
	IsOpen       bool                   `gorm:"default:true" json:"isOpen"`
	OpeningHours map[string]interface{} `gorm:"type:jsonb" json:"openingHours"`
	ManagerID    *uint                  `gorm:"index" json:"managerId"`
	Inventory    []StoreInventory       `json:"inventory"`
	CreatedAt    time.Time              `json:"createdAt"`
	UpdatedAt    time.Time              `json:"updatedAt"`
}

// StoreInventory represents inventory items in a store
type StoreInventory struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	StoreID   uint      `gorm:"not null;index" json:"storeId"`
	ProductID uint      `gorm:"not null;index" json:"productId"`
	Quantity  int       `gorm:"not null;default:0" json:"quantity"`
	MinStock  int       `gorm:"default:10" json:"minStock"`
	MaxStock  int       `gorm:"default:1000" json:"maxStock"`
	Price     float64   `gorm:"type:decimal(10,2)" json:"price"`
	IsActive  bool      `gorm:"default:true" json:"isActive"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// StoreHours represents store operating hours
type StoreHours struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	StoreID   uint   `gorm:"not null;index" json:"storeId"`
	DayOfWeek int    `gorm:"not null" json:"dayOfWeek"` // 0=Sunday, 1=Monday, etc.
	OpenTime  string `gorm:"size:10" json:"openTime"`   // "09:00"
	CloseTime string `gorm:"size:10" json:"closeTime"`  // "18:00"
	IsClosed  bool   `gorm:"default:false" json:"isClosed"`
}

// StoreManager represents store management assignments
type StoreManager struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	StoreID   uint      `gorm:"not null;index" json:"storeId"`
	UserID    uint      `gorm:"not null;index" json:"userId"`
	Role      string    `gorm:"size:50;default:'manager'" json:"role"` // manager, assistant_manager
	IsActive  bool      `gorm:"default:true" json:"isActive"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// TableName methods for custom table names
func (Store) TableName() string {
	return "stores"
}

func (StoreInventory) TableName() string {
	return "store_inventory"
}

func (StoreHours) TableName() string {
	return "store_hours"
}

func (StoreManager) TableName() string {
	return "store_managers"
}
