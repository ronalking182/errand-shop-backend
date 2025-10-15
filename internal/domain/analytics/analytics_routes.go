package analytics

import (
	"errandShop/config"
	"errandShop/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

func SetupAnalyticsRoutes(app *fiber.App, handler *AnalyticsHandler, cfg *config.Config) {
	// Create analytics group with JWT middleware
	analytics := app.Group("/api/v1/analytics")
	analytics.Use(middleware.JWTMiddleware(cfg))
	analytics.Use(middleware.RBACMiddleware("admin", "superadmin"))

	// Create dashboard group with JWT middleware
	dashboard := app.Group("/api/v1/dashboard")
	dashboard.Use(middleware.JWTMiddleware(cfg))
	dashboard.Use(middleware.RBACMiddleware("admin", "superadmin"))

	// Dashboard endpoints - Individual metrics endpoints as per specification
	dashboard.Get("/data", handler.GetDashboardData)                    // Combined dashboard data
	dashboard.Get("/today-sales", handler.GetTodaySales)                 // Today's Sales
	dashboard.Get("/active-users", handler.GetActiveUsers)               // Active Users
	dashboard.Get("/total-products", handler.GetTotalProducts)           // Total Products
	dashboard.Get("/coupons-analytics", handler.GetCouponsAnalytics)     // Coupons Analytics
	dashboard.Get("/recent-orders", handler.GetRecentOrdersEndpoint)     // Recent Orders
	dashboard.Get("/low-stock-alerts", handler.GetLowStockAlerts)        // Low Stock Alerts
	dashboard.Get("/sales-overview", handler.GetSalesOverviewEndpoint)   // Sales Overview

	// Legacy analytics endpoints
	analytics.Get("/dashboard", handler.GetDashboard)

	// New reports endpoints
	analytics.Get("/reports/sales", handler.GetSalesReport)
	analytics.Get("/reports/products", handler.GetProductsReport)
	analytics.Get("/reports/customers", handler.GetCustomersReport)
	analytics.Get("/reports/orders", handler.GetOrdersReport)
	analytics.Get("/reports/delivery", handler.GetDeliveryReport)
	analytics.Get("/reports/payments", handler.GetPaymentsReport)

	// Legacy individual report endpoints (keeping for backward compatibility)
	analytics.Get("/customer", handler.GetCustomerReport)
	analytics.Get("/product", handler.GetProductReport)
	analytics.Get("/order", handler.GetOrderReport)
	analytics.Get("/payment", handler.GetPaymentReport)
}
