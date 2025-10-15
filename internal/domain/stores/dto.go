package stores

import "time"

// Request DTOs
type CreateStoreRequest struct {
	Name         string                 `json:"name" binding:"required"`
	Description  string                 `json:"description"`
	Address      string                 `json:"address" binding:"required"`
	City         string                 `json:"city" binding:"required"`
	State        string                 `json:"state" binding:"required"`
	ZipCode      string                 `json:"zipCode" binding:"required"`
	Country      string                 `json:"country"`
	Lat          *float64               `json:"lat"`
	Lng          *float64               `json:"lng"`
	Phone        string                 `json:"phone"`
	Email        string                 `json:"email" binding:"omitempty,email"`
	Website      string                 `json:"website"`
	OpeningHours map[string]interface{} `json:"openingHours"`
	ManagerID    *uint                  `json:"managerId"`
}

type UpdateStoreRequest struct {
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Address      string                 `json:"address"`
	City         string                 `json:"city"`
	State        string                 `json:"state"`
	ZipCode      string                 `json:"zipCode"`
	Country      string                 `json:"country"`
	Lat          *float64               `json:"lat"`
	Lng          *float64               `json:"lng"`
	Phone        string                 `json:"phone"`
	Email        string                 `json:"email" binding:"omitempty,email"`
	Website      string                 `json:"website"`
	OpeningHours map[string]interface{} `json:"openingHours"`
	ManagerID    *uint                  `json:"managerId"`
}

type UpdateStoreStatusRequest struct {
	IsActive bool `json:"isActive"`
	IsOpen   bool `json:"isOpen"`
}

type AddInventoryRequest struct {
	ProductID uint    `json:"productId" binding:"required"`
	Quantity  int     `json:"quantity" binding:"required,min=0"`
	MinStock  int     `json:"minStock" binding:"min=0"`
	MaxStock  int     `json:"maxStock" binding:"min=1"`
	Price     float64 `json:"price" binding:"required,min=0"`
}

type UpdateInventoryRequest struct {
	Quantity int     `json:"quantity" binding:"min=0"`
	MinStock int     `json:"minStock" binding:"min=0"`
	MaxStock int     `json:"maxStock" binding:"min=1"`
	Price    float64 `json:"price" binding:"min=0"`
	IsActive bool    `json:"isActive"`
}

type StoreHoursRequest struct {
	DayOfWeek int    `json:"dayOfWeek" binding:"required,min=0,max=6"`
	OpenTime  string `json:"openTime" binding:"required"`
	CloseTime string `json:"closeTime" binding:"required"`
	IsClosed  bool   `json:"isClosed"`
}

type AssignManagerRequest struct {
	UserID uint   `json:"userId" binding:"required"`
	Role   string `json:"role" binding:"required,oneof=manager assistant_manager"`
}

// Response DTOs
type StoreResponse struct {
	ID           uint                   `json:"id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Address      string                 `json:"address"`
	City         string                 `json:"city"`
	State        string                 `json:"state"`
	ZipCode      string                 `json:"zipCode"`
	Country      string                 `json:"country"`
	Lat          *float64               `json:"lat"`
	Lng          *float64               `json:"lng"`
	Phone        string                 `json:"phone"`
	Email        string                 `json:"email"`
	Website      string                 `json:"website"`
	IsActive     bool                   `json:"isActive"`
	IsOpen       bool                   `json:"isOpen"`
	OpeningHours map[string]interface{} `json:"openingHours"`
	ManagerID    *uint                  `json:"managerId"`
	CreatedAt    time.Time              `json:"createdAt"`
	UpdatedAt    time.Time              `json:"updatedAt"`
}

type StoreListResponse struct {
	Stores     []StoreResponse `json:"stores"`
	Total      int64           `json:"total"`
	Page       int             `json:"page"`
	Limit      int             `json:"limit"`
	TotalPages int             `json:"totalPages"`
}

type StoreInventoryResponse struct {
	ID        uint      `json:"id"`
	StoreID   uint      `json:"storeId"`
	ProductID uint      `json:"productId"`
	Quantity  int       `json:"quantity"`
	MinStock  int       `json:"minStock"`
	MaxStock  int       `json:"maxStock"`
	Price     float64   `json:"price"`
	IsActive  bool      `json:"isActive"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type StoreHoursResponse struct {
	ID        uint   `json:"id"`
	StoreID   uint   `json:"storeId"`
	DayOfWeek int    `json:"dayOfWeek"`
	OpenTime  string `json:"openTime"`
	CloseTime string `json:"closeTime"`
	IsClosed  bool   `json:"isClosed"`
}

type StoreManagerResponse struct {
	ID        uint      `json:"id"`
	StoreID   uint      `json:"storeId"`
	UserID    uint      `json:"userId"`
	Role      string    `json:"role"`
	IsActive  bool      `json:"isActive"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type StoreStatsResponse struct {
	TotalStores    int64 `json:"totalStores"`
	ActiveStores   int64 `json:"activeStores"`
	OpenStores     int64 `json:"openStores"`
	TotalInventory int64 `json:"totalInventory"`
	LowStockItems  int64 `json:"lowStockItems"`
}

// Helper functions
func ToStoreResponse(store *Store) StoreResponse {
	return StoreResponse{
		ID:           store.ID,
		Name:         store.Name,
		Description:  store.Description,
		Address:      store.Address,
		City:         store.City,
		State:        store.State,
		ZipCode:      store.ZipCode,
		Country:      store.Country,
		Lat:          store.Lat,
		Lng:          store.Lng,
		Phone:        store.Phone,
		Email:        store.Email,
		Website:      store.Website,
		IsActive:     store.IsActive,
		IsOpen:       store.IsOpen,
		OpeningHours: store.OpeningHours,
		ManagerID:    store.ManagerID,
		CreatedAt:    store.CreatedAt,
		UpdatedAt:    store.UpdatedAt,
	}
}

func ToStoreResponses(stores []Store) []StoreResponse {
	responses := make([]StoreResponse, len(stores))
	for i, store := range stores {
		responses[i] = ToStoreResponse(&store)
	}
	return responses
}

func ToStoreInventoryResponse(inventory *StoreInventory) StoreInventoryResponse {
	return StoreInventoryResponse{
		ID:        inventory.ID,
		StoreID:   inventory.StoreID,
		ProductID: inventory.ProductID,
		Quantity:  inventory.Quantity,
		MinStock:  inventory.MinStock,
		MaxStock:  inventory.MaxStock,
		Price:     inventory.Price,
		IsActive:  inventory.IsActive,
		CreatedAt: inventory.CreatedAt,
		UpdatedAt: inventory.UpdatedAt,
	}
}

func ToStoreInventoryResponses(inventories []StoreInventory) []StoreInventoryResponse {
	responses := make([]StoreInventoryResponse, len(inventories))
	for i, inventory := range inventories {
		responses[i] = ToStoreInventoryResponse(&inventory)
	}
	return responses
}
