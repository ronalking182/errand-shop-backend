package products

import (
	"errors"
	"log"
	"strconv"
	"strings"

	"errandShop/internal/services/validation"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Handler struct {
	svc           *Service
	urlValidator  *validation.URLValidator
	logger        *log.Logger
}

func NewHandler(s *Service) *Handler {
	return &Handler{
		svc:          s,
		urlValidator: validation.NewURLValidator(),
		logger:       log.New(log.Writer(), "[PRODUCTS_HANDLER] ", log.LstdFlags|log.Lshortfile),
	}
}

var validate = validator.New()

// Error response helper
func (h *Handler) errorResponse(c *fiber.Ctx, statusCode int, message string, err error) error {
	h.logger.Printf("Error [%d]: %s - %v", statusCode, message, err)
	return c.Status(statusCode).JSON(fiber.Map{
		"success": false,
		"error":   message,
		"details": err.Error(),
	})
}

// Success response helper
func (h *Handler) successResponse(c *fiber.Ctx, data interface{}, message string) error {
	response := fiber.Map{
		"success": true,
		"data":    data,
	}
	if message != "" {
		response["message"] = message
	}
	return c.JSON(response)
}

func (h *Handler) List(c *fiber.Ctx) error {
	q := ListQuery{
		Q:     strings.TrimSpace(c.Query("q")),
		Page:  atoiDefault(c.Query("page"), 1),
		Limit: atoiDefault(c.Query("limit"), 20),
	}

	// Validate pagination
	if q.Page < 1 {
		q.Page = 1
	}
	if q.Limit < 1 || q.Limit > 100 {
		q.Limit = 20
	}

	res, err := h.svc.List(c.Context(), q)
	if err != nil {
		return h.errorResponse(c, fiber.StatusInternalServerError, "Failed to list products", err)
	}

	// Debug logging for response structure
	h.logger.Printf("[PRODUCTS-DEBUG] Response structure: %+v", res)
	if res.Data != nil {
		h.logger.Printf("[PRODUCTS-DEBUG] Number of products: %d", len(res.Data))
		if len(res.Data) > 0 {
			h.logger.Printf("[PRODUCTS-DEBUG] First product sample: %+v", res.Data[0])
		}
	}

	return h.successResponse(c, res, "Products retrieved successfully")
}

func (h *Handler) Get(c *fiber.Ctx) error {
	idStr := c.Params("id")
	if idStr == "" {
		return h.errorResponse(c, fiber.StatusBadRequest, "Product ID is required", errors.New("missing id parameter"))
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "Invalid product ID format", err)
	}

	product, err := h.svc.Get(c.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return h.errorResponse(c, fiber.StatusNotFound, "Product not found", err)
		}
		return h.errorResponse(c, fiber.StatusInternalServerError, "Failed to get product", err)
	}

	return h.successResponse(c, product, "")
}

func (h *Handler) GetBySKU(c *fiber.Ctx) error {
	sku := c.Params("sku")
	if sku == "" {
		return h.errorResponse(c, fiber.StatusBadRequest, "Product SKU is required", errors.New("missing sku parameter"))
	}

	product, err := h.svc.GetBySKU(c.Context(), sku)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return h.errorResponse(c, fiber.StatusNotFound, "Product not found", err)
		}
		return h.errorResponse(c, fiber.StatusInternalServerError, "Failed to get product", err)
	}

	return h.successResponse(c, product, "")
}

func (h *Handler) Create(c *fiber.Ctx) error {
	var req CreateProductRequest

	if err := c.BodyParser(&req); err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "Invalid request body", err)
	}

	if err := validate.Struct(&req); err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "Validation failed", err)
	}

	// Validate image URL if provided
	if req.ImageURL != "" {
		if err := h.urlValidator.ValidateImageURL(req.ImageURL); err != nil {
			return h.errorResponse(c, fiber.StatusBadRequest, "Invalid image URL", err)
		}
	}

	product, err := h.svc.Create(c.Context(), req)
	if err != nil {
		return h.errorResponse(c, fiber.StatusInternalServerError, "Failed to create product", err)
	}

	return h.successResponse(c, product, "Product created successfully")
}

func (h *Handler) Update(c *fiber.Ctx) error {
	idStr := c.Params("id")
	if idStr == "" {
		return h.errorResponse(c, fiber.StatusBadRequest, "Product ID is required", errors.New("missing id parameter"))
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "Invalid product ID format", err)
	}

	var req UpdateProductRequest
	if err := c.BodyParser(&req); err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "Invalid request body", err)
	}

	if err := validate.Struct(&req); err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "Validation failed", err)
	}

	// Validate image URL if provided
	if req.ImageURL != nil && *req.ImageURL != "" {
		if err := h.urlValidator.ValidateImageURL(*req.ImageURL); err != nil {
			return h.errorResponse(c, fiber.StatusBadRequest, "Invalid image URL", err)
		}
	}

	product, err := h.svc.Update(c.Context(), id, req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || strings.Contains(err.Error(), "not found") {
			return h.errorResponse(c, fiber.StatusNotFound, "Product not found", err)
		}
		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "required") {
			return h.errorResponse(c, fiber.StatusBadRequest, "Invalid product data", err)
		}
		return h.errorResponse(c, fiber.StatusInternalServerError, "Failed to update product", err)
	}

	h.logger.Printf("Product updated successfully: %s (ID: %s)", product.Name, product.ID.String())
	return h.successResponse(c, product, "Product updated successfully")
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	idStr := c.Params("id")
	if idStr == "" {
		return h.errorResponse(c, fiber.StatusBadRequest, "Product ID is required", errors.New("missing id parameter"))
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "Invalid product ID format", err)
	}

	if err := h.svc.Delete(c.Context(), id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || strings.Contains(err.Error(), "not found") {
			return h.errorResponse(c, fiber.StatusNotFound, "Product not found", err)
		}
		return h.errorResponse(c, fiber.StatusInternalServerError, "Failed to delete product", err)
	}

	h.logger.Printf("Product deleted successfully (ID: %s)", id.String())
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Product deleted successfully",
	})
}

func (h *Handler) AdminList(c *fiber.Ctx) error {
	var query AdminListQuery
	if err := c.QueryParser(&query); err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "Invalid query parameters", err)
	}

	// Set defaults and validate
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.Limit <= 0 || query.Limit > 100 {
		query.Limit = 20
	}

	result, err := h.svc.AdminList(c.Context(), query)
	if err != nil {
		return h.errorResponse(c, fiber.StatusInternalServerError, "Failed to list products", err)
	}

	// Debug logging for admin response structure
	h.logger.Printf("[PRODUCTS-ADMIN-DEBUG] Admin response structure: %+v", result)
	if result.Data != nil {
		h.logger.Printf("[PRODUCTS-ADMIN-DEBUG] Number of products: %d", len(result.Data))
		if len(result.Data) > 0 {
			h.logger.Printf("[PRODUCTS-ADMIN-DEBUG] First product sample: %+v", result.Data[0])
		}
	}

	return h.successResponse(c, result, "Products retrieved successfully")
}

func (h *Handler) BulkUpdateStock(c *fiber.Ctx) error {
	var req BulkUpdateStockRequest
	if err := c.BodyParser(&req); err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "Invalid request body", err)
	}

	if err := validate.Struct(&req); err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "Validation failed", err)
	}

	// Extract user ID from context (assuming it's set by middleware)
	userID := uuid.New() // TODO: Get from JWT middleware context
	if err := h.svc.BulkUpdateStock(c.Context(), req, userID); err != nil {
		if strings.Contains(err.Error(), "invalid") {
			return h.errorResponse(c, fiber.StatusBadRequest, "Invalid stock update data", err)
		}
		return h.errorResponse(c, fiber.StatusInternalServerError, "Failed to update stock", err)
	}

	h.logger.Printf("Bulk stock update completed for %d products", len(req.Updates))
	return h.successResponse(c, nil, "Stock updated successfully")
}

func (h *Handler) GetLowStock(c *fiber.Ctx) error {
	limitStr := c.Query("limit", "50")
	limit := atoiDefault(limitStr, 50)

	products, err := h.svc.GetLowStockProducts(c.Context(), limit)
	if err != nil {
		return h.errorResponse(c, fiber.StatusInternalServerError, "Failed to get low stock products", err)
	}

	return h.successResponse(c, products, "")
}

func (h *Handler) GetCategories(c *fiber.Ctx) error {
	categories, err := h.svc.GetCategories(c.Context())
	if err != nil {
		return h.errorResponse(c, fiber.StatusInternalServerError, "Failed to get categories", err)
	}
	return h.successResponse(c, categories, "Categories retrieved successfully")
}

// Superadmin-only: Create Category
func (h *Handler) CreateCategory(c *fiber.Ctx) error {
	var req CreateCategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "Invalid request body", err)
	}
	if err := validate.Struct(req); err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "Validation failed", err)
	}
	cat := &Category{
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
		IsActive:    true,
	}
	if err := h.svc.CreateCategory(c.Context(), cat); err != nil {
		return h.errorResponse(c, fiber.StatusInternalServerError, "Failed to create category", err)
	}
	return h.successResponse(c, cat, "Category created successfully")
}

// Superadmin-only: List Categories (from categories table)
func (h *Handler) ListCategories(c *fiber.Ctx) error {
	cats, err := h.svc.ListCategories(c.Context())
	if err != nil {
		return h.errorResponse(c, fiber.StatusInternalServerError, "Failed to list categories", err)
	}
	return h.successResponse(c, cats, "Categories retrieved successfully")
}

// Superadmin-only: Get Category by ID
func (h *Handler) GetCategory(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "Invalid category ID", err)
	}
	cat, err := h.svc.GetCategory(c.Context(), id)
	if err != nil {
		return h.errorResponse(c, fiber.StatusNotFound, "Category not found", err)
	}
	return h.successResponse(c, cat, "Category retrieved successfully")
}

// Superadmin-only: Update Category
func (h *Handler) UpdateCategory(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "Invalid category ID", err)
	}
	var req UpdateCategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "Invalid request body", err)
	}
	if err := validate.Struct(req); err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "Validation failed", err)
	}
	cat, err := h.svc.UpdateCategory(c.Context(), id, req)
	if err != nil {
		return h.errorResponse(c, fiber.StatusInternalServerError, "Failed to update category", err)
	}
	return h.successResponse(c, cat, "Category updated successfully")
}

// Superadmin-only: Delete Category (soft delete)
func (h *Handler) DeleteCategory(c *fiber.Ctx) error {
    idStr := c.Params("id")
    if strings.TrimSpace(idStr) == "" {
        return h.errorResponse(c, fiber.StatusBadRequest, "Category ID is required", errors.New("missing id parameter"))
    }

    id, err := uuid.Parse(idStr)
    if err != nil {
        return h.errorResponse(c, fiber.StatusBadRequest, "Invalid category ID", err)
    }
    if id == uuid.Nil {
        return h.errorResponse(c, fiber.StatusBadRequest, "Invalid category ID", errors.New("uuid is nil"))
    }

    if err := h.svc.DeleteCategory(c.Context(), id); err != nil {
        // Map common error cases to appropriate status codes
        if errors.Is(err, gorm.ErrRecordNotFound) || strings.Contains(err.Error(), "not found") {
            return h.errorResponse(c, fiber.StatusNotFound, "Category not found", err)
        }
        if strings.Contains(strings.ToLower(err.Error()), "invalid category id") {
            return h.errorResponse(c, fiber.StatusBadRequest, "Invalid category ID", err)
        }
        return h.errorResponse(c, fiber.StatusInternalServerError, "Failed to delete category", err)
    }

    return h.successResponse(c, fiber.Map{"id": id}, "Category deleted successfully")
}

func atoiDefault(s string, d int) int {
	if s == "" {
		return d
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		return d
	}
	return i
}

// Image upload/delete methods removed - using external image hosting (Cloudinary)
