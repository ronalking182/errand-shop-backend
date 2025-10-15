package custom_requests

import (
	"time"

	"github.com/google/uuid"
)

// CreateCustomRequestReq represents the request to create a custom request
type CreateCustomRequestReq struct {
	DeliveryAddressID  *uuid.UUID           `json:"deliveryAddressId"`
	Items              []CreateRequestItem  `json:"items" validate:"required,min=1,dive"`
	AllowSubstitutions bool                 `json:"allowSubstitutions"`
	Notes              string               `json:"notes"`
	Priority           RequestPriority      `json:"priority"`
}

// CreateRequestItem represents an item in a custom request creation
type CreateRequestItem struct {
	Name           string   `json:"name" validate:"required,min=1,max=255"`
	Description    string   `json:"description"`
	Quantity       float64  `json:"quantity" validate:"required,gt=0"`
	Unit           string   `json:"unit"`
	PreferredBrand string   `json:"preferredBrand"`
	EstimatedPrice *int64   `json:"estimatedPrice"` // in kobo
	Images         []string `json:"images"`
}

// UpdateCustomRequestReq represents the request to update a custom request
type UpdateCustomRequestReq struct {
	DeliveryAddressID  *uuid.UUID           `json:"deliveryAddressId"`
	Items              []UpdateRequestItem  `json:"items"`
	AllowSubstitutions *bool                `json:"allowSubstitutions"`
	Notes              *string              `json:"notes"`
	Priority           *RequestPriority     `json:"priority"`
}

// UpdateRequestItem represents an item update in a custom request
type UpdateRequestItem struct {
	ID             *uuid.UUID `json:"id"` // nil for new items
	Name           string     `json:"name" validate:"required,min=1,max=255"`
	Description    string     `json:"description"`
	Quantity       float64    `json:"quantity" validate:"required,gt=0"`
	Unit           string     `json:"unit"`
	PreferredBrand string     `json:"preferredBrand"`
	EstimatedPrice *int64     `json:"estimatedPrice"` // in kobo
	Images         []string   `json:"images"`
}

// CreateQuoteReq represents the request to create a quote for a custom request
type CreateQuoteReq struct {
	CustomRequestID uuid.UUID         `json:"customRequestId" validate:"required"`
	Items           []QuoteItemReq    `json:"items" validate:"required,min=1,dive"`
	Fees            QuoteFees         `json:"fees" validate:"required"`
	ValidUntil      *time.Time        `json:"validUntil"`
}

// QuoteItemReq represents an item in a quote creation request
type QuoteItemReq struct {
	RequestItemID uuid.UUID `json:"requestItemId" validate:"required"`
	QuotedPrice   int64     `json:"quotedPrice" validate:"required,gte=0"` // in kobo
	AdminNotes    string    `json:"adminNotes"`
}

// AcceptQuoteReq represents the request to accept a quote
type AcceptQuoteReq struct {
	QuoteID uuid.UUID `json:"quoteId" validate:"required"`
}

// UpdateRequestStatusReq represents admin request to update custom request status
type UpdateRequestStatusReq struct {
	Status     RequestStatus `json:"status" validate:"required"`
	AssigneeID *uuid.UUID    `json:"assigneeId"`
	Notes      string        `json:"notes"`
}

// SendMessageReq represents a request to send a message
type SendMessageReq struct {
	Message string `json:"message" validate:"required,min=1"`
}

// CustomRequestListQuery represents query parameters for listing custom requests
type CustomRequestListQuery struct {
	Status     *RequestStatus   `json:"status"`
	Priority   *RequestPriority `json:"priority"`
	AssigneeID *uuid.UUID       `json:"assigneeId"`
	UserID     *uuid.UUID       `json:"userId"`
	Page       int              `json:"page" validate:"min=1"`
	Limit      int              `json:"limit" validate:"min=1,max=100"`
	SortBy     string           `json:"sortBy"`
	SortOrder  string           `json:"sortOrder" validate:"oneof=asc desc"`
}

// Response DTOs

// CustomRequestRes represents a custom request response
type CustomRequestRes struct {
	ID                 uuid.UUID              `json:"id"`
	UserID             uuid.UUID              `json:"userId"`
	DeliveryAddressID  *uuid.UUID             `json:"deliveryAddressId"`
	Status             RequestStatus          `json:"status"`
	Priority           RequestPriority        `json:"priority"`
	AllowSubstitutions bool                   `json:"allowSubstitutions"`
	Notes              string                 `json:"notes"`
	AssigneeID         *uuid.UUID             `json:"assigneeId"`
	SubmittedAt        time.Time              `json:"submittedAt"`
	UpdatedAt          time.Time              `json:"updatedAt"`
	ExpiresAt          *time.Time             `json:"expiresAt"`
	Items              []RequestItemRes       `json:"items"`
	Quotes             []QuoteRes             `json:"quotes,omitempty"`
	Messages           []CustomRequestMsgRes  `json:"messages,omitempty"`
	ActiveQuote        *QuoteRes              `json:"activeQuote,omitempty"`
}

// RequestItemRes represents a request item response
type RequestItemRes struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	Quantity       float64   `json:"quantity"`
	Unit           string    `json:"unit"`
	PreferredBrand string    `json:"preferredBrand"`
	EstimatedPrice *int64    `json:"estimatedPrice"` // in kobo
	QuotedPrice    *int64    `json:"quotedPrice"`    // in kobo
	AdminNotes     string    `json:"adminNotes"`
	Images         []string  `json:"images"`
}

// QuoteRes represents a quote response
type QuoteRes struct {
	ID            uuid.UUID      `json:"id"`
	ItemsSubtotal int64          `json:"itemsSubtotal"` // in kobo
	Fees          QuoteFees      `json:"fees"`
	FeesTotal     int64          `json:"feesTotal"`     // in kobo
	GrandTotal    int64          `json:"grandTotal"`    // in kobo
	Status        QuoteStatus    `json:"status"`
	ValidUntil    *time.Time     `json:"validUntil"`
	SentAt        *time.Time     `json:"sentAt"`
	AcceptedAt    *time.Time     `json:"acceptedAt"`
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
	Items         []QuoteItemRes `json:"items"`
}

// QuoteItemRes represents a quote item response
type QuoteItemRes struct {
	ID            uuid.UUID `json:"id"`
	RequestItemID uuid.UUID `json:"requestItemId"`
	UnitPrice     int64     `json:"unitPrice"`   // in kobo - this is the quoted price
	QuotedPrice   int64     `json:"quotedPrice"` // in kobo - same as unitPrice for compatibility
	AdminNotes    string    `json:"adminNotes"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
	// Include request item details for mobile app
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	Quantity       float64  `json:"quantity"`
	Unit           string   `json:"unit"`
	PreferredBrand string   `json:"preferredBrand"`
	Images         []string `json:"images"`
}

// CustomRequestMsgRes represents a custom request message response
type CustomRequestMsgRes struct {
	ID         uuid.UUID  `json:"id"`
	SenderType SenderType `json:"senderType"`
	SenderID   uuid.UUID  `json:"senderId"`
	Message    string     `json:"message"`
	CreatedAt  time.Time  `json:"createdAt"`
}

// CustomRequestListRes represents a paginated list of custom requests
type CustomRequestListRes struct {
	Data       []CustomRequestRes `json:"data"`
	Total      int64              `json:"total"`
	Page       int                `json:"page"`
	Limit      int                `json:"limit"`
	TotalPages int                `json:"totalPages"`
}

// CustomRequestStatsRes represents statistics for custom requests
type CustomRequestStatsRes struct {
	TotalRequests     int64                        `json:"totalRequests"`
	ByStatus          map[RequestStatus]int64      `json:"byStatus"`
	ByPriority        map[RequestPriority]int64    `json:"byPriority"`
	AverageItems      float64                      `json:"averageItems"`
	AverageValue      int64                        `json:"averageValue"` // in kobo
	ResponseTime      CustomRequestResponseTime    `json:"responseTime"`
}

// Bulk operation DTOs

// BulkUpdateStatusReq represents a bulk status update request
type BulkUpdateStatusReq struct {
	RequestIDs []uuid.UUID    `json:"requestIds" validate:"required,min=1"`
	Status     RequestStatus  `json:"status" validate:"required"`
	AssigneeID *uuid.UUID     `json:"assigneeId"`
	Notes      string         `json:"notes"`
}

// BulkUpdateStatusRes represents a bulk status update response
type BulkUpdateStatusRes struct {
	UpdatedCount int           `json:"updatedCount"`
	RequestIDs   []uuid.UUID   `json:"requestIds"`
	Status       RequestStatus `json:"status"`
}

// BulkAssignReq represents a bulk assign request
type BulkAssignReq struct {
	RequestIDs []uuid.UUID `json:"requestIds" validate:"required,min=1"`
	AssigneeID uuid.UUID   `json:"assigneeId" validate:"required"`
	Notes      string      `json:"notes"`
}

// BulkAssignRes represents a bulk assign response
type BulkAssignRes struct {
	AssignedCount int           `json:"assignedCount"`
	RequestIDs    []uuid.UUID   `json:"requestIds"`
	AssigneeID    uuid.UUID     `json:"assigneeId"`
}

// CustomRequestResponseTime represents response time statistics
type CustomRequestResponseTime struct {
	AverageHours float64 `json:"averageHours"`
	MedianHours  float64 `json:"medianHours"`
}

// Conversion methods

// ToCustomRequestRes converts a CustomRequest model to response DTO
func (cr *CustomRequest) ToCustomRequestRes() CustomRequestRes {
	res := CustomRequestRes{
		ID:                 cr.ID,
		UserID:             cr.UserID,
		DeliveryAddressID:  cr.DeliveryAddressID,
		Status:             cr.Status,
		Priority:           cr.Priority,
		AllowSubstitutions: cr.AllowSubstitutions,
		Notes:              cr.Notes,
		AssigneeID:         cr.AssigneeID,
		SubmittedAt:        cr.SubmittedAt,
		UpdatedAt:          cr.UpdatedAt,
		ExpiresAt:          cr.ExpiresAt,
	}

	// Convert items
	for _, item := range cr.Items {
		res.Items = append(res.Items, item.ToRequestItemRes())
	}

	// Convert quotes
	for _, quote := range cr.Quotes {
		res.Quotes = append(res.Quotes, quote.ToQuoteRes())
	}

	// Convert messages
	for _, msg := range cr.Messages {
		res.Messages = append(res.Messages, msg.ToCustomRequestMsgRes())
	}

	// Set active quote
	if activeQuote := cr.GetActiveQuote(); activeQuote != nil {
		quoteRes := activeQuote.ToQuoteRes()
		res.ActiveQuote = &quoteRes
	}

	return res
}

// ToRequestItemRes converts a RequestItem model to response DTO
func (ri *RequestItem) ToRequestItemRes() RequestItemRes {
	return RequestItemRes{
		ID:             ri.ID,
		Name:           ri.Name,
		Description:    ri.Description,
		Quantity:       ri.Quantity,
		Unit:           ri.Unit,
		PreferredBrand: ri.PreferredBrand,
		EstimatedPrice: ri.EstimatedPrice,
		QuotedPrice:    ri.QuotedPrice,
		AdminNotes:     ri.AdminNotes,
		Images:         ri.Images,
	}
}

// ToQuoteRes converts a Quote model to response DTO
func (q *Quote) ToQuoteRes() QuoteRes {
	// Convert quote items to response DTOs
	itemRes := make([]QuoteItemRes, len(q.Items))
	for i, item := range q.Items {
		itemRes[i] = item.ToQuoteItemRes()
	}

	return QuoteRes{
		ID:            q.ID,
		ItemsSubtotal: q.ItemsSubtotal,
		Fees:          q.Fees,
		FeesTotal:     q.FeesTotal,
		GrandTotal:    q.GrandTotal,
		Status:        q.Status,
		ValidUntil:    q.ValidUntil,
		SentAt:        q.SentAt,
		AcceptedAt:    q.AcceptedAt,
		CreatedAt:     q.CreatedAt,
		UpdatedAt:     q.UpdatedAt,
		Items:         itemRes,
	}
}

// ToQuoteItemRes converts a QuoteItem model to response DTO
func (qi *QuoteItem) ToQuoteItemRes() QuoteItemRes {
	return QuoteItemRes{
		ID:            qi.ID,
		RequestItemID: qi.RequestItemID,
		UnitPrice:     qi.QuotedPrice,
		QuotedPrice:   qi.QuotedPrice,
		AdminNotes:    qi.AdminNotes,
		CreatedAt:     qi.CreatedAt,
		UpdatedAt:     qi.UpdatedAt,
		// Include request item details - these will be populated via preloading
		Name:           qi.RequestItem.Name,
		Description:    qi.RequestItem.Description,
		Quantity:       qi.RequestItem.Quantity,
		Unit:           qi.RequestItem.Unit,
		PreferredBrand: qi.RequestItem.PreferredBrand,
		Images:         qi.RequestItem.Images,
	}
}

// ToCustomRequestMsgRes converts a CustomRequestMessage model to response DTO
func (crm *CustomRequestMessage) ToCustomRequestMsgRes() CustomRequestMsgRes {
	return CustomRequestMsgRes{
		ID:         crm.ID,
		SenderType: crm.SenderType,
		SenderID:   crm.SenderID,
		Message:    crm.Message,
		CreatedAt:  crm.CreatedAt,
	}
}

// Default values for query parameters
func (q *CustomRequestListQuery) SetDefaults() {
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.Limit <= 0 {
		q.Limit = 20
	}
	if q.SortBy == "" {
		q.SortBy = "submitted_at"
	}
	if q.SortOrder == "" {
		q.SortOrder = "desc"
	}
}

// Validation helpers

// ValidateStatus checks if the status transition is valid
func ValidateStatusTransition(from, to RequestStatus) bool {
	validTransitions := map[RequestStatus][]RequestStatus{
		RequestSubmitted: {
			RequestUnderReview,
			RequestNeedsInfo,
			RequestCancelled,
		},
		RequestUnderReview: {
			RequestQuoteSent,
			RequestNeedsInfo,
			RequestCancelled,
		},
		RequestNeedsInfo: {
			RequestUnderReview,
			RequestSubmitted,
			RequestCancelled,
		},
		RequestQuoteSent: {
			RequestCustomerAccepted,
			RequestCustomerDeclined,
			RequestCancelled,
		},
		RequestCustomerAccepted: {
			RequestApproved,
			RequestInCart,
			RequestCancelled,
		},
		RequestCustomerDeclined: {
			RequestUnderReview,
			RequestCancelled,
		},
	}

	allowedTransitions, exists := validTransitions[from]
	if !exists {
		return false
	}

	for _, allowed := range allowedTransitions {
		if allowed == to {
			return true
		}
	}
	return false
}

// IsValidPriority checks if the priority is valid
func IsValidPriority(priority RequestPriority) bool {
	switch priority {
	case PriorityLow, PriorityMedium, PriorityHigh, PriorityUrgent:
		return true
	default:
		return false
	}
}

// IsValidQuoteStatus checks if the quote status is valid
func IsValidQuoteStatus(status QuoteStatus) bool {
	switch status {
	case QuoteDraft, QuoteSent, QuoteAccepted, QuoteDeclined:
		return true
	default:
		return false
	}
}