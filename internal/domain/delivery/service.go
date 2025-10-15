package delivery

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"time"
	"errandShop/internal/domain/notifications"
	"errandShop/internal/domain/orders"
	"errandShop/internal/domain/customers"
	"github.com/google/uuid"
)

// DeliveryService interface defines delivery service methods
type DeliveryService interface {
	// Delivery methods
	CreateDelivery(req *CreateDeliveryRequest) (*DeliveryResponse, error)
	GetDelivery(id uint) (*DeliveryResponse, error)
	GetDeliveryByTrackingNumber(trackingNumber string) (*DeliveryResponse, error)
	GetDeliveryByOrderID(orderID string) (*DeliveryResponse, error)
	UpdateDeliveryStatus(id uint, req *UpdateDeliveryStatusRequest) (*DeliveryResponse, error)
	AssignDriver(id uint, driverID uint) (*DeliveryResponse, error)
	ListDeliveries(limit, offset int, status *DeliveryStatus) ([]DeliveryResponse, int64, error)
	GetDeliveriesByDriver(driverID uint, limit, offset int) ([]DeliveryResponse, int64, error)
	CancelDelivery(id uint, reason string) (*DeliveryResponse, error)

	// Driver methods
	CreateDriver(req *CreateDriverRequest) (*DeliveryDriverResponse, error)
	GetDriver(id uint) (*DeliveryDriverResponse, error)
	GetDriverByUserID(userID uint) (*DeliveryDriverResponse, error)
	UpdateDriverLocation(driverID uint, req *UpdateDriverLocationRequest) error
	ListDrivers(limit, offset int, isActive *bool) ([]DeliveryDriverResponse, int64, error)
	GetAvailableDrivers(vehicleType *VehicleType, lat, lng *float64, radius float64) ([]DeliveryDriverResponse, error)
	ToggleDriverAvailability(driverID uint, available bool) error

	// Tracking methods
	GetTrackingUpdates(deliveryID uint) ([]TrackingUpdateResponse, error)
	AddTrackingUpdate(deliveryID uint, status DeliveryStatus, message string, lat, lng *float64) error

	// Quote and pricing
	GetDeliveryQuote(req *DeliveryQuoteRequest) (*DeliveryQuoteResponse, error)
	CalculateDeliveryFee(distance float64, deliveryType string) int64

	// Stats methods
	GetDeliveryStats(startDate, endDate *time.Time) (*DeliveryStatsResponse, error)
	GetDriverStats(driverID uint, startDate, endDate *time.Time) (map[string]interface{}, error)
	GetDeliveryByID(id uint) (*DeliveryResponse, error)
}

// deliveryService implements DeliveryService
type deliveryService struct {
	repo                DeliveryRepository
	notificationService notifications.NotificationService
	ordersRepo          *orders.Repository
	customersService    customers.Service
}

// NewDeliveryService creates a new delivery service
func NewDeliveryService(repo DeliveryRepository, notificationService notifications.NotificationService, ordersRepo *orders.Repository, customersService customers.Service) DeliveryService {
	return &deliveryService{
		repo: repo,
		notificationService: notificationService,
		ordersRepo: ordersRepo,
		customersService: customersService,
	}
}

// generateTrackingNumber generates a unique tracking number
func (s *deliveryService) generateTrackingNumber() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("TRK%d%04d", time.Now().Unix(), rand.Intn(10000))
}

// Delivery methods implementation
func (s *deliveryService) CreateDelivery(req *CreateDeliveryRequest) (*DeliveryResponse, error) {
	// Calculate delivery fee and estimated time
	distance := s.calculateDistance(req.PickupLatitude, req.PickupLongitude, req.DeliveryLatitude, req.DeliveryLongitude)
	deliveryFee := s.CalculateDeliveryFee(distance, string(req.DeliveryType))
	estimatedTime := s.calculateEstimatedTime(distance, req.DeliveryType)

	delivery := &Delivery{
		OrderID:           req.OrderID,
		TrackingNumber:    s.generateTrackingNumber(),
		DeliveryType:      req.DeliveryType,
		Status:            DeliveryStatusPending,
		PickupAddress:     req.PickupAddress,
		PickupLatitude:    req.PickupLatitude,
		PickupLongitude:   req.PickupLongitude,
		PickupNotes:       req.PickupNotes,
		DeliveryAddress:   req.DeliveryAddress,
		DeliveryLatitude:  req.DeliveryLatitude,
		DeliveryLongitude: req.DeliveryLongitude,
		DeliveryNotes:     req.DeliveryNotes,
		RecipientName:     req.RecipientName,
		RecipientPhone:    req.RecipientPhone,
		ScheduledDate:     req.ScheduledDate,
		EstimatedTime:     &estimatedTime,
		DeliveryFee:       deliveryFee,
		Distance:          &distance,
	}

	err := s.repo.CreateDelivery(delivery)
	if err != nil {
		return nil, err
	}

	// Add initial tracking update
	s.AddTrackingUpdate(delivery.ID, DeliveryStatusPending, "Delivery order created", nil, nil)

	return s.mapDeliveryToResponse(delivery), nil
}

func (s *deliveryService) GetDelivery(id uint) (*DeliveryResponse, error) {
	delivery, err := s.repo.GetDeliveryByID(id)
	if err != nil {
		return nil, err
	}
	return s.mapDeliveryToResponse(delivery), nil
}

func (s *deliveryService) GetDeliveryByTrackingNumber(trackingNumber string) (*DeliveryResponse, error) {
	delivery, err := s.repo.GetDeliveryByTrackingNumber(trackingNumber)
	if err != nil {
		return nil, err
	}
	return s.mapDeliveryToResponse(delivery), nil
}

func (s *deliveryService) GetDeliveryByOrderID(orderID string) (*DeliveryResponse, error) {
	delivery, err := s.repo.GetDeliveryByOrderID(orderID)
	if err != nil {
		return nil, err
	}
	return s.mapDeliveryToResponse(delivery), nil
}

func (s *deliveryService) UpdateDeliveryStatus(id uint, req *UpdateDeliveryStatusRequest) (*DeliveryResponse, error) {
	delivery, err := s.repo.GetDeliveryByID(id)
	if err != nil {
		return nil, err
	}

	// Update status and timing
	delivery.Status = req.Status
	now := time.Now()

	switch req.Status {
	case DeliveryStatusPickedUp:
		delivery.PickupTime = &now
	case DeliveryStatusDelivered:
		delivery.DeliveryTime = &now
		delivery.ActualTime = &now
	}

	err = s.repo.UpdateDelivery(delivery)
	if err != nil {
		return nil, err
	}

	// Add tracking update
	s.AddTrackingUpdate(id, req.Status, req.Message, req.Latitude, req.Longitude)

	// Send notification to customer about delivery status update
	if s.notificationService != nil {
		// Get customer ID from order - we need to add this method to get order details
		if customerID, err := s.getCustomerIDFromDelivery(delivery); err == nil {
			// Send delivery update notification
			if notifErr := s.notificationService.SendDeliveryUpdate(id, string(req.Status), customerID); notifErr != nil {
				// Log error but don't fail the delivery update
				fmt.Printf("Failed to send delivery notification: %v\n", notifErr)
			}
		}
	}

	return s.mapDeliveryToResponse(delivery), nil
}

func (s *deliveryService) AssignDriver(id uint, driverID uint) (*DeliveryResponse, error) {
	delivery, err := s.repo.GetDeliveryByID(id)
	if err != nil {
		return nil, err
	}

	// Verify driver exists and is available
	driver, err := s.repo.GetDriverByID(driverID)
	if err != nil {
		return nil, err
	}

	if !driver.IsActive || !driver.IsAvailable {
		return nil, errors.New("driver is not available")
	}

	delivery.DriverID = &driverID
	delivery.Status = DeliveryStatusAssigned

	err = s.repo.UpdateDelivery(delivery)
	if err != nil {
		return nil, err
	}

	// Update driver availability
	driver.IsAvailable = false
	s.repo.UpdateDriver(driver)

	// Add tracking update
	s.AddTrackingUpdate(id, DeliveryStatusAssigned, fmt.Sprintf("Driver %s assigned", driver.VehicleNumber), nil, nil)

	return s.mapDeliveryToResponse(delivery), nil
}

func (s *deliveryService) ListDeliveries(limit, offset int, status *DeliveryStatus) ([]DeliveryResponse, int64, error) {
	deliveries, total, err := s.repo.ListDeliveries(limit, offset, status)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]DeliveryResponse, len(deliveries))
	for i, delivery := range deliveries {
		responses[i] = *s.mapDeliveryToResponse(&delivery)
	}

	return responses, total, nil
}

func (s *deliveryService) GetDeliveriesByDriver(driverID uint, limit, offset int) ([]DeliveryResponse, int64, error) {
	deliveries, total, err := s.repo.GetDeliveriesByDriver(driverID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]DeliveryResponse, len(deliveries))
	for i, delivery := range deliveries {
		responses[i] = *s.mapDeliveryToResponse(&delivery)
	}

	return responses, total, nil
}

func (s *deliveryService) CancelDelivery(id uint, reason string) (*DeliveryResponse, error) {
	delivery, err := s.repo.GetDeliveryByID(id)
	if err != nil {
		return nil, err
	}

	if delivery.Status == DeliveryStatusDelivered {
		return nil, errors.New("cannot cancel delivered order")
	}

	delivery.Status = DeliveryStatusCancelled
	err = s.repo.UpdateDelivery(delivery)
	if err != nil {
		return nil, err
	}

	// Free up driver if assigned
	if delivery.DriverID != nil {
		driver, _ := s.repo.GetDriverByID(*delivery.DriverID)
		if driver != nil {
			driver.IsAvailable = true
			s.repo.UpdateDriver(driver)
		}
	}

	// Add tracking update
	s.AddTrackingUpdate(id, DeliveryStatusCancelled, fmt.Sprintf("Delivery cancelled: %s", reason), nil, nil)

	return s.mapDeliveryToResponse(delivery), nil
}

// Driver methods implementation
func (s *deliveryService) CreateDriver(req *CreateDriverRequest) (*DeliveryDriverResponse, error) {
	driver := &DeliveryDriver{
		UserID:           req.UserID,
		LicenseNumber:    req.LicenseNumber,
		VehicleType:      req.VehicleType,
		VehicleNumber:    req.VehicleNumber,
		VehiclePlate:     req.VehiclePlate,
		VehicleModel:     req.VehicleModel,
		VehicleColor:     req.VehicleColor,
		EmergencyContact: req.EmergencyContact,
		IsActive:         req.IsActive,
		IsAvailable:      true,
	}

	err := s.repo.CreateDriver(driver)
	if err != nil {
		return nil, err
	}

	return s.mapDriverToResponse(driver), nil
}

func (s *deliveryService) GetDriver(id uint) (*DeliveryDriverResponse, error) {
	driver, err := s.repo.GetDriverByID(id)
	if err != nil {
		return nil, err
	}
	return s.mapDriverToResponse(driver), nil
}

func (s *deliveryService) GetDriverByUserID(userID uint) (*DeliveryDriverResponse, error) {
	driver, err := s.repo.GetDriverByUserID(userID)
	if err != nil {
		return nil, err
	}
	return s.mapDriverToResponse(driver), nil
}

func (s *deliveryService) UpdateDriverLocation(driverID uint, req *UpdateDriverLocationRequest) error {
	return s.repo.UpdateDriverLocation(driverID, req.Latitude, req.Longitude)
}

func (s *deliveryService) ListDrivers(limit, offset int, isActive *bool) ([]DeliveryDriverResponse, int64, error) {
	drivers, total, err := s.repo.ListDrivers(limit, offset, isActive)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]DeliveryDriverResponse, len(drivers))
	for i, driver := range drivers {
		responses[i] = *s.mapDriverToResponse(&driver)
	}

	return responses, total, nil
}

func (s *deliveryService) GetAvailableDrivers(vehicleType *VehicleType, lat, lng *float64, radius float64) ([]DeliveryDriverResponse, error) {
	drivers, err := s.repo.GetAvailableDrivers(vehicleType, lat, lng, radius)
	if err != nil {
		return nil, err
	}

	responses := make([]DeliveryDriverResponse, len(drivers))
	for i, driver := range drivers {
		responses[i] = *s.mapDriverToResponse(&driver)
	}

	return responses, nil
}

func (s *deliveryService) ToggleDriverAvailability(driverID uint, available bool) error {
	driver, err := s.repo.GetDriverByID(driverID)
	if err != nil {
		return err
	}

	driver.IsAvailable = available
	return s.repo.UpdateDriver(driver)
}

// Tracking methods implementation
func (s *deliveryService) GetTrackingUpdates(deliveryID uint) ([]TrackingUpdateResponse, error) {
	updates, err := s.repo.GetTrackingUpdatesByDeliveryID(deliveryID)
	if err != nil {
		return nil, err
	}

	responses := make([]TrackingUpdateResponse, len(updates))
	for i, update := range updates {
		responses[i] = TrackingUpdateResponse{
			ID:          update.ID,
			Status:      update.Status,
			Message:     update.Message,
			IsAutomatic: update.IsAutomatic,
			Timestamp:   update.Timestamp,
		}
	}

	return responses, nil
}

func (s *deliveryService) AddTrackingUpdate(deliveryID uint, status DeliveryStatus, message string, lat, lng *float64) error {
	update := &TrackingUpdate{
		DeliveryID:  deliveryID,
		Status:      status,
		Message:     message,
		IsAutomatic: false,
		Timestamp:   time.Now(),
	}

	return s.repo.CreateTrackingUpdate(update)
}

// Quote and pricing implementation
func (s *deliveryService) GetDeliveryQuote(req *DeliveryQuoteRequest) (*DeliveryQuoteResponse, error) {
	distance := s.calculateDistance(&req.PickupLatitude, &req.PickupLongitude, &req.DeliveryLatitude, &req.DeliveryLongitude)
	estimatedTime := s.calculateEstimatedTime(distance, req.DeliveryType)
	deliveryFee := s.CalculateDeliveryFee(distance, string(req.DeliveryType))

	now := time.Now()
	estimatedPickup := now.Add(30 * time.Minute) // 30 minutes from now
	estimatedDelivery := estimatedPickup.Add(time.Duration(estimatedTime.Sub(time.Now()).Minutes()) * time.Minute)

	if req.ScheduledDate != nil {
		estimatedPickup = *req.ScheduledDate
		estimatedDelivery = req.ScheduledDate.Add(time.Duration(estimatedTime.Sub(time.Now()).Minutes()) * time.Minute)
	}

	return &DeliveryQuoteResponse{
		Distance:          distance,
		EstimatedTime:     int(estimatedTime.Sub(time.Now()).Minutes()),
		DeliveryFee:       deliveryFee,
		EstimatedPickup:   &estimatedPickup,
		EstimatedDelivery: &estimatedDelivery,
	}, nil
}

func (s *deliveryService) CalculateDeliveryFee(distance float64, deliveryType string) int64 {
	baseFee := int64(50000)  // ₦500 base fee in kobo
	perKmFee := int64(10000) // ₦100 per km in kobo

	// Apply multipliers based on delivery type
	multiplier := 1.0
	switch deliveryType {
	case "express":
		multiplier = 1.5
	case "same_day":
		multiplier = 2.0
	case "scheduled":
		multiplier = 0.8
	case "standard":
		multiplier = 1.0
	}

	totalFee := float64(baseFee) + (distance * float64(perKmFee))
	return int64(totalFee * multiplier)
}

// Analytics implementation
func (s *deliveryService) GetDeliveryStats(startDate, endDate *time.Time) (*DeliveryStatsResponse, error) {
	return s.repo.GetDeliveryStats(startDate, endDate)
}

func (s *deliveryService) GetDriverStats(driverID uint, startDate, endDate *time.Time) (map[string]interface{}, error) {
	return s.repo.GetDriverStats(driverID, startDate, endDate)
}

// Helper methods
func (s *deliveryService) calculateDistance(lat1, lng1, lat2, lng2 *float64) float64 {
	if lat1 == nil || lng1 == nil || lat2 == nil || lng2 == nil {
		return 5.0 // Default 5km if coordinates not provided
	}

	// Haversine formula for calculating distance between two points
	const earthRadius = 6371 // Earth's radius in kilometers

	dLat := (*lat2 - *lat1) * math.Pi / 180
	dLng := (*lng2 - *lng1) * math.Pi / 180

	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(*lat1*math.Pi/180)*math.Cos(*lat2*math.Pi/180)*math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}

func (s *deliveryService) calculateEstimatedTime(distance float64, deliveryType DeliveryType) time.Time {
	// Base speed in km/h
	speed := 30.0

	switch deliveryType {
	case DeliveryTypeExpress:
		speed = 45.0
	case DeliveryTypeSameDay:
		speed = 50.0
	}

	timeInHours := distance / speed
	timeInMinutes := timeInHours * 60

	return time.Now().Add(time.Duration(timeInMinutes) * time.Minute)
}

func (s *deliveryService) mapDeliveryToResponse(delivery *Delivery) *DeliveryResponse {
	response := &DeliveryResponse{
		ID:                delivery.ID,
		OrderID:           delivery.OrderID,
		TrackingNumber:    delivery.TrackingNumber,
		DeliveryType:      delivery.DeliveryType,
		Status:            delivery.Status,
		PickupAddress:     delivery.PickupAddress,
		PickupLatitude:    delivery.PickupLatitude,
		PickupLongitude:   delivery.PickupLongitude,
		PickupTime:        delivery.PickupTime,
		PickupNotes:       delivery.PickupNotes,
		DeliveryAddress:   delivery.DeliveryAddress,
		DeliveryLatitude:  delivery.DeliveryLatitude,
		DeliveryLongitude: delivery.DeliveryLongitude,
		DeliveryTime:      delivery.DeliveryTime,
		DeliveryNotes:     delivery.DeliveryNotes,
		RecipientName:     delivery.RecipientName,
		RecipientPhone:    delivery.RecipientPhone,
		ScheduledDate:     delivery.ScheduledDate,
		EstimatedTime:     delivery.EstimatedTime,
		ActualTime:        delivery.ActualTime,
		DeliveryFee:       delivery.DeliveryFee,
		Distance:          delivery.Distance,
		Duration:          delivery.Duration,
		CreatedAt:         delivery.CreatedAt,
		UpdatedAt:         delivery.UpdatedAt,
	}

	if len(delivery.TrackingUpdates) > 0 {
		response.TrackingUpdates = make([]TrackingUpdateResponse, len(delivery.TrackingUpdates))
		for i, update := range delivery.TrackingUpdates {
			response.TrackingUpdates[i] = TrackingUpdateResponse{
				ID:          update.ID,
				Status:      update.Status,
				Message:     update.Message,
				IsAutomatic: update.IsAutomatic,
				Timestamp:   update.Timestamp,
			}
		}
	}

	return response
}

func (s *deliveryService) GetDeliveryByID(id uint) (*DeliveryResponse, error) {
	delivery, err := s.repo.GetDeliveryByID(id)
	if err != nil {
		return nil, err
	}
	return s.mapDeliveryToResponse(delivery), nil
}

func (s *deliveryService) mapDriverToResponse(driver *DeliveryDriver) *DeliveryDriverResponse {
	return &DeliveryDriverResponse{
		ID:                 driver.ID,
		UserID:             driver.UserID,
		LicenseNumber:      driver.LicenseNumber,
		VehicleType:        driver.VehicleType,
		VehicleNumber:      driver.VehicleNumber,
		VehiclePlate:       driver.VehiclePlate,
		VehicleModel:       driver.VehicleModel,
		VehicleColor:       driver.VehicleColor,
		EmergencyContact:   driver.EmergencyContact,
		IsActive:           driver.IsActive,
		IsAvailable:        driver.IsAvailable,
		CurrentLatitude:    driver.CurrentLatitude,
		CurrentLongitude:   driver.CurrentLongitude,
		LastLocationUpdate: driver.LastLocationUpdate,
		Rating:             driver.Rating,
		TotalDeliveries:    driver.TotalDeliveries,
		CreatedAt:          driver.CreatedAt,
		UpdatedAt:          driver.UpdatedAt,
	}
}

// getCustomerIDFromDelivery gets the customer ID from the delivery's associated order
func (s *deliveryService) getCustomerIDFromDelivery(delivery *Delivery) (uuid.UUID, error) {
	// Parse the order UUID from the delivery
	orderUUID, err := uuid.Parse(delivery.OrderID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid order UUID: %w", err)
	}

	// Get the order from the orders repository
	order, err := s.ordersRepo.AdminGet(context.Background(), orderUUID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get order: %w", err)
	}

	// Get the customer using the customer service
	customer, err := s.customersService.GetCustomerByUserID(order.CustomerID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get customer: %w", err)
	}

	return customer.UserID, nil
}
