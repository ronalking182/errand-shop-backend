package products_test

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	products "errandShop/internal/domain/products"
	v1 "errandShop/internal/transport/http/v1"
)

func setupTestApp(t *testing.T) (*fiber.App, *gorm.DB) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite in-memory db: %v", err)
	}
	// Create minimal schema manually to avoid Postgres-specific defaults in model tags
	if err := db.Exec(`
		CREATE TABLE categories (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			description TEXT,
			is_active BOOLEAN DEFAULT 1,
			created_at DATETIME,
			updated_at DATETIME
		);
	`).Error; err != nil {
		t.Fatalf("failed to create categories table: %v", err)
	}
	if err := db.Exec(`
		CREATE TABLE products (
			id TEXT PRIMARY KEY,
			name TEXT,
			category TEXT,
			is_active BOOLEAN DEFAULT 1,
			created_at DATETIME
		);
	`).Error; err != nil {
		t.Fatalf("failed to create products table: %v", err)
	}

	repo := products.NewRepository(db)
	svc := products.NewService(repo)
	h := products.NewHandler(svc)

	app := fiber.New()
	// Mount superadmin category routes without middleware for unit tests
	v1.MountSuperAdminCategoryRoutes(app, h)
	return app, db
}

func TestCreateCategory(t *testing.T) {
	app, _ := setupTestApp(t)

	body := `{"name":"Condiments","description":"Sauces, spices and dressings"}`
	req := httptest.NewRequest("POST", "/categories", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestListAndGetCategory(t *testing.T) {
	app, db := setupTestApp(t)
	// Seed two categories
	cats := []products.Category{
		{Name: "Oils", Description: "Cooking oils and sprays", IsActive: true},
		{Name: "Proteins", Description: "Meat, fish, eggs", IsActive: true},
	}
	for i := range cats {
		cats[i].ID = uuid.New()
	}
	if err := db.Create(&cats).Error; err != nil {
		t.Fatalf("seed failed: %v", err)
	}

	// List
	req := httptest.NewRequest("GET", "/categories", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("list request failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200 for list, got %d", resp.StatusCode)
	}
	var list struct{
		Success bool `json:"success"`
		Data    []products.CategoryResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		t.Fatalf("decode list failed: %v", err)
	}
	if len(list.Data) < 2 {
		t.Fatalf("expected at least 2 categories, got %d", len(list.Data))
	}

	// Get by ID (use first seeded)
	id := cats[0].ID.String()
	req = httptest.NewRequest("GET", "/categories/"+id, nil)
	resp, err = app.Test(req, -1)
	if err != nil {
		t.Fatalf("get request failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200 for get, got %d", resp.StatusCode)
	}
	var getRes struct{
		Success bool `json:"success"`
		Data    products.Category `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&getRes); err != nil {
		t.Fatalf("decode get failed: %v", err)
	}
	if getRes.Data.ID != cats[0].ID {
		t.Fatalf("expected ID %s, got %s", cats[0].ID, getRes.Data.ID)
	}
}

func TestUpdateAndDeleteCategory(t *testing.T) {
	app, db := setupTestApp(t)
	// Seed
	c := products.Category{ID: uuid.New(), Name: "Grains and Staples", Description: "Rice, pasta", IsActive: true}
	if err := db.Create(&c).Error; err != nil {
		t.Fatalf("seed failed: %v", err)
	}

	// Update name
	body := `{"name":"Grains & Staples"}`
	req := httptest.NewRequest("PUT", "/categories/"+c.ID.String(), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("update request failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200 for update, got %d", resp.StatusCode)
	}

	// Delete
	req = httptest.NewRequest("DELETE", "/categories/"+c.ID.String(), nil)
	resp, err = app.Test(req, -1)
	if err != nil {
		t.Fatalf("delete request failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200 for delete, got %d", resp.StatusCode)
	}
}