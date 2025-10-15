package notifications

import (
	"errandShop/config"
	"errandShop/internal/middleware"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// SetupFCMRoutes sets up FCM-specific routes for the dashboard
func SetupFCMRoutes(app *fiber.App, db *gorm.DB, cfg *config.Config) {
	// Initialize FCM repositories
	fcmTokenRepo := NewFCMTokenRepository(db)
	fcmMessageRepo := NewFCMMessageRepository(db)
	fcmRecipientRepo := NewFCMMessageRecipientRepository(db)

	// Initialize FCM service
	fcmService := NewFCMService(fcmTokenRepo, fcmMessageRepo, fcmRecipientRepo)

	// Initialize FCM handler
	fcmHandler := NewFCMHandler(fcmService)

	// Setup FCM routes
	fcm := app.Group("/api/fcm")
	fcm.Use(middleware.JWTMiddleware(cfg))
	fcm.Use(middleware.AdminMiddleware()) // Only admins can access FCM endpoints

	// FCM endpoints as required by the dashboard
	fcm.Post("/send", fcmHandler.SendToSingleUser)           // POST /api/fcm/send - Send to single user
	fcm.Post("/send-multiple", fcmHandler.SendToMultipleUsers) // POST /api/fcm/send-multiple - Send to multiple users
	fcm.Post("/broadcast", fcmHandler.BroadcastToAllUsers)     // POST /api/fcm/broadcast - Send to all users
	fcm.Post("/register-token", fcmHandler.RegisterToken)      // POST /api/fcm/register-token - Register device tokens
	fcm.Delete("/unregister-token", fcmHandler.UnregisterToken) // DELETE /api/fcm/unregister-token - Remove tokens
	fcm.Get("/messages", fcmHandler.GetMessages)               // GET /api/fcm/messages - Message history
	fcm.Get("/stats", fcmHandler.GetStats)                    // GET /api/fcm/stats - Analytics data
	fcm.Post("/test", fcmHandler.TestMessage)                  // POST /api/fcm/test - Test message delivery
}

// SetupPublicFCMRoutes sets up public FCM routes (for mobile apps to register tokens)
func SetupPublicFCMRoutes(app *fiber.App, db *gorm.DB, cfg *config.Config) {
	// Initialize FCM repositories
	fcmTokenRepo := NewFCMTokenRepository(db)
	fcmMessageRepo := NewFCMMessageRepository(db)
	fcmRecipientRepo := NewFCMMessageRecipientRepository(db)

	// Initialize FCM service
	fcmService := NewFCMService(fcmTokenRepo, fcmMessageRepo, fcmRecipientRepo)

	// Initialize FCM handler
	fcmHandler := NewFCMHandler(fcmService)

	// Setup public FCM routes (for mobile apps)
	publicFCM := app.Group("/api/v1/fcm")
	publicFCM.Use(middleware.JWTMiddleware(cfg)) // Require authentication but not admin

	// Public endpoints for mobile apps
	publicFCM.Post("/register-token", fcmHandler.RegisterToken)      // Mobile apps can register their tokens
	publicFCM.Delete("/unregister-token", fcmHandler.UnregisterToken) // Mobile apps can unregister their tokens
}