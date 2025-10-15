package analytics

import (
	"errandShop/internal/presenter"
	"errandShop/internal/validation"

	"github.com/gofiber/fiber/v2"
)

type AnalyticsHandler struct {
	service AnalyticsService
}

func NewAnalyticsHandler(service AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{service: service}
}

// GET /api/v1/dashboard/data - New dashboard endpoint matching frontend requirements
func (h *AnalyticsHandler) GetDashboardData(c *fiber.Ctx) error {
	timeRange := c.Query("timeRange", "month")
	
	// Validate timeRange
	validRanges := map[string]bool{"today": true, "week": true, "month": true}
	if !validRanges[timeRange] {
		timeRange = "month"
	}

	dashboard, err := h.service.GetDashboardData(timeRange)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to get dashboard data",
		})
	}

	return c.JSON(dashboard)
}

// GET /api/v1/dashboard/today-sales - Today's Sales endpoint
func (h *AnalyticsHandler) GetTodaySales(c *fiber.Ctx) error {
	sales, err := h.service.GetTodaySales()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to get today's sales",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    sales,
	})
}

// GET /api/v1/dashboard/active-users - Active Users endpoint
func (h *AnalyticsHandler) GetActiveUsers(c *fiber.Ctx) error {
	users, err := h.service.GetActiveUsers()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to get active users",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    users,
	})
}

// GET /api/v1/dashboard/total-products - Total Products endpoint
func (h *AnalyticsHandler) GetTotalProducts(c *fiber.Ctx) error {
	products, err := h.service.GetTotalProducts()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to get total products",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    products,
	})
}

// GET /api/v1/dashboard/coupons-analytics - Coupons Analytics endpoint
func (h *AnalyticsHandler) GetCouponsAnalytics(c *fiber.Ctx) error {
	coupons, err := h.service.GetCouponsAnalytics()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to get coupons analytics",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    coupons,
	})
}

// GET /api/v1/dashboard/recent-orders - Recent Orders endpoint
func (h *AnalyticsHandler) GetRecentOrdersEndpoint(c *fiber.Ctx) error {
	limit := 10
	if l := c.QueryInt("limit", 10); l > 0 && l <= 50 {
		limit = l
	}

	orders, err := h.service.GetRecentOrdersData(limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to get recent orders",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    orders,
	})
}

// GET /api/v1/dashboard/low-stock-alerts - Low Stock Alerts endpoint
func (h *AnalyticsHandler) GetLowStockAlerts(c *fiber.Ctx) error {
	threshold := 10
	if t := c.QueryInt("threshold", 10); t > 0 {
		threshold = t
	}

	alerts, err := h.service.GetLowStockAlerts(threshold)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to get low stock alerts",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    alerts,
	})
}

// GET /api/v1/dashboard/sales-overview - Sales Overview endpoint
func (h *AnalyticsHandler) GetSalesOverviewEndpoint(c *fiber.Ctx) error {
	period := c.Query("period", "daily")
	validPeriods := map[string]bool{"daily": true, "weekly": true, "monthly": true}
	if !validPeriods[period] {
		period = "daily"
	}

	overview, err := h.service.GetSalesOverviewData(period)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to get sales overview",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    overview,
	})
}

// GET /api/v1/analytics/dashboard - Legacy endpoint
func (h *AnalyticsHandler) GetDashboard(c *fiber.Ctx) error {
	var req AnalyticsRequest
	if err := c.QueryParser(&req); err != nil {
		return presenter.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request parameters")
	}

	if err := validation.ValidateStruct(&req); err != nil {
		return presenter.ErrorResponse(c, fiber.StatusBadRequest, "Validation failed")
	}

	dashboard, err := h.service.GetDashboard(&req)
	if err != nil {
		return presenter.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to get dashboard data")
	}

	return presenter.SuccessResponse(c, "Dashboard data retrieved successfully", dashboard)
}

// GET /analytics/reports/sales - New sales report endpoint
func (h *AnalyticsHandler) GetSalesReport(c *fiber.Ctx) error {
	var req ReportRequest
	req.ReportType = ReportSales
	if err := c.QueryParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request parameters",
		})
	}

	report, err := h.service.GetSalesReport(&req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to get sales report",
		})
	}

	return c.JSON(report)
}

// GET /analytics/reports/products - New products report endpoint
func (h *AnalyticsHandler) GetProductsReport(c *fiber.Ctx) error {
	var req ReportRequest
	req.ReportType = ReportProducts
	if err := c.QueryParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request parameters",
		})
	}

	report, err := h.service.GetProductsReport(&req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to get products report",
		})
	}

	return c.JSON(report)
}

// GET /analytics/reports/customers - New customers report endpoint
func (h *AnalyticsHandler) GetCustomersReport(c *fiber.Ctx) error {
	var req ReportRequest
	req.ReportType = ReportCustomers
	if err := c.QueryParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request parameters",
		})
	}

	report, err := h.service.GetCustomersReport(&req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to get customers report",
		})
	}

	return c.JSON(report)
}

// GET /analytics/reports/orders - New orders report endpoint
func (h *AnalyticsHandler) GetOrdersReport(c *fiber.Ctx) error {
	var req ReportRequest
	req.ReportType = ReportOrders
	if err := c.QueryParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request parameters",
		})
	}

	report, err := h.service.GetOrdersReport(&req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to get orders report",
		})
	}

	return c.JSON(report)
}

// GET /analytics/reports/delivery - New delivery report endpoint
func (h *AnalyticsHandler) GetDeliveryReport(c *fiber.Ctx) error {
	var req ReportRequest
	req.ReportType = ReportDelivery
	if err := c.QueryParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request parameters",
		})
	}

	report, err := h.service.GetDeliveryReport(&req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to get delivery report",
		})
	}

	return c.JSON(report)
}

// GET /analytics/reports/payments - New payments report endpoint
func (h *AnalyticsHandler) GetPaymentsReport(c *fiber.Ctx) error {
	var req ReportRequest
	req.ReportType = ReportPayments
	if err := c.QueryParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request parameters",
		})
	}

	report, err := h.service.GetPaymentsReport(&req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to get payments report",
		})
	}

	return c.JSON(report)
}

// Legacy handlers for backward compatibility

// GET /api/v1/analytics/reports/customer (legacy)
func (h *AnalyticsHandler) GetCustomerReport(c *fiber.Ctx) error {
	var req ReportRequest
	req.ReportType = ReportCustomers
	if err := c.QueryParser(&req); err != nil {
		return presenter.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request parameters")
	}

	if err := validation.ValidateStruct(&req); err != nil {
		return presenter.ErrorResponse(c, fiber.StatusBadRequest, "Validation failed")
	}

	report, err := h.service.GetCustomerReport(&req)
	if err != nil {
		return presenter.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to get customer report")
	}

	return presenter.SuccessResponse(c, "Customer report retrieved successfully", report)
}

// GET /api/v1/analytics/reports/product (legacy)
func (h *AnalyticsHandler) GetProductReport(c *fiber.Ctx) error {
	var req ReportRequest
	req.ReportType = ReportProducts
	if err := c.QueryParser(&req); err != nil {
		return presenter.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request parameters")
	}

	if err := validation.ValidateStruct(&req); err != nil {
		return presenter.ErrorResponse(c, fiber.StatusBadRequest, "Validation failed")
	}

	report, err := h.service.GetProductReport(&req)
	if err != nil {
		return presenter.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to get product report")
	}

	return presenter.SuccessResponse(c, "Product report retrieved successfully", report)
}

// GET /api/v1/analytics/reports/order (legacy)
func (h *AnalyticsHandler) GetOrderReport(c *fiber.Ctx) error {
	var req ReportRequest
	req.ReportType = ReportOrders
	if err := c.QueryParser(&req); err != nil {
		return presenter.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request parameters")
	}

	if err := validation.ValidateStruct(&req); err != nil {
		return presenter.ErrorResponse(c, fiber.StatusBadRequest, "Validation failed")
	}

	report, err := h.service.GetOrderReport(&req)
	if err != nil {
		return presenter.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to get order report")
	}

	return presenter.SuccessResponse(c, "Order report retrieved successfully", report)
}

// GET /api/v1/analytics/reports/payment (legacy)
func (h *AnalyticsHandler) GetPaymentReport(c *fiber.Ctx) error {
	var req ReportRequest
	req.ReportType = ReportPayments
	if err := c.QueryParser(&req); err != nil {
		return presenter.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request parameters")
	}

	if err := validation.ValidateStruct(&req); err != nil {
		return presenter.ErrorResponse(c, fiber.StatusBadRequest, "Validation failed")
	}

	report, err := h.service.GetPaymentReport(&req)
	if err != nil {
		return presenter.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to get payment report")
	}

	return presenter.SuccessResponse(c, "Payment report retrieved successfully", report)
}
