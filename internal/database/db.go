package database

import (
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB is the global database connection instance
var DB *gorm.DB

// ConnectDB establishes a connection to the PostgreSQL database and runs migrations
// Parameters:
//   - dsn: Data Source Name (e.g., "host=localhost user=postgres dbname=stackle port=5432 sslmode=disable")
//
// Returns:
//   - *gorm.DB: The database connection instance
func ConnectDB(dsn string) *gorm.DB {
	var err error

	// =============================================
	// DATABASE CONNECTION SETUP
	// =============================================

	// Configure GORM for optimal performance
	gormConfig := &gorm.Config{
		SkipDefaultTransaction: true,                                // Disable automatic transaction wrapping for individual operations
		CreateBatchSize:        1000,                                // Optimize bulk inserts
		Logger:                 logger.Default.LogMode(logger.Warn), // Only log warnings and errors
	}

	// Establish database connection with retry logic
	for i := 0; i < 3; i++ {
		DB, err = gorm.Open(postgres.Open(dsn), gormConfig)
		if err == nil {
			break
		}
		log.Printf("Database connection attempt %d failed: %v", i+1, err)
		time.Sleep(2 * time.Second) // Wait before retrying
	}
	if err != nil {
		log.Fatal("Failed to connect to database after retries:", err)
	}

	// =============================================
	// CONNECTION POOL CONFIGURATION
	// =============================================

	// Get underlying SQL DB instance for connection pooling
	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatal("Failed to get underlying DB instance:", err)
	}

	// Configure connection pool settings
	sqlDB.SetMaxIdleConns(10)           // Maximum idle connections
	sqlDB.SetMaxOpenConns(100)          // Maximum open connections
	sqlDB.SetConnMaxLifetime(time.Hour) // Maximum connection lifetime

	// =============================================
	// MIGRATION EXECUTION
	// =============================================

	// Run database migrations
	if err := RunMigrations(DB); err != nil {
		log.Fatal("Database migrations failed:", err)
	}

	log.Println("✅ Database connection established and migrations completed")
	return DB
}

// PingDB checks if the database connection is alive
func PingDB() error {
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

// CloseDB closes the database connection
func CloseDB() {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			log.Printf("Error getting SQL DB: %v", err)
			return
		}
		if err := sqlDB.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
		log.Println("Database connection closed")
	}
}
