package main

import (
	"errandShop/config"
	"errandShop/internal/database"
	"log"
)

func main() {
	log.Println("🌱 Manual User Seeding...")

	cfg := config.LoadConfig()
	db := database.ConnectDB(cfg.DatabaseUrl)

	if err := database.SeedUsers(db); err != nil {
		log.Fatalf("❌ Seeding failed: %v", err)
	}

	log.Println("✅ Seeding completed successfully!")
}
