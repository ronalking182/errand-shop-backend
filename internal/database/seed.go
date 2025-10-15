package database

import (
	"log"
	"time"

	"errandShop/internal/domain/auth"

	"gorm.io/gorm"
)

// SeedUsers creates both super admin and regular user
func SeedUsers(db *gorm.DB) error {
	log.Println("🌱 Seeding users...")

	// Create Super Admin for Dashboard
	if err := seedSuperAdmin(db); err != nil {
		return err
	}

	// Create Regular User for Mobile App
	if err := seedRegularUser(db); err != nil {
		return err
	}

	log.Println("✅ All users seeded successfully")
	return nil
}

// seedSuperAdmin creates a super admin user for dashboard access
func seedSuperAdmin(db *gorm.DB) error {
	log.Println("👑 Seeding super admin user...")

	// Check if super admin already exists
	var existing auth.User
	if err := db.Where("email = ?", "admin@errandshop.com").First(&existing).Error; err == nil {
		log.Println("⚠️  Super admin already exists. Skipping seeding.")
		return nil
	}

	admin := &auth.User{
		Email:      "admin@errandshop.com",
		Name:       "Super Admin",
		Phone:      "+1234567890",
		IsVerified: true,
		Role:       "superadmin", // Super admin role for dashboard
		Status:     "active",
		Password:   "Admin123!", // Will be hashed by BeforeCreate hook
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := db.Create(admin).Error; err != nil {
		log.Println("🔥 Failed to create super admin:", err)
		return err
	}

	log.Println("✅ Super admin created successfully")
	log.Println("📧 Email: admin@errandshop.com")
	log.Println("🔑 Password: Admin123!")
	return nil
}

// seedRegularUser creates a regular customer user for mobile app
func seedRegularUser(db *gorm.DB) error {
	log.Println("👤 Seeding regular user...")

	// Check if regular user already exists
	var existing auth.User
	if err := db.Where("email = ?", "user@errandshop.com").First(&existing).Error; err == nil {
		log.Println("⚠️  Regular user already exists. Skipping seeding.")
		return nil
	}

	user := &auth.User{
		Email:      "user@errandshop.com",
		Name:       "John Doe",
		Phone:      "+1987654321",
		IsVerified: true,
		Role:       "customer", // Regular customer role for mobile app
		Status:     "active",
		Password:   "User123!", // Will be hashed by BeforeCreate hook
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := db.Create(user).Error; err != nil {
		log.Println("🔥 Failed to create regular user:", err)
		return err
	}

	log.Println("✅ Regular user created successfully")
	log.Println("📧 Email: user@errandshop.com")
	log.Println("🔑 Password: User123!")
	return nil
}

// Legacy function - kept for backward compatibility
func SeedTestUser(db *gorm.DB) error {
	return SeedUsers(db)
}
