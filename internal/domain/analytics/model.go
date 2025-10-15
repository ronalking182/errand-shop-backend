package analytics

import (
	"gorm.io/gorm"
	"time"
)

type ReportType string
type MetricType string
type TimeRange string

const (
	// Report Types
	ReportSales     ReportType = "sales"
	ReportProducts  ReportType = "products"
	ReportCustomers ReportType = "customers"
	ReportOrders    ReportType = "orders"
	ReportDelivery  ReportType = "delivery"
	ReportPayments  ReportType = "payments"

	// Metric Types
	MetricRevenue    MetricType = "revenue"
	MetricOrders     MetricType = "orders"
	MetricCustomers  MetricType = "customers"
	MetricProducts   MetricType = "products"
	MetricConversion MetricType = "conversion"

	// Time Ranges
	TimeRangeToday   TimeRange = "today"
	TimeRangeWeek    TimeRange = "week"
	TimeRangeMonth   TimeRange = "month"
	TimeRangeQuarter TimeRange = "quarter"
	TimeRangeYear    TimeRange = "year"
	TimeRangeCustom  TimeRange = "custom"
)

// Dashboard metrics
type DashboardMetrics struct {
	TotalRevenue      float64 `json:"totalRevenue"`
	TotalOrders       int64   `json:"totalOrders"`
	TotalCustomers    int64   `json:"totalCustomers"`
	TotalProducts     int64   `json:"totalProducts"`
	AverageOrderValue float64 `json:"averageOrderValue"`
	ConversionRate    float64 `json:"conversionRate"`
	GrowthRate        float64 `json:"growthRate"`
}

// Time series data point
type DataPoint struct {
	Date  time.Time `json:"date"`
	Value float64   `json:"value"`
	Count int64     `json:"count,omitempty"`
}

// Sales analytics
type SalesAnalytics struct {
	TotalRevenue      float64         `json:"totalRevenue"`
	TotalOrders       int64           `json:"totalOrders"`
	AverageOrderValue float64         `json:"averageOrderValue"`
	RevenueByDay      []DataPoint     `json:"revenueByDay"`
	OrdersByDay       []DataPoint     `json:"ordersByDay"`
	TopProducts       []ProductSales  `json:"topProducts"`
	TopCategories     []CategorySales `json:"topCategories"`
}

type ProductSales struct {
	ProductID    uint    `json:"productId"`
	ProductName  string  `json:"productName"`
	QuantitySold int64   `json:"quantitySold"`
	Revenue      float64 `json:"revenue"`
}

type CategorySales struct {
	CategoryID   uint    `json:"categoryId"`
	CategoryName string  `json:"categoryName"`
	QuantitySold int64   `json:"quantitySold"`
	Revenue      float64 `json:"revenue"`
}

// Customer analytics
type CustomerAnalytics struct {
	TotalCustomers     int64              `json:"totalCustomers"`
	NewCustomers       int64              `json:"newCustomers"`
	ReturningCustomers int64              `json:"returningCustomers"`
	CustomersByDay     []DataPoint        `json:"customersByDay"`
	TopCustomers       []CustomerSpending `json:"topCustomers"`
	CustomerRetention  float64            `json:"customerRetention"`
}

type CustomerSpending struct {
	CustomerID   uint    `json:"customerId"`
	CustomerName string  `json:"customerName"`
	TotalOrders  int64   `json:"totalOrders"`
	TotalSpent   float64 `json:"totalSpent"`
}

// Product analytics
type ProductAnalytics struct {
	TotalProducts      int64          `json:"totalProducts"`
	ActiveProducts     int64          `json:"activeProducts"`
	LowStockProducts   int64          `json:"lowStockProducts"`
	TopSellingProducts []ProductSales `json:"topSellingProducts"`
	SlowMovingProducts []ProductSales `json:"slowMovingProducts"`
	InventoryValue     float64        `json:"inventoryValue"`
}

// Order analytics
type OrderAnalytics struct {
	TotalOrders       int64         `json:"totalOrders"`
	PendingOrders     int64         `json:"pendingOrders"`
	CompletedOrders   int64         `json:"completedOrders"`
	CancelledOrders   int64         `json:"cancelledOrders"`
	OrdersByStatus    []StatusCount `json:"ordersByStatus"`
	OrdersByDay       []DataPoint   `json:"ordersByDay"`
	AverageOrderValue float64       `json:"averageOrderValue"`
}

type StatusCount struct {
	Status string `json:"status"`
	Count  int64  `json:"count"`
}

// Delivery analytics
type DeliveryAnalytics struct {
	TotalDeliveries     int64       `json:"totalDeliveries"`
	PendingDeliveries   int64       `json:"pendingDeliveries"`
	CompletedDeliveries int64       `json:"completedDeliveries"`
	FailedDeliveries    int64       `json:"failedDeliveries"`
	AverageDeliveryTime float64     `json:"averageDeliveryTime"`
	DeliverysByZone     []ZoneStats `json:"deliverysByZone"`
	DeliverysByDay      []DataPoint `json:"deliverysByDay"`
}

type ZoneStats struct {
	ZoneID      uint    `json:"zoneId"`
	ZoneName    string  `json:"zoneName"`
	Deliveries  int64   `json:"deliveries"`
	AverageTime float64 `json:"averageTime"`
}

// Payment analytics
type PaymentAnalytics struct {
	TotalPayments      int64         `json:"totalPayments"`
	SuccessfulPayments int64         `json:"successfulPayments"`
	FailedPayments     int64         `json:"failedPayments"`
	RefundedPayments   int64         `json:"refundedPayments"`
	PaymentsByMethod   []MethodStats `json:"paymentsByMethod"`
	PaymentsByDay      []DataPoint   `json:"paymentsByDay"`
	SuccessRate        float64       `json:"successRate"`
}

type MethodStats struct {
	Method string  `json:"method"`
	Count  int64   `json:"count"`
	Amount float64 `json:"amount"`
}

type TimePeriod string

const (
	PeriodDaily   TimePeriod = "daily"
	PeriodWeekly  TimePeriod = "weekly"
	PeriodMonthly TimePeriod = "monthly"
	PeriodYearly  TimePeriod = "yearly"
	PeriodCustom  TimePeriod = "custom"
)

// Analytics Event for tracking user actions
type AnalyticsEvent struct {
	ID         uint                   `gorm:"primaryKey" json:"id"`
	UserID     *uint                  `json:"userId,omitempty"`
	UserType   string                 `gorm:"type:varchar(20)" json:"userType,omitempty"`
	EventType  string                 `gorm:"type:varchar(50);not null" json:"eventType"`
	EventName  string                 `gorm:"type:varchar(100);not null" json:"eventName"`
	Properties map[string]interface{} `gorm:"type:jsonb" json:"properties,omitempty"`
	SessionID  string                 `gorm:"type:varchar(100)" json:"sessionId,omitempty"`
	IPAddress  string                 `gorm:"type:varchar(45)" json:"ipAddress,omitempty"`
	UserAgent  string                 `gorm:"type:text" json:"userAgent,omitempty"`
	CreatedAt  time.Time              `json:"createdAt"`
}

// Saved Reports
type SavedReport struct {
	ID          uint                   `gorm:"primaryKey" json:"id"`
	Name        string                 `gorm:"size:200;not null" json:"name"`
	Description string                 `gorm:"type:text" json:"description,omitempty"`
	ReportType  ReportType             `gorm:"type:varchar(20);not null" json:"reportType"`
	Filters     map[string]interface{} `gorm:"type:jsonb" json:"filters,omitempty"`
	Schedule    string                 `gorm:"type:varchar(50)" json:"schedule,omitempty"` // cron expression
	CreatedBy   uint                   `gorm:"not null" json:"createdBy"`
	IsActive    bool                   `gorm:"default:true" json:"isActive"`
	CreatedAt   time.Time              `json:"createdAt"`
	UpdatedAt   time.Time              `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt         `gorm:"index" json:"-"`
}

// Dashboard Widgets
type DashboardWidget struct {
	ID         uint                   `gorm:"primaryKey" json:"id"`
	UserID     uint                   `gorm:"not null" json:"userId"`
	WidgetType string                 `gorm:"type:varchar(50);not null" json:"widgetType"`
	Title      string                 `gorm:"size:200;not null" json:"title"`
	Config     map[string]interface{} `gorm:"type:jsonb" json:"config,omitempty"`
	Position   int                    `gorm:"default:0" json:"position"`
	IsVisible  bool                   `gorm:"default:true" json:"isVisible"`
	CreatedAt  time.Time              `json:"createdAt"`
	UpdatedAt  time.Time              `json:"updatedAt"`
}
