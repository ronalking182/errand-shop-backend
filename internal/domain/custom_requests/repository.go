package custom_requests

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	// Custom Request operations
	CreateCustomRequest(req *CustomRequest) error
	GetCustomRequestByID(id uuid.UUID) (*CustomRequest, error)
	GetCustomRequestByIDWithDetails(id uuid.UUID) (*CustomRequest, error)
	UpdateCustomRequest(req *CustomRequest) error
	DeleteCustomRequest(id uuid.UUID) error
	ListCustomRequests(query CustomRequestListQuery) ([]CustomRequest, int64, error)
	GetCustomRequestsByUserID(userID uuid.UUID, query CustomRequestListQuery) ([]CustomRequest, int64, error)
	GetCustomRequestsByAssigneeID(assigneeID uuid.UUID, query CustomRequestListQuery) ([]CustomRequest, int64, error)

	// Request Item operations
	CreateRequestItem(item *RequestItem) error
	UpdateRequestItem(item *RequestItem) error
	DeleteRequestItem(id uuid.UUID) error
	GetRequestItemsByCustomRequestID(customRequestID uuid.UUID) ([]RequestItem, error)

	// Quote operations
	CreateQuote(quote *Quote) error
	GetQuoteByID(id uuid.UUID) (*Quote, error)
	GetQuoteByIDWithItems(id uuid.UUID) (*Quote, error)
	UpdateQuote(quote *Quote) error
	DeleteQuote(id uuid.UUID) error
	GetQuotesByCustomRequestID(customRequestID uuid.UUID) ([]Quote, error)
	GetActiveQuoteByCustomRequestID(customRequestID uuid.UUID) (*Quote, error)

	// Quote Item operations
	CreateQuoteItem(item *QuoteItem) error
	GetQuoteItemsByQuoteID(quoteID uuid.UUID) ([]QuoteItem, error)
	UpdateQuoteItem(item *QuoteItem) error
	DeleteQuoteItem(id uuid.UUID) error

	// Message operations
	CreateMessage(message *CustomRequestMessage) error
	GetMessagesByCustomRequestID(customRequestID uuid.UUID) ([]CustomRequestMessage, error)
	DeleteMessage(id uuid.UUID) error

	// Statistics and analytics
	GetCustomRequestStats() (*CustomRequestStatsRes, error)
	GetCustomRequestStatsByDateRange(startDate, endDate time.Time) (*CustomRequestStatsRes, error)
	GetExpiredCustomRequests() ([]CustomRequest, error)
	GetExpiredQuotes() ([]Quote, error)

	// Bulk operations
	BulkUpdateCustomRequestStatus(ids []uuid.UUID, status RequestStatus, assigneeID *uuid.UUID) error
	BulkDeleteCustomRequests(ids []uuid.UUID) error

	// Transaction operations
	CreateCustomRequestWithItems(req *CustomRequest, items []RequestItem) error
	UpdateCustomRequestWithItems(req *CustomRequest, items []RequestItem) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// Custom Request operations

func (r *repository) CreateCustomRequest(req *CustomRequest) error {
	return r.db.Create(req).Error
}

func (r *repository) GetCustomRequestByID(id uuid.UUID) (*CustomRequest, error) {
	var req CustomRequest
	err := r.db.Where("id = ?", id).First(&req).Error
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *repository) GetCustomRequestByIDWithDetails(id uuid.UUID) (*CustomRequest, error) {
	var req CustomRequest
	err := r.db.Preload("Items").
		Preload("Quotes").
		Preload("Quotes.Items").
		Preload("Quotes.Items.RequestItem").
		Preload("Messages").
		Where("id = ?", id).
		First(&req).Error
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *repository) UpdateCustomRequest(req *CustomRequest) error {
	return r.db.Save(req).Error
}

func (r *repository) DeleteCustomRequest(id uuid.UUID) error {
	// Use Unscoped() to permanently delete the record and all related data
	// This will bypass GORM's soft delete and actually remove the record from the database
	return r.db.Unscoped().Delete(&CustomRequest{}, id).Error
}

func (r *repository) ListCustomRequests(query CustomRequestListQuery) ([]CustomRequest, int64, error) {
	var requests []CustomRequest
	var total int64

	db := r.db.Model(&CustomRequest{})

	// Apply filters
	db = r.applyCustomRequestFilters(db, query)

	// Count total records
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and sorting
	db = r.applyPaginationAndSorting(db, query)

	// Preload relationships
	db = db.Preload("Items").Preload("Quotes").Preload("Messages")

	err := db.Find(&requests).Error
	return requests, total, err
}

func (r *repository) GetCustomRequestsByUserID(userID uuid.UUID, query CustomRequestListQuery) ([]CustomRequest, int64, error) {
	query.UserID = &userID
	return r.ListCustomRequests(query)
}

func (r *repository) GetCustomRequestsByAssigneeID(assigneeID uuid.UUID, query CustomRequestListQuery) ([]CustomRequest, int64, error) {
	query.AssigneeID = &assigneeID
	return r.ListCustomRequests(query)
}

// Request Item operations

func (r *repository) CreateRequestItem(item *RequestItem) error {
	return r.db.Create(item).Error
}

func (r *repository) UpdateRequestItem(item *RequestItem) error {
	return r.db.Save(item).Error
}

func (r *repository) DeleteRequestItem(id uuid.UUID) error {
	return r.db.Delete(&RequestItem{}, id).Error
}

func (r *repository) GetRequestItemsByCustomRequestID(customRequestID uuid.UUID) ([]RequestItem, error) {
	var items []RequestItem
	err := r.db.Where("custom_request_id = ?", customRequestID).Find(&items).Error
	return items, err
}

// Quote operations

func (r *repository) CreateQuote(quote *Quote) error {
	return r.db.Create(quote).Error
}

func (r *repository) GetQuoteByID(id uuid.UUID) (*Quote, error) {
	var quote Quote
	err := r.db.Where("id = ?", id).First(&quote).Error
	if err != nil {
		return nil, err
	}
	return &quote, nil
}

func (r *repository) UpdateQuote(quote *Quote) error {
	return r.db.Save(quote).Error
}

func (r *repository) DeleteQuote(id uuid.UUID) error {
	return r.db.Delete(&Quote{}, id).Error
}

func (r *repository) GetQuotesByCustomRequestID(customRequestID uuid.UUID) ([]Quote, error) {
	var quotes []Quote
	err := r.db.Where("custom_request_id = ?", customRequestID).
		Order("created_at DESC").
		Find(&quotes).Error
	return quotes, err
}

func (r *repository) GetActiveQuoteByCustomRequestID(customRequestID uuid.UUID) (*Quote, error) {
	var quote Quote
	err := r.db.Where("custom_request_id = ? AND status = ? AND (valid_until IS NULL OR valid_until > ?)",
		customRequestID, QuoteSent, time.Now()).
		Order("created_at DESC").
		First(&quote).Error
	if err != nil {
		return nil, err
	}
	return &quote, nil
}

func (r *repository) GetQuoteByIDWithItems(id uuid.UUID) (*Quote, error) {
	var quote Quote
	err := r.db.Preload("Items").Preload("Items.RequestItem").Where("id = ?", id).First(&quote).Error
	if err != nil {
		return nil, err
	}
	return &quote, nil
}

// Quote Item operations

func (r *repository) CreateQuoteItem(item *QuoteItem) error {
	return r.db.Create(item).Error
}

func (r *repository) GetQuoteItemsByQuoteID(quoteID uuid.UUID) ([]QuoteItem, error) {
	var items []QuoteItem
	err := r.db.Where("quote_id = ?", quoteID).Preload("RequestItem").Find(&items).Error
	return items, err
}

func (r *repository) UpdateQuoteItem(item *QuoteItem) error {
	return r.db.Save(item).Error
}

func (r *repository) DeleteQuoteItem(id uuid.UUID) error {
	return r.db.Delete(&QuoteItem{}, id).Error
}

// Message operations

func (r *repository) CreateMessage(message *CustomRequestMessage) error {
	return r.db.Create(message).Error
}

func (r *repository) GetMessagesByCustomRequestID(customRequestID uuid.UUID) ([]CustomRequestMessage, error) {
	var messages []CustomRequestMessage
	err := r.db.Where("custom_request_id = ?", customRequestID).
		Order("created_at ASC").
		Find(&messages).Error
	return messages, err
}

func (r *repository) DeleteMessage(id uuid.UUID) error {
	return r.db.Delete(&CustomRequestMessage{}, id).Error
}

// Statistics and analytics

func (r *repository) GetCustomRequestStats() (*CustomRequestStatsRes, error) {
	return r.GetCustomRequestStatsByDateRange(time.Time{}, time.Now())
}

func (r *repository) GetCustomRequestStatsByDateRange(startDate, endDate time.Time) (*CustomRequestStatsRes, error) {
	stats := &CustomRequestStatsRes{
		ByStatus:   make(map[RequestStatus]int64),
		ByPriority: make(map[RequestPriority]int64),
	}

	// Base query
	db := r.db.Model(&CustomRequest{})
	if !startDate.IsZero() {
		db = db.Where("submitted_at >= ?", startDate)
	}
	if !endDate.IsZero() {
		db = db.Where("submitted_at <= ?", endDate)
	}

	// Total requests
	if err := db.Count(&stats.TotalRequests).Error; err != nil {
		return nil, err
	}

	// By status
	var statusCounts []struct {
		Status RequestStatus `json:"status"`
		Count  int64         `json:"count"`
	}
	if err := db.Select("status, COUNT(*) as count").Group("status").Find(&statusCounts).Error; err != nil {
		return nil, err
	}
	for _, sc := range statusCounts {
		stats.ByStatus[sc.Status] = sc.Count
	}

	// By priority
	var priorityCounts []struct {
		Priority RequestPriority `json:"priority"`
		Count    int64           `json:"count"`
	}
	if err := db.Select("priority, COUNT(*) as count").Group("priority").Find(&priorityCounts).Error; err != nil {
		return nil, err
	}
	for _, pc := range priorityCounts {
		stats.ByPriority[pc.Priority] = pc.Count
	}

	// Average items per request
	var avgItems struct {
		Average float64 `json:"average"`
	}
	if err := r.db.Raw(`
		SELECT AVG(item_count) as average
		FROM (
			SELECT COUNT(*) as item_count
			FROM request_items ri
			JOIN custom_requests cr ON ri.custom_request_id = cr.id
			WHERE cr.submitted_at >= ? AND cr.submitted_at <= ?
			GROUP BY ri.custom_request_id
		) as item_counts
	`, startDate, endDate).Scan(&avgItems).Error; err != nil {
		return nil, err
	}
	stats.AverageItems = avgItems.Average

	// Average value (from accepted quotes)
	var avgValue struct {
		Average int64 `json:"average"`
	}
	if err := r.db.Raw(`
		SELECT AVG(q.grand_total) as average
		FROM quotes q
		JOIN custom_requests cr ON q.custom_request_id = cr.id
		WHERE q.status = ? AND cr.submitted_at >= ? AND cr.submitted_at <= ?
	`, QuoteAccepted, startDate, endDate).Scan(&avgValue).Error; err != nil {
		return nil, err
	}
	stats.AverageValue = avgValue.Average

	// Response time statistics
	var responseTime struct {
		AverageHours float64 `json:"average_hours"`
		MedianHours  float64 `json:"median_hours"`
	}
	if err := r.db.Raw(`
		SELECT 
			AVG(EXTRACT(EPOCH FROM (q.sent_at - cr.submitted_at))/3600) as average_hours,
			PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY EXTRACT(EPOCH FROM (q.sent_at - cr.submitted_at))/3600) as median_hours
		FROM quotes q
		JOIN custom_requests cr ON q.custom_request_id = cr.id
		WHERE q.sent_at IS NOT NULL AND cr.submitted_at >= ? AND cr.submitted_at <= ?
	`, startDate, endDate).Scan(&responseTime).Error; err != nil {
		return nil, err
	}
	stats.ResponseTime = CustomRequestResponseTime{
		AverageHours: responseTime.AverageHours,
		MedianHours:  responseTime.MedianHours,
	}

	return stats, nil
}

func (r *repository) GetExpiredCustomRequests() ([]CustomRequest, error) {
	var requests []CustomRequest
	err := r.db.Where("expires_at IS NOT NULL AND expires_at < ?", time.Now()).
		Find(&requests).Error
	return requests, err
}

func (r *repository) GetExpiredQuotes() ([]Quote, error) {
	var quotes []Quote
	err := r.db.Where("valid_until IS NOT NULL AND valid_until < ? AND status = ?",
		time.Now(), QuoteSent).Find(&quotes).Error
	return quotes, err
}

// Bulk operations

func (r *repository) BulkUpdateCustomRequestStatus(ids []uuid.UUID, status RequestStatus, assigneeID *uuid.UUID) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}
	if assigneeID != nil {
		updates["assignee_id"] = *assigneeID
	}

	return r.db.Model(&CustomRequest{}).Where("id IN ?", ids).Updates(updates).Error
}

func (r *repository) BulkDeleteCustomRequests(ids []uuid.UUID) error {
	return r.db.Where("id IN ?", ids).Delete(&CustomRequest{}).Error
}

// Helper methods

func (r *repository) applyCustomRequestFilters(db *gorm.DB, query CustomRequestListQuery) *gorm.DB {
	if query.Status != nil {
		db = db.Where("status = ?", *query.Status)
	}
	if query.Priority != nil {
		db = db.Where("priority = ?", *query.Priority)
	}
	if query.AssigneeID != nil {
		db = db.Where("assignee_id = ?", *query.AssigneeID)
	}
	if query.UserID != nil {
		db = db.Where("user_id = ?", *query.UserID)
	}
	return db
}

func (r *repository) applyPaginationAndSorting(db *gorm.DB, query CustomRequestListQuery) *gorm.DB {
	// Set defaults
	query.SetDefaults()

	// Apply sorting
	orderClause := fmt.Sprintf("%s %s", r.sanitizeSortField(query.SortBy), strings.ToUpper(query.SortOrder))
	db = db.Order(orderClause)

	// Apply pagination
	offset := (query.Page - 1) * query.Limit
	db = db.Offset(offset).Limit(query.Limit)

	return db
}

func (r *repository) sanitizeSortField(field string) string {
	allowedFields := map[string]string{
		"submitted_at": "submitted_at",
		"updated_at":   "updated_at",
		"status":       "status",
		"priority":     "priority",
		"user_id":      "user_id",
	}

	if sanitized, exists := allowedFields[field]; exists {
		return sanitized
	}
	return "submitted_at" // default
}

// Transaction helpers

func (r *repository) WithTransaction(fn func(Repository) error) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		txRepo := &repository{db: tx}
		return fn(txRepo)
	})
}

// CreateCustomRequestWithItems creates a custom request with its items in a transaction
func (r *repository) CreateCustomRequestWithItems(req *CustomRequest, items []RequestItem) error {
	return r.WithTransaction(func(txRepo Repository) error {
		if err := txRepo.CreateCustomRequest(req); err != nil {
			return err
		}

		for i := range items {
			items[i].CustomRequestID = req.ID
			if err := txRepo.CreateRequestItem(&items[i]); err != nil {
				return err
			}
		}

		return nil
	})
}

// UpdateCustomRequestWithItems updates a custom request and its items in a transaction
func (r *repository) UpdateCustomRequestWithItems(req *CustomRequest, items []RequestItem) error {
	return r.WithTransaction(func(txRepo Repository) error {
		if err := txRepo.UpdateCustomRequest(req); err != nil {
			return err
		}

		// Delete existing items
		if err := r.db.Where("custom_request_id = ?", req.ID).Delete(&RequestItem{}).Error; err != nil {
			return err
		}

		// Create new items
		for i := range items {
			items[i].CustomRequestID = req.ID
			if err := txRepo.CreateRequestItem(&items[i]); err != nil {
				return err
			}
		}

		return nil
	})
}