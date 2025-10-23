package products

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
)

type Service struct {
	repo   *Repository
	logger *log.Logger
}

func NewService(r *Repository) *Service {
	return &Service{
		repo:   r,
		logger: log.New(log.Writer(), "[PRODUCTS] ", log.LstdFlags|log.Lshortfile),
	}
}

// Types are defined in dto.go

func (s *Service) List(ctx context.Context, q ListQuery) (ListResult, error) {
	s.logger.Printf("Listing products with query: %+v", q)

	items, total, err := s.repo.List(ctx, q)
	if err != nil {
		s.logger.Printf("Error listing products: %v", err)
		return ListResult{}, fmt.Errorf("failed to list products: %w", err)
	}

	if q.Page <= 0 {
		q.Page = 1
	}
	if q.Limit <= 0 {
		q.Limit = 20
	}

	totalPages := (int(total) + q.Limit - 1) / q.Limit

	s.logger.Printf("Successfully listed %d products (page %d/%d)", len(items), q.Page, totalPages)

	// Convert to ProductResponse slice
	responses := make([]ProductResponse, len(items))
	for i, item := range items {
		responses[i] = *s.toProductResponse(&item)
	}

	return ListResult{
		Data: responses,
		Meta: PageMeta{Page: q.Page, Limit: q.Limit, Total: int(total), TotalPages: totalPages},
	}, nil
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (*ProductResponse, error) {
	s.logger.Printf("Getting product with ID: %s", id.String())

	if id == uuid.Nil {
		s.logger.Printf("Invalid product ID: %s", id.String())
		return nil, errors.New("invalid product ID")
	}

	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Printf("Error getting product %s: %v", id.String(), err)
		return nil, fmt.Errorf("product not found: %w", err)
	}

	s.logger.Printf("Successfully retrieved product: %s (ID: %s)", product.Name, product.ID.String())
	return s.toProductResponse(product), nil
}

func (s *Service) GetBySKU(ctx context.Context, sku string) (*ProductResponse, error) {
	s.logger.Printf("Getting product with SKU: %s", sku)

	if strings.TrimSpace(sku) == "" {
		s.logger.Printf("Invalid product SKU: %s", sku)
		return nil, errors.New("invalid product SKU")
	}

	product, err := s.repo.GetBySKU(ctx, sku)
	if err != nil {
		s.logger.Printf("Error getting product %s: %v", sku, err)
		return nil, fmt.Errorf("product not found: %w", err)
	}

	s.logger.Printf("Successfully retrieved product: %s (SKU: %s)", product.Name, product.SKU)
	return s.toProductResponse(product), nil
}

func (s *Service) Create(ctx context.Context, req CreateProductRequest) (*ProductResponse, error) {
	s.logger.Printf("Creating new product: %s", req.Name)

	// Validate required fields
	if strings.TrimSpace(req.Name) == "" {
		return nil, errors.New("product name is required")
	}
	if strings.TrimSpace(req.Category) == "" {
		return nil, errors.New("product category is required")
	}
	if req.CostPrice <= 0 {
		return nil, errors.New("cost price must be greater than 0")
	}
	if req.SellingPrice <= 0 {
		return nil, errors.New("selling price must be greater than 0")
	}
	if req.StockQuantity < 0 {
		return nil, errors.New("stock quantity cannot be negative")
	}

	product := &Product{
		Name:              strings.TrimSpace(req.Name),
		Slug:              generateSlug(req.Name),
		Description:       strings.TrimSpace(req.Description),
		CostPrice:         req.CostPrice,
		SellingPrice:      req.SellingPrice,
		StockQuantity:     req.StockQuantity,
		ImageURL:          strings.TrimSpace(req.ImageURL),
		ImagePublicID:     req.ImagePublicID,
		Category:          strings.TrimSpace(req.Category),
		Tags:              req.Tags,
		LowStockThreshold: req.LowStockThreshold,
		IsActive:          true,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := s.repo.Create(ctx, product); err != nil {
		s.logger.Printf("Error creating product %s: %v", req.Name, err)
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	s.logger.Printf("Successfully created product: %s (ID: %s)", product.Name, product.ID.String())
	return s.toProductResponse(product), nil
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, req UpdateProductRequest) (*ProductResponse, error) {
	s.logger.Printf("Updating product ID: %s", id.String())

	if id == uuid.Nil {
		return nil, errors.New("invalid product ID")
	}

	// Check if product exists
	existingProduct, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Printf("Product %s not found for update: %v", id.String(), err)
		return nil, fmt.Errorf("product not found: %w", err)
	}

	updates := make(map[string]interface{})

	if req.Name != nil && strings.TrimSpace(*req.Name) != "" {
		updates["name"] = strings.TrimSpace(*req.Name)
		updates["slug"] = generateSlug(*req.Name)
	}
	if req.Description != nil {
		updates["description"] = strings.TrimSpace(*req.Description)
	}
	if req.CostPrice != nil {
		if *req.CostPrice <= 0 {
			return nil, errors.New("cost price must be greater than 0")
		}
		updates["cost_price"] = *req.CostPrice
	}
	if req.SellingPrice != nil {
		if *req.SellingPrice <= 0 {
			return nil, errors.New("selling price must be greater than 0")
		}
		updates["selling_price"] = *req.SellingPrice
	}
	if req.StockQuantity != nil {
		if *req.StockQuantity < 0 {
			return nil, errors.New("stock quantity cannot be negative")
		}
		updates["stock_quantity"] = *req.StockQuantity
	}
	if req.ImageURL != nil {
		updates["image_url"] = strings.TrimSpace(*req.ImageURL)
	}
	if req.ImagePublicID != nil {
		updates["image_public_id"] = *req.ImagePublicID
	}
	if req.Category != nil && strings.TrimSpace(*req.Category) != "" {
		updates["category"] = strings.TrimSpace(*req.Category)
	}
	if req.Tags != nil {
		updates["tags"] = *req.Tags
	}
	if req.LowStockThreshold != nil {
		if *req.LowStockThreshold < 0 {
			return nil, errors.New("low stock threshold cannot be negative")
		}
		updates["low_stock_threshold"] = *req.LowStockThreshold
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	if len(updates) == 0 {
		s.logger.Printf("No fields to update for product %s", id.String())
		return nil, errors.New("no fields to update")
	}

	updates["updated_at"] = time.Now()

	if err := s.repo.Update(ctx, id, updates); err != nil {
		s.logger.Printf("Error updating product %s: %v", id.String(), err)
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	updatedProduct, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Printf("Error retrieving updated product %s: %v", id.String(), err)
		return nil, fmt.Errorf("failed to retrieve updated product: %w", err)
	}

	s.logger.Printf("Successfully updated product: %s (ID: %s)", existingProduct.Name, id.String())
	return s.toProductResponse(updatedProduct), nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	s.logger.Printf("Deleting product ID: %s", id.String())

	if id == uuid.Nil {
		return errors.New("invalid product ID")
	}

	// Check if product exists
	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Printf("Product %s not found for deletion: %v", id.String(), err)
		return fmt.Errorf("product not found: %w", err)
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Printf("Error deleting product %s: %v", id.String(), err)
		return fmt.Errorf("failed to delete product: %w", err)
	}

	s.logger.Printf("Successfully deleted product: %s (ID: %s)", product.Name, id.String())
	return nil
}

func (s *Service) AdminList(ctx context.Context, query AdminListQuery) (*ListResult, error) {
	s.logger.Printf("Admin listing products with query: %+v", query)

	if query.Page <= 0 {
		query.Page = 1
	}
	if query.Limit <= 0 || query.Limit > 100 {
		query.Limit = 20
	}

	products, total, err := s.repo.AdminList(ctx, query)
	if err != nil {
		s.logger.Printf("Error admin listing products: %v", err)
		return nil, fmt.Errorf("failed to list products: %w", err)
	}

	responses := make([]ProductResponse, len(products))
	for i, product := range products {
		responses[i] = *s.toProductResponse(&product)
	}

	totalPages := int((total + int64(query.Limit) - 1) / int64(query.Limit))

	s.logger.Printf("Successfully admin listed %d products (page %d/%d)", len(products), query.Page, totalPages)

	return &ListResult{
		Data: responses,
		Meta: PageMeta{
			Page:       query.Page,
			Limit:      query.Limit,
			Total:      int(total),
			TotalPages: totalPages,
		},
	}, nil
}

// Stock Management Methods
func (s *Service) UpdateStock(ctx context.Context, productID uuid.UUID, req StockUpdateRequest, userID uuid.UUID) (*StockUpdateResponse, error) {
	s.logger.Printf("Updating stock for product %s", productID.String())

	if productID == uuid.Nil {
		return nil, errors.New("invalid product ID")
	}
	if req.Quantity <= 0 {
		return nil, errors.New("quantity must be greater than 0")
	}
	if req.ChangeType == "" {
		return nil, errors.New("change type is required")
	}

	response, err := s.repo.UpdateStock(ctx, productID, req, userID)
	if err != nil {
		s.logger.Printf("Error updating stock for product %s: %v", productID.String(), err)
		return nil, fmt.Errorf("failed to update stock: %w", err)
	}

	s.logger.Printf("Successfully updated stock for product %s", productID.String())
	return response, nil
}

func (s *Service) BulkUpdateStock(ctx context.Context, req BulkUpdateStockRequest, userID uuid.UUID) error {
	s.logger.Printf("Bulk updating stock for %d products", len(req.Updates))

	if len(req.Updates) == 0 {
		return errors.New("no stock updates provided")
	}

	// Validate all updates first
	for i, update := range req.Updates {
		if update.ProductID == uuid.Nil {
			return fmt.Errorf("invalid product ID at index %d", i)
		}
		if update.Quantity <= 0 {
			return fmt.Errorf("quantity must be greater than 0 at index %d", i)
		}
		if update.ChangeType == "" {
			return fmt.Errorf("change type is required at index %d", i)
		}
	}

	if err := s.repo.BulkUpdateStock(ctx, req, userID); err != nil {
		s.logger.Printf("Error bulk updating stock: %v", err)
		return fmt.Errorf("failed to bulk update stock: %w", err)
	}

	s.logger.Printf("Successfully bulk updated stock for %d products", len(req.Updates))
	return nil
}

func (s *Service) GetLowStockProducts(ctx context.Context, limit int) ([]ProductResponse, error) {
	s.logger.Printf("Getting low stock products")

	products, err := s.repo.GetLowStockProducts(ctx, limit)
	if err != nil {
		s.logger.Printf("Error getting low stock products: %v", err)
		return nil, fmt.Errorf("failed to get low stock products: %w", err)
	}

	responses := make([]ProductResponse, len(products))
	for i, product := range products {
		responses[i] = *s.toProductResponse(&product)
	}

	s.logger.Printf("Successfully retrieved %d low stock products", len(products))
	return responses, nil
}

func (s *Service) GetOutOfStockProducts(ctx context.Context, limit int) ([]ProductResponse, error) {
	s.logger.Printf("Getting out of stock products")

	products, err := s.repo.GetOutOfStockProducts(ctx, limit)
	if err != nil {
		s.logger.Printf("Error getting out of stock products: %v", err)
		return nil, fmt.Errorf("failed to get out of stock products: %w", err)
	}

	responses := make([]ProductResponse, len(products))
	for i, product := range products {
		responses[i] = *s.toProductResponse(&product)
	}

	s.logger.Printf("Successfully retrieved %d out of stock products", len(products))
	return responses, nil
}

// Category Management
func (s *Service) GetCategories(ctx context.Context) ([]CategoryResponse, error) {
	s.logger.Printf("Getting product categories")

	categories, err := s.repo.GetCategories(ctx)
	if err != nil {
		s.logger.Printf("Error getting categories: %v", err)
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}

	s.logger.Printf("Successfully retrieved %d categories", len(categories))
	return categories, nil
}

func (s *Service) CreateCategory(ctx context.Context, category *Category) error {
	s.logger.Printf("Creating category: %s", category.Name)

	name := strings.TrimSpace(category.Name)
	if name == "" {
		return errors.New("category name is required")
	}
	category.Name = name

	// Set portable ID and timestamps to avoid DB-specific defaults
	if category.ID == uuid.Nil {
		category.ID = uuid.New()
	}
	now := time.Now()
	if category.CreatedAt.IsZero() {
		category.CreatedAt = now
	}
	category.UpdatedAt = now
	if !category.IsActive {
		category.IsActive = true
	}

	if err := s.repo.CreateCategory(ctx, category); err != nil {
		s.logger.Printf("Error creating category: %v", err)
		return fmt.Errorf("failed to create category: %w", err)
	}

	s.logger.Printf("Successfully created category: %s", category.Name)
	return nil
}

func (s *Service) ListCategories(ctx context.Context) ([]CategoryResponse, error) {
	s.logger.Printf("Listing categories")
	cats, err := s.repo.ListCategories(ctx)
	if err != nil {
		s.logger.Printf("Error listing categories: %v", err)
		return nil, fmt.Errorf("failed to list categories: %w", err)
	}
	return cats, nil
}

func (s *Service) GetCategory(ctx context.Context, id uuid.UUID) (*Category, error) {
	if id == uuid.Nil {
		return nil, errors.New("invalid category ID")
	}
	cat, err := s.repo.GetCategoryByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("category not found: %w", err)
	}
	return cat, nil
}

func (s *Service) UpdateCategory(ctx context.Context, id uuid.UUID, req UpdateCategoryRequest) (*Category, error) {
	if id == uuid.Nil {
		return nil, errors.New("invalid category ID")
	}
	_, err := s.repo.GetCategoryByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("category not found: %w", err)
	}
	updates := map[string]interface{}{}
	if req.Name != nil && strings.TrimSpace(*req.Name) != "" {
		updates["name"] = strings.TrimSpace(*req.Name)
	}
	if req.Description != nil {
		updates["description"] = strings.TrimSpace(*req.Description)
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if len(updates) == 0 {
		return nil, errors.New("no fields to update")
	}
	updates["updated_at"] = time.Now()
	if err := s.repo.UpdateCategory(ctx, id, updates); err != nil {
		return nil, fmt.Errorf("failed to update category: %w", err)
	}
	cat, err := s.repo.GetCategoryByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch updated category: %w", err)
	}
	return cat, nil
}

func (s *Service) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return errors.New("invalid category ID")
	}
	// ensure exists
	_, err := s.repo.GetCategoryByID(ctx, id)
	if err != nil {
		return fmt.Errorf("category not found: %w", err)
	}
	if err := s.repo.DeleteCategory(ctx, id); err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}
	return nil
}
// Analytics
func (s *Service) GetProductAnalytics(ctx context.Context) (*ProductAnalytics, error) {
	s.logger.Printf("Getting product analytics")

	analytics, err := s.repo.GetProductAnalytics(ctx)
	if err != nil {
		s.logger.Printf("Error getting product analytics: %v", err)
		return nil, fmt.Errorf("failed to get product analytics: %w", err)
	}

	s.logger.Printf("Successfully retrieved product analytics")
	return analytics, nil
}

func (s *Service) GetStockHistory(ctx context.Context, productID uuid.UUID, limit int) ([]StockHistory, error) {
	s.logger.Printf("Getting stock history for product %s", productID.String())

	history, err := s.repo.GetStockHistory(ctx, productID, limit)
	if err != nil {
		s.logger.Printf("Error getting stock history: %v", err)
		return nil, fmt.Errorf("failed to get stock history: %w", err)
	}

	s.logger.Printf("Successfully retrieved %d stock history records", len(history))
	return history, nil
}

func (s *Service) toProductResponse(product *Product) *ProductResponse {
	return &ProductResponse{
		ID:                product.ID,
		SKU:               product.SKU,
		Name:              product.Name,
		Slug:              product.Slug,
		Description:       product.Description,
		CostPrice:         product.CostPrice,
		SellingPrice:      product.SellingPrice,
		Profit:            product.Profit(),
		StockQuantity:     product.StockQuantity,
		ImageURL:          product.ImageURL,
		ImagePublicID:     product.ImagePublicID,
		Category:          product.Category,
		Tags:              product.Tags,
		LowStockThreshold: product.LowStockThreshold,
		IsLowStock:        product.IsLowStock(),
		IsActive:          product.IsActive,
		CreatedAt:         product.CreatedAt,
		UpdatedAt:         product.UpdatedAt,
	}
}

func generateSlug(name string) string {
	slug := strings.ToLower(strings.TrimSpace(name))
	slug = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return r
		}
		if unicode.IsSpace(r) || r == '-' || r == '_' {
			return '-'
		}
		return -1
	}, slug)

	// Remove multiple consecutive dashes
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}

	return strings.Trim(slug, "-")
}
