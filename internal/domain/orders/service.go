package orders

import (
    "context"
    "errors"
    "fmt"
    "strconv"
    "time"

    "errandShop/internal/domain/auth"
    "errandShop/internal/domain/products"
    "errandShop/internal/domain/coupons"
    "errandShop/internal/domain/customers"
    "errandShop/internal/domain/notifications"
    "errandShop/internal/domain/custom_requests"
    "errandShop/internal/domain/payments"
    "errandShop/internal/core/types"
    "github.com/google/uuid"
    "gorm.io/gorm"
)

// Service interfaces
type AuthServiceInterface interface {
	GetUserByID(ctx context.Context, userID uuid.UUID) (*auth.UserResponse, error)
}

type PaymentServiceInterface interface {
	InitializePayment(req payments.CreatePaymentRequest, customerID uint) (*payments.PaymentInitResponse, error)
}

type DeliveryServiceInterface interface {
	CalculateDeliveryFee(distance float64, deliveryType string) int64
}

type AddressRepoInterface interface {
	GetByID(userID, addressID string) (*types.Address, error)
}

type DeliveryMatcherInterface interface {
	MatchAddress(address string) (*types.MatchResult, *types.NoMatchResult)
}

type Service struct {
	repo        *Repository
	productRepo *products.Repository
	couponService coupons.Service
	customerService customers.Service
	authService AuthServiceInterface
	paymentService PaymentServiceInterface
	deliveryService DeliveryServiceInterface
	addressRepo AddressRepoInterface
	deliveryMatcher DeliveryMatcherInterface
	cartService *CartService
	notificationService notifications.NotificationService
	customRequestService custom_requests.Service
	db          *gorm.DB
}

func NewService(repo *Repository, productRepo *products.Repository, couponService coupons.Service, customerService customers.Service, authService AuthServiceInterface, paymentService PaymentServiceInterface, deliveryService DeliveryServiceInterface, addressRepo AddressRepoInterface, deliveryMatcher DeliveryMatcherInterface, notificationService notifications.NotificationService, customRequestService custom_requests.Service, db *gorm.DB) *Service {
	return &Service{
		repo:        repo,
		productRepo: productRepo,
		couponService: couponService,
		customerService: customerService,
		authService: authService,
		paymentService: paymentService,
		deliveryService: deliveryService,
		addressRepo: addressRepo,
		deliveryMatcher: deliveryMatcher,
		cartService: NewCartService(db, productRepo),
		notificationService: notificationService,
		customRequestService: customRequestService,
		db:          db,
	}
}

type PageMeta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"totalPages"`
}

type ListResult struct {
	Data []OrderResponse `json:"data"`
	Meta PageMeta        `json:"meta"`
}

// Cart methods
func (s *Service) GetCart(ctx context.Context, userID uuid.UUID) (*CartResponse, error) {
	cart, err := s.cartService.GetCart(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cart: %w", err)
	}
	return s.toCartResponse(*cart), nil
}

func (s *Service) AddToCart(ctx context.Context, userID uuid.UUID, req AddToCartRequest) (*CartResponse, error) {
	cart, err := s.cartService.AddToCart(userID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to add to cart: %w", err)
	}
	return s.toCartResponse(*cart), nil
}

func (s *Service) UpdateCartItem(ctx context.Context, userID uuid.UUID, itemID uuid.UUID, req UpdateCartItemRequest) (*CartResponse, error) {
	cart, err := s.cartService.UpdateCartItem(userID, itemID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update cart item: %w", err)
	}
	return s.toCartResponse(*cart), nil
}

func (s *Service) RemoveFromCart(ctx context.Context, userID uuid.UUID, itemID uuid.UUID) (*CartResponse, error) {
	cart, err := s.cartService.RemoveFromCart(userID, itemID)
	if err != nil {
		return nil, fmt.Errorf("failed to remove from cart: %w", err)
	}
	return s.toCartResponse(*cart), nil
}

func (s *Service) ClearCart(ctx context.Context, userID uuid.UUID) error {
	return s.cartService.ClearCart(userID)
}

// Order methods
func (s *Service) List(ctx context.Context, userID uuid.UUID, query ListQuery) (*ListResult, error) {
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.Limit <= 0 || query.Limit > 100 {
		query.Limit = 20
	}

	orders, total, err := s.repo.List(ctx, userID, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list orders: %w", err)
	}

	responses := make([]OrderResponse, len(orders))
	for i, order := range orders {
		responses[i] = *s.toOrderResponseWithContext(ctx, &order)
	}

	totalPages := int((total + int64(query.Limit) - 1) / int64(query.Limit))

	return &ListResult{
		Data: responses,
		Meta: PageMeta{
			Page:       query.Page,
			Limit:      query.Limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

func (s *Service) Get(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*OrderResponse, error) {
	order, err := s.repo.Get(ctx, id, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("order not found")
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	response := s.toOrderResponseWithContext(ctx, order)
	return response, nil
}

func (s *Service) CreateFromCart(ctx context.Context, userID uuid.UUID, req CreateOrderFromCartRequest) (*OrderResponse, error) {
	// Get and validate cart
	cart, warnings, err := s.cartService.ValidateCartForCheckout(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to validate cart: %w", err)
	}
	if len(warnings) > 0 {
		return nil, fmt.Errorf("cart validation failed: %v", warnings)
	}

	// Convert cart to order items
	orderItems, err := s.cartService.ConvertCartToOrderItems(cart)
	if err != nil {
		return nil, fmt.Errorf("failed to convert cart to order items: %w", err)
	}

	// Create order request
	createReq := CreateOrderRequest{
		DeliveryAddressID: req.DeliveryAddressID,
		Items:             orderItems,
		CouponCode:        req.CouponCode,
		Notes:             req.Notes,
		IdempotencyKey:    req.IdempotencyKey,
	}

	// Create order
	order, err := s.Create(ctx, userID, createReq)
	if err != nil {
		return nil, err
	}

	// Clear cart after successful order creation
	if err := s.cartService.ClearCart(userID); err != nil {
		// Log warning but don't fail the order
		fmt.Printf("Warning: failed to clear cart after order creation: %v\n", err)
	}

	return order, nil
}

func (s *Service) Create(ctx context.Context, userID uuid.UUID, req CreateOrderRequest) (*OrderResponse, error) {
	// Validate that order has either items or custom requests
	if len(req.Items) == 0 && len(req.CustomRequests) == 0 {
		return nil, fmt.Errorf("order must contain at least one item or custom request")
	}

	// Check for duplicate order using idempotency key
	if req.IdempotencyKey != "" {
		existingOrder, err := s.repo.CheckIdempotency(ctx, userID, req.IdempotencyKey)
		if err != nil && err != gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("failed to check idempotency: %w", err)
		}
		if existingOrder != nil {
			response := s.toOrderResponseWithContext(ctx, existingOrder)
			return response, nil
		}
	}

	// Validate and calculate total
	var subtotalKobo int64
	var customRequestsTotal int64
	var orderItems []OrderItem
	if len(req.Items) > 0 {
		orderItems = make([]OrderItem, len(req.Items))
	}

	// Process custom requests if provided
	if len(req.CustomRequests) > 0 {
		for _, customReq := range req.CustomRequests {
			// Get custom request with details
			customRequest, err := s.customRequestService.GetCustomRequest(userID, customReq.CustomRequestID)
			if err != nil {
				return nil, fmt.Errorf("failed to get custom request %s: %w", customReq.CustomRequestID, err)
			}

			// Validate that the custom request is accepted and has an active quote
			if customRequest.Status != custom_requests.RequestCustomerAccepted {
				return nil, fmt.Errorf("custom request %s is not in accepted status", customReq.CustomRequestID)
			}

			// Validate that the custom request is not expired
			if customRequest.ExpiresAt != nil && time.Now().After(*customRequest.ExpiresAt) {
				return nil, fmt.Errorf("custom request %s has expired", customReq.CustomRequestID)
			}

			if customRequest.ActiveQuote == nil {
				return nil, fmt.Errorf("custom request %s has no active quote", customReq.CustomRequestID)
			}

			// Add the quote total to the custom requests total
			customRequestsTotal += customRequest.ActiveQuote.GrandTotal
		}
	}

	for i, item := range req.Items {
		// Get product to validate and get current price
		product, err := s.productRepo.GetByID(ctx, item.ProductID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, fmt.Errorf("product with ID %s not found", item.ProductID)
			}
			return nil, fmt.Errorf("failed to get product: %w", err)
		}

		// Check stock availability
		if product.StockQuantity < item.Quantity {
			return nil, fmt.Errorf("insufficient stock for product %s. Available: %d, Requested: %d", product.Name, product.StockQuantity, item.Quantity)
		}

		// Convert product price from naira to kobo
		unitPriceKobo := int64(product.SellingPrice * 100)
		itemTotal := unitPriceKobo * int64(item.Quantity)
		subtotalKobo += itemTotal

		orderItems[i] = OrderItem{
			ProductID:  item.ProductID,
			Name:       product.Name,
			SKU:        product.SKU,
			Quantity:   item.Quantity,
			UnitPrice:  unitPriceKobo,
			TotalPrice: itemTotal,
			Source:     "catalog",
		}
	}

	// Apply coupon if provided
	var discountKobo int64
	if req.CouponCode != nil && *req.CouponCode != "" {
		// Validate coupon
		validationReq := coupons.ValidateCouponRequest{
			Code:        *req.CouponCode,
			UserID:      userID,
			OrderAmount: float64(subtotalKobo), // Keep in kobo format
		}
		
		validation, err := s.couponService.ValidateCoupon(validationReq)
		if err != nil {
			return nil, fmt.Errorf("failed to validate coupon: %w", err)
		}
		
		if !validation.Valid {
			return nil, fmt.Errorf("invalid coupon: %s", validation.Message)
		}
		
		// Keep discount amount as is (already in kobo)
		discountKobo = int64(validation.DiscountAmount)
	}

	// Calculate delivery fee based on delivery zone
	var deliveryFeeKobo int64 = 0
	if req.DeliveryAddressID != nil && *req.DeliveryAddressID != "" {
		// Get the delivery address
		address, err := s.addressRepo.GetByID(userID.String(), *req.DeliveryAddressID)
		if err != nil {
			return nil, fmt.Errorf("failed to get delivery address: %w", err)
		}
		
		// Match address to delivery zone
		fullAddress := address.Text
		matchResult, noMatchResult := s.deliveryMatcher.MatchAddress(fullAddress)

		if matchResult != nil {
			// Use zone-based pricing
			deliveryFeeKobo = int64(matchResult.Price * 100) // Convert to kobo
		} else if noMatchResult != nil {
			// Use fallback pricing for unmatched zones
			deliveryFeeKobo = s.deliveryService.CalculateDeliveryFee(5.0, "standard")
		} else {
			// Default fallback
			deliveryFeeKobo = s.deliveryService.CalculateDeliveryFee(5.0, "standard")
		}
	} else {
		// No delivery address provided, use default
		deliveryFeeKobo = s.deliveryService.CalculateDeliveryFee(5.0, "standard")
	}
	
	// Service fee (can be a fixed amount or percentage - using 5% of subtotal as example)
	serviceFeeKobo := subtotalKobo * 5 / 100 // 5% service fee

	// Calculate total including custom requests, delivery fee, and service fee
	totalKobo := subtotalKobo + customRequestsTotal + deliveryFeeKobo + serviceFeeKobo - discountKobo
	if totalKobo < 0 {
		totalKobo = 0
	}

	// Convert delivery address ID from string to uint
	var deliveryAddressID *uint
	if req.DeliveryAddressID != nil && *req.DeliveryAddressID != "" {
		if addressID, err := strconv.ParseUint(*req.DeliveryAddressID, 10, 32); err != nil {
			return nil, fmt.Errorf("invalid delivery address ID: %w", err)
		} else {
			uintAddressID := uint(addressID)
			deliveryAddressID = &uintAddressID
		}
	}

	order := &Order{
		CustomerID:        userID,
		DeliveryAddressID: deliveryAddressID,
		Status:            OrderStatusPending,
		PaymentStatus:     PaymentStatusUnpaid,
		ItemsSubtotal:     subtotalKobo,
		DeliveryFee:       deliveryFeeKobo,
		ServiceFee:        serviceFeeKobo,
		CouponDiscount:    discountKobo,
		TotalAmount:       totalKobo,
		CustomRequests:    extractCustomRequestIDs(req.CustomRequests),
		CouponCode:        req.CouponCode,
		Notes:             req.Notes,
		IdempotencyKey:    req.IdempotencyKey,
	}

	if err := s.repo.Create(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Set OrderID for each item and create them
	for i := range orderItems {
		orderItems[i].OrderID = order.ID
	}
	order.Items = orderItems

	// Save the order items
	if len(orderItems) > 0 {
		if err := s.db.WithContext(ctx).Create(&orderItems).Error; err != nil {
			return nil, fmt.Errorf("failed to create order items: %w", err)
		}
	}

	// Update product stock
	for _, item := range req.Items {
		stockReq := products.StockUpdateRequest{
			Quantity:   item.Quantity,
			ChangeType: "REMOVE",
			Reason:     "Order creation",
		}
		// userID is already uuid.UUID, use it directly
		if _, err := s.productRepo.UpdateStock(ctx, item.ProductID, stockReq, userID); err != nil {
			// Log error but don't fail the order creation
			fmt.Printf("Warning: failed to update stock for product %s: %v\n", item.ProductID, err)
		}
	}

	// Apply coupon usage if coupon was used
	if order.CouponCode != nil && *order.CouponCode != "" {
		applyReq := coupons.ApplyCouponRequest{
			Code:        *order.CouponCode,
			UserID:      userID,
			OrderID:     order.ID,
			OrderAmount: float64(order.TotalAmount), // Keep in kobo format
		}
		
		if _, err := s.couponService.ApplyCoupon(applyReq); err != nil {
			// Log error but don't fail the order creation
			fmt.Printf("Warning: failed to apply coupon usage tracking: %v\n", err)
		}
	}

	// Update custom request status to indicate they are now in an order (cart)
	for _, customReq := range req.CustomRequests {
		updateReq := custom_requests.UpdateRequestStatusReq{
			Status: custom_requests.RequestInCart,
			Notes:  "Custom request included in order",
		}
		if _, err := s.customRequestService.UpdateCustomRequestStatus(customReq.CustomRequestID, updateReq); err != nil {
			// Log error but don't fail the order creation
			fmt.Printf("Warning: failed to update custom request %s status to 'in_cart': %v\n", customReq.CustomRequestID, err)
		}
	}

	response := s.toOrderResponseWithContext(ctx, order)
	return response, nil
}

// CreateWithPayment creates an order and initializes payment, returning payment initialization data
func (s *Service) CreateWithPayment(ctx context.Context, userID uuid.UUID, req CreateOrderRequest) (*CreateOrderResponse, error) {
	// First create the order using the existing Create method
	orderResponse, err := s.Create(ctx, userID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Get customer information for payment initialization
	customer, err := s.customerService.GetCustomerByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer information: %w", err)
	}

	// Initialize payment
	paymentReq := payments.CreatePaymentRequest{
		OrderID:       orderResponse.ID.String(),
		PaymentMethod: payments.PaymentMethod(req.PaymentMethod),
		ReturnURL:     "", // Will be set by payment service
		CancelURL:     "", // Will be set by payment service
	}

	paymentInit, err := s.paymentService.InitializePayment(paymentReq, customer.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize payment: %w", err)
	}

	// Generate order number (simple implementation)
	orderNumber := fmt.Sprintf("ORD-%06d", orderResponse.ID.ID()%1000000)

	// Create response with payment information
	response := &CreateOrderResponse{
		OrderID:     orderResponse.ID,
		OrderNumber: orderNumber,
		Payment: PaymentInfo{
			Provider:   string(req.PaymentMethod),
			Reference:  paymentInit.TransactionRef,
			PaymentURL: paymentInit.PaymentURL,
		},
	}

	return response, nil
}

func (s *Service) UpdateStatus(ctx context.Context, id uuid.UUID, userID uuid.UUID, status OrderStatus) error {
	// Get the order first to validate ownership and current status
	order, err := s.repo.Get(ctx, id, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("order not found")
		}
		return fmt.Errorf("failed to get order: %w", err)
	}

	// Validate status transition
	if !s.isValidStatusTransition(order.Status, status) {
		return fmt.Errorf("invalid status transition from %s to %s", order.Status, status)
	}

	// Update the status
	if err := s.repo.UpdateStatus(ctx, id, userID, status, ""); err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	// Send notification about order status change
	s.sendOrderStatusNotification(order.CustomerID, id, status)

	return nil
}

func (s *Service) CancelOrder(ctx context.Context, id uuid.UUID, userID uuid.UUID, reason string) error {
	// Get the order first to validate ownership and current status
	order, err := s.repo.Get(ctx, id, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("order not found")
		}
		return fmt.Errorf("failed to get order: %w", err)
	}

	// Check if order can be cancelled
	if order.Status == OrderStatusDelivered || order.Status == OrderStatusCancelled {
		return fmt.Errorf("order cannot be cancelled in %s status", order.Status)
	}

	// Cancel the order
	if err := s.repo.CancelOrder(ctx, id, userID, reason); err != nil {
		return fmt.Errorf("failed to cancel order: %w", err)
	}

	// Restore product stock
	for _, item := range order.Items {
		stockReq := products.StockUpdateRequest{
			Quantity:   item.Quantity,
			ChangeType: "ADD",
			Reason:     "Order cancellation",
		}
		// userID is already uuid.UUID, use it directly
		if _, err := s.productRepo.UpdateStock(ctx, item.ProductID, stockReq, userID); err != nil {
			fmt.Printf("Warning: failed to restore stock for product %s: %v\n", item.ProductID, err)
		}
	}
	
	// TODO: Implement coupon usage restoration if needed
	// if order.CouponCode != nil {
	//     // Restore coupon usage
	// }

	return nil
}

func (s *Service) isValidStatusTransition(currentStatus, newStatus OrderStatus) bool {
	switch currentStatus {
	case OrderStatusPending:
		return newStatus == OrderStatusConfirmed || newStatus == OrderStatusCancelled
	case OrderStatusConfirmed:
		return newStatus == OrderStatusPreparing || newStatus == OrderStatusCancelled
	case OrderStatusPreparing:
		return newStatus == OrderStatusOutForDelivery || newStatus == OrderStatusCancelled
	case OrderStatusOutForDelivery:
		return newStatus == OrderStatusDelivered
	case OrderStatusDelivered, OrderStatusCancelled:
		return false // Terminal states
	default:
		return false
	}
}

// Admin methods
func (s *Service) AdminList(ctx context.Context, query AdminListQuery) (*ListResult, error) {
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.Limit <= 0 || query.Limit > 100 {
		query.Limit = 20
	}

	orders, total, err := s.repo.AdminList(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list orders: %w", err)
	}

	responses := make([]OrderResponse, len(orders))
	for i, order := range orders {
		responses[i] = *s.toOrderResponseWithContext(ctx, &order)
	}

	totalPages := int((total + int64(query.Limit) - 1) / int64(query.Limit))

	return &ListResult{
		Data: responses,
		Meta: PageMeta{
			Page:       query.Page,
			Limit:      query.Limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

func (s *Service) AdminGet(ctx context.Context, id uuid.UUID) (*OrderResponse, error) {
	order, err := s.repo.AdminGet(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("order not found")
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	response := s.toOrderResponseWithContext(ctx, order)
	return response, nil
}

func (s *Service) AdminUpdateStatus(ctx context.Context, id uuid.UUID, status OrderStatus) error {
	// Get the order first to get customer ID for notification
	order, err := s.repo.AdminGet(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	// Update the status
	if err := s.repo.AdminUpdateStatus(ctx, id, status); err != nil {
		return err
	}

	// Send notification about order status change
	s.sendOrderStatusNotification(order.CustomerID, id, status)

	return nil
}

func (s *Service) AdminUpdatePaymentStatus(ctx context.Context, id uuid.UUID, paymentStatus interface{}) error {
	// Convert payment status to our internal PaymentStatus type
	var internalStatus PaymentStatus
	
	// Handle different types that might be passed
	switch v := paymentStatus.(type) {
	case string:
		if v == "paid" {
			internalStatus = PaymentStatusPaid
		} else {
			internalStatus = PaymentStatusPaid // default
		}
	case PaymentStatus:
		internalStatus = v
	default:
		internalStatus = PaymentStatusPaid // default fallback
	}
	return s.repo.AdminUpdatePaymentStatus(ctx, id, internalStatus)
}

func (s *Service) AdminCancelOrder(ctx context.Context, id uuid.UUID, reason string) error {
	// Get the order first to validate current status
	order, err := s.repo.AdminGet(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("order not found")
		}
		return fmt.Errorf("failed to get order: %w", err)
	}

	// Check if order can be cancelled
	if order.Status == OrderStatusDelivered || order.Status == OrderStatusCancelled {
		return fmt.Errorf("order cannot be cancelled in %s status", order.Status)
	}

	// Cancel the order using admin method
	if err := s.repo.AdminCancelOrder(ctx, id, reason); err != nil {
		return fmt.Errorf("failed to cancel order: %w", err)
	}

	// Restore product stock
	for _, item := range order.Items {
		stockReq := products.StockUpdateRequest{
			Quantity:   item.Quantity,
			ChangeType: "ADD",
			Reason:     "Order cancellation by admin",
		}
		// Use a system UUID for admin operations
		systemUserID := uuid.MustParse("00000000-0000-0000-0000-000000000000")
		if _, err := s.productRepo.UpdateStock(ctx, item.ProductID, stockReq, systemUserID); err != nil {
			fmt.Printf("Warning: failed to restore stock for product %s: %v\n", item.ProductID, err)
		}
	}

	return nil
}

func (s *Service) GetStats(ctx context.Context) (*OrderStats, error) {
	// Create empty query for all stats
	query := OrderStatsQuery{}
	return s.repo.GetStats(ctx, query)
}

// Helper methods
func (s *Service) toCartResponse(cart Cart) *CartResponse {
	items := make([]CartItemResponse, len(cart.Items))
	for i, item := range cart.Items {
		itemResponse := CartItemResponse{
			ID:         item.ID,
			ProductID:  item.ProductID,
			Quantity:   item.Quantity,
			// Price will be set from product info below
			PriceKobo:  0,
			PriceNaira: 0,
			SubtotalKobo: 0,
			SubtotalNaira: 0,
			CreatedAt:  item.CreatedAt,
			UpdatedAt:  item.UpdatedAt,
		}

		// Add product info if loaded and set prices
		if item.Product.ID != uuid.Nil {
			priceKobo := int64(item.Product.SellingPrice) // Keep as is, no conversion
			itemResponse.PriceKobo = priceKobo
			itemResponse.PriceNaira = item.Product.SellingPrice
			itemResponse.SubtotalKobo = priceKobo * int64(item.Quantity)
			itemResponse.SubtotalNaira = item.Product.SellingPrice * float64(item.Quantity)
			
			itemResponse.Product = &ProductInfo{
				ID:       item.Product.ID,
				Name:     item.Product.Name,
				Slug:     item.Product.Slug,
				ImageURL: item.Product.ImageURL,
			}
		}

		items[i] = itemResponse
	}

	// Calculate total from items
	totalKobo := int64(0)
	for _, item := range items {
		totalKobo += item.SubtotalKobo
	}
	
	return &CartResponse{
		ID:            cart.ID,
		UserID:        cart.UserID,
		Items:         items,
		TotalItems:    len(items),
		TotalKobo:     totalKobo,
		TotalNaira:    float64(totalKobo),
		CreatedAt:     cart.CreatedAt,
		UpdatedAt:     cart.UpdatedAt,
	}
}

func (s *Service) toOrderResponse(order *Order) *OrderResponse {
	return s.toOrderResponseWithContext(context.Background(), order)
}

func (s *Service) toOrderResponseWithContext(ctx context.Context, order *Order) *OrderResponse {
	items := make([]OrderItemResponse, len(order.Items))
	for i, item := range order.Items {
		itemResponse := OrderItemResponse{
			ID:              item.ID,
			ProductID:       item.ProductID,
			Name:            item.Name,
			SKU:             item.SKU,
			Quantity:        item.Quantity,
			UnitPrice:       item.UnitPrice,
			UnitPriceNaira:  float64(item.UnitPrice) / 100.0,
			TotalPrice:      item.TotalPrice,
			TotalPriceNaira: float64(item.TotalPrice) / 100.0,
			CreatedAt:       item.CreatedAt,
			UpdatedAt:       item.UpdatedAt,
		}

		// Add product info if loaded
		if item.Product.ID != uuid.Nil {
			// Convert product price from float64 to kobo (int64)
			priceKobo := int64(item.Product.SellingPrice * 100)
			itemResponse.Product = &ProductInfo{
				ID:         item.Product.ID,
				Name:       item.Product.Name,
				Slug:       item.Product.Slug,
				ImageURL:   item.Product.ImageURL,
				Price:      priceKobo,
				PriceNaira: item.Product.SellingPrice,
			}
		}

		items[i] = itemResponse
	}

	// Convert CustomerID from uuid.UUID to uint
	var customerID uint
	if customer, err := s.customerService.GetCustomerByUserID(order.CustomerID); err == nil {
		customerID = customer.ID
	} else {
		// Fallback: use 0 if customer not found
		customerID = 0
	}

	// Use delivery address ID directly from order
	var deliveryAddressID *uint = order.DeliveryAddressID

	response := OrderResponse{
		ID:                    order.ID,
		CustomerID:            customerID,
		DeliveryAddressID:     deliveryAddressID,
		Status:                order.Status,
		PaymentStatus:         order.PaymentStatus,
		IdempotencyKey:        order.IdempotencyKey,
		CouponCode:            order.CouponCode,
		CouponDiscount:        order.CouponDiscount,
		CouponDiscountNaira:   float64(order.CouponDiscount) / 100.0,
		ItemsSubtotal:         order.ItemsSubtotal,
		ItemsSubtotalNaira:    float64(order.ItemsSubtotal) / 100.0,
		DeliveryFee:           order.DeliveryFee,
		DeliveryFeeNaira:      float64(order.DeliveryFee) / 100.0,
		ServiceFee:            order.ServiceFee,
		ServiceFeeNaira:       float64(order.ServiceFee) / 100.0,
		TotalAmount:           order.TotalAmount,
		TotalAmountNaira:      float64(order.TotalAmount) / 100.0,
		CustomRequests:        func() []uuid.UUID {
			if order.CustomRequests == nil {
				return []uuid.UUID{}
			}
			return []uuid.UUID(order.CustomRequests)
		}(),
		CustomRequestDetails:  []CustomRequestInfo{},
		Notes:                 order.Notes,
		EstimatedDelivery:     order.EstimatedDelivery,
		DeliveredAt:           order.DeliveredAt,
		CancelledAt:           order.CancelledAt,
		CancellationReason:    order.CancellationReason,
		Items:                 items,
		CreatedAt:             order.CreatedAt,
		UpdatedAt:             order.UpdatedAt,
	}

	// Fetch real customer data using customer service
	// Get customer information with proper email from auth service
	if customer, err := s.customerService.GetCustomerByUserID(order.CustomerID); err == nil {
		// Get user email from auth service
		userEmail := "N/A"
		if user, userErr := s.authService.GetUserByID(ctx, order.CustomerID); userErr == nil {
			userEmail = user.Email
		}

		response.Customer = &CustomerInfo{
			ID:        customer.ID,
			FirstName: customer.FirstName,
			LastName:  customer.LastName,
			Phone:     customer.Phone,
			Email:     userEmail,
		}

		// Get delivery address if specified
		if deliveryAddressID != nil {
			// Find the address in customer's addresses
			for _, addr := range customer.Addresses {
				if addr.ID == *deliveryAddressID {
					response.DeliveryAddress = &AddressInfo{
						ID:         addr.ID,
						Label:      addr.Label,
						Street:     addr.Street,
						City:       addr.City,
						State:      addr.State,
						Country:    addr.Country,
						PostalCode: addr.PostalCode,
					}
					break
				}
			}
		}
	} else {
		// Fallback to placeholder if customer fetch fails
		// Still try to get email from auth service
		userEmail := "N/A"
		if user, userErr := s.authService.GetUserByID(ctx, order.CustomerID); userErr == nil {
			userEmail = user.Email
		}

		response.Customer = &CustomerInfo{
			ID:        customerID,
			FirstName: "Unknown",
			LastName:  "User",
			Phone:     "N/A",
			Email:     userEmail,
		}
	}

	// Fetch custom request details if any
	if len(order.CustomRequests) > 0 {
		customRequestDetails := make([]CustomRequestInfo, 0, len(order.CustomRequests))
		for _, requestID := range order.CustomRequests {
			if customRequest, err := s.customRequestService.GetCustomRequestAdmin(requestID); err == nil {
				// Convert custom request items
				items := make([]CustomRequestItemInfo, len(customRequest.Items))
				for i, item := range customRequest.Items {
					var quotedPriceNaira *float64
					if item.QuotedPrice != nil {
						nairaValue := float64(*item.QuotedPrice)
						quotedPriceNaira = &nairaValue
					}

					items[i] = CustomRequestItemInfo{
						ID:             item.ID,
						Name:           item.Name,
						Description:    item.Description,
						Quantity:       item.Quantity,
						Unit:           item.Unit,
						PreferredBrand: item.PreferredBrand,
						QuotedPrice:    item.QuotedPrice,
						QuotedPriceNaira: quotedPriceNaira,
						AdminNotes:     item.AdminNotes,
						Images:         item.Images,
					}
				}

				// Convert active quote if exists
				var activeQuote *CustomRequestQuoteInfo
				if customRequest.ActiveQuote != nil {
					activeQuote = &CustomRequestQuoteInfo{
						ID:            customRequest.ActiveQuote.ID,
						ItemsSubtotal: customRequest.ActiveQuote.ItemsSubtotal,
						ItemsSubtotalNaira: float64(customRequest.ActiveQuote.ItemsSubtotal),
						GrandTotal:    customRequest.ActiveQuote.GrandTotal,
						GrandTotalNaira: float64(customRequest.ActiveQuote.GrandTotal),
						Status:        string(customRequest.ActiveQuote.Status),
						ValidUntil:    customRequest.ActiveQuote.ValidUntil,
						AcceptedAt:    customRequest.ActiveQuote.AcceptedAt,
					}
				}

				customRequestInfo := CustomRequestInfo{
					ID:                 customRequest.ID,
					Status:             string(customRequest.Status),
					Priority:           string(customRequest.Priority),
					AllowSubstitutions: customRequest.AllowSubstitutions,
					Notes:              customRequest.Notes,
					SubmittedAt:        customRequest.SubmittedAt,
					Items:              items,
					ActiveQuote:        activeQuote,
				}
				customRequestDetails = append(customRequestDetails, customRequestInfo)
			}
		}
		response.CustomRequestDetails = customRequestDetails
	}

	return &response
}

// sendOrderStatusNotification sends a notification when order status changes
func (s *Service) sendOrderStatusNotification(customerID uuid.UUID, orderID uuid.UUID, status OrderStatus) {
	if s.notificationService == nil {
		return
	}

	// Get customer by user ID to get the uint customer ID
	customer, err := s.customerService.GetCustomerByUserID(customerID)
	if err != nil {
		// Log error but don't fail the order update
		fmt.Printf("Failed to get customer for notification: %v\n", err)
		return
	}

	// Create status-specific notification content
	title, body := s.getNotificationContent(status, orderID.String())

	// Create notification request
	notificationReq := &notifications.CreateNotificationRequest{
		RecipientID:   customer.UserID,
		RecipientType: notifications.RecipientCustomer,
		Type:          notifications.TypeOrderUpdate,
		Title:         title,
		Body:          body,
		Data: map[string]interface{}{
			"orderId": orderID.String(),
			"status":  string(status),
		},
	}

	// Send notification asynchronously to avoid blocking the main operation
	go func() {
		if _, err := s.notificationService.CreateNotification(notificationReq); err != nil {
			// Log error but don't fail the order update
			fmt.Printf("Failed to send order status notification: %v\n", err)
		}
	}()
}

// getNotificationContent returns appropriate title and body for order status
func (s *Service) getNotificationContent(status OrderStatus, orderID string) (string, string) {
	switch status {
	case OrderStatusConfirmed:
		return "Order Confirmed", fmt.Sprintf("Your order %s has been confirmed and is being prepared.", orderID)
	case OrderStatusPreparing:
		return "Order Being Prepared", fmt.Sprintf("Your order %s is now being prepared.", orderID)
	case OrderStatusOutForDelivery:
		return "Out for Delivery", fmt.Sprintf("Your order %s is out for delivery and will arrive soon.", orderID)
	case OrderStatusDelivered:
		return "Order Delivered", fmt.Sprintf("Your order %s has been successfully delivered. Thank you for your business!", orderID)
	case OrderStatusCancelled:
		return "Order Cancelled", fmt.Sprintf("Your order %s has been cancelled. If you have any questions, please contact support.", orderID)
	default:
		return "Order Update", fmt.Sprintf("Your order %s status has been updated to %s.", orderID, string(status))
	}
}

// Helper function to extract UUID slice from CreateOrderCustomRequest slice
func extractCustomRequestIDs(customRequests []CreateOrderCustomRequest) UUIDSlice {
	if len(customRequests) == 0 {
		return UUIDSlice{}
	}
	
	ids := make([]uuid.UUID, len(customRequests))
	for i, req := range customRequests {
		ids[i] = req.CustomRequestID
	}
	return UUIDSlice(ids)
}
