package main

import (
	"errandShop/config"
	"errandShop/internal/database"
	"errandShop/internal/domain/auth"
	"log"
)

func main() {
	log.Println("🔧 Creating users table with UUID support...")

	// Load configuration
	cfg := config.LoadConfig()

	// Connect to database
	db := database.ConnectDB(cfg.DatabaseUrl)

	log.Println("✅ Database connection established")

	// Create users table directly
	err := db.AutoMigrate(
		&auth.User{},
		&auth.OTP{},
		&auth.Address{},
		&auth.RefreshToken{},
	)
	if err != nil {
		log.Fatalf("❌ Failed to create tables: %v", err)
	}

	log.Println("✅ Users table created successfully with UUID support!")
	log.Println("🔄 You can now restart the server and seed users.")
}