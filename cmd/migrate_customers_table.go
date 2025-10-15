package main

import (
	"errandShop/config"
	"errandShop/internal/database"
	"errandShop/internal/domain/customers"
	"fmt"
	"log"
	"os"
)

func main() {
	// Load config to get database URL
	cfg := config.LoadConfig()
	
	// Use default DSN if DATABASE env var is not set
	dsn := cfg.DatabaseUrl
	if dsn == "" {
		dsn = os.Getenv("DATABASE_URL")
	}
	if dsn == "" {
		// Default connection string for local development
		dsn = "host=localhost user=kingdavidabutanko dbname=errand_shop port=5432 sslmode=disable"
	}
	
	fmt.Printf("Using database DSN: %s\n", dsn)
	
	// Connect to database
	database.ConnectDB(dsn)
	db := database.DB

	if db == nil {
		log.Fatal("Failed to connect to database")
	}

	fmt.Println("Connected to database successfully")

	// Drop and recreate customers table with correct UUID schema
	fmt.Println("Dropping existing customers and addresses tables...")
	err := db.Exec("DROP TABLE IF EXISTS addresses CASCADE").Error
	if err != nil {
		log.Printf("Warning: Failed to drop addresses table: %v", err)
	}

	err = db.Exec("DROP TABLE IF EXISTS customers CASCADE").Error
	if err != nil {
		log.Printf("Warning: Failed to drop customers table: %v", err)
	}

	// Create tables with correct schema
	fmt.Println("Creating customers and addresses tables with UUID support...")
	err = db.AutoMigrate(&customers.Customer{}, &customers.Address{})
	if err != nil {
		log.Fatalf("Failed to migrate customers tables: %v", err)
	}

	fmt.Println("Customers table migration completed successfully!")
	fmt.Println("Tables created with proper UUID support for user_id fields")
}