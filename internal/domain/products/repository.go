package products

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Product CRUD Operations
func (r *Repository) List(ctx context.Context, q ListQuery) (items []Product, total int64, err error) {
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.Limit <= 0 || q.Limit > 100 {
		q.Limit = 20
	}

	tx := r.db.WithContext(ctx).Model(&Product{}).Where(&Product{IsActive: true})
	if q.Q != "" {
		tx = tx.Where("name ILIKE ? OR category ILIKE ? OR description ILIKE ?", 
			"%"+q.Q+"%", "%"+q.Q+"%", "%"+q.Q+"%")
	}
	if err = tx.Count(&total).Error; err != nil {
		return
	}
	offset := (q.Page - 1) * q.Limit
	err = tx.Order("created_at DESC").Limit(q.Limit).Offset(offset).Find(&items).Error
	return
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*Product, error) {
	var p Product
	if err := r.db.WithContext(ctx).Where("id = ?", id).Where(&Product{IsActive: true}).First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repository) GetBySKU(ctx context.Context, sku string) (*Product, error) {
	var p Product
	if err := r.db.WithContext(ctx).Where("sku = ?", sku).Where(&Product{IsActive: true}).First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repository) Create(ctx context.Context, product *Product) error {
	return r.db.WithContext(ctx).Create(product).Error
}

func (r *Repository) Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	updates["updated_at"] = "NOW()"
	return r.db.WithContext(ctx).Model(&Product{}).Where("id = ?", id).Where(&Product{IsActive: true}).Updates(updates).Error
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&Product{}).Where("id = ?", id).Update("is_active", false).Error
}

// Admin List with Advanced Filtering
func (r *Repository) AdminList(ctx context.Context, query AdminListQuery) ([]Product, int64, error) {
	var products []Product
	var total int64

	db := r.db.WithContext(ctx).Model(&Product{}).Where(&Product{IsActive: true})

	// Apply filters
	if query.Q != "" || query.Search != "" {
		searchTerm := query.Q
		if query.Search != "" {
			searchTerm = query.Search
		}
		db = db.Where("name ILIKE ? OR description ILIKE ? OR category ILIKE ? OR sku ILIKE ?",
			"%"+searchTerm+"%", "%"+searchTerm+"%", "%"+searchTerm+"%", "%"+searchTerm+"%")
	}

	if query.Category != "" {
		db = db.Where("category = ?", query.Category)
	}

	if query.LowStock {
		db = db.Where("stock_quantity <= low_stock_threshold AND stock_quantity > 0")
	}

	if query.OutOfStock {
		db = db.Where("stock_quantity = 0")
	}

	// Count total
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	sortBy := "created_at"
	if query.SortBy != "" {
		sortBy = query.SortBy
		if sortBy == "price" {
			sortBy = "selling_price"
		} else if sortBy == "stock" {
			sortBy = "stock_quantity"
		}
	}
	sortOrder := "desc"
	if query.SortOrder != "" {
		sortOrder = query.SortOrder
	}
	db = db.Order(fmt.Sprintf("%s %s", sortBy, sortOrder))

	// Apply pagination
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.Limit <= 0 || query.Limit > 100 {
		query.Limit = 20
	}
	offset := (query.Page - 1) * query.Limit
	if err := db.Offset(offset).Limit(query.Limit).Find(&products).Error; err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

// Stock Management
func (r *Repository) UpdateStock(ctx context.Context, productID uuid.UUID, req StockUpdateRequest, userID uuid.UUID) (*StockUpdateResponse, error) {
	tx := r.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get current product
	var product Product
	if err := tx.Where("id = ?", productID).Where(&Product{IsActive: true}).First(&product).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	previousQuantity := product.StockQuantity
	var newQuantity int

	// Calculate new quantity based on change type
	switch req.ChangeType {
	case "ADD":
		newQuantity = previousQuantity + req.Quantity
	case "REMOVE":
		newQuantity = previousQuantity - req.Quantity
		if newQuantity < 0 {
			newQuantity = 0
		}
	case "ADJUST":
		newQuantity = req.Quantity
	default:
		tx.Rollback()
		return nil, fmt.Errorf("invalid change type: %s", req.ChangeType)
	}

	// Update product stock
	if err := tx.Model(&product).Update("stock_quantity", newQuantity).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Create stock history record
	stockHistory := StockHistory{
		ProductID:        productID,
		ChangeType:       req.ChangeType,
		QuantityChange:   req.Quantity,
		PreviousQuantity: previousQuantity,
		NewQuantity:      newQuantity,
		Reason:           req.Reason,
		CreatedBy:        userID,
	}

	if err := tx.Create(&stockHistory).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &StockUpdateResponse{
		PreviousQuantity: previousQuantity,
		NewQuantity:      newQuantity,
		Change:           newQuantity - previousQuantity,
	}, nil
}

func (r *Repository) BulkUpdateStock(ctx context.Context, req BulkUpdateStockRequest, userID uuid.UUID) error {
	tx := r.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, update := range req.Updates {
		// Get current product
		var product Product
		if err := tx.Where("id = ?", update.ProductID).Where(&Product{IsActive: true}).First(&product).Error; err != nil {
			tx.Rollback()
			return err
		}

		previousQuantity := product.StockQuantity
		var newQuantity int

		// Calculate new quantity
		switch update.ChangeType {
		case "ADD":
			newQuantity = previousQuantity + update.Quantity
		case "REMOVE":
			newQuantity = previousQuantity - update.Quantity
			if newQuantity < 0 {
				newQuantity = 0
			}
		case "ADJUST":
			newQuantity = update.Quantity
		default:
			tx.Rollback()
			return fmt.Errorf("invalid change type: %s", update.ChangeType)
		}

		// Update product stock
		if err := tx.Model(&product).Update("stock_quantity", newQuantity).Error; err != nil {
			tx.Rollback()
			return err
		}

		// Create stock history record
		stockHistory := StockHistory{
			ProductID:        update.ProductID,
			ChangeType:       update.ChangeType,
			QuantityChange:   update.Quantity,
			PreviousQuantity: previousQuantity,
			NewQuantity:      newQuantity,
			Reason:           req.Reason,
			CreatedBy:        userID,
		}

		if err := tx.Create(&stockHistory).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

func (r *Repository) GetLowStockProducts(ctx context.Context, limit int) ([]Product, error) {
	var products []Product
	query := r.db.WithContext(ctx).Where("stock_quantity <= low_stock_threshold AND stock_quantity > 0").Where(&Product{IsActive: true}).Order("stock_quantity ASC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	return products, query.Find(&products).Error
}

func (r *Repository) GetOutOfStockProducts(ctx context.Context, limit int) ([]Product, error) {
	var products []Product
	query := r.db.WithContext(ctx).Where("stock_quantity = 0").Where(&Product{IsActive: true}).Order("updated_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	return products, query.Find(&products).Error
}

// Category Management
func (r *Repository) GetCategories(ctx context.Context) ([]CategoryResponse, error) {
	var categories []CategoryResponse
	err := r.db.WithContext(ctx).Model(&Product{}).
		Select("category as name, COUNT(*) as product_count").
		Where("category != ''").Where(&Product{IsActive: true}).
		Group("category").
		Order("product_count DESC").
		Scan(&categories).Error
	return categories, err
}

func (r *Repository) CreateCategory(ctx context.Context, category *Category) error {
	return r.db.WithContext(ctx).Create(category).Error
}

// Analytics
func (r *Repository) GetProductAnalytics(ctx context.Context) (*ProductAnalytics, error) {
	var analytics ProductAnalytics

	// Use int64 variables for Count operations
	var totalProducts, lowStockCount, outOfStockCount int64

	// Get total products
	r.db.WithContext(ctx).Model(&Product{}).Where(&Product{IsActive: true}).Count(&totalProducts)
	analytics.TotalProducts = int(totalProducts)

	// Get low stock count
	r.db.WithContext(ctx).Model(&Product{}).Where("stock_quantity <= low_stock_threshold AND stock_quantity > 0").Where(&Product{IsActive: true}).Count(&lowStockCount)
	analytics.LowStockCount = int(lowStockCount)

	// Get out of stock count
	r.db.WithContext(ctx).Model(&Product{}).Where("stock_quantity = 0").Where(&Product{IsActive: true}).Count(&outOfStockCount)
	analytics.OutOfStockCount = int(outOfStockCount)

	// Get total value and average price
	var result struct {
		TotalValue   float64 `json:"total_value"`
		AveragePrice float64 `json:"average_price"`
	}
	r.db.WithContext(ctx).Model(&Product{}).
		Select("SUM(selling_price * stock_quantity) as total_value, AVG(selling_price) as average_price").
		Where(&Product{IsActive: true}).
		Scan(&result)

	analytics.TotalValue = result.TotalValue
	analytics.AveragePrice = result.AveragePrice

	// Get top categories
	var categoryStats []CategoryStats
	r.db.WithContext(ctx).Model(&Product{}).
		Select("category, COUNT(*) as product_count, SUM(selling_price * stock_quantity) as total_value").
		Where("category != ''").Where(&Product{IsActive: true}).
		Group("category").
		Order("product_count DESC").
		Limit(5).
		Scan(&categoryStats)

	analytics.TopCategories = categoryStats

	return &analytics, nil
}

// Stock History
func (r *Repository) GetStockHistory(ctx context.Context, productID uuid.UUID, limit int) ([]StockHistory, error) {
	var history []StockHistory
	query := r.db.WithContext(ctx).Where("product_id = ?", productID).Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	return history, query.Find(&history).Error
}
