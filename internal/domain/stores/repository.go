package stores

import (
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Store operations
func (r *Repository) CreateStore(store *Store) error {
	return r.db.Create(store).Error
}

func (r *Repository) GetStoreByID(id uint) (*Store, error) {
	var store Store
	err := r.db.Preload("Inventory").First(&store, id).Error
	if err != nil {
		return nil, err
	}
	return &store, nil
}

func (r *Repository) GetStores(limit, offset int, isActive *bool) ([]Store, int64, error) {
	var stores []Store
	var total int64

	query := r.db.Model(&Store{})
	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	// Get total count
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err = query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&stores).Error
	return stores, total, err
}

func (r *Repository) UpdateStore(store *Store) error {
	return r.db.Save(store).Error
}

func (r *Repository) DeleteStore(id uint) error {
	return r.db.Delete(&Store{}, id).Error
}

func (r *Repository) GetStoresByLocation(lat, lng float64, radius float64) ([]Store, error) {
	var stores []Store
	// Using Haversine formula for distance calculation
	query := `
		SELECT *, (
			6371 * acos(
				cos(radians(?)) * cos(radians(lat)) *
				cos(radians(lng) - radians(?)) +
				sin(radians(?)) * sin(radians(lat))
			)
		) AS distance
		FROM stores
		WHERE is_active = true
		HAVING distance < ?
		ORDER BY distance
	`
	err := r.db.Raw(query, lat, lng, lat, radius).Scan(&stores).Error
	return stores, err
}

// Store Inventory operations
func (r *Repository) AddInventoryItem(inventory *StoreInventory) error {
	return r.db.Create(inventory).Error
}

func (r *Repository) GetStoreInventory(storeID uint) ([]StoreInventory, error) {
	var inventory []StoreInventory
	err := r.db.Where("store_id = ?", storeID).Find(&inventory).Error
	return inventory, err
}

func (r *Repository) GetInventoryItem(storeID, productID uint) (*StoreInventory, error) {
	var inventory StoreInventory
	err := r.db.Where("store_id = ? AND product_id = ?", storeID, productID).First(&inventory).Error
	if err != nil {
		return nil, err
	}
	return &inventory, nil
}

func (r *Repository) UpdateInventoryItem(inventory *StoreInventory) error {
	return r.db.Save(inventory).Error
}

func (r *Repository) DeleteInventoryItem(storeID, productID uint) error {
	return r.db.Where("store_id = ? AND product_id = ?", storeID, productID).Delete(&StoreInventory{}).Error
}

func (r *Repository) GetLowStockItems(storeID uint) ([]StoreInventory, error) {
	var inventory []StoreInventory
	query := r.db.Where("store_id = ? AND quantity <= min_stock AND is_active = true", storeID)
	err := query.Find(&inventory).Error
	return inventory, err
}

// Store Hours operations
func (r *Repository) SetStoreHours(storeID uint, hours []StoreHours) error {
	// Delete existing hours
	if err := r.db.Where("store_id = ?", storeID).Delete(&StoreHours{}).Error; err != nil {
		return err
	}

	// Create new hours
	for i := range hours {
		hours[i].StoreID = storeID
	}
	return r.db.Create(&hours).Error
}

func (r *Repository) GetStoreHours(storeID uint) ([]StoreHours, error) {
	var hours []StoreHours
	err := r.db.Where("store_id = ?", storeID).Order("day_of_week").Find(&hours).Error
	return hours, err
}

// Store Manager operations
func (r *Repository) AssignManager(manager *StoreManager) error {
	return r.db.Create(manager).Error
}

func (r *Repository) GetStoreManagers(storeID uint) ([]StoreManager, error) {
	var managers []StoreManager
	err := r.db.Where("store_id = ? AND is_active = true", storeID).Find(&managers).Error
	return managers, err
}

func (r *Repository) RemoveManager(storeID, userID uint) error {
	return r.db.Where("store_id = ? AND user_id = ?", storeID, userID).Delete(&StoreManager{}).Error
}

// Statistics
func (r *Repository) GetStoreStats() (*StoreStatsResponse, error) {
	stats := &StoreStatsResponse{}

	// Total stores
	r.db.Model(&Store{}).Count(&stats.TotalStores)

	// Active stores
	r.db.Model(&Store{}).Where("is_active = true").Count(&stats.ActiveStores)

	// Open stores
	r.db.Model(&Store{}).Where("is_active = true AND is_open = true").Count(&stats.OpenStores)

	// Total inventory items
	r.db.Model(&StoreInventory{}).Where("is_active = true").Count(&stats.TotalInventory)

	// Low stock items
	r.db.Model(&StoreInventory{}).Where("quantity <= min_stock AND is_active = true").Count(&stats.LowStockItems)

	return stats, nil
}
