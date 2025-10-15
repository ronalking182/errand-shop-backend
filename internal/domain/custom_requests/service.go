package custom_requests

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrCustomRequestNotFound     = errors.New("custom request not found")
	ErrQuoteNotFound            = errors.New("quote not found")
	ErrInvalidStatusTransition  = errors.New("invalid status transition")
	ErrQuoteExpired             = errors.New("quote has expired")
	ErrQuoteNotActive           = errors.New("quote is not active")
	ErrUnauthorizedAccess       = errors.New("unauthorized access to custom request")
	ErrCannotModifyRequest      = errors.New("custom request cannot be modified in current status")
	ErrInvalidQuoteStatus       = errors.New("invalid quote status")
	ErrDuplicateActiveQuote     = errors.New("custom request already has an active quote")
	ErrEmptyRequestItems        = errors.New("custom request must have at least one item")
	ErrInvalidPriority          = errors.New("invalid priority level")
)

type Service interface {
	// User operations
	CreateCustomRequest(userID uuid.UUID, req CreateCustomRequestReq) (*CustomRequestRes, error)
	GetCustomRequest(userID uuid.UUID, requestID uuid.UUID) (*CustomRequestRes, error)
	UpdateCustomRequest(userID uuid.UUID, requestID uuid.UUID, req UpdateCustomRequestReq) (*CustomRequestRes, error)
	DeleteCustomRequest(userID uuid.UUID, requestID uuid.UUID) error
	PermanentlyDeleteCustomRequest(userID uuid.UUID, requestID uuid.UUID) error
	CancelCustomRequest(userID uuid.UUID, requestID uuid.UUID, reason string) (*CustomRequestRes, error)
	ListUserCustomRequests(userID uuid.UUID, query CustomRequestListQuery) (*CustomRequestListRes, error)
	AcceptQuote(userID uuid.UUID, req AcceptQuoteReq) (*CustomRequestRes, error)
	AcceptQuoteByRequestID(userID uuid.UUID, requestID uuid.UUID) (*CustomRequestRes, error)
	SendMessage(userID uuid.UUID, requestID uuid.UUID, req SendMessageReq) (*CustomRequestMsgRes, error)

	// Admin operations
	GetCustomRequestAdmin(requestID uuid.UUID) (*CustomRequestRes, error)
	ListCustomRequestsAdmin(query CustomRequestListQuery) (*CustomRequestListRes, error)
	UpdateCustomRequestStatus(requestID uuid.UUID, req UpdateRequestStatusReq) (*CustomRequestRes, error)
	AssignCustomRequest(requestID uuid.UUID, assigneeID uuid.UUID) (*CustomRequestRes, error)
	CreateQuote(adminID uuid.UUID, req CreateQuoteReq) (*QuoteRes, error)
	UpdateQuote(quoteID uuid.UUID, req CreateQuoteReq) (*QuoteRes, error)
	SendQuote(quoteID uuid.UUID) (*QuoteRes, error)
	SendMessageAdmin(adminID uuid.UUID, requestID uuid.UUID, req SendMessageReq) (*CustomRequestMsgRes, error)
	// Admin cancel and delete operations (no status restrictions)
	CancelCustomRequestAdmin(adminID uuid.UUID, requestID uuid.UUID, reason string) (*CustomRequestRes, error)
	PermanentlyDeleteCustomRequestAdmin(adminID uuid.UUID, requestID uuid.UUID) error

	// Analytics and reporting
	GetCustomRequestStats() (*CustomRequestStatsRes, error)
	GetCustomRequestStatsByDateRange(startDate, endDate time.Time) (*CustomRequestStatsRes, error)

	// Background tasks
	ProcessExpiredRequests() error
	ProcessExpiredQuotes() error
	CleanupOldMessages(olderThan time.Time) error

	// Bulk operations
	BulkUpdateStatus(req BulkUpdateStatusReq) (*BulkUpdateStatusRes, error)
	BulkAssign(req BulkAssignReq) (*BulkAssignRes, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// User operations

func (s *service) CreateCustomRequest(userID uuid.UUID, req CreateCustomRequestReq) (*CustomRequestRes, error) {
	// Validate request
	if len(req.Items) == 0 {
		return nil, ErrEmptyRequestItems
	}

	if req.Priority != "" && !IsValidPriority(req.Priority) {
		return nil, ErrInvalidPriority
	}

	// Set default priority if not provided
	if req.Priority == "" {
		req.Priority = PriorityMedium
	}

	// Create custom request
	customRequest := &CustomRequest{
		ID:                 uuid.New(),
		UserID:             userID,
		DeliveryAddressID:  req.DeliveryAddressID,
		Status:             RequestSubmitted,
		Priority:           req.Priority,
		AllowSubstitutions: req.AllowSubstitutions,
		Notes:              req.Notes,
		SubmittedAt:        time.Now(),
		UpdatedAt:          time.Now(),
	}

	// Create request items
	var items []RequestItem
	for _, itemReq := range req.Items {
		item := RequestItem{
			ID:               uuid.New(),
			CustomRequestID:  customRequest.ID,
			Name:             itemReq.Name,
			Description:      itemReq.Description,
			Quantity:         itemReq.Quantity,
			Unit:             itemReq.Unit,
			PreferredBrand:   itemReq.PreferredBrand,
			EstimatedPrice:   itemReq.EstimatedPrice,
			Images:           itemReq.Images,
		}
		items = append(items, item)
	}

	// Save to database
	if err := s.repo.CreateCustomRequestWithItems(customRequest, items); err != nil {
		return nil, fmt.Errorf("failed to create custom request: %w", err)
	}

	// Fetch the created request with details
	createdRequest, err := s.repo.GetCustomRequestByIDWithDetails(customRequest.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch created custom request: %w", err)
	}

	res := createdRequest.ToCustomRequestRes()
	return &res, nil
}

func (s *service) GetCustomRequest(userID uuid.UUID, requestID uuid.UUID) (*CustomRequestRes, error) {
	request, err := s.repo.GetCustomRequestByIDWithDetails(requestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCustomRequestNotFound
		}
		return nil, fmt.Errorf("failed to get custom request: %w", err)
	}

	// Check if user owns the request
	if request.UserID != userID {
		return nil, ErrUnauthorizedAccess
	}

	res := request.ToCustomRequestRes()
	return &res, nil
}

func (s *service) UpdateCustomRequest(userID uuid.UUID, requestID uuid.UUID, req UpdateCustomRequestReq) (*CustomRequestRes, error) {
	request, err := s.repo.GetCustomRequestByIDWithDetails(requestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCustomRequestNotFound
		}
		return nil, fmt.Errorf("failed to get custom request: %w", err)
	}

	// Check if user owns the request
	if request.UserID != userID {
		return nil, ErrUnauthorizedAccess
	}

	// Check if request can be modified
	if !request.CanBeModified() {
		return nil, ErrCannotModifyRequest
	}

	// Update fields
	if req.DeliveryAddressID != nil {
		request.DeliveryAddressID = req.DeliveryAddressID
	}
	if req.AllowSubstitutions != nil {
		request.AllowSubstitutions = *req.AllowSubstitutions
	}
	if req.Notes != nil {
		request.Notes = *req.Notes
	}
	if req.Priority != nil {
		if !IsValidPriority(*req.Priority) {
			return nil, ErrInvalidPriority
		}
		request.Priority = *req.Priority
	}

	// Update items if provided
	var items []RequestItem
	if req.Items != nil {
		for _, itemReq := range req.Items {
			item := RequestItem{
				ID:               uuid.New(),
				CustomRequestID:  request.ID,
				Name:             itemReq.Name,
				Description:      itemReq.Description,
				Quantity:         itemReq.Quantity,
				Unit:             itemReq.Unit,
				PreferredBrand:   itemReq.PreferredBrand,
				EstimatedPrice:   itemReq.EstimatedPrice,
				Images:           itemReq.Images,
			}
			// If ID is provided, use it (for updates)
			if itemReq.ID != nil {
				item.ID = *itemReq.ID
			}
			items = append(items, item)
		}

		// Update with items
		if err := s.repo.UpdateCustomRequestWithItems(request, items); err != nil {
			return nil, fmt.Errorf("failed to update custom request with items: %w", err)
		}
	} else {
		// Update without items
		if err := s.repo.UpdateCustomRequest(request); err != nil {
			return nil, fmt.Errorf("failed to update custom request: %w", err)
		}
	}

	// Fetch updated request
	updatedRequest, err := s.repo.GetCustomRequestByIDWithDetails(requestID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch updated custom request: %w", err)
	}

	res := updatedRequest.ToCustomRequestRes()
	return &res, nil
}

func (s *service) DeleteCustomRequest(userID uuid.UUID, requestID uuid.UUID) error {
	request, err := s.repo.GetCustomRequestByID(requestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCustomRequestNotFound
		}
		return fmt.Errorf("failed to get custom request: %w", err)
	}

	// Check if user owns the request
	if request.UserID != userID {
		return ErrUnauthorizedAccess
	}

	// Allow deletion regardless of status for better user experience
	// Users can now delete their requests at any stage

	return s.repo.DeleteCustomRequest(requestID)
}

func (s *service) PermanentlyDeleteCustomRequest(userID uuid.UUID, requestID uuid.UUID) error {
	request, err := s.repo.GetCustomRequestByID(requestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCustomRequestNotFound
		}
		return fmt.Errorf("failed to get custom request: %w", err)
	}

	// Check if user owns the request
	if request.UserID != userID {
		return ErrUnauthorizedAccess
	}

	// Allow permanent deletion regardless of status for better user experience
	// Users can now permanently delete their requests at any stage

	// Permanently delete the request and all related data
	return s.repo.DeleteCustomRequest(requestID)
}

func (s *service) CancelCustomRequest(userID uuid.UUID, requestID uuid.UUID, reason string) (*CustomRequestRes, error) {
	request, err := s.repo.GetCustomRequestByID(requestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCustomRequestNotFound
		}
		return nil, fmt.Errorf("failed to get custom request: %w", err)
	}

	// Check if user owns the request
	if request.UserID != userID {
		return nil, ErrUnauthorizedAccess
	}

	// Check if request is already cancelled
	if request.Status == RequestCancelled {
		return nil, errors.New("custom request is already cancelled")
	}

	// Allow cancellation regardless of status for better user experience
	// Users can now cancel their requests at any stage

	// Update status to cancelled
	request.Status = RequestCancelled
	request.UpdatedAt = time.Now()

	// Add a message about the cancellation
	message := &CustomRequestMessage{
		ID:               uuid.New(),
		CustomRequestID:  requestID,
		SenderType:       SenderUser,
		Message:          fmt.Sprintf("Request cancelled by user. Reason: %s", reason),
		CreatedAt:        time.Now(),
	}
	request.Messages = append(request.Messages, *message)

	if err := s.repo.UpdateCustomRequest(request); err != nil {
		return nil, fmt.Errorf("failed to cancel custom request: %w", err)
	}

	res := request.ToCustomRequestRes()
	return &res, nil
}

func (s *service) ListUserCustomRequests(userID uuid.UUID, query CustomRequestListQuery) (*CustomRequestListRes, error) {
	requests, total, err := s.repo.GetCustomRequestsByUserID(userID, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list user custom requests: %w", err)
	}

	// Convert to response DTOs
	var data []CustomRequestRes
	for _, req := range requests {
		data = append(data, req.ToCustomRequestRes())
	}

	// Calculate total pages
	query.SetDefaults()
	totalPages := int((total + int64(query.Limit) - 1) / int64(query.Limit))

	return &CustomRequestListRes{
		Data:       data,
		Total:      total,
		Page:       query.Page,
		Limit:      query.Limit,
		TotalPages: totalPages,
	}, nil
}

func (s *service) AcceptQuote(userID uuid.UUID, req AcceptQuoteReq) (*CustomRequestRes, error) {
	quote, err := s.repo.GetQuoteByID(req.QuoteID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrQuoteNotFound
		}
		return nil, fmt.Errorf("failed to get quote: %w", err)
	}

	// Get the custom request
	customRequest, err := s.repo.GetCustomRequestByIDWithDetails(quote.CustomRequestID)
	if err != nil {
		return nil, fmt.Errorf("failed to get custom request: %w", err)
	}

	// Check if user owns the request
	if customRequest.UserID != userID {
		return nil, ErrUnauthorizedAccess
	}

	// Check if request can be accepted
	if !customRequest.CanBeAccepted() {
		return nil, ErrCannotModifyRequest
	}

	// Check if quote is valid
	if quote.Status != QuoteSent {
		return nil, ErrQuoteNotActive
	}
	if quote.IsExpired() {
		return nil, ErrQuoteExpired
	}

	// Update quote status
	now := time.Now()
	quote.Status = QuoteAccepted
	quote.AcceptedAt = &now
	if err := s.repo.UpdateQuote(quote); err != nil {
		return nil, fmt.Errorf("failed to update quote: %w", err)
	}

	// Update custom request status
	customRequest.Status = RequestCustomerAccepted
	if err := s.repo.UpdateCustomRequest(customRequest); err != nil {
		return nil, fmt.Errorf("failed to update custom request: %w", err)
	}

	// Fetch updated request
	updatedRequest, err := s.repo.GetCustomRequestByIDWithDetails(customRequest.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch updated custom request: %w", err)
	}

	res := updatedRequest.ToCustomRequestRes()
	return &res, nil
}

func (s *service) AcceptQuoteByRequestID(userID uuid.UUID, requestID uuid.UUID) (*CustomRequestRes, error) {
	// Get the custom request with details
	customRequest, err := s.repo.GetCustomRequestByIDWithDetails(requestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCustomRequestNotFound
		}
		return nil, fmt.Errorf("failed to get custom request: %w", err)
	}

	// Check if user owns the request
	if customRequest.UserID != userID {
		return nil, ErrUnauthorizedAccess
	}

	// Check if request can be accepted
	if !customRequest.CanBeAccepted() {
		return nil, ErrCannotModifyRequest
	}

	// Get the active quote
	activeQuote := customRequest.GetActiveQuote()
	if activeQuote == nil {
		return nil, ErrQuoteNotFound
	}

	// Check if quote is valid
	if activeQuote.Status != QuoteSent {
		return nil, ErrQuoteNotActive
	}
	if activeQuote.IsExpired() {
		return nil, ErrQuoteExpired
	}

	// Update quote status
	now := time.Now()
	activeQuote.Status = QuoteAccepted
	activeQuote.AcceptedAt = &now
	if err := s.repo.UpdateQuote(activeQuote); err != nil {
		return nil, fmt.Errorf("failed to update quote: %w", err)
	}

	// Update custom request status
	customRequest.Status = RequestCustomerAccepted
	if err := s.repo.UpdateCustomRequest(customRequest); err != nil {
		return nil, fmt.Errorf("failed to update custom request: %w", err)
	}

	// Fetch updated request
	updatedRequest, err := s.repo.GetCustomRequestByIDWithDetails(customRequest.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch updated custom request: %w", err)
	}

	res := updatedRequest.ToCustomRequestRes()
	return &res, nil
}

func (s *service) SendMessage(userID uuid.UUID, requestID uuid.UUID, req SendMessageReq) (*CustomRequestMsgRes, error) {
	customRequest, err := s.repo.GetCustomRequestByID(requestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCustomRequestNotFound
		}
		return nil, fmt.Errorf("failed to get custom request: %w", err)
	}

	// Check if user owns the request
	if customRequest.UserID != userID {
		return nil, ErrUnauthorizedAccess
	}

	// Create message
	message := &CustomRequestMessage{
		ID:              uuid.New(),
		CustomRequestID: requestID,
		SenderType:      SenderUser,
		SenderID:        userID,
		Message:         req.Message,
		CreatedAt:       time.Now(),
	}

	if err := s.repo.CreateMessage(message); err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	res := message.ToCustomRequestMsgRes()
	return &res, nil
}

// Admin operations

func (s *service) GetCustomRequestAdmin(requestID uuid.UUID) (*CustomRequestRes, error) {
	request, err := s.repo.GetCustomRequestByIDWithDetails(requestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCustomRequestNotFound
		}
		return nil, fmt.Errorf("failed to get custom request: %w", err)
	}

	res := request.ToCustomRequestRes()
	return &res, nil
}

func (s *service) ListCustomRequestsAdmin(query CustomRequestListQuery) (*CustomRequestListRes, error) {
	requests, total, err := s.repo.ListCustomRequests(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list custom requests: %w", err)
	}

	// Convert to response DTOs
	var data []CustomRequestRes
	for _, req := range requests {
		data = append(data, req.ToCustomRequestRes())
	}

	// Calculate total pages
	query.SetDefaults()
	totalPages := int((total + int64(query.Limit) - 1) / int64(query.Limit))

	return &CustomRequestListRes{
		Data:       data,
		Total:      total,
		Page:       query.Page,
		Limit:      query.Limit,
		TotalPages: totalPages,
	}, nil
}

func (s *service) UpdateCustomRequestStatus(requestID uuid.UUID, req UpdateRequestStatusReq) (*CustomRequestRes, error) {
	customRequest, err := s.repo.GetCustomRequestByID(requestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCustomRequestNotFound
		}
		return nil, fmt.Errorf("failed to get custom request: %w", err)
	}

	// Validate status transition
	if !ValidateStatusTransition(customRequest.Status, req.Status) {
		return nil, ErrInvalidStatusTransition
	}

	// Update status
	customRequest.Status = req.Status
	if req.AssigneeID != nil {
		customRequest.AssigneeID = req.AssigneeID
	}

	if err := s.repo.UpdateCustomRequest(customRequest); err != nil {
		return nil, fmt.Errorf("failed to update custom request status: %w", err)
	}

	// Add message if notes provided
	if req.Notes != "" {
		message := &CustomRequestMessage{
			ID:              uuid.New(),
			CustomRequestID: requestID,
			SenderType:      SenderAdmin,
			SenderID:        uuid.New(), // This should be the admin ID from context
			Message:         req.Notes,
			CreatedAt:       time.Now(),
		}
		s.repo.CreateMessage(message) // Ignore error for message creation
	}

	// Fetch updated request
	updatedRequest, err := s.repo.GetCustomRequestByIDWithDetails(requestID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch updated custom request: %w", err)
	}

	res := updatedRequest.ToCustomRequestRes()
	return &res, nil
}

func (s *service) AssignCustomRequest(requestID uuid.UUID, assigneeID uuid.UUID) (*CustomRequestRes, error) {
	customRequest, err := s.repo.GetCustomRequestByID(requestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCustomRequestNotFound
		}
		return nil, fmt.Errorf("failed to get custom request: %w", err)
	}

	customRequest.AssigneeID = &assigneeID
	if err := s.repo.UpdateCustomRequest(customRequest); err != nil {
		return nil, fmt.Errorf("failed to assign custom request: %w", err)
	}

	// Fetch updated request
	updatedRequest, err := s.repo.GetCustomRequestByIDWithDetails(requestID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch updated custom request: %w", err)
	}

	res := updatedRequest.ToCustomRequestRes()
	return &res, nil
}

func (s *service) CreateQuote(adminID uuid.UUID, req CreateQuoteReq) (*QuoteRes, error) {
	// Check if custom request exists
	customRequest, err := s.repo.GetCustomRequestByIDWithDetails(req.CustomRequestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCustomRequestNotFound
		}
		return nil, fmt.Errorf("failed to get custom request: %w", err)
	}

	// Check if there's already an active quote
	activeQuote := customRequest.GetActiveQuote()
	if activeQuote != nil {
		return nil, ErrDuplicateActiveQuote
	}

	// Calculate totals
	var itemsSubtotal int64
	for _, item := range req.Items {
		itemsSubtotal += item.QuotedPrice
	}

	// Create quote
	quote := &Quote{
		ID:              uuid.New(),
		CustomRequestID: req.CustomRequestID,
		ItemsSubtotal:   itemsSubtotal,
		Fees:            req.Fees,
		Status:          QuoteDraft,
		ValidUntil:      req.ValidUntil,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	quote.CalculateTotal()

	if err := s.repo.CreateQuote(quote); err != nil {
		return nil, fmt.Errorf("failed to create quote: %w", err)
	}

	// Create quote items for each requested item
	for _, itemReq := range req.Items {
		quoteItem := &QuoteItem{
			ID:            uuid.New(),
			QuoteID:       quote.ID,
			RequestItemID: itemReq.RequestItemID,
			QuotedPrice:   itemReq.QuotedPrice,
			AdminNotes:    itemReq.AdminNotes,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		if err := s.repo.CreateQuoteItem(quoteItem); err != nil {
			return nil, fmt.Errorf("failed to create quote item: %w", err)
		}
	}

	// Update request items with quoted prices (for backward compatibility)
	for _, itemReq := range req.Items {
		for i, item := range customRequest.Items {
			if item.ID == itemReq.RequestItemID {
				customRequest.Items[i].QuotedPrice = &itemReq.QuotedPrice
				customRequest.Items[i].AdminNotes = itemReq.AdminNotes
				if err := s.repo.UpdateRequestItem(&customRequest.Items[i]); err != nil {
					return nil, fmt.Errorf("failed to update request item: %w", err)
				}
				break
			}
		}
	}

	// Fetch the quote with items for the response
	quoteWithItems, err := s.repo.GetQuoteByIDWithItems(quote.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch quote with items: %w", err)
	}

	res := quoteWithItems.ToQuoteRes()
	return &res, nil
}

func (s *service) UpdateQuote(quoteID uuid.UUID, req CreateQuoteReq) (*QuoteRes, error) {
	quote, err := s.repo.GetQuoteByID(quoteID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrQuoteNotFound
		}
		return nil, fmt.Errorf("failed to get quote: %w", err)
	}

	// Only draft quotes can be updated
	if quote.Status != QuoteDraft {
		return nil, ErrInvalidQuoteStatus
	}

	// Calculate totals
	var itemsSubtotal int64
	for _, item := range req.Items {
		itemsSubtotal += item.QuotedPrice
	}

	// Update quote
	quote.ItemsSubtotal = itemsSubtotal
	quote.Fees = req.Fees
	quote.ValidUntil = req.ValidUntil
	quote.CalculateTotal()

	if err := s.repo.UpdateQuote(quote); err != nil {
		return nil, fmt.Errorf("failed to update quote: %w", err)
	}

	// Get existing quote items to delete them
	existingItems, err := s.repo.GetQuoteItemsByQuoteID(quoteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing quote items: %w", err)
	}

	// Delete existing quote items
	for _, item := range existingItems {
		if err := s.repo.DeleteQuoteItem(item.ID); err != nil {
			return nil, fmt.Errorf("failed to delete existing quote item: %w", err)
		}
	}

	// Create new quote items
	for _, itemReq := range req.Items {
		quoteItem := &QuoteItem{
			ID:            uuid.New(),
			QuoteID:       quote.ID,
			RequestItemID: itemReq.RequestItemID,
			QuotedPrice:   itemReq.QuotedPrice,
			AdminNotes:    itemReq.AdminNotes,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		if err := s.repo.CreateQuoteItem(quoteItem); err != nil {
			return nil, fmt.Errorf("failed to create quote item: %w", err)
		}
	}

	// Small delay to ensure items are committed
	time.Sleep(10 * time.Millisecond)

	// Get updated quote with items for response
	updatedQuote, err := s.repo.GetQuoteByIDWithItems(quoteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated quote: %w", err)
	}

	res := updatedQuote.ToQuoteRes()
	return &res, nil
}

func (s *service) SendQuote(quoteID uuid.UUID) (*QuoteRes, error) {
	quote, err := s.repo.GetQuoteByID(quoteID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrQuoteNotFound
		}
		return nil, fmt.Errorf("failed to get quote: %w", err)
	}

	// Only draft quotes can be sent
	if quote.Status != QuoteDraft {
		return nil, ErrInvalidQuoteStatus
	}

	// Update quote status
	now := time.Now()
	quote.Status = QuoteSent
	quote.SentAt = &now

	if err := s.repo.UpdateQuote(quote); err != nil {
		return nil, fmt.Errorf("failed to send quote: %w", err)
	}

	// Update custom request status
	customRequest, err := s.repo.GetCustomRequestByID(quote.CustomRequestID)
	if err == nil {
		customRequest.Status = RequestQuoteSent
		if updateErr := s.repo.UpdateCustomRequest(customRequest); updateErr != nil {
			// Log error but don't fail the quote sending
			fmt.Printf("Warning: failed to update custom request status: %v\n", updateErr)
		}
	}

	res := quote.ToQuoteRes()
	return &res, nil
}

func (s *service) SendMessageAdmin(adminID uuid.UUID, requestID uuid.UUID, req SendMessageReq) (*CustomRequestMsgRes, error) {
	// Check if custom request exists
	if _, err := s.repo.GetCustomRequestByID(requestID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCustomRequestNotFound
		}
		return nil, fmt.Errorf("failed to get custom request: %w", err)
	}

	// Create message
	message := &CustomRequestMessage{
		ID:              uuid.New(),
		CustomRequestID: requestID,
		SenderType:      SenderAdmin,
		SenderID:        adminID,
		Message:         req.Message,
		CreatedAt:       time.Now(),
	}

	if err := s.repo.CreateMessage(message); err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	res := message.ToCustomRequestMsgRes()
	return &res, nil
}

// Analytics and reporting

func (s *service) GetCustomRequestStats() (*CustomRequestStatsRes, error) {
	return s.repo.GetCustomRequestStats()
}

func (s *service) GetCustomRequestStatsByDateRange(startDate, endDate time.Time) (*CustomRequestStatsRes, error) {
	return s.repo.GetCustomRequestStatsByDateRange(startDate, endDate)
}

// Background tasks

func (s *service) ProcessExpiredRequests() error {
	expiredRequests, err := s.repo.GetExpiredCustomRequests()
	if err != nil {
		return fmt.Errorf("failed to get expired requests: %w", err)
	}

	for _, request := range expiredRequests {
		request.Status = RequestCancelled
		if err := s.repo.UpdateCustomRequest(&request); err != nil {
			// Log error but continue processing
			continue
		}
	}

	return nil
}

func (s *service) ProcessExpiredQuotes() error {
	expiredQuotes, err := s.repo.GetExpiredQuotes()
	if err != nil {
		return fmt.Errorf("failed to get expired quotes: %w", err)
	}

	for _, quote := range expiredQuotes {
		quote.Status = QuoteDeclined
		if err := s.repo.UpdateQuote(&quote); err != nil {
			// Log error but continue processing
			continue
		}

		// Update custom request status back to under review
		customRequest, err := s.repo.GetCustomRequestByID(quote.CustomRequestID)
		if err == nil {
			customRequest.Status = RequestUnderReview
			s.repo.UpdateCustomRequest(customRequest) // Ignore error
		}
	}

	return nil
}

func (s *service) CleanupOldMessages(olderThan time.Time) error {
	// This would require a new repository method to delete messages older than a certain date
	// For now, we'll leave this as a placeholder
	return nil
}

// Bulk operations

func (s *service) BulkUpdateStatus(req BulkUpdateStatusReq) (*BulkUpdateStatusRes, error) {
	if len(req.RequestIDs) == 0 {
		return nil, ErrEmptyRequestItems
	}

	err := s.repo.BulkUpdateCustomRequestStatus(req.RequestIDs, req.Status, req.AssigneeID)
	if err != nil {
		return nil, fmt.Errorf("failed to bulk update status: %w", err)
	}

	return &BulkUpdateStatusRes{
		UpdatedCount: len(req.RequestIDs),
		RequestIDs:   req.RequestIDs,
		Status:       req.Status,
	}, nil
}

func (s *service) BulkAssign(req BulkAssignReq) (*BulkAssignRes, error) {
	if len(req.RequestIDs) == 0 {
		return nil, ErrEmptyRequestItems
	}

	err := s.repo.BulkUpdateCustomRequestStatus(req.RequestIDs, "", &req.AssigneeID)
	if err != nil {
		return nil, fmt.Errorf("failed to bulk assign: %w", err)
	}

	return &BulkAssignRes{
		AssignedCount: len(req.RequestIDs),
		RequestIDs:    req.RequestIDs,
		AssigneeID:    req.AssigneeID,
	}, nil
}

// CancelCustomRequestAdmin allows admins to cancel any custom request regardless of status
func (s *service) CancelCustomRequestAdmin(adminID uuid.UUID, requestID uuid.UUID, reason string) (*CustomRequestRes, error) {
	request, err := s.repo.GetCustomRequestByID(requestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCustomRequestNotFound
		}
		return nil, fmt.Errorf("failed to get custom request: %w", err)
	}

	// Check if request is already cancelled
	if request.Status == RequestCancelled {
		return nil, errors.New("custom request is already cancelled")
	}

	// Admin can cancel any request regardless of status
	// No ownership check required for admin operations

	// Update status to cancelled
	request.Status = RequestCancelled
	request.UpdatedAt = time.Now()

	// Add a message about the cancellation
	message := &CustomRequestMessage{
		ID:               uuid.New(),
		CustomRequestID:  requestID,
		SenderType:       SenderAdmin,
		SenderID:         adminID,
		Message:          fmt.Sprintf("Request cancelled by admin. Reason: %s", reason),
		CreatedAt:        time.Now(),
	}
	request.Messages = append(request.Messages, *message)

	if err := s.repo.UpdateCustomRequest(request); err != nil {
		return nil, fmt.Errorf("failed to cancel custom request: %w", err)
	}

	res := request.ToCustomRequestRes()
	return &res, nil
}

// PermanentlyDeleteCustomRequestAdmin allows admins to permanently delete any custom request regardless of status
func (s *service) PermanentlyDeleteCustomRequestAdmin(adminID uuid.UUID, requestID uuid.UUID) error {
	// Check if request exists
	_, err := s.repo.GetCustomRequestByID(requestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCustomRequestNotFound
		}
		return fmt.Errorf("failed to get custom request: %w", err)
	}

	// Admin can permanently delete any request regardless of status
	// No ownership check required for admin operations
	// No status restrictions for admin operations

	// Permanently delete the request and all related data
	return s.repo.DeleteCustomRequest(requestID)
}