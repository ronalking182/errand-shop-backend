package main

import (
	"errandShop/config"
	"errandShop/internal/database"
	"errandShop/internal/domain/analytics"
	"errandShop/internal/domain/auth"
	"errandShop/internal/domain/chat"
	"errandShop/internal/domain/coupons"
	"errandShop/internal/domain/custom_requests"
	"errandShop/internal/domain/customers"
	"errandShop/internal/domain/delivery"
	"errandShop/internal/domain/notifications"
	"errandShop/internal/domain/orders"
	"errandShop/internal/domain/payments"
	"errandShop/internal/domain/products"

	"errandShop/internal/middleware"
	"errandShop/internal/services/audit"
	"errandShop/internal/services/email"
	v1 "errandShop/internal/transport/http/v1"
	"fmt"
	"log"

	"errandShop/internal/http/handlers"
	"errandShop/internal/repos"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

// Create a temporary payments service interface for orders initialization
type tempPaymentService struct{}

func (t *tempPaymentService) InitializePayment(req payments.CreatePaymentRequest, customerID uint) (*payments.PaymentInitResponse, error) {
	return nil, fmt.Errorf("payment service not yet initialized")
}

func main() {
	// 🔧 Configuration Setup
	log.Println("🚀 Starting Errand Shop Backend...")
	cfg := config.LoadConfig() // ✅ Fixed: was config.Load()
	log.Println("✅ Configuration loaded successfully")

	// 🗄️ Database Connection & Migration
	log.Println("🔌 Connecting to database...")
	db := database.ConnectDB(cfg.DatabaseUrl) // ✅ Fixed: was database.Connect(cfg)
	log.Println("✅ Database connected successfully")

	// 🔄 Database migrations are handled automatically by ConnectDB
	log.Println("✅ Database migrations completed automatically")

	// 🌱 Seed initial users
	log.Println("🌱 Seeding initial users...")
	if err := database.SeedUsers(db); err != nil {
		log.Printf("⚠️ Seeding failed: %v", err)
	} else {
		log.Println("✅ User seeding completed")
	}

	// 🌐 Initialize Fiber Web Framework
	log.Println("🌐 Initializing Fiber app...")
	app := fiber.New(fiber.Config{
		// 🚨 Global Error Handler
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			log.Printf("❌ Error [%d]: %s", code, err.Error())
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// 🛡️ Middleware Setup
	log.Println("🛡️ Setting up middleware...")
	app.Use(logger.New())         // 📝 Request logging
	app.Use(recover.New())        // 🔄 Panic recovery
    // 🌍 CORS configuration for all /api/* routes
    app.Use("/api", cors.New(cors.Config{
        AllowOrigins:     "https://v0-errand-shop-dashboard.vercel.app,https://v0-errand-shop-dashboard-git-main-ronalking182s-projects.vercel.app,https://v0-errand-shop-dashboard-jcjvf4fer-ronalking182s-projects.vercel.app,http://localhost:5173",
        AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
        AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
        ExposeHeaders:    "Set-Cookie",
        AllowCredentials: true,
    }))
    // 🔒 Security headers and caching policy for /api/*
    app.Use("/api", middleware.APISecurityHeaders())
    // ✅ Ensure preflight (OPTIONS) returns 204 with CORS headers
    app.Options("/api/*", func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusNoContent) })
	log.Println("✅ Middleware configured")

	// 👥 Initialize Customers Domain (needed for auth service)
	log.Println("👥 Setting up customers domain...")
	customersRepo := customers.NewRepository(db)
	customersService := customers.NewService(customersRepo)
	log.Println("✅ Customers domain initialized")

	// 🔐 Initialize Authentication Domain
	log.Println("🔐 Setting up authentication domain...")
	authRepo := auth.NewRepository(db)

	// Initialize Email Service
	emailService := email.NewResendService(cfg.ResendAPIKey, cfg.FromEmail)

	// Initialize Audit Service
	auditService := audit.NewAuditService(db)

	// Update auth service initialization with customer service
	authService := auth.NewService(authRepo, cfg, emailService, auditService, customersService)

	// Add rate limiting
	app.Use("/api/v1/auth", middleware.AuthRateLimit())
	app.Use("/api/v1", middleware.APIRateLimit())
	authHandler := auth.NewHandler(authService)
	log.Println("✅ Authentication domain initialized")

	// 🛣️ API Routes Setup
	log.Println("🛣️ Setting up API routes...")
	api := app.Group("/api/v1")

	// Add base API info endpoint
	api.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "🚀 Errand Shop API v1",
			"status":  "active",
			"version": "1.0.0",
			"endpoints": fiber.Map{
				"auth":   "/api/v1/auth",
				"admin":  "/api/v1/admin",
				"health": "/health",
			},
		})
	})

	// 🌐 Public Authentication Routes (No JWT Required)
	log.Println("🌐 Configuring public auth routes...")
	authRoutes := api.Group("/auth")
	authRoutes.Post("/register", authHandler.Register)              // 📝 User registration
	authRoutes.Post("/login", authHandler.Login)                    // 🔑 User login
	authRoutes.Post("/verify-email", authHandler.VerifyEmail)       // ✉️ Email verification
	authRoutes.Post("/resend-otp", authHandler.ResendOTP)           // 🔄 Resend OTP
	authRoutes.Post("/refresh-token", authHandler.RefreshToken)     // 🔄 Token refresh
	authRoutes.Post("/forgot-password", authHandler.ForgotPassword) // 🔒 Password reset request
	authRoutes.Post("/reset-password", authHandler.ResetPassword)   // 🔓 Password reset

	// 🔒 Protected Authentication Routes (JWT Required)
	log.Println("🔒 Configuring protected auth routes...")
	protectedAuth := authRoutes.Group("", middleware.JWTMiddleware(cfg))
	protectedAuth.Post("/logout", authHandler.Logout)                  // 🚪 User logout
	protectedAuth.Get("/me", authHandler.Me)                           // 👤 Get current user info
	protectedAuth.Post("/password/change", authHandler.ChangePassword) // 🔑 Change password

	// 👑 Admin Routes (JWT + Admin Role Required)
	log.Println("👑 Configuring admin routes...")
	adminRoutes := api.Group("/admin", middleware.JWTMiddleware(cfg), middleware.AdminMiddleware())
	adminRoutes.Get("/users", authHandler.GetUsers)                              // 👥 List all users
	adminRoutes.Post("/users", authHandler.CreateUser)                           // ➕ Create new user
	adminRoutes.Get("/users/:id", authHandler.GetUserByID)                       // 👤 Get user by ID
	adminRoutes.Put("/users/:id", authHandler.UpdateUser)                        // ✏️ Update user
	adminRoutes.Delete("/users/:id", authHandler.DeleteUser)                     // 🗑️ Delete user
	adminRoutes.Patch("/users/:id/status", authHandler.UpdateUserStatus)         // 🔄 Update user status
	adminRoutes.Get("/permissions/available", authHandler.GetAvailablePermissions) // 📋 Get available permissions
	adminRoutes.Put("/users/:id/permissions", authHandler.UpdateUserPermissions)   // 🔐 Update permissions
	adminRoutes.Put("/users/:id/force-reset", authHandler.ForcePasswordReset)      // 🔒 Force password reset

	// 🔍 Admin-only DB introspection endpoint for incident diagnostics
	adminRoutes.Get("/system/db", func(c *fiber.Ctx) error {
		var dbName string
		var dbUser string
		var searchPath string
		var publicOrders string
		var unqualifiedOrders string
		var migrationsCount int64
		var migrationsError string
		var pgcryptoInstalled bool

		// Collect diagnostics
		query := func(q string, dest interface{}) error {
			res := db.Raw(q).Scan(dest)
			return res.Error
		}

		_ = query("SELECT current_database()", &dbName)
		_ = query("SELECT current_user", &dbUser)
		_ = query("SHOW search_path", &searchPath)
		_ = query("SELECT to_regclass('public.orders')", &publicOrders)
		_ = query("SELECT to_regclass('orders')", &unqualifiedOrders)

		// gorm_migrations count (optional, may be missing)
		if res := db.Raw("SELECT COUNT(*) FROM gorm_migrations").Scan(&migrationsCount); res.Error != nil {
			migrationsError = res.Error.Error()
		}

		// pgcrypto extension presence
		if res := db.Raw("SELECT EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'pgcrypto')").Scan(&pgcryptoInstalled); res.Error != nil {
			// keep default false; include error in response for visibility
			if migrationsError == "" {
				migrationsError = res.Error.Error()
			} else {
				migrationsError = migrationsError + "; " + res.Error.Error()
			}
		}

		return c.JSON(fiber.Map{
			"db_name":                 dbName,
			"db_user":                 dbUser,
			"search_path":            searchPath,
			"to_regclass":            fiber.Map{"public.orders": publicOrders, "orders": unqualifiedOrders},
			"gorm_migrations_count":  migrationsCount,
			"gorm_migrations_error":  migrationsError,
			"pgcrypto_installed":     pgcryptoInstalled,
		})
	})

	// ❤️ Health Check Endpoint
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "message": "🟢 Server is healthy"})
	})

	log.Println("✅ All routes configured successfully")

	// 🚀 Start HTTP Server
	log.Printf("🚀 Server starting on port %s", cfg.Port)
	log.Printf("🌐 Health check: http://localhost:%s/health", cfg.Port)

	// Move these route definitions inside the main function
	// Add them after your existing adminRoutes setup and before app.Listen()

	// Add these routes to your admin routes group
	// adminRoutes.GET("/permissions/available", userHandler.GetAvailablePermissions)
	// adminRoutes.PUT("/users/:id/permissions", middleware.SuperAdminOnly(), userHandler.UpdateUserPermissions)
	// adminRoutes.POST("/users/:id/permissions/toggle", middleware.SuperAdminOnly(), userHandler.ToggleUserPermission)

	// Example of using permission middleware on other endpoints
	// productRoutes := app.Group("/api/v1/products")
	// productRoutes.Use(middleware.JWTMiddleware(cfg))
	// productRoutes.GET("/", middleware.RequirePermission(users.PermProductsRead), productHandler.GetProducts)
	// productRoutes.POST("/", middleware.RequirePermission(users.PermProductsWrite), productHandler.CreateProduct)
	// productRoutes.DELETE("/:id", middleware.RequirePermission(users.PermProductsDelete), productHandler.DeleteProduct)

	log.Printf("📚 API base URL: http://localhost:%s/api/v1", cfg.Port)
	log.Printf("🔐 Auth endpoints: http://localhost:%s/api/v1/auth", cfg.Port)
	log.Printf("👑 Admin endpoints: http://localhost:%s/api/v1/admin", cfg.Port)

	// 🛍️ Initialize Products Domain
	log.Println("🛍️ Setting up products domain...")
	productsRepo := products.NewRepository(db)
	productsService := products.NewService(productsRepo)
	productsHandler := products.NewHandler(productsService)
	log.Println("✅ Products domain initialized (using external image hosting)")

	// Configure product routes
	log.Println("🛍️ Configuring public product routes...")
	v1.MountProductRoutes(api, productsHandler)
	log.Println("👑 Configuring admin product routes...")
	v1.MountAdminProductRoutes(adminRoutes, productsHandler)

	// 👥 Setup Customers Routes (service already initialized above)
	log.Println("👥 Setting up customers routes...")
	customersHandler := customers.NewHandler(customersService)
	customers.SetupRoutes(api, customersHandler, cfg)
	log.Println("✅ Customers routes configured")

	// 💳 Initialize Payments Repository and Client (service will be initialized after orders)
	log.Println("💳 Setting up payments repository and client...")
	paymentsRepo := payments.NewRepository(db)
	
	// Initialize Paystack client
	paystackClient := payments.NewPaystackClient(cfg.PaystackSecretKey, cfg.PaystackWebhookSecret, cfg.AppBaseURL, cfg.CallbackURL)
	log.Println("✅ Payments repository and client initialized")

	// 🎫 Initialize Coupons Domain (moved before orders)
	log.Println("🎫 Setting up coupons domain...")
	couponsRepo := coupons.NewRepository(db)
	couponsService := coupons.NewService(couponsRepo)
	couponsHandler := coupons.NewHandler(couponsService)
	coupons.SetupPublicRoutes(app, couponsHandler)
	coupons.SetupRoutes(app, couponsHandler, cfg)
	log.Println("✅ Coupons domain initialized")

	// 🔔 Initialize Notifications Domain (moved before orders)
	log.Println("🔔 Setting up notifications domain...")
	notificationRepo := notifications.NewNotificationRepository(db)
	templateRepo := notifications.NewTemplateRepository(db)
	pushTokenRepo := notifications.NewPushTokenRepository(db)
	notificationService := notifications.NewNotificationService(notificationRepo, templateRepo, pushTokenRepo)
	notificationHandler := notifications.NewNotificationHandler(notificationService)
	notifications.SetupRoutes(app, cfg, notificationHandler)
	notifications.SetupAdminRoutes(app, cfg, notificationHandler)
	
	// 🔥 Setup FCM Routes for Dashboard
	log.Println("🔥 Setting up FCM routes for dashboard...")
	notifications.SetupFCMRoutes(app, db, cfg)
	notifications.SetupPublicFCMRoutes(app, db, cfg)
	log.Println("✅ FCM routes initialized")
	
	log.Println("✅ Notifications domain initialized")

	// 💬 Initialize Chat Domain
	log.Println("💬 Setting up chat domain...")
	chat.SetupRoutes(app, db, cfg)
	chat.SetupAdminRoutes(app, db, cfg)
	log.Println("✅ Chat domain initialized")

	// 📦 Initialize Orders Domain with Cart and Coupon Integration
	log.Println("📦 Setting up orders domain with cart functionality...")
	ordersRepo := orders.NewRepository(db)
	
	// 🎯 Initialize Custom Requests Domain (needed by orders)
	log.Println("🎯 Setting up custom requests domain...")
	customRequestsRepo := custom_requests.NewRepository(db)
	customRequestsService := custom_requests.NewService(customRequestsRepo)
	customRequestsHandler := custom_requests.NewHandler(customRequestsService)
	custom_requests.SetupRoutes(api, customRequestsHandler, cfg)
	custom_requests.SetupAdminRoutes(adminRoutes, customRequestsHandler, cfg)
	v1.MountCustomRequestRoutes(api, customRequestsHandler, middleware.JWTMiddleware(cfg), middleware.AdminMiddleware())
	log.Println("✅ Custom requests domain initialized")
	
	// Initialize delivery costing functionality first (needed by orders)
	log.Println("💰 Setting up delivery costing system...")
	addressRepo := repos.NewDBAddressRepo(db)
	deliveryHandler, err := handlers.NewDeliveryHandler("./data/delivery_zones.json", addressRepo)
	var deliveryMatcher orders.DeliveryMatcherInterface
	if err != nil {
		log.Printf("⚠️ Failed to initialize delivery costing: %v", err)
		deliveryMatcher = nil // Will use fallback pricing
	} else {
		deliveryMatcher = deliveryHandler.GetMatcher()
		log.Println("✅ Delivery costing system initialized")
	}
	
	// Initialize delivery service (needed by orders)
	deliveryRepo := delivery.NewDeliveryRepository(db)
	deliveryService := delivery.NewDeliveryService(deliveryRepo, notificationService, ordersRepo, customersService)
	
	// Initialize orders service first (without payments service)
	var ordersService *orders.Service
	ordersService = orders.NewService(ordersRepo, productsRepo, couponsService, customersService, authService, &tempPaymentService{}, deliveryService, addressRepo, deliveryMatcher, notificationService, customRequestsService, db)
	
	// Now initialize payments service with orders service
	paymentsService := payments.NewService(paymentsRepo, paystackClient, ordersService)
	
	// Update orders service with real payments service
	ordersService = orders.NewService(ordersRepo, productsRepo, couponsService, customersService, authService, paymentsService, deliveryService, addressRepo, deliveryMatcher, notificationService, customRequestsService, db)
	
	// Setup payments routes
	paymentsHandler := payments.NewHandler(paymentsService)
	payments.SetupRoutes(app, cfg, paymentsHandler)
	payments.SetupAdminRoutes(app, cfg, paymentsHandler)
	log.Println("✅ Payments domain initialized with Paystack integration")
	
	// Setup orders routes
	ordersHandler := orders.NewHandler(ordersService)
	cartHandler := orders.NewCartHandler(ordersService)
	couponHandler := orders.NewCouponHandler(couponsService)
	orders.SetupRoutes(app, cfg, ordersHandler, cartHandler, couponHandler)
	log.Println("✅ Orders domain with cart functionality initialized")

	// 🚚 Setup Delivery Routes (service and costing already initialized above)
	log.Println("🚚 Setting up delivery routes...")
	if deliveryHandler != nil {
		// Create enhanced delivery handler with costing functionality
		enhancedHandler := delivery.NewDeliveryHandlerWithCosting(deliveryService, deliveryHandler.GetMatcher(), addressRepo)
		delivery.SetupDeliveryRoutes(app, enhancedHandler, cfg)
	} else {
		// Fallback to basic delivery handler without costing
		basicHandler := delivery.NewDeliveryHandler(deliveryService)
		delivery.SetupDeliveryRoutes(app, basicHandler, cfg)
	}
	log.Println("✅ Delivery routes initialized")

	// 📊 Initialize Analytics Domain
	log.Println("📊 Setting up analytics domain...")
	analyticsRepo := analytics.NewAnalyticsRepository(db)
	analyticsService := analytics.NewAnalyticsService(analyticsRepo)
	analyticsHandler := analytics.NewAnalyticsHandler(analyticsService)
	analytics.SetupAnalyticsRoutes(app, analyticsHandler, cfg)
	log.Println("✅ Analytics domain initialized")



	// 👤 Old Users Domain - DISABLED (replaced by auth domain)
	// The old users domain conflicts with the new auth domain
	// All user management is now handled by the auth domain
	log.Println("ℹ️ Old users domain disabled - using auth domain instead")

	// 📁 Static file serving for existing uploads (while transitioning to Cloudinary)
	app.Static("/uploads", "./uploads")
	log.Println("✅ Static file serving configured for /uploads")

	log.Fatal(app.Listen(":" + cfg.Port))
}
