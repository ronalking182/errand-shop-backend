package analytics

import (
	"time"

	"gorm.io/gorm"
)

// revenueOrderStatuses are order rows counted toward revenue (amounts live in orders.total_amount, kobo).
// Align with GetTodaySales(); avoids legacy "completed" / total_kobo which do not exist in the orders schema.
var revenueOrderStatuses = []string{"delivered", "confirmed"}

// recentOrderScan maps DB rows before formatting API RecentOrder values (avoid scanning timestamps into strings).
type recentOrderScan struct {
	ID           string    `gorm:"column:id"`
	CustomerName string    `gorm:"column:customer_name"`
	AmountKobo   int64     `gorm:"column:amount"`
	Status       string    `gorm:"column:status"`
	CreatedAt    time.Time `gorm:"column:created_at"`
}

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
	GetRevenueAndOrderTotals(startDate, endDate time.Time) (revenue float64, orderCount int64, err error)

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

	// Total Revenue (total_amount is kobo; legacy queries used missing total_kobo + status "completed")
	if err := r.db.Table("orders").
		Select("COALESCE(SUM(total_amount), 0) as total_revenue").
		Where("status IN ?", revenueOrderStatuses).
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Scan(&metrics.TotalRevenue).Error; err != nil {
		return nil, err
	}
	metrics.TotalRevenue /= 100.0

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

	// Average Order Value (revenue numerator is fulfilled subset; denominator is all orders in range — legacy behavior)
	if metrics.TotalOrders > 0 {
		metrics.AverageOrderValue = metrics.TotalRevenue / float64(metrics.TotalOrders)
	}

	return &metrics, nil
}

func (r *analyticsRepository) GetSalesAnalytics(startDate, endDate time.Time) (*SalesAnalytics, error) {
	var analytics SalesAnalytics

	// Total Revenue and Orders (total_amount in kobo)
	if err := r.db.Table("orders").
		Select("COALESCE(SUM(total_amount), 0) as total_revenue, COUNT(*) as total_orders").
		Where("status IN ?", revenueOrderStatuses).
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Scan(&analytics).Error; err != nil {
		return nil, err
	}
	analytics.TotalRevenue /= 100.0

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

	if err := r.db.Table("orders").
		Select("DATE(created_at) AS date, COALESCE(SUM(total_amount), 0) AS value, COUNT(*) AS count").
		Where("status IN ?", revenueOrderStatuses).
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Group("DATE(created_at)").
		Order("date").
		Scan(&dataPoints).Error; err != nil {
		return nil, err
	}

	for i := range dataPoints {
		dataPoints[i].Value /= 100.0
	}

	return dataPoints, nil
}

func (r *analyticsRepository) GetRevenueAndOrderTotals(startDate, endDate time.Time) (float64, int64, error) {
	type row struct {
		TotalKobo float64 `gorm:"column:total"`
		N         int64   `gorm:"column:n"`
	}
	var out row
	if err := r.db.Table("orders").
		Select("COALESCE(SUM(total_amount), 0) AS total, COUNT(*) AS n").
		Where("status IN ?", revenueOrderStatuses).
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Scan(&out).Error; err != nil {
		return 0, 0, err
	}
	return out.TotalKobo / 100.0, out.N, nil
}

func (r *analyticsRepository) GetTopProducts(startDate, endDate time.Time, limit int) ([]ProductSales, error) {
	var products []ProductSales

	if err := r.db.Table("order_items oi").
		Select("p.id as product_id, p.name as product_name, SUM(oi.quantity) as quantity_sold, SUM(COALESCE(oi.unit_price, 0) * oi.quantity) as revenue").
		Joins("JOIN products p ON oi.product_id = p.id").
		Joins("JOIN orders o ON oi.order_id = o.id").
		Where("o.status IN ?", revenueOrderStatuses).
		Where("o.created_at BETWEEN ? AND ?", startDate, endDate).
		Group("p.id, p.name").
		Order("revenue DESC").
		Limit(limit).
		Scan(&products).Error; err != nil {
		return nil, err
	}

	for i := range products {
		products[i].Revenue /= 100.0
	}

	return products, nil
}

func (r *analyticsRepository) GetTopCustomers(startDate, endDate time.Time, limit int) ([]CustomerSpending, error) {
	var customers []CustomerSpending

	if err := r.db.Table("orders o").
		Select("c.id as customer_id, CONCAT(c.first_name, ' ', c.last_name) as customer_name, COUNT(o.id) as total_orders, SUM(o.total_amount) as total_spent").
		Joins("JOIN customers c ON o.customer_id = c.user_id").
		Where("o.status IN ?", revenueOrderStatuses).
		Where("o.created_at BETWEEN ? AND ?", startDate, endDate).
		Group("c.id, c.first_name, c.last_name").
		Order("total_spent DESC").
		Limit(limit).
		Scan(&customers).Error; err != nil {
		return nil, err
	}

	for i := range customers {
		customers[i].TotalSpent /= 100.0
	}

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

	// Total Sales (total_amount kobo → naira on dashboard KPIs)
	var totalKobo float64
	if err := r.db.Table("orders").
		Select("COALESCE(SUM(total_amount), 0)").
		Where("status IN ?", revenueOrderStatuses).
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Scan(&totalKobo).Error; err != nil {
		return nil, err
	}
	kpis.TotalSales = totalKobo / 100.0

	activeUsers, err := r.GetMobileAppUsersCount(startDate, endDate)
	if err != nil {
		activeUsers = 0
	}
	kpis.ActiveUsers = activeUsers

	if err := r.db.Table("products").
		Where("is_active = ?", true).
		Count(&kpis.TotalProducts).Error; err != nil {
		kpis.TotalProducts = 0
	}

	if err := r.db.Table("coupons").
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Count(&kpis.CouponsIssued).Error; err != nil {
		kpis.CouponsIssued = 0
	}

	return &kpis, nil
}

func (r *analyticsRepository) GetMobileAppUsersCount(startDate, endDate time.Time) (int64, error) {
	var count int64

	db := r.db.Table("users").
		Where("role = ? AND status = ? AND (last_login_at BETWEEN ? AND ? OR created_at BETWEEN ? AND ?)",
			"customer", "active", startDate, endDate, startDate, endDate).
		Count(&count)

	return count, db.Error
}

func (r *analyticsRepository) GetRecentOrders(limit int) ([]RecentOrder, error) {
	var rows []recentOrderScan

	err := r.db.Table("orders o").
		Select(`
			o.id::text AS id,
			TRIM(CONCAT(COALESCE(c.first_name, ''), ' ', COALESCE(c.last_name, ''))) AS customer_name,
			o.total_amount AS amount,
			o.status::text AS status,
			o.created_at`).
		Joins("LEFT JOIN customers c ON o.customer_id = c.user_id").
		Order("o.created_at DESC").
		Limit(limit).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	out := make([]RecentOrder, len(rows))
	for i, row := range rows {
		name := row.CustomerName
		if name == "" {
			name = "Guest"
		}
		out[i] = RecentOrder{
			ID:           row.ID,
			CustomerName: name,
			Amount:       float64(row.AmountKobo) / 100.0,
			Status:       row.Status,
			CreatedAt:    row.CreatedAt.UTC().Format(time.RFC3339),
		}
	}

	return out, nil
}

func (r *analyticsRepository) GetLowStockProducts(threshold int) ([]LowStockProduct, error) {
	var products []LowStockProduct

	err := r.db.Table("products").
		Select("id::text AS id, COALESCE(TRIM(name), '') AS name, COALESCE(TRIM(sku), '') AS sku, stock_quantity AS current_stock").
		Where("stock_quantity <= ? AND is_active = ?", threshold, true).
		Order("stock_quantity ASC").
		Scan(&products).Error

	if err != nil {
		return nil, err
	}
	for i := range products {
		products[i].LowStockThreshold = threshold
		if products[i].Name == "" {
			products[i].Name = "Unnamed product"
		}
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
	if err := r.db.Table("orders").
		Select("COALESCE(SUM(total_amount), 0)").
		Where("status IN ?", revenueOrderStatuses).
		Where("created_at >= ?", today).
		Scan(&todayValue).Error; err != nil {
		return nil, err
	}
	overview.TodaySales = TrendData{Value: todayValue / 100.0, Trend: "up", Change: 5.2}

	// Weekly sales
	var weeklyValue float64
	if err := r.db.Table("orders").
		Select("COALESCE(SUM(total_amount), 0)").
		Where("status IN ?", revenueOrderStatuses).
		Where("created_at >= ?", weekStart).
		Scan(&weeklyValue).Error; err != nil {
		return nil, err
	}
	overview.WeeklySales = TrendData{Value: weeklyValue / 100.0, Trend: "up", Change: 12.5}

	// Monthly sales
	var monthlyValue float64
	if err := r.db.Table("orders").
		Select("COALESCE(SUM(total_amount), 0)").
		Where("status IN ?", revenueOrderStatuses).
		Where("created_at >= ?", monthStart).
		Scan(&monthlyValue).Error; err != nil {
		return nil, err
	}
	overview.MonthlySales = TrendData{Value: monthlyValue / 100.0, Trend: "stable", Change: 0.8}

	// Yearly sales
	var yearlyValue float64
	if err := r.db.Table("orders").
		Select("COALESCE(SUM(total_amount), 0)").
		Where("status IN ?", revenueOrderStatuses).
		Where("created_at >= ?", yearStart).
		Scan(&yearlyValue).Error; err != nil {
		return nil, err
	}
	overview.YearlySales = TrendData{Value: yearlyValue / 100.0, Trend: "up", Change: 18.3}

	return &overview, nil
}

func (r *analyticsRepository) GetTopProductsReport(startDate, endDate time.Time, limit int) ([]TopProduct, error) {
	var products []TopProduct

	r.db.Table("order_items oi").
		Select("p.name, SUM(oi.quantity) as units_sold, SUM(COALESCE(oi.unit_price, 0) * oi.quantity) as revenue, 'up' as trend").
		Joins("JOIN products p ON oi.product_id = p.id").
		Joins("JOIN orders o ON oi.order_id = o.id").
		Where("o.status IN ?", revenueOrderStatuses).
		Where("o.created_at BETWEEN ? AND ?", startDate, endDate).
		Group("p.id, p.name").
		Order("revenue DESC").
		Limit(limit).
		Scan(&products)

	for i := range products {
		products[i].Revenue /= 100.0
	}

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
		Select("s.name, COALESCE(SUM(o.total_amount), 0) as revenue, COUNT(o.id) as orders, 15.5 as growth, COALESCE(AVG(o.total_amount), 0) as average_order_value").
		Joins(`LEFT JOIN orders o ON o.store_id = s.id AND o.status IN ('delivered','confirmed') AND o.created_at BETWEEN ? AND ?`, startDate, endDate).
		Where("s.is_active = ?", true).
		Group("s.id, s.name").
		Order("revenue DESC").
		Scan(&stores)

	for i := range stores {
		stores[i].Revenue /= 100.0
		stores[i].AverageOrderValue /= 100.0
	}

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
