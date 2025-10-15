package analytics

import (
	"fmt"
	"time"
)

type AnalyticsService interface {
	// New dashboard endpoint
	GetDashboardData(timeRange string) (*DashboardResponse, error)
	
	// Individual dashboard metrics endpoints
	GetTodaySales() (*TodaySalesResponse, error)
	GetActiveUsers() (*ActiveUsersResponse, error)
	GetTotalProducts() (*TotalProductsResponse, error)
	GetCouponsAnalytics() (*CouponsAnalyticsResponse, error)
	GetRecentOrdersData(limit int) (*RecentOrdersResponse, error)
	GetLowStockAlerts(threshold int) (*LowStockAlertsResponse, error)
	GetSalesOverviewData(period string) (*SalesOverviewResponse, error)
	
	// New reports endpoints
	GetSalesReport(req *ReportRequest) (*SalesReportResponse, error)
	GetProductsReport(req *ReportRequest) (*ProductReportResponse, error)
	GetCustomersReport(req *ReportRequest) (*CustomerReportResponse, error)
	GetOrdersReport(req *ReportRequest) (*OrderReportResponse, error)
	GetDeliveryReport(req *ReportRequest) (*DeliveryReportResponse, error)
	GetPaymentsReport(req *ReportRequest) (*PaymentReportResponse, error)
	
	// Legacy methods (keeping for backward compatibility)
	GetDashboard(req *AnalyticsRequest) (*DashboardResponse, error)
	GetCustomerReport(req *ReportRequest) (*CustomerReportResponse, error)
	GetProductReport(req *ReportRequest) (*ProductReportResponse, error)
	GetOrderReport(req *ReportRequest) (*OrderReportResponse, error)
	GetPaymentReport(req *ReportRequest) (*PaymentReportResponse, error)
}

type analyticsService struct {
	repo AnalyticsRepository
}

func NewAnalyticsService(repo AnalyticsRepository) AnalyticsService {
	return &analyticsService{repo: repo}
}

func (s *analyticsService) GetDashboard(req *AnalyticsRequest) (*DashboardResponse, error) {
	// Legacy endpoint - returns basic dashboard data
	startDate, endDate := s.getDateRange(req.TimeRange, req.StartDate, req.EndDate)

	metrics, err := s.repo.GetDashboardMetrics(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get dashboard metrics: %w", err)
	}

	return &DashboardResponse{
		Success: true,
		Data: DashboardData{
			KPIs: DashboardKPIs{
				TotalSales:    metrics.TotalRevenue,
				ActiveUsers:   metrics.TotalCustomers,
				TotalProducts: metrics.TotalProducts,
				CouponsIssued: 0, // TODO: implement coupon count
			},
			RecentOrders:     []RecentOrder{},
			LowStockProducts: []LowStockProduct{},
		},
	}, nil
}

// New GetSalesReport method (used by new endpoints)
func (s *analyticsService) GetSalesReport(req *ReportRequest) (*SalesReportResponse, error) {
	startDate, endDate := s.getDateRange(req.TimeRange, req.StartDate, req.EndDate)

	analytics, err := s.repo.GetSalesAnalytics(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get sales analytics: %w", err)
	}

	return &SalesReportResponse{
		Success: true,
		Data: SalesOverviewData{
			TodaySales:   TrendData{Value: analytics.TotalRevenue, Trend: "up", Change: 0},
			WeeklySales:  TrendData{Value: analytics.TotalRevenue, Trend: "up", Change: 0},
			MonthlySales: TrendData{Value: analytics.TotalRevenue, Trend: "up", Change: 0},
			YearlySales:  TrendData{Value: analytics.TotalRevenue, Trend: "up", Change: 0},
		},
	}, nil
}

func (s *analyticsService) GetCustomerReport(req *ReportRequest) (*CustomerReportResponse, error) {
	startDate, endDate := s.getDateRange(req.TimeRange, req.StartDate, req.EndDate)

	analytics, err := s.repo.GetCustomerAnalytics(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer analytics: %w", err)
	}

	return &CustomerReportResponse{
		Success: true,
		Data:    *analytics,
	}, nil
}

func (s *analyticsService) GetProductReport(req *ReportRequest) (*ProductReportResponse, error) {
	startDate, endDate := s.getDateRange(req.TimeRange, req.StartDate, req.EndDate)

	analytics, err := s.repo.GetProductAnalytics(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get product analytics: %w", err)
	}

	return &ProductReportResponse{
		Success: true,
		Data:    analytics.TopSellingProducts,
	}, nil
}

func (s *analyticsService) GetOrderReport(req *ReportRequest) (*OrderReportResponse, error) {
	startDate, endDate := s.getDateRange(req.TimeRange, req.StartDate, req.EndDate)

	analytics, err := s.repo.GetOrderAnalytics(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get order analytics: %w", err)
	}

	return &OrderReportResponse{
		Success: true,
		Data:    *analytics,
	}, nil
}

func (s *analyticsService) GetDeliveryReport(req *ReportRequest) (*DeliveryReportResponse, error) {
	startDate, endDate := s.getDateRange(req.TimeRange, req.StartDate, req.EndDate)

	analytics, err := s.repo.GetDeliveryAnalytics(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get delivery analytics: %w", err)
	}

	return &DeliveryReportResponse{
		Success: true,
		Data:    *analytics,
	}, nil
}

func (s *analyticsService) GetPaymentReport(req *ReportRequest) (*PaymentReportResponse, error) {
	startDate, endDate := s.getDateRange(req.TimeRange, req.StartDate, req.EndDate)

	analytics, err := s.repo.GetPaymentAnalytics(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment analytics: %w", err)
	}

	return &PaymentReportResponse{
		Success: true,
		Data:    *analytics,
	}, nil
}

// New Dashboard Method
func (s *analyticsService) GetDashboardData(timeRange string) (*DashboardResponse, error) {
	data, err := s.repo.GetDashboardData(timeRange)
	if err != nil {
		return nil, fmt.Errorf("failed to get dashboard data: %w", err)
	}

	return &DashboardResponse{
		Success: true,
		Data:    *data,
	}, nil
}

// New Reports Methods

func (s *analyticsService) GetProductsReport(req *ReportRequest) (*ProductReportResponse, error) {
	startDate, endDate := s.getDateRange(req.TimeRange, req.StartDate, req.EndDate)
	limit := req.Limit
	if limit == 0 {
		limit = 10
	}

	data, err := s.repo.GetTopProducts(startDate, endDate, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get top products: %w", err)
	}

	return &ProductReportResponse{
		Success: true,
		Data:    data,
	}, nil
}

func (s *analyticsService) GetCustomersReport(req *ReportRequest) (*CustomerReportResponse, error) {
	startDate, endDate := s.getDateRange(req.TimeRange, req.StartDate, req.EndDate)

	analytics, err := s.repo.GetCustomerAnalytics(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer analytics: %w", err)
	}

	return &CustomerReportResponse{
		Success: true,
		Data:    *analytics,
	}, nil
}

func (s *analyticsService) GetOrdersReport(req *ReportRequest) (*OrderReportResponse, error) {
	startDate, endDate := s.getDateRange(req.TimeRange, req.StartDate, req.EndDate)

	analytics, err := s.repo.GetOrderAnalytics(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get order analytics: %w", err)
	}

	return &OrderReportResponse{
		Success: true,
		Data:    *analytics,
	}, nil
}

func (s *analyticsService) GetPaymentsReport(req *ReportRequest) (*PaymentReportResponse, error) {
	startDate, endDate := s.getDateRange(req.TimeRange, req.StartDate, req.EndDate)

	analytics, err := s.repo.GetPaymentAnalytics(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment analytics: %w", err)
	}

	return &PaymentReportResponse{
		Success: true,
		Data:    *analytics,
	}, nil
}

// Individual Dashboard Metrics Service Methods
func (s *analyticsService) GetTodaySales() (*TodaySalesResponse, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	yesterday := startOfDay.AddDate(0, 0, -1)

	// Get today's sales
	todayMetrics, err := s.repo.GetDashboardKPIs(startOfDay, now)
	if err != nil {
		return nil, fmt.Errorf("failed to get today's sales: %w", err)
	}

	// Get yesterday's sales for comparison
	yesterdayMetrics, err := s.repo.GetDashboardKPIs(yesterday, startOfDay)
	if err != nil {
		return nil, fmt.Errorf("failed to get yesterday's sales: %w", err)
	}

	// Calculate growth rate and trend
	growthRate := 0.0
	trend := "stable"
	if yesterdayMetrics.TotalSales > 0 {
		growthRate = ((todayMetrics.TotalSales - yesterdayMetrics.TotalSales) / yesterdayMetrics.TotalSales) * 100
		if growthRate > 0 {
			trend = "up"
		} else if growthRate < 0 {
			trend = "down"
		}
	}

	// Get order count for today
	orders, err := s.repo.GetRecentOrders(1000) // Get all orders to count today's
	if err != nil {
		return nil, fmt.Errorf("failed to get orders count: %w", err)
	}

	var todayOrdersCount int64
	for _, order := range orders {
		if orderTime, err := time.Parse(time.RFC3339, order.CreatedAt); err == nil {
			if orderTime.After(startOfDay) {
				todayOrdersCount++
			}
		}
	}

	return &TodaySalesResponse{
		Success: true,
		Data: TodaySalesData{
			Amount:      todayMetrics.TotalSales,
			Currency:    "NGN",
			OrdersCount: todayOrdersCount,
			GrowthRate:  growthRate,
			Trend:       trend,
			LastUpdated: now.Format(time.RFC3339),
		},
	}, nil
}

func (s *analyticsService) GetActiveUsers() (*ActiveUsersResponse, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	yesterday := startOfDay.AddDate(0, 0, -1)

	// Get active users count (users who placed orders in the last 30 days)
	last30Days := now.AddDate(0, 0, -30)
	activeCount, err := s.repo.GetMobileAppUsersCount(last30Days, now)
	if err != nil {
		return nil, fmt.Errorf("failed to get active users: %w", err)
	}

	// Get yesterday's active users for comparison
	yesterdayCount, err := s.repo.GetMobileAppUsersCount(last30Days.AddDate(0, 0, -1), yesterday)
	if err != nil {
		return nil, fmt.Errorf("failed to get yesterday's active users: %w", err)
	}

	// Calculate growth rate and trend
	growthRate := 0.0
	trend := "stable"
	if yesterdayCount > 0 {
		growthRate = ((float64(activeCount) - float64(yesterdayCount)) / float64(yesterdayCount)) * 100
		if growthRate > 0 {
			trend = "up"
		} else if growthRate < 0 {
			trend = "down"
		}
	}

	return &ActiveUsersResponse{
		Success: true,
		Data: ActiveUsersData{
			Count:       activeCount,
			GrowthRate:  growthRate,
			Trend:       trend,
			LastUpdated: now.Format(time.RFC3339),
		},
	}, nil
}

func (s *analyticsService) GetTotalProducts() (*TotalProductsResponse, error) {
	now := time.Now()

	// Get product analytics
	productAnalytics, err := s.repo.GetProductAnalytics(now.AddDate(0, 0, -1), now)
	if err != nil {
		return nil, fmt.Errorf("failed to get product analytics: %w", err)
	}

	return &TotalProductsResponse{
		Success: true,
		Data: TotalProductsData{
			Total:       productAnalytics.TotalProducts,
			Active:      productAnalytics.ActiveProducts,
			Inactive:    productAnalytics.TotalProducts - productAnalytics.ActiveProducts,
			LowStock:    productAnalytics.LowStockProducts,
			LastUpdated: now.Format(time.RFC3339),
		},
	}, nil
}

func (s *analyticsService) GetCouponsAnalytics() (*CouponsAnalyticsResponse, error) {
	now := time.Now()
	last30Days := now.AddDate(0, 0, -30)

	// Get coupon performance data
	couponPerformance, err := s.repo.GetCouponPerformance(last30Days, now)
	if err != nil {
		return nil, fmt.Errorf("failed to get coupon performance: %w", err)
	}

	// Calculate totals
	var totalIssued, totalUsed int64
	for _, coupon := range couponPerformance {
		totalUsed += coupon.Usage
		totalIssued++ // Each coupon in the list represents an issued coupon
	}

	// Calculate usage rate
	usageRate := 0.0
	if totalIssued > 0 {
		usageRate = (float64(totalUsed) / float64(totalIssued)) * 100
	}

	// Get top 5 performing coupons
	topPerforming := couponPerformance
	if len(topPerforming) > 5 {
		topPerforming = topPerforming[:5]
	}

	return &CouponsAnalyticsResponse{
		Success: true,
		Data: CouponsAnalyticsData{
			TotalIssued:   totalIssued,
			TotalUsed:     totalUsed,
			ActiveCount:   totalIssued, // Assuming all issued coupons are active
			UsageRate:     usageRate,
			TopPerforming: topPerforming,
			LastUpdated:   now.Format(time.RFC3339),
		},
	}, nil
}

func (s *analyticsService) GetRecentOrdersData(limit int) (*RecentOrdersResponse, error) {
	now := time.Now()

	// Get recent orders
	orders, err := s.repo.GetRecentOrders(limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent orders: %w", err)
	}

	// Get total orders count for today
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	_, err = s.repo.GetDashboardKPIs(startOfDay, now)
	if err != nil {
		return nil, fmt.Errorf("failed to get total orders count: %w", err)
	}

	return &RecentOrdersResponse{
		Success: true,
		Data: RecentOrdersData{
			Orders:      orders,
			TotalCount:  int64(len(orders)),
			LastUpdated: now.Format(time.RFC3339),
		},
	}, nil
}

func (s *analyticsService) GetLowStockAlerts(threshold int) (*LowStockAlertsResponse, error) {
	now := time.Now()

	// Get low stock products
	products, err := s.repo.GetLowStockProducts(threshold)
	if err != nil {
		return nil, fmt.Errorf("failed to get low stock products: %w", err)
	}

	// Count critical stock products (less than 5)
	var criticalCount int64
	for _, product := range products {
		if product.CurrentStock < 5 {
			criticalCount++
		}
	}

	return &LowStockAlertsResponse{
		Success: true,
		Data: LowStockAlertsData{
			Products:      products,
			TotalCount:    int64(len(products)),
			CriticalCount: criticalCount,
			LastUpdated:   now.Format(time.RFC3339),
		},
	}, nil
}

func (s *analyticsService) GetSalesOverviewData(period string) (*SalesOverviewResponse, error) {
	now := time.Now()

	// Get sales overview data
	var startDate time.Time
	switch period {
	case "daily":
		startDate = now.AddDate(0, 0, -7) // Last 7 days
	case "weekly":
		startDate = now.AddDate(0, 0, -28) // Last 4 weeks
	case "monthly":
		startDate = now.AddDate(0, -12, 0) // Last 12 months
	default:
		startDate = now.AddDate(0, 0, -7) // Default to daily
		period = "daily"
	}

	_, err := s.repo.GetSalesOverview(startDate, now, period)
	if err != nil {
		return nil, fmt.Errorf("failed to get sales overview: %w", err)
	}

	// Convert to chart data format
	chartData := []ChartPoint{}
	// This would need to be implemented based on the actual data structure returned
	// For now, we'll return empty chart data

	// Calculate totals and averages
	metrics, err := s.repo.GetDashboardKPIs(startDate, now)
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics for sales overview: %w", err)
	}

	averageOrder := 0.0
	if metrics.ActiveUsers > 0 {
		averageOrder = metrics.TotalSales / float64(metrics.ActiveUsers)
	}

	return &SalesOverviewResponse{
		Success: true,
		Data: SalesOverviewChart{
			Period:       period,
			ChartData:    chartData,
			TotalRevenue: metrics.TotalSales,
			TotalOrders:  metrics.ActiveUsers, // Using active users as proxy for orders
			AverageOrder: averageOrder,
			GrowthRate:   0.0, // Would need previous period data to calculate
			LastUpdated:  now.Format(time.RFC3339),
		},
	}, nil
}

func (s *analyticsService) getDateRange(timeRange TimeRange, startDate, endDate *time.Time) (time.Time, time.Time) {
	now := time.Now()

	switch timeRange {
	case TimeRangeToday:
		return now.Truncate(24 * time.Hour), now
	case TimeRangeWeek:
		return now.AddDate(0, 0, -7), now
	case TimeRangeMonth:
		return now.AddDate(0, -1, 0), now
	case TimeRangeQuarter:
		return now.AddDate(0, -3, 0), now
	case TimeRangeYear:
		return now.AddDate(-1, 0, 0), now
	case TimeRangeCustom:
		if startDate != nil && endDate != nil {
			return *startDate, *endDate
		}
		fallthrough
	default:
		return now.AddDate(0, -1, 0), now
	}
}
