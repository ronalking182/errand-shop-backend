package chat

import (
	"errandShop/config"
	"errandShop/internal/domain/notifications"
	"errandShop/internal/middleware"
	"errandShop/internal/pkg/jwt"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	jwtlib "github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

func SetupRoutes(app *fiber.App, db *gorm.DB, cfg *config.Config) {
	// Initialize repositories
	roomRepo := NewChatRoomRepository(db)
	messageRepo := NewChatMessageRepository(db)

	// Initialize notification service
	notificationRepo := notifications.NewNotificationRepository(db)
	templateRepo := notifications.NewTemplateRepository(db)
	pushTokenRepo := notifications.NewPushTokenRepository(db)
	notificationSvc := notifications.NewNotificationService(notificationRepo, templateRepo, pushTokenRepo)

	// Initialize chat service
	chatSvc := NewChatService(roomRepo, messageRepo, notificationSvc)

	// Initialize WebSocket hub
	hub := NewHub()
	go hub.Run()

	// Connect hub to chat service
	chatSvc.SetHub(hub)

	// Initialize handlers
	handler := NewChatHandler(chatSvc)
	wsHandler := NewWebSocketHandler(hub, chatSvc, cfg.JWTSecret)

	// Setup routes
	chat := app.Group("/api/v1/chat")
	chat.Use(middleware.JWTMiddleware(cfg))

	// Chat room routes
	chat.Post("/rooms", handler.CreateChatRoom)
	chat.Get("/rooms", handler.GetChatRooms)
	chat.Get("/rooms/:id", handler.GetChatRoom)
	chat.Put("/rooms/:id", handler.UpdateChatRoom)
	chat.Delete("/rooms/:id", handler.DeleteChatRoom)
	chat.Post("/rooms/:id/assign", handler.AssignAdminToRoom)
	chat.Post("/rooms/:id/unassign", handler.UnassignAdminFromRoom)

	// Chat message routes
	chat.Post("/messages", handler.SendMessage)
	chat.Get("/rooms/:id/messages", handler.GetMessages)
	chat.Put("/rooms/:id/read", handler.MarkMessagesAsRead)
	chat.Delete("/messages/:id", handler.DeleteMessage)

	// Utility routes
	chat.Get("/stats", handler.GetChatStats)
	chat.Post("/typing", handler.SendTypingIndicator)

	// WebSocket routes with authentication
	app.Use("/ws", func(c *fiber.Ctx) error {
		// IsWebSocketUpgrade returns true if the client
		// requested upgrade to the WebSocket protocol.
		if websocket.IsWebSocketUpgrade(c) {
			// Check for token before allowing WebSocket upgrade
			token := c.Query("token")
			if token == "" {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": "Token required for WebSocket connection",
				})
			}
			

			
			// Validate JWT token using the same method as middleware
			parsedToken, err := jwtlib.ParseWithClaims(token, &jwt.JWTClaims{}, func(token *jwtlib.Token) (interface{}, error) {
				return []byte(cfg.JWTSecret), nil
			})
			if err != nil {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": "Invalid token",
				})
			}
			
			_, ok := parsedToken.Claims.(*jwt.JWTClaims)
			if !ok || !parsedToken.Valid {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": "Invalid token claims",
				})
			}
			
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/ws/chat", websocket.New(wsHandler.HandleWebSocket))
}

func SetupAdminRoutes(app *fiber.App, db *gorm.DB, cfg *config.Config) {
	// Initialize repositories
	roomRepo := NewChatRoomRepository(db)
	messageRepo := NewChatMessageRepository(db)

	// Initialize notification service
	notificationRepo := notifications.NewNotificationRepository(db)
	templateRepo := notifications.NewTemplateRepository(db)
	pushTokenRepo := notifications.NewPushTokenRepository(db)
	notificationSvc := notifications.NewNotificationService(notificationRepo, templateRepo, pushTokenRepo)

	// Initialize chat service
	chatSvc := NewChatService(roomRepo, messageRepo, notificationSvc)

	// Initialize handler
	handler := NewChatHandler(chatSvc)

	// Setup admin routes
	admin := app.Group("/api/admin/chat")
	admin.Use(middleware.JWTMiddleware(cfg))
	admin.Use(middleware.AdminMiddleware())

	// Admin chat management
	admin.Get("/rooms", handler.GetChatRooms)
	admin.Get("/rooms/:id", handler.GetChatRoom)
	admin.Put("/rooms/:id", handler.UpdateChatRoom)
	admin.Delete("/rooms/:id", handler.DeleteChatRoom)
	admin.Post("/rooms/:id/assign", handler.AssignAdminToRoom)
	admin.Post("/rooms/:id/unassign", handler.UnassignAdminFromRoom)

	// Admin message management
	admin.Get("/rooms/:id/messages", handler.GetMessages)
	admin.Delete("/messages/:id", handler.DeleteMessage)

	// Admin statistics
	admin.Get("/stats", handler.GetChatStats)
}