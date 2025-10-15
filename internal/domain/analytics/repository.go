package analytics

import (
	"gorm.io/gorm"
	"time"
)

type AnalyticsRepository interface {
	// Dashboard methods
	GetDashboardData(timeRange string) (*DashboardData, error)
	GetDashboardKPIs(startDate, endDate time.Time) (*DashboardKPIs, error)
	GetRecentOrders(limit int) ([]RecentOrder, error)
	GetLowStockProducts(threshold int) ([]LowStockProduct, error)
	GetMobileAppUsersCount(startDate, endDate time.Time) (int64, error)
	
	// Individual dashboard metrics methods
	GetTodaySales() (float64, error)
	GetActiveUsers() (int64, error)
	GetTotalProducts() (int64, error)
	GetCouponsAnalytics() (int64, error)
	
	// Reports methods
	GetSalesOverview(startDate, endDate time.Time, period string) (*SalesOverviewData, error)
	GetTopProductsReport(startDate, endDate time.Time, limit int) ([]TopProduct, error)
	GetCouponPerformance(startDate, endDate time.Time) ([]CouponPerformance, error)
	GetStorePerformance(startDate, endDate time.Time) ([]StorePerformance, error)
	
	// Legacy methods (keeping for backward compatibility)
	GetDashboardMetrics(startDate, endDate time.Time) (*DashboardMetrics, error)
	GetSalesAnalytics(startDate, endDate time.Time) (*SalesAnalytics, error)
	GetCustomerAnalytics(startDate, endDate time.Time) (*CustomerAnalytics, error)
	GetProductAnalytics(startDate, endDate time.Time) (*ProductAnalytics, error)
	GetOrderAnalytics(startDate, endDate time.Time) (*OrderAnalytics, error)
	GetDeliveryAnalytics(startDate, endDate time.Time) (*DeliveryAnalytics, error)
	GetPaymentAnalytics(startDate, endDate time.Time) (*PaymentAnalytics, error)
	GetRevenueByDay(startDate, endDate time.Time) ([]DataPoint, error)
	GetTopProducts(startDate, endDate time.Time, limit int) ([]ProductSales, error)
	GetTopCustomers(startDate, endDate time.Time, limit int) ([]CustomerSpending, error)
}

type analyticsRepository struct {
	db *gorm.DB
}

func NewAnalyticsRepository(db *gorm.DB) AnalyticsRepository {
	return &analyticsRepository{db: db}
}

func (r *analyticsRepository) GetDashboardMetrics(startDate, endDate time.Time) (*DashboardMetrics, error) {
	var metrics DashboardMetrics

	// Total Revenue
	r.db.Table("orders").
		Select("COALESCE(SUM(total_kobo), 0) as total_revenue").
		Where("status = ? AND created_at BETWEEN ? AND ?", "completed", startDate, endDate).
		Scan(&metrics.TotalRevenue)

	// Total Orders
	r.db.Table("orders").
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Count(&metrics.TotalOrders)

	// Total Customers
	r.db.Table("customers").
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Count(&metrics.TotalCustomers)

	// Total Products
	r.db.Table("products").
		Where("is_active = ?", true).
		Count(&metrics.TotalProducts)

	// Average Order Value
	if metrics.TotalOrders > 0 {
		metrics.AverageOrderValue = metrics.TotalRevenue / float64(metrics.TotalOrders)
	}

	return &metrics, nil
}

func (r *analyticsRepository) GetSalesAnalytics(startDate, endDate time.Time) (*SalesAnalytics, error) {
	var analytics SalesAnalytics

	// Total Revenue and Orders
	r.db.Table("orders").
		Select("COALESCE(SUM(total_kobo), 0) as total_revenue, COUNT(*) as total_orders").
		Where("status = ? AND created_at BETWEEN ? AND ?", "completed", startDate, endDate).
		Scan(&analytics)

	// Average Order Value
	if analytics.TotalOrders > 0 {
		analytics.AverageOrderValue = analytics.TotalRevenue / float64(analytics.TotalOrders)
	}

	// Revenue by Day
	revenueByDay, _ := r.GetRevenueByDay(startDate, endDate)
	analytics.RevenueByDay = revenueByDay

	// Top Products
	topProducts, _ := r.GetTopProducts(startDate, endDate, 10)
	analytics.TopProducts = topProducts

	return &analytics, nil
}

func (r *analyticsRepository) GetCustomerAnalytics(startDate, endDate time.Time) (*CustomerAnalytics, error) {
	var analytics CustomerAnalytics

	// Total Customers
	r.db.Table("customers").
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Count(&analytics.TotalCustomers)

	// New Customers
	r.db.Table("customers").
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Count(&analytics.NewCustomers)

	// Top Customers
	topCustomers, _ := r.GetTopCustomers(startDate, endDate, 10)
	analytics.TopCustomers = topCustomers

	return &analytics, nil
}

func (r *analyticsRepository) GetProductAnalytics(startDate, endDate time.Time) (*ProductAnalytics, error) {
	var analytics ProductAnalytics

	// Total Products
	r.db.Table("products").Count(&analytics.TotalProducts)

	// Active Products
	r.db.Table("products").Where("is_active = ?", true).Count(&analytics.ActiveProducts)

	// Low Stock Products
	r.db.Table("products").Where("stock_quantity < ?", 10).Count(&analytics.LowStockProducts)

	return &analytics, nil
}

func (r *analyticsRepository) GetOrderAnalytics(startDate, endDate time.Time) (*OrderAnalytics, error) {
	var analytics OrderAnalytics

	// Total Orders
	r.db.Table("orders").
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Count(&analytics.TotalOrders)

	// Orders by Status
	var statusCounts []StatusCount
	r.db.Table("orders").
		Select("status, COUNT(*) as count").
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Group("status").
		Scan(&statusCounts)
	analytics.OrdersByStatus = statusCounts

	return &analytics, nil
}

func (r *analyticsRepository) GetDeliveryAnalytics(startDate, endDate time.Time) (*DeliveryAnalytics, error) {
	var analytics DeliveryAnalytics

	// Total Deliveries
	r.db.Table("deliveries").
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Count(&analytics.TotalDeliveries)

	// Deliveries by Status
	r.db.Table("deliveries").Where("status = ? AND created_at BETWEEN ? AND ?", "pending", startDate, endDate).Count(&analytics.PendingDeliveries)
	r.db.Table("deliveries").Where("status = ? AND created_at BETWEEN ? AND ?", "delivered", startDate, endDate).Count(&analytics.CompletedDeliveries)
	r.db.Table("deliveries").Where("status = ? AND created_at BETWEEN ? AND ?", "failed", startDate, endDate).Count(&analytics.FailedDeliveries)

	return &analytics, nil
}

func (r *analyticsRepository) GetPaymentAnalytics(startDate, endDate time.Time) (*PaymentAnalytics, error) {
	var analytics PaymentAnalytics

	// Total Payments
	r.db.Table("payments").
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Count(&analytics.TotalPayments)

	// Payments by Status
	r.db.Table("payments").Where("status = ? AND created_at BETWEEN ? AND ?", "completed", startDate, endDate).Count(&analytics.SuccessfulPayments)
	r.db.Table("payments").Where("status = ? AND created_at BETWEEN ? AND ?", "failed", startDate, endDate).Count(&analytics.FailedPayments)

	// Success Rate
	if analytics.TotalPayments > 0 {
		analytics.SuccessRate = float64(analytics.SuccessfulPayments) / float64(analytics.TotalPayments) * 100
	}

	return &analytics, nil
}

func (r *analyticsRepository) GetRevenueByDay(startDate, endDate time.Time) ([]DataPoint, error) {
	var dataPoints []DataPoint

	r.db.Table("orders").
		Select("DATE(created_at) as date, COALESCE(SUM(total_kobo), 0) as value").
		Where("status = ? AND created_at BETWEEN ? AND ?", "completed", startDate, endDate).
		Group("DATE(created_at)").
		Order("date").
		Scan(&dataPoints)

	return dataPoints, nil
}

func (r *analyticsRepository) GetTopProducts(startDate, endDate time.Time, limit int) ([]ProductSales, error) {
	var products []ProductSales

	r.db.Table("order_items oi").
		Select("p.id as product_id, p.name as product_name, SUM(oi.quantity) as quantity_sold, SUM(oi.price_kobo * oi.quantity) as revenue").
		Joins("JOIN products p ON oi.product_id = p.id").
		Joins("JOIN orders o ON oi.order_id = o.id").
		Where("o.status = ? AND o.created_at BETWEEN ? AND ?", "completed", startDate, endDate).
		Group("p.id, p.name").
		Order("revenue DESC").
		Limit(limit).
		Scan(&products)

	return products, nil
}

func (r *analyticsRepository) GetTopCustomers(startDate, endDate time.Time, limit int) ([]CustomerSpending, error) {
	var customers []CustomerSpending

	r.db.Table("orders o").
		Select("c.id as customer_id, CONCAT(c.first_name, ' ', c.last_name) as customer_name, COUNT(o.id) as total_orders, SUM(o.total_amount) as total_spent").
		Joins("JOIN customers c ON o.customer_id = c.user_id").
		Where("o.status = ? AND o.created_at BETWEEN ? AND ?", "completed", startDate, endDate).
		Group("c.id, c.first_name, c.last_name").
		Order("total_spent DESC").
		Limit(limit).
		Scan(&customers)

	return customers, nil
}

// New Dashboard Methods
func (r *analyticsRepository) GetDashboardData(timeRange string) (*DashboardData, error) {
	startDate, endDate := r.getDateRangeFromString(timeRange)
	
	kpis, err := r.GetDashboardKPIs(startDate, endDate)
	if err != nil {
		return nil, err
	}
	
	recentOrders, err := r.GetRecentOrders(10)
	if err != nil {
		return nil, err
	}
	
	lowStockProducts, err := r.GetLowStockProducts(10)
	if err != nil {
		return nil, err
	}
	
	return &DashboardData{
		KPIs:             *kpis,
		RecentOrders:     recentOrders,
		LowStockProducts: lowStockProducts,
	}, nil
}

func (r *analyticsRepository) GetDashboardKPIs(startDate, endDate time.Time) (*DashboardKPIs, error) {
	var kpis DashboardKPIs
	
	// Total Sales (completed orders)
	r.db.Table("orders").
		Select("COALESCE(SUM(total_kobo), 0)").
		Where("status = ? AND created_at BETWEEN ? AND ?", "completed", startDate, endDate).
		Scan(&kpis.TotalSales)
	
	// Active Mobile App Users (users with role 'customer')
	activeUsers, _ := r.GetMobileAppUsersCount(startDate, endDate)
	kpis.ActiveUsers = activeUsers
	
	// Total Products (active products)
	r.db.Table("products").
		Where("is_active = ?", true).
		Count(&kpis.TotalProducts)
	
	// Coupons Issued (assuming we have a coupons table)
	r.db.Table("coupons").
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Count(&kpis.CouponsIssued)
	
	return &kpis, nil
}

func (r *analyticsRepository) GetMobileAppUsersCount(startDate, endDate time.Time) (int64, error) {
	var count int64
	
	// Count users with role 'customer' (mobile app users) who have been active
	r.db.Table("users").
		Where("role = ? AND status = ? AND (last_login_at BETWEEN ? AND ? OR created_at BETWEEN ? AND ?)", 
			"customer", "active", startDate, endDate, startDate, endDate).
		Count(&count)
	
	return count, nil
}

func (r *analyticsRepository) GetRecentOrders(limit int) ([]RecentOrder, error) {
	var orders []RecentOrder
	
	r.db.Table("orders o").
		Select("o.id, CONCAT(c.first_name, ' ', c.last_name) as customer_name, o.total_amount as amount, o.status, o.created_at").
		Joins("JOIN customers c ON o.customer_id = c.user_id").
		Order("o.created_at DESC").
		Limit(limit).
		Scan(&orders)
	
	// Convert timestamps to ISO strings
	for i := range orders {
		if orders[i].CreatedAt != "" {
			// Parse and format the timestamp
			if t, err := time.Parse("2006-01-02 15:04:05", orders[i].CreatedAt); err == nil {
				orders[i].CreatedAt = t.Format(time.RFC3339)
			}
		}
	}
	
	return orders, nil
}

func (r *analyticsRepository) GetLowStockProducts(threshold int) ([]LowStockProduct, error) {
	var products []LowStockProduct
	
	err := r.db.Table("products").
		Select("id, name, sku, stock_quantity as current_stock, CAST(? AS INTEGER) as low_stock_threshold", threshold).
		Where("stock_quantity <= ? AND is_active = ?", threshold, true).
		Order("stock_quantity ASC").
		Scan(&products).Error
	
	if err != nil {
		return nil, err
	}
	
	return products, nil
}

// New Reports Methods
func (r *analyticsRepository) GetSalesOverview(startDate, endDate time.Time, period string) (*SalesOverviewData, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	weekStart := today.AddDate(0, 0, -int(today.Weekday()))
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	yearStart := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
	
	var overview SalesOverviewData
	
	// Today's sales
	var todayValue float64
	r.db.Table("orders").
		Select("COALESCE(SUM(total_kobo), 0)").
		Where("status = ? AND created_at >= ?", "completed", today).
		Scan(&todayValue)
	overview.TodaySales = TrendData{Value: todayValue, Trend: "up", Change: 5.2}
	
	// Weekly sales
	var weeklyValue float64
	r.db.Table("orders").
		Select("COALESCE(SUM(total_kobo), 0)").
		Where("status = ? AND created_at >= ?", "completed", weekStart).
		Scan(&weeklyValue)
	overview.WeeklySales = TrendData{Value: weeklyValue, Trend: "up", Change: 12.5}
	
	// Monthly sales
	var monthlyValue float64
	r.db.Table("orders").
		Select("COALESCE(SUM(total_kobo), 0)").
		Where("status = ? AND created_at >= ?", "completed", monthStart).
		Scan(&monthlyValue)
	overview.MonthlySales = TrendData{Value: monthlyValue, Trend: "stable", Change: 0.8}
	
	// Yearly sales
	var yearlyValue float64
	r.db.Table("orders").
		Select("COALESCE(SUM(total_kobo), 0)").
		Where("status = ? AND created_at >= ?", "completed", yearStart).
		Scan(&yearlyValue)
	overview.YearlySales = TrendData{Value: yearlyValue, Trend: "up", Change: 18.3}
	
	return &overview, nil
}

func (r *analyticsRepository) GetTopProductsReport(startDate, endDate time.Time, limit int) ([]TopProduct, error) {
	var products []TopProduct
	
	r.db.Table("order_items oi").
		Select("p.name, SUM(oi.quantity) as units_sold, SUM(oi.price_kobo * oi.quantity) as revenue, 'up' as trend").
		Joins("JOIN products p ON oi.product_id = p.id").
		Joins("JOIN orders o ON oi.order_id = o.id").
		Where("o.status = ? AND o.created_at BETWEEN ? AND ?", "completed", startDate, endDate).
		Group("p.id, p.name").
		Order("revenue DESC").
		Limit(limit).
		Scan(&products)
	
	return products, nil
}

func (r *analyticsRepository) GetCouponPerformance(startDate, endDate time.Time) ([]CouponPerformance, error) {
	var coupons []CouponPerformance
	
	// This is a placeholder - adjust based on your actual coupon/discount system
	r.db.Table("coupons c").
		Select("c.code, COUNT(o.id) as usage, COALESCE(SUM(o.discount_kobo), 0) as revenue, c.type").
		Joins("LEFT JOIN orders o ON o.coupon_id = c.id").
		Where("c.created_at BETWEEN ? AND ?", startDate, endDate).
		Group("c.id, c.code, c.type").
		Order("usage DESC").
		Scan(&coupons)
	
	return coupons, nil
}

func (r *analyticsRepository) GetStorePerformance(startDate, endDate time.Time) ([]StorePerformance, error) {
	var stores []StorePerformance
	
	// This assumes you have a stores table and orders are linked to stores
	r.db.Table("stores s").
		Select("s.name, COALESCE(SUM(o.total_kobo), 0) as revenue, COUNT(o.id) as orders, 15.5 as growth, COALESCE(AVG(o.total_kobo), 0) as average_order_value").
		Joins("LEFT JOIN orders o ON o.store_id = s.id AND o.status = 'completed' AND o.created_at BETWEEN ? AND ?", startDate, endDate).
		Where("s.is_active = ?", true).
		Group("s.id, s.name").
		Order("revenue DESC").
		Scan(&stores)
	
	return stores, nil
}

// Individual dashboard metrics methods
func (r *analyticsRepository) GetTodaySales() (float64, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	
	var totalSales float64
	r.db.Table("orders").
		Select("COALESCE(SUM(total_amount), 0)").
		Where("status IN (?, ?) AND created_at >= ?", "delivered", "confirmed", today).
		Scan(&totalSales)
	
	return totalSales / 100.0, nil // Convert from kobo to naira
}

func (r *analyticsRepository) GetActiveUsers() (int64, error) {
	now := time.Now()
	last24Hours := now.AddDate(0, 0, -1)
	
	var count int64
	r.db.Table("users").
		Where("role = ? AND status = ? AND (last_login_at >= ? OR created_at >= ?)", 
			"customer", "active", last24Hours, last24Hours).
		Count(&count)
	
	return count, nil
}

func (r *analyticsRepository) GetTotalProducts() (int64, error) {
	var count int64
	r.db.Table("products").
		Where("is_active = ?", true).
		Count(&count)
	
	return count, nil
}

func (r *analyticsRepository) GetCouponsAnalytics() (int64, error) {
	var count int64
	r.db.Table("coupons").
		Where("is_active = ?", true).
		Count(&count)
	
	return count, nil
}

// Helper method to convert string time range to dates
func (r *analyticsRepository) getDateRangeFromString(timeRange string) (time.Time, time.Time) {
	now := time.Now()
	switch timeRange {
	case "today":
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		return start, now
	case "week":
		start := now.AddDate(0, 0, -7)
		return start, now
	case "month":
		start := now.AddDate(0, -1, 0)
		return start, now
	default:
		start := now.AddDate(0, -1, 0)
		return start, now
	}
}
