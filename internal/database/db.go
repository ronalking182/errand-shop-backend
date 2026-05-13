package database

import (
	"database/sql"
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
	dsn = PostgresDSNWithPoolerCompat(dsn)
	log.Printf("ℹ️ postgres PreferSimpleProtocol=%v (set PG_USE_SIMPLE_PROTOCOL=false to disable)", PreferSimpleProtocolEnabled())

	// Configure GORM for optimal performance
	gormConfig := &gorm.Config{
		SkipDefaultTransaction: true,                                // Disable automatic transaction wrapping for individual operations
		CreateBatchSize:        1000,                                // Optimize bulk inserts
		Logger:                 logger.Default.LogMode(logger.Warn), // Only log warnings and errors
	}

	// PreferSimpleProtocol avoids prepared-statement clashes with poolers (e.g. PgBouncer / some cloud proxies).
	pgCfg := postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: PreferSimpleProtocolEnabled(),
	}

	// Establish database connection with retry logic
	for i := 0; i < 3; i++ {
		DB, err = gorm.Open(postgres.New(pgCfg), gormConfig)
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
	// DB INTROSPECTION (diagnostics)
	// =============================================
	// Log current database, user, and search_path to aid triage in production
	func() {
		// current_database(), current_user
		var currentDB, currentUser string
		row := DB.Raw("select current_database(), current_user").Row()
		if err := row.Scan(&currentDB, &currentUser); err == nil {
			log.Printf("ℹ️ Connected to DB='%s' as user='%s'", currentDB, currentUser)
		} else {
			log.Printf("⚠️ Failed to read current_database/current_user: %v", err)
		}

		// search_path
		var searchPath string
		row = DB.Raw("show search_path").Row()
		if err := row.Scan(&searchPath); err == nil {
			log.Printf("ℹ️ search_path='%s'", searchPath)
		} else {
			log.Printf("⚠️ Failed to read search_path: %v", err)
		}

		// Check orders table visibility (public and default schema resolution)
		var publicOrders, defaultOrders sql.NullString
		row = DB.Raw("select to_regclass('public.orders'), to_regclass('orders')").Row()
		if err := row.Scan(&publicOrders, &defaultOrders); err == nil {
			log.Printf("ℹ️ to_regclass public.orders='%v', orders='%v'", publicOrders.String, defaultOrders.String)
		} else {
			log.Printf("⚠️ Failed to check to_regclass for orders: %v", err)
		}

		// Check categories table visibility (public and default schema resolution)
		var publicCategories, defaultCategories sql.NullString
		row = DB.Raw("select to_regclass('public.categories'), to_regclass('categories')").Row()
		if err := row.Scan(&publicCategories, &defaultCategories); err == nil {
			log.Printf("ℹ️ to_regclass public.categories='%v', categories='%v'", publicCategories.String, defaultCategories.String)
		} else {
			log.Printf("⚠️ Failed to check to_regclass for categories: %v", err)
		}

		var publicCustomReq sql.NullString
		row = DB.Raw(`SELECT to_regclass('public.custom_requests')::text`).Row()
		if err := row.Scan(&publicCustomReq); err == nil {
			log.Printf(`ℹ️ to_regclass public.custom_requests=%q`, publicCustomReq.String)
		} else {
			log.Printf("⚠️ Failed to_regclass custom_requests: %v", err)
		}

		// go-gormigrate uses table name "migrations" by default
		var migCount int
		if err := DB.Raw("SELECT COUNT(*) FROM migrations").Scan(&migCount).Error; err == nil {
			log.Printf("ℹ️ migrations rows=%d", migCount)
		} else {
			log.Printf("ℹ️ migrations table not readable yet (normal on first boot): %v", err)
		}
	}()

	// =============================================
	// MIGRATION EXECUTION
	// =============================================

	// Run database migrations
	if err := RunMigrations(DB); err != nil {
		log.Fatal("Database migrations failed:", err)
	}

	if err := EnsureMinimalDashboardTables(DB); err != nil {
		log.Fatal("Failed to ensure dashboard schema:", err)
	}

	if err := RepairCustomRequestsIfMissing(DB); err != nil {
		log.Fatal("custom_requests verification/repair failed:", err)
	}

	// Confirm orders resolved after migrations + ensure step
	func() {
		var publicOrders sql.NullString
		row := DB.Raw("SELECT to_regclass('public.orders')::text").Row()
		if err := row.Scan(&publicOrders); err == nil {
			log.Printf("ℹ️ after migrate+ensure: public.orders=%q", publicOrders.String)
		}
	}()
	func() {
		var cr sql.NullString
		row := DB.Raw(`SELECT to_regclass('public.custom_requests')::text`).Row()
		if err := row.Scan(&cr); err == nil {
			log.Printf("ℹ️ after migrate+ensure+repair: public.custom_requests=%q", cr.String)
		}
	}()

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
