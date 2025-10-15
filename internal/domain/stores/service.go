package stores

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// Store operations
func (s *Service) CreateStore(req CreateStoreRequest) (*Store, error) {
	store := &Store{
		Name:         req.Name,
		Description:  req.Description,
		Address:      req.Address,
		City:         req.City,
		State:        req.State,
		ZipCode:      req.ZipCode,
		Country:      req.Country,
		Lat:          req.Lat,
		Lng:          req.Lng,
		Phone:        req.Phone,
		Email:        req.Email,
		Website:      req.Website,
		IsActive:     true,
		IsOpen:       true,
		OpeningHours: req.OpeningHours,
		ManagerID:    req.ManagerID,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if req.Country == "" {
		store.Country = "USA"
	}

	err := s.repo.CreateStore(store)
	if err != nil {
		return nil, err
	}

	return store, nil
}

func (s *Service) GetStore(id uint) (*Store, error) {
	return s.repo.GetStoreByID(id)
}

func (s *Service) GetStores(page, limit int, isActive *bool) ([]Store, int64, error) {
	offset := (page - 1) * limit
	return s.repo.GetStores(limit, offset, isActive)
}

func (s *Service) UpdateStore(id uint, req UpdateStoreRequest) (*Store, error) {
	store, err := s.repo.GetStoreByID(id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.Name != "" {
		store.Name = req.Name
	}
	if req.Description != "" {
		store.Description = req.Description
	}
	if req.Address != "" {
		store.Address = req.Address
	}
	if req.City != "" {
		store.City = req.City
	}
	if req.State != "" {
		store.State = req.State
	}
	if req.ZipCode != "" {
		store.ZipCode = req.ZipCode
	}
	if req.Country != "" {
		store.Country = req.Country
	}
	if req.Lat != nil {
		store.Lat = req.Lat
	}
	if req.Lng != nil {
		store.Lng = req.Lng
	}
	if req.Phone != "" {
		store.Phone = req.Phone
	}
	if req.Email != "" {
		store.Email = req.Email
	}
	if req.Website != "" {
		store.Website = req.Website
	}
	if req.OpeningHours != nil {
		store.OpeningHours = req.OpeningHours
	}
	if req.ManagerID != nil {
		store.ManagerID = req.ManagerID
	}

	store.UpdatedAt = time.Now()

	err = s.repo.UpdateStore(store)
	if err != nil {
		return nil, err
	}

	return store, nil
}

func (s *Service) UpdateStoreStatus(id uint, req UpdateStoreStatusRequest) error {
	store, err := s.repo.GetStoreByID(id)
	if err != nil {
		return err
	}

	store.IsActive = req.IsActive
	store.IsOpen = req.IsOpen
	store.UpdatedAt = time.Now()

	return s.repo.UpdateStore(store)
}

func (s *Service) DeleteStore(id uint) error {
	return s.repo.DeleteStore(id)
}

func (s *Service) GetStoresByLocation(lat, lng, radius float64) ([]Store, error) {
	return s.repo.GetStoresByLocation(lat, lng, radius)
}

// Inventory operations
func (s *Service) AddInventoryItem(storeID uint, req AddInventoryRequest) (*StoreInventory, error) {
	// Check if item already exists
	existing, err := s.repo.GetInventoryItem(storeID, req.ProductID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if existing != nil {
		return nil, errors.New("inventory item already exists for this product")
	}

	inventory := &StoreInventory{
		StoreID:   storeID,
		ProductID: req.ProductID,
		Quantity:  req.Quantity,
		MinStock:  req.MinStock,
		MaxStock:  req.MaxStock,
		Price:     req.Price,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = s.repo.AddInventoryItem(inventory)
	if err != nil {
		return nil, err
	}

	return inventory, nil
}

func (s *Service) GetStoreInventory(storeID uint) ([]StoreInventory, error) {
	return s.repo.GetStoreInventory(storeID)
}

func (s *Service) UpdateInventoryItem(storeID, productID uint, req UpdateInventoryRequest) (*StoreInventory, error) {
	inventory, err := s.repo.GetInventoryItem(storeID, productID)
	if err != nil {
		return nil, err
	}

	inventory.Quantity = req.Quantity
	inventory.MinStock = req.MinStock
	inventory.MaxStock = req.MaxStock
	inventory.Price = req.Price
	inventory.IsActive = req.IsActive
	inventory.UpdatedAt = time.Now()

	err = s.repo.UpdateInventoryItem(inventory)
	if err != nil {
		return nil, err
	}

	return inventory, nil
}

func (s *Service) DeleteInventoryItem(storeID, productID uint) error {
	return s.repo.DeleteInventoryItem(storeID, productID)
}

func (s *Service) GetLowStockItems(storeID uint) ([]StoreInventory, error) {
	return s.repo.GetLowStockItems(storeID)
}

// Store Hours operations
func (s *Service) SetStoreHours(storeID uint, hoursReq []StoreHoursRequest) error {
	hours := make([]StoreHours, len(hoursReq))
	for i, req := range hoursReq {
		hours[i] = StoreHours{
			StoreID:   storeID,
			DayOfWeek: req.DayOfWeek,
			OpenTime:  req.OpenTime,
			CloseTime: req.CloseTime,
			IsClosed:  req.IsClosed,
		}
	}

	return s.repo.SetStoreHours(storeID, hours)
}

func (s *Service) GetStoreHours(storeID uint) ([]StoreHours, error) {
	return s.repo.GetStoreHours(storeID)
}

// Manager operations
func (s *Service) AssignManager(storeID uint, req AssignManagerRequest) (*StoreManager, error) {
	manager := &StoreManager{
		StoreID:   storeID,
		UserID:    req.UserID,
		Role:      req.Role,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := s.repo.AssignManager(manager)
	if err != nil {
		return nil, err
	}

	return manager, nil
}

func (s *Service) GetStoreManagers(storeID uint) ([]StoreManager, error) {
	return s.repo.GetStoreManagers(storeID)
}

func (s *Service) RemoveManager(storeID, userID uint) error {
	return s.repo.RemoveManager(storeID, userID)
}

// Statistics
func (s *Service) GetStoreStats() (*StoreStatsResponse, error) {
	return s.repo.GetStoreStats()
}
