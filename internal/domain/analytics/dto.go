package analytics

import "time"

// Common data structures
type PeriodData struct {
	Period string  `json:"period"`
	Value  float64 `json:"value"`
	Count  int64   `json:"count"`
}

type ProductSalesData struct {
	ProductID    uint    `json:"productId"`
	ProductName  string  `json:"productName"`
	QuantitySold int64   `json:"quantitySold"`
	Revenue      float64 `json:"revenue"`
	Category     string  `json:"category"`
}

type CategoryPerformanceData struct {
	CategoryID   uint    `json:"categoryId"`
	CategoryName string  `json:"categoryName"`
	TotalSales   int64   `json:"totalSales"`
	Revenue      float64 `json:"revenue"`
	GrowthRate   float64 `json:"growthRate"`
}

type CustomerData struct {
	CustomerID    uint    `json:"customerId"`
	CustomerName  string  `json:"customerName"`
	TotalOrders   int64   `json:"totalOrders"`
	TotalSpent    float64 `json:"totalSpent"`
	LastOrderDate string  `json:"lastOrderDate"`
}

// Request DTOs
type AnalyticsEventRequest struct {
	EventType  string                 `json:"eventType" validate:"required"`
	EventName  string                 `json:"eventName" validate:"required"`
	Properties map[string]interface{} `json:"properties,omitempty"`
	SessionID  string                 `json:"sessionId,omitempty"`
}

type AnalyticsRequest struct {
	TimeRange TimeRange  `json:"timeRange" validate:"required,oneof=today week month quarter year custom"`
	StartDate *time.Time `json:"startDate,omitempty"`
	EndDate   *time.Time `json:"endDate,omitempty"`
	StoreID   *uint      `json:"storeId,omitempty"`
}

type ReportRequest struct {
	ReportType ReportType `json:"reportType" validate:"required"`
	TimeRange  TimeRange  `json:"timeRange" validate:"required"`
	StartDate  *time.Time `json:"startDate,omitempty"`
	EndDate    *time.Time `json:"endDate,omitempty"`
	StoreID    *uint      `json:"storeId,omitempty"`
	Limit      int        `json:"limit,omitempty"`
}

type CustomReportRequest struct {
	Name        string                 `json:"name" validate:"required,max=100"`
	Description string                 `json:"description,omitempty"`
	Metrics     []MetricType           `json:"metrics" validate:"required,min=1"`
	Filters     map[string]interface{} `json:"filters,omitempty"`
	TimeRange   TimeRange              `json:"timeRange" validate:"required"`
	StartDate   *time.Time             `json:"startDate,omitempty"`
	EndDate     *time.Time             `json:"endDate,omitempty"`
}

type SaveReportRequest struct {
	Name        string                 `json:"name" validate:"required,max=200"`
	Description string                 `json:"description,omitempty"`
	ReportType  ReportType             `json:"reportType" validate:"required"`
	Filters     map[string]interface{} `json:"filters,omitempty"`
	Schedule    string                 `json:"schedule,omitempty"`
}

type DashboardWidgetRequest struct {
	WidgetType string                 `json:"widgetType" validate:"required"`
	Title      string                 `json:"title" validate:"required,max=200"`
	Config     map[string]interface{} `json:"config,omitempty"`
	Position   int                    `json:"position"`
}

// Dashboard Data Structures (matching frontend requirements)
type DashboardKPIs struct {
	TotalSales    float64 `json:"totalSales"`    // Total sales amount in currency
	ActiveUsers   int64   `json:"activeUsers"`   // Count of active mobile app users
	TotalProducts int64   `json:"totalProducts"` // Total number of products
	CouponsIssued int64   `json:"couponsIssued"` // Number of coupons issued
}

type RecentOrder struct {
	ID           string  `json:"id"`
	CustomerName string  `json:"customerName"`
	Amount       float64 `json:"amount"`
	Status       string  `json:"status"` // 'pending' | 'processing' | 'shipped' | 'delivered' | 'cancelled'
	CreatedAt    string  `json:"createdAt"` // ISO date string
}

type LowStockProduct struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	SKU                string `json:"sku"`
	CurrentStock       int    `json:"currentStock"`
	LowStockThreshold  int    `json:"lowStockThreshold"`
}

type DashboardData struct {
	KPIs              DashboardKPIs      `json:"kpis"`
	RecentOrders      []RecentOrder      `json:"recentOrders"`
	LowStockProducts  []LowStockProduct  `json:"lowStockProducts"`
}

// Response DTOs
type DashboardResponse struct {
	Success bool          `json:"success"`
	Data    DashboardData `json:"data"`
}

// Reports Data Structures (matching frontend requirements)
type SalesOverviewData struct {
	TodaySales   TrendData `json:"todaySales"`
	WeeklySales  TrendData `json:"weeklySales"`
	MonthlySales TrendData `json:"monthlySales"`
	YearlySales  TrendData `json:"yearlySales"`
}

type TrendData struct {
	Value  float64 `json:"value"`
	Trend  string  `json:"trend"`  // "up" | "down" | "stable"
	Change float64 `json:"change"` // percentage change
}

type TopProduct struct {
	Name      string  `json:"name"`
	UnitsSold int64   `json:"unitsSold"`
	Revenue   float64 `json:"revenue"`
	Trend     string  `json:"trend"` // "up" | "down" | "stable"
}

type CouponPerformance struct {
	Code    string  `json:"code"`
	Usage   int64   `json:"usage"`
	Revenue float64 `json:"revenue"`
	Type    string  `json:"type"` // "percentage" | "fixed"
}

type StorePerformance struct {
	Name              string  `json:"name"`
	Revenue           float64 `json:"revenue"`
	Orders            int64   `json:"orders"`
	Growth            float64 `json:"growth"`
	AverageOrderValue float64 `json:"averageOrderValue"`
}

type SalesReportResponse struct {
	Success bool               `json:"success"`
	Data    SalesOverviewData  `json:"data"`
}

type ProductReportResponse struct {
	Success bool           `json:"success"`
	Data    []ProductSales `json:"data"`
}

type CustomerReportResponse struct {
	Success bool                `json:"success"`
	Data    CustomerAnalytics   `json:"data"`
}

type OrderReportResponse struct {
	Success bool           `json:"success"`
	Data    OrderAnalytics `json:"data"`
}

type DeliveryReportResponse struct {
	Success bool              `json:"success"`
	Data    DeliveryAnalytics `json:"data"`
}

type PaymentReportResponse struct {
	Success bool             `json:"success"`
	Data    PaymentAnalytics `json:"data"`
}

// Individual Dashboard Endpoint Response DTOs
type TodaySalesResponse struct {
	Success bool            `json:"success"`
	Data    TodaySalesData  `json:"data"`
}

type TodaySalesData struct {
	Amount       float64 `json:"amount"`       // Total sales amount in kobo
	Currency     string  `json:"currency"`     // "NGN"
	OrdersCount  int64   `json:"ordersCount"`  // Number of orders today
	GrowthRate   float64 `json:"growthRate"`   // Percentage growth compared to yesterday
	Trend        string  `json:"trend"`        // "up" | "down" | "stable"
	LastUpdated  string  `json:"lastUpdated"`  // ISO timestamp
}

type ActiveUsersResponse struct {
	Success bool             `json:"success"`
	Data    ActiveUsersData  `json:"data"`
}

type ActiveUsersData struct {
	Count        int64   `json:"count"`        // Number of active users
	GrowthRate   float64 `json:"growthRate"`   // Percentage growth
	Trend        string  `json:"trend"`        // "up" | "down" | "stable"
	LastUpdated  string  `json:"lastUpdated"`  // ISO timestamp
}

type TotalProductsResponse struct {
	Success bool               `json:"success"`
	Data    TotalProductsData  `json:"data"`
}

type TotalProductsData struct {
	Total        int64   `json:"total"`        // Total number of products
	Active       int64   `json:"active"`       // Active products
	Inactive     int64   `json:"inactive"`     // Inactive products
	LowStock     int64   `json:"lowStock"`     // Products with low stock
	LastUpdated  string  `json:"lastUpdated"`  // ISO timestamp
}

type CouponsAnalyticsResponse struct {
	Success bool                  `json:"success"`
	Data    CouponsAnalyticsData  `json:"data"`
}

type CouponsAnalyticsData struct {
	TotalIssued    int64                 `json:"totalIssued"`    // Total coupons issued
	TotalUsed      int64                 `json:"totalUsed"`      // Total coupons used
	ActiveCount    int64                 `json:"activeCount"`    // Currently active coupons
	UsageRate      float64               `json:"usageRate"`      // Usage rate percentage
	TopPerforming  []CouponPerformance   `json:"topPerforming"`  // Top 5 performing coupons
	LastUpdated    string                `json:"lastUpdated"`    // ISO timestamp
}

type RecentOrdersResponse struct {
	Success bool              `json:"success"`
	Data    RecentOrdersData  `json:"data"`
}

type RecentOrdersData struct {
	Orders       []RecentOrder `json:"orders"`       // List of recent orders
	TotalCount   int64         `json:"totalCount"`   // Total orders count
	LastUpdated  string        `json:"lastUpdated"`  // ISO timestamp
}

type LowStockAlertsResponse struct {
	Success bool                `json:"success"`
	Data    LowStockAlertsData  `json:"data"`
}

type LowStockAlertsData struct {
	Products     []LowStockProduct `json:"products"`     // List of low stock products
	TotalCount   int64             `json:"totalCount"`   // Total low stock products
	CriticalCount int64            `json:"criticalCount"` // Products with critical stock levels
	LastUpdated  string            `json:"lastUpdated"`  // ISO timestamp
}

type SalesOverviewResponse struct {
	Success bool               `json:"success"`
	Data    SalesOverviewChart `json:"data"`
}

type SalesOverviewChart struct {
	Period       string        `json:"period"`       // "daily" | "weekly" | "monthly"
	ChartData    []ChartPoint  `json:"chartData"`    // Time series data for chart
	TotalRevenue float64       `json:"totalRevenue"` // Total revenue for the period
	TotalOrders  int64         `json:"totalOrders"`  // Total orders for the period
	AverageOrder float64       `json:"averageOrder"` // Average order value
	GrowthRate   float64       `json:"growthRate"`   // Growth rate compared to previous period
	LastUpdated  string        `json:"lastUpdated"`  // ISO timestamp
}

type ChartPoint struct {
	Label    string  `json:"label"`    // Date/time label
	Value    float64 `json:"value"`    // Sales value
	Orders   int64   `json:"orders"`   // Number of orders
	Date     string  `json:"date"`     // ISO date string
}
