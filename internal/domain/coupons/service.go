package coupons

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Service interface {
	// Admin Coupon Management
	CreateCoupon(req CreateCouponRequest, createdByUserID *uuid.UUID) (*CouponResponse, error)
	GetCoupon(id uuid.UUID) (*CouponResponse, error)
	UpdateCoupon(id uuid.UUID, req UpdateCouponRequest) (*CouponResponse, error)
	DeleteCoupon(id uuid.UUID) error
	ListCoupons(page, limit int, filters map[string]interface{}) (*CouponListResponse, error)
	ToggleCouponActive(id uuid.UUID) (*CouponResponse, error)
	
	// User Coupon Operations
	GetAvailableCoupons(userID uuid.UUID, page, limit int) (*CouponListResponse, error)
	ValidateCoupon(req ValidateCouponRequest) (*CouponValidationResponse, error)
	ApplyCoupon(req ApplyCouponRequest) (*CouponValidationResponse, error)
	
	// Refund Credits
	GetUserRefundCredits(userID uuid.UUID, page, limit int) (*RefundCreditListResponse, error)
	CreateRefundCredit(req CreateRefundCreditRequest) (*UserRefundCreditResponse, error)
	ConvertRefundToCredit(req ConvertRefundRequest) (*CouponResponse, error)
	
	// System Operations
	GenerateRefundCoupon(orderID, userID uuid.UUID, refundAmount float64) (*CouponResponse, error)
	AutoGenerateUserCoupon(userID uuid.UUID, couponType CouponType, value float64, description string) (*CouponResponse, error)
	
	// Mobile App Operations
	MobileAutoGenerateCoupon(userID uuid.UUID, req MobileAutoGenerateCouponRequest) (*CouponResponse, error)
	
	// Analytics
	GetCouponStats() (*CouponStatsResponse, error)
	
	// Public operations
	GetPublicCoupons() ([]CouponResponse, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// Admin Coupon Management
func (s *service) CreateCoupon(req CreateCouponRequest, createdByUserID *uuid.UUID) (*CouponResponse, error) {
	// Validate coupon code uniqueness
	existingCoupon, err := s.repo.GetByCode(req.Code)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("error checking coupon code: %w", err)
	}
	if existingCoupon != nil {
		return nil, errors.New("coupon code already exists")
	}
	
	// Validate percentage value
	if req.Type == CouponPercentage && req.Value > 100 {
		return nil, errors.New("percentage discount cannot exceed 100%")
	}
	
	// Set default values
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	
	coupon := &Coupon{
		ID:                 uuid.New(),
		Code:               strings.ToUpper(req.Code),
		Type:               req.Type,
		Value:              req.Value,
		Description:        req.Description,
		MaxUsage:           req.MaxUsage,
		UsageCount:         0,
		ExpiryDate:         req.ExpiryDate,
		IsActive:           isActive,
		CreatedBy:          string(CreatedByOwner),
		CreatedByUserID:    createdByUserID,
		LinkedUserID:       req.LinkedUserID,
		MinimumOrderAmount: req.MinimumOrderAmount,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
	
	err = s.repo.Create(coupon)
	if err != nil {
		return nil, fmt.Errorf("error creating coupon: %w", err)
	}
	
	return s.toCouponResponse(coupon), nil
}

func (s *service) GetCoupon(id uuid.UUID) (*CouponResponse, error) {
	coupon, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("coupon not found")
		}
		return nil, fmt.Errorf("error getting coupon: %w", err)
	}
	
	return s.toCouponResponse(coupon), nil
}

func (s *service) UpdateCoupon(id uuid.UUID, req UpdateCouponRequest) (*CouponResponse, error) {
	coupon, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("coupon not found")
		}
		return nil, fmt.Errorf("error getting coupon: %w", err)
	}
	
	// Update fields if provided
	if req.Description != nil {
		coupon.Description = *req.Description
	}
	if req.MaxUsage != nil {
		coupon.MaxUsage = req.MaxUsage
	}
	if req.ExpiryDate != nil {
		coupon.ExpiryDate = req.ExpiryDate
	}
	if req.IsActive != nil {
		coupon.IsActive = *req.IsActive
	}
	if req.MinimumOrderAmount != nil {
		coupon.MinimumOrderAmount = *req.MinimumOrderAmount
	}
	
	coupon.UpdatedAt = time.Now()
	
	err = s.repo.Update(coupon)
	if err != nil {
		return nil, fmt.Errorf("error updating coupon: %w", err)
	}
	
	return s.toCouponResponse(coupon), nil
}

func (s *service) DeleteCoupon(id uuid.UUID) error {
	_, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("coupon not found")
		}
		return fmt.Errorf("error getting coupon: %w", err)
	}
	
	return s.repo.Delete(id)
}

func (s *service) ListCoupons(page, limit int, filters map[string]interface{}) (*CouponListResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	
	coupons, total, err := s.repo.List(page, limit, filters)
	if err != nil {
		return nil, fmt.Errorf("error listing coupons: %w", err)
	}
	
	responses := make([]CouponResponse, len(coupons))
	for i, coupon := range coupons {
		responses[i] = *s.toCouponResponse(&coupon)
	}
	
	return &CouponListResponse{
		Coupons: responses,
		Total:   total,
		Page:    page,
		Limit:   limit,
	}, nil
}

func (s *service) ToggleCouponActive(id uuid.UUID) (*CouponResponse, error) {
	err := s.repo.ToggleActive(id)
	if err != nil {
		return nil, fmt.Errorf("error toggling coupon status: %w", err)
	}
	
	return s.GetCoupon(id)
}

// User Coupon Operations
func (s *service) GetAvailableCoupons(userID uuid.UUID, page, limit int) (*CouponListResponse, error) {
	filters := map[string]interface{}{
		"is_active": true,
	}
	
	// Include general coupons and user-specific coupons
	coupons, _, err := s.repo.List(page, limit, filters)
	if err != nil {
		return nil, fmt.Errorf("error getting available coupons: %w", err)
	}
	
	// Filter out expired coupons and user-specific coupons for other users
	var availableCoupons []CouponResponse
	now := time.Now()
	
	for _, coupon := range coupons {
		// Skip expired coupons
		if coupon.ExpiryDate != nil && coupon.ExpiryDate.Before(now) {
			continue
		}
		
		// Skip user-specific coupons for other users
		if coupon.LinkedUserID != nil && *coupon.LinkedUserID != userID {
			continue
		}
		
		// Skip coupons that have reached usage limit
		if coupon.MaxUsage != nil && coupon.UsageCount >= *coupon.MaxUsage {
			continue
		}
		
		availableCoupons = append(availableCoupons, *s.toCouponResponse(&coupon))
	}
	
	return &CouponListResponse{
		Coupons: availableCoupons,
		Total:   int64(len(availableCoupons)),
		Page:    page,
		Limit:   limit,
	}, nil
}

func (s *service) ValidateCoupon(req ValidateCouponRequest) (*CouponValidationResponse, error) {
	coupon, err := s.repo.GetByCode(req.Code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &CouponValidationResponse{
				Valid:   false,
				Message: "Coupon code not found",
			}, nil
		}
		return nil, fmt.Errorf("error getting coupon: %w", err)
	}
	
	validation := s.validateCouponForUser(coupon, req.UserID, req.OrderAmount)
	validation.Coupon = s.toCouponResponse(coupon)
	
	return validation, nil
}

func (s *service) ApplyCoupon(req ApplyCouponRequest) (*CouponValidationResponse, error) {
	coupon, err := s.repo.GetByCode(req.Code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &CouponValidationResponse{
				Valid:   false,
				Message: "Coupon code not found",
			}, nil
		}
		return nil, fmt.Errorf("error getting coupon: %w", err)
	}
	
	validation := s.validateCouponForUser(coupon, req.UserID, req.OrderAmount)
	if !validation.Valid {
		return validation, nil
	}
	
	// Create usage record
	usage := &CouponUsage{
		ID:             uuid.New(),
		CouponID:       coupon.ID,
		UserID:         req.UserID,
		OrderID:        req.OrderID,
		DiscountAmount: validation.DiscountAmount,
		UsedAt:         time.Now(),
	}
	
	err = s.repo.CreateUsage(usage)
	if err != nil {
		return nil, fmt.Errorf("error creating coupon usage: %w", err)
	}
	
	// Update coupon usage count
	coupon.UsageCount++
	err = s.repo.Update(coupon)
	if err != nil {
		return nil, fmt.Errorf("error updating coupon usage count: %w", err)
	}
	
	validation.Coupon = s.toCouponResponse(coupon)
	return validation, nil
}

// Refund Credits
func (s *service) GetUserRefundCredits(userID uuid.UUID, page, limit int) (*RefundCreditListResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	
	credits, total, err := s.repo.GetRefundCreditsByUser(userID, page, limit)
	if err != nil {
		return nil, fmt.Errorf("error getting refund credits: %w", err)
	}
	
	responses := make([]UserRefundCreditResponse, len(credits))
	for i, credit := range credits {
		responses[i] = s.toRefundCreditResponse(&credit)
	}
	
	return &RefundCreditListResponse{
		Credits: responses,
		Total:   total,
		Page:    page,
		Limit:   limit,
	}, nil
}

func (s *service) CreateRefundCredit(req CreateRefundCreditRequest) (*UserRefundCreditResponse, error) {
	credit := &UserRefundCredit{
		ID:                uuid.New(),
		UserID:            req.UserID,
		OriginalOrderID:   req.OrderID,
		RefundAmount:      req.RefundAmount,
		ConvertedToCoupon: false,
		CreatedAt:         time.Now(),
	}
	
	err := s.repo.CreateRefundCredit(credit)
	if err != nil {
		return nil, fmt.Errorf("error creating refund credit: %w", err)
	}
	
	response := s.toRefundCreditResponse(credit)
	return &response, nil
}

func (s *service) ConvertRefundToCredit(req ConvertRefundRequest) (*CouponResponse, error) {
	credit, err := s.repo.GetRefundCreditByID(req.RefundCreditID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("refund credit not found")
		}
		return nil, fmt.Errorf("error getting refund credit: %w", err)
	}
	
	if credit.ConvertedToCoupon {
		return nil, errors.New("refund credit already converted to coupon")
	}
	
	// Generate coupon code if not provided
	couponCode := fmt.Sprintf("REFUND-%s-001", credit.UserID.String()[:8])
	if req.CouponCode != nil && *req.CouponCode != "" {
		couponCode = *req.CouponCode
	}
	
	// Create coupon
	coupon := &Coupon{
		ID:                 uuid.New(),
		Code:               strings.ToUpper(couponCode),
		Type:               CouponFixed,
		Value:              credit.RefundAmount,
		Description:        fmt.Sprintf("Refund credit from order %s", credit.OriginalOrderID.String()[:8]),
		MaxUsage:           &[]int{1}[0], // One-time use
		UsageCount:         0,
		IsActive:           true,
		CreatedBy:          string(CreatedBySystem),
		LinkedOrderID:      &credit.OriginalOrderID,
		LinkedUserID:       &credit.UserID,
		MinimumOrderAmount: 0,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
	
	err = s.repo.Create(coupon)
	if err != nil {
		return nil, fmt.Errorf("error creating refund coupon: %w", err)
	}
	
	// Update refund credit
	credit.ConvertedToCoupon = true
	credit.CouponID = &coupon.ID
	err = s.repo.UpdateRefundCredit(credit)
	if err != nil {
		return nil, fmt.Errorf("error updating refund credit: %w", err)
	}
	
	return s.toCouponResponse(coupon), nil
}

// System Operations
func (s *service) GenerateRefundCoupon(orderID, userID uuid.UUID, refundAmount float64) (*CouponResponse, error) {
	couponCode := fmt.Sprintf("SORRY-%s", orderID.String()[:8])
	
	coupon := &Coupon{
		ID:                 uuid.New(),
		Code:               couponCode,
		Type:               CouponFixed,
		Value:              refundAmount,
		Description:        fmt.Sprintf("Apology coupon for order %s", orderID.String()[:8]),
		MaxUsage:           &[]int{1}[0], // One-time use
		UsageCount:         0,
		IsActive:           true,
		CreatedBy:          string(CreatedBySystem),
		LinkedOrderID:      &orderID,
		LinkedUserID:       &userID,
		MinimumOrderAmount: 0,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
	
	err := s.repo.Create(coupon)
	if err != nil {
		return nil, fmt.Errorf("error creating refund coupon: %w", err)
	}
	
	return s.toCouponResponse(coupon), nil
}

func (s *service) AutoGenerateUserCoupon(userID uuid.UUID, couponType CouponType, value float64, description string) (*CouponResponse, error) {
	couponCode := fmt.Sprintf("USER-%s-%d", userID.String()[:8], time.Now().Unix())
	
	coupon := &Coupon{
		ID:                 uuid.New(),
		Code:               couponCode,
		Type:               couponType,
		Value:              value,
		Description:        description,
		MaxUsage:           &[]int{1}[0], // One-time use
		UsageCount:         0,
		IsActive:           true,
		CreatedBy:          string(CreatedBySystem),
		LinkedUserID:       &userID,
		MinimumOrderAmount: 0,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
	
	err := s.repo.Create(coupon)
	if err != nil {
		return nil, fmt.Errorf("error creating user coupon: %w", err)
	}
	
	return s.toCouponResponse(coupon), nil
}

// MobileAutoGenerateCoupon allows mobile users to create their own coupons with restrictions
func (s *service) MobileAutoGenerateCoupon(userID uuid.UUID, req MobileAutoGenerateCouponRequest) (*CouponResponse, error) {
	// Check if user has reached daily limit (max 3 coupons per day)
	startOfDay := time.Now().Truncate(24 * time.Hour)
	filters := map[string]interface{}{
		"linked_user_id": userID,
		"created_by":     string(CreatedBySystem),
		"created_at >=":  startOfDay,
	}
	
	coupons, _, err := s.repo.List(1, 10, filters)
	if err != nil {
		return nil, fmt.Errorf("error checking daily coupon limit: %w", err)
	}
	
	if len(coupons) >= 3 {
		return nil, errors.New("daily coupon generation limit reached (3 per day)")
	}
	
	// Generate unique coupon code
	couponCode := fmt.Sprintf("MOBILE-%s-%d", userID.String()[:8], time.Now().Unix())
	
	// Set expiry date
	var expiryDate *time.Time
	if req.ExpiryDays != nil {
		expiry := time.Now().AddDate(0, 0, *req.ExpiryDays)
		expiryDate = &expiry
	} else {
		// Default to 30 days
		expiry := time.Now().AddDate(0, 0, 30)
		expiryDate = &expiry
	}
	
	// Additional validation for percentage coupons
	if req.Type == CouponPercentage && req.Value > 50 {
		return nil, errors.New("mobile-generated percentage coupons cannot exceed 50%")
	}
	
	// Additional validation for fixed coupons
	if req.Type == CouponFixed && req.Value > 50 {
		return nil, errors.New("mobile-generated fixed coupons cannot exceed $50")
	}
	
	coupon := &Coupon{
		ID:                 uuid.New(),
		Code:               couponCode,
		Type:               req.Type,
		Value:              req.Value,
		Description:        req.Description,
		MaxUsage:           &[]int{1}[0], // One-time use for mobile generated
		UsageCount:         0,
		ExpiryDate:         expiryDate,
		IsActive:           true,
		CreatedBy:          string(CreatedBySystem),
		LinkedUserID:       &userID,
		MinimumOrderAmount: 0, // No minimum for mobile generated
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
	
	err = s.repo.Create(coupon)
	if err != nil {
		return nil, fmt.Errorf("error creating mobile coupon: %w", err)
	}
	
	return s.toCouponResponse(coupon), nil
}

// Analytics
func (s *service) GetCouponStats() (*CouponStatsResponse, error) {
	return s.repo.GetCouponStats()
}

// GetPublicCoupons gets publicly available coupons
func (s *service) GetPublicCoupons() ([]CouponResponse, error) {
	coupons, err := s.repo.GetPublicCoupons()
	if err != nil {
		return nil, err
	}
	
	var responses []CouponResponse
	for _, coupon := range coupons {
		responses = append(responses, *s.toCouponResponse(&coupon))
	}
	
	return responses, nil
}

// Helper methods
func (s *service) validateCouponForUser(coupon *Coupon, userID uuid.UUID, orderAmount float64) *CouponValidationResponse {
	now := time.Now()
	
	// Check if coupon is active
	if !coupon.IsActive {
		return &CouponValidationResponse{
			Valid:   false,
			Message: "Coupon is not active",
		}
	}
	
	// Check expiry date
	if coupon.ExpiryDate != nil && coupon.ExpiryDate.Before(now) {
		return &CouponValidationResponse{
			Valid:   false,
			Message: "Coupon has expired",
		}
	}
	
	// Check usage limit
	if coupon.MaxUsage != nil && coupon.UsageCount >= *coupon.MaxUsage {
		return &CouponValidationResponse{
			Valid:   false,
			Message: "Coupon usage limit reached",
		}
	}
	
	// Check user-specific coupon
	if coupon.LinkedUserID != nil && *coupon.LinkedUserID != userID {
		return &CouponValidationResponse{
			Valid:   false,
			Message: "Coupon is not valid for this user",
		}
	}
	
	// Check minimum order amount
	if orderAmount < coupon.MinimumOrderAmount {
		return &CouponValidationResponse{
			Valid:   false,
			Message: fmt.Sprintf("Minimum order amount is %.2f", coupon.MinimumOrderAmount),
		}
	}
	
	// Check if user has already used this coupon (for one-time use coupons)
	if coupon.LinkedUserID != nil {
		usageCount, err := s.repo.GetUsageCountByUserAndCoupon(userID, coupon.ID)
		if err == nil && usageCount > 0 {
			return &CouponValidationResponse{
				Valid:   false,
				Message: "Coupon already used by this user",
			}
		}
	}
	
	// Calculate discount amount
	discountAmount := s.calculateDiscount(coupon, orderAmount)
	
	return &CouponValidationResponse{
		Valid:          true,
		DiscountAmount: discountAmount,
		Message:        "Coupon is valid",
	}
}

func (s *service) calculateDiscount(coupon *Coupon, orderAmount float64) float64 {
	switch coupon.Type {
	case CouponPercentage:
		return math.Round((orderAmount * coupon.Value / 100) * 100) / 100
	case CouponFixed:
		if coupon.Value > orderAmount {
			return orderAmount
		}
		return coupon.Value
	default:
		return 0
	}
}

func (s *service) toCouponResponse(coupon *Coupon) *CouponResponse {
	return &CouponResponse{
		ID:                 coupon.ID,
		Code:               coupon.Code,
		Type:               coupon.Type,
		Value:              coupon.Value,
		Description:        coupon.Description,
		MaxUsage:           coupon.MaxUsage,
		UsageCount:         coupon.UsageCount,
		ExpiryDate:         coupon.ExpiryDate,
		IsActive:           coupon.IsActive,
		CreatedBy:          coupon.CreatedBy,
		LinkedOrderID:      coupon.LinkedOrderID,
		LinkedUserID:       coupon.LinkedUserID,
		MinimumOrderAmount: coupon.MinimumOrderAmount,
		CreatedAt:          coupon.CreatedAt,
		UpdatedAt:          coupon.UpdatedAt,
	}
}

func (s *service) toRefundCreditResponse(credit *UserRefundCredit) UserRefundCreditResponse {
	response := UserRefundCreditResponse{
		ID:                credit.ID,
		UserID:            credit.UserID,
		OriginalOrderID:   credit.OriginalOrderID,
		RefundAmount:      credit.RefundAmount,
		ConvertedToCoupon: credit.ConvertedToCoupon,
		CouponID:          credit.CouponID,
		CreatedAt:         credit.CreatedAt,
	}
	
	if credit.Coupon != nil {
		response.Coupon = s.toCouponResponse(credit.Coupon)
	}
	
	return response
}