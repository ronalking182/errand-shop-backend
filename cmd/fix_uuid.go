package main

import (
	"errandShop/config"
	"errandShop/internal/database"
	"log"
)

func main() {
	log.Println("🔧 Fixing UUID issues by recreating tables...")

	cfg := config.LoadConfig()
	db := database.ConnectDB(cfg.DatabaseUrl)

	// Drop all auth-related tables to start fresh
	tablesToDrop := []string{
		"refresh_tokens",
		"addresses", 
		"otps",
		"users",
	}

	for _, table := range tablesToDrop {
		log.Printf("🗑️  Dropping table: %s", table)
		if err := db.Exec("DROP TABLE IF EXISTS " + table + " CASCADE").Error; err != nil {
			log.Printf("⚠️  Warning dropping %s: %v", table, err)
		}
	}

	// Run migrations to recreate tables with proper UUIDs
	log.Println("🔄 Running migrations to recreate tables...")
	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("❌ Failed to run migrations: %v", err)
	}

	// Seed users again
	log.Println("🌱 Seeding users...")
	if err := database.SeedUsers(db); err != nil {
		log.Printf("⚠️  Warning seeding users: %v", err)
	}

	log.Println("✅ UUID fix completed successfully!")
	log.Println("🔄 You can now restart the server.")
}