package delivery

import (
	"time"

	"gorm.io/gorm"
)

// DeliveryRepository interface defines delivery repository methods
type DeliveryRepository interface {
	// Delivery methods
	CreateDelivery(delivery *Delivery) error
	GetDeliveryByID(id uint) (*Delivery, error)
	GetDeliveryByTrackingNumber(trackingNumber string) (*Delivery, error)
	GetDeliveryByOrderID(orderID string) (*Delivery, error)
	UpdateDelivery(delivery *Delivery) error
	DeleteDelivery(id uint) error
	ListDeliveries(limit, offset int, status *DeliveryStatus) ([]Delivery, int64, error)
	GetDeliveriesByDriver(driverID uint, limit, offset int) ([]Delivery, int64, error)
	GetDeliveriesByDateRange(startDate, endDate time.Time, limit, offset int) ([]Delivery, int64, error)

	// Driver methods
	CreateDriver(driver *DeliveryDriver) error
	GetDriverByID(id uint) (*DeliveryDriver, error)
	GetDriverByUserID(userID uint) (*DeliveryDriver, error)
	UpdateDriver(driver *DeliveryDriver) error
	DeleteDriver(id uint) error
	ListDrivers(limit, offset int, isActive *bool) ([]DeliveryDriver, int64, error)
	GetAvailableDrivers(vehicleType *VehicleType, lat, lng *float64, radius float64) ([]DeliveryDriver, error)
	UpdateDriverLocation(driverID uint, lat, lng float64) error

	// Tracking methods
	CreateTrackingUpdate(update *TrackingUpdate) error
	GetTrackingUpdatesByDeliveryID(deliveryID uint) ([]TrackingUpdate, error)

	// Zone methods
	CreateDeliveryZone(zone *DeliveryZone) error
	GetDeliveryZoneByID(id uint) (*DeliveryZone, error)
	UpdateDeliveryZone(zone *DeliveryZone) error
	DeleteDeliveryZone(id uint) error
	ListDeliveryZones(limit, offset int, isActive *bool) ([]DeliveryZone, int64, error)
	GetZoneByCoordinates(lat, lng float64) (*DeliveryZone, error)

	// Analytics methods
	GetDeliveryStats(startDate, endDate *time.Time) (*DeliveryStatsResponse, error)
	GetDriverStats(driverID uint, startDate, endDate *time.Time) (map[string]interface{}, error)
}

// deliveryRepository implements DeliveryRepository
type deliveryRepository struct {
	db *gorm.DB
}

// NewDeliveryRepository creates a new delivery repository
func NewDeliveryRepository(db *gorm.DB) DeliveryRepository {
	return &deliveryRepository{db: db}
}

// Delivery methods implementation
func (r *deliveryRepository) CreateDelivery(delivery *Delivery) error {
	return r.db.Create(delivery).Error
}

func (r *deliveryRepository) GetDeliveryByID(id uint) (*Delivery, error) {
	var delivery Delivery
	err := r.db.Preload("Driver").Preload("TrackingUpdates").First(&delivery, id).Error
	if err != nil {
		return nil, err
	}
	return &delivery, nil
}

func (r *deliveryRepository) GetDeliveryByTrackingNumber(trackingNumber string) (*Delivery, error) {
	var delivery Delivery
	err := r.db.Preload("Driver").Preload("TrackingUpdates").Where("tracking_number = ?", trackingNumber).First(&delivery).Error
	if err != nil {
		return nil, err
	}
	return &delivery, nil
}

func (r *deliveryRepository) GetDeliveryByOrderID(orderID string) (*Delivery, error) {
	var delivery Delivery
	err := r.db.Preload("Driver").Preload("TrackingUpdates").Where("order_id = ?", orderID).First(&delivery).Error
	if err != nil {
		return nil, err
	}
	return &delivery, nil
}

func (r *deliveryRepository) UpdateDelivery(delivery *Delivery) error {
	return r.db.Save(delivery).Error
}

func (r *deliveryRepository) DeleteDelivery(id uint) error {
	return r.db.Delete(&Delivery{}, id).Error
}

func (r *deliveryRepository) ListDeliveries(limit, offset int, status *DeliveryStatus) ([]Delivery, int64, error) {
	var deliveries []Delivery
	var total int64

	query := r.db.Model(&Delivery{}).Preload("Driver")
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Limit(limit).Offset(offset).Order("created_at DESC").Find(&deliveries).Error
	return deliveries, total, err
}

func (r *deliveryRepository) GetDeliveriesByDriver(driverID uint, limit, offset int) ([]Delivery, int64, error) {
	var deliveries []Delivery
	var total int64

	query := r.db.Model(&Delivery{}).Where("driver_id = ?", driverID)

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Limit(limit).Offset(offset).Order("created_at DESC").Find(&deliveries).Error
	return deliveries, total, err
}

func (r *deliveryRepository) GetDeliveriesByDateRange(startDate, endDate time.Time, limit, offset int) ([]Delivery, int64, error) {
	var deliveries []Delivery
	var total int64

	query := r.db.Model(&Delivery{}).Where("created_at BETWEEN ? AND ?", startDate, endDate)

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Preload("Driver").Limit(limit).Offset(offset).Order("created_at DESC").Find(&deliveries).Error
	return deliveries, total, err
}

// Driver methods implementation
func (r *deliveryRepository) CreateDriver(driver *DeliveryDriver) error {
	return r.db.Create(driver).Error
}

func (r *deliveryRepository) GetDriverByID(id uint) (*DeliveryDriver, error) {
	var driver DeliveryDriver
	err := r.db.First(&driver, id).Error
	if err != nil {
		return nil, err
	}
	return &driver, nil
}

func (r *deliveryRepository) GetDriverByUserID(userID uint) (*DeliveryDriver, error) {
	var driver DeliveryDriver
	err := r.db.Where("user_id = ?", userID).First(&driver).Error
	if err != nil {
		return nil, err
	}
	return &driver, nil
}

func (r *deliveryRepository) UpdateDriver(driver *DeliveryDriver) error {
	return r.db.Save(driver).Error
}

func (r *deliveryRepository) DeleteDriver(id uint) error {
	return r.db.Delete(&DeliveryDriver{}, id).Error
}

func (r *deliveryRepository) ListDrivers(limit, offset int, isActive *bool) ([]DeliveryDriver, int64, error) {
	var drivers []DeliveryDriver
	var total int64

	query := r.db.Model(&DeliveryDriver{})
	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Limit(limit).Offset(offset).Order("created_at DESC").Find(&drivers).Error
	return drivers, total, err
}

func (r *deliveryRepository) GetAvailableDrivers(vehicleType *VehicleType, lat, lng *float64, radius float64) ([]DeliveryDriver, error) {
	var drivers []DeliveryDriver
	query := r.db.Where("is_active = ? AND is_available = ?", true, true)

	if vehicleType != nil {
		query = query.Where("vehicle_type = ?", *vehicleType)
	}

	// If coordinates provided, filter by radius (simplified - in production use PostGIS)
	if lat != nil && lng != nil && radius > 0 {
		query = query.Where("current_lat IS NOT NULL AND current_lng IS NOT NULL")
	}

	err := query.Find(&drivers).Error
	return drivers, err
}

func (r *deliveryRepository) UpdateDriverLocation(driverID uint, lat, lng float64) error {
	now := time.Now()
	return r.db.Model(&DeliveryDriver{}).Where("id = ?", driverID).Updates(map[string]interface{}{
		"current_lat":          lat,
		"current_lng":          lng,
		"last_location_update": now,
	}).Error
}

// Tracking methods implementation
func (r *deliveryRepository) CreateTrackingUpdate(update *TrackingUpdate) error {
	return r.db.Create(update).Error
}

func (r *deliveryRepository) GetTrackingUpdatesByDeliveryID(deliveryID uint) ([]TrackingUpdate, error) {
	var updates []TrackingUpdate
	err := r.db.Where("delivery_id = ?", deliveryID).Order("timestamp DESC").Find(&updates).Error
	return updates, err
}

// Zone methods implementation
func (r *deliveryRepository) CreateDeliveryZone(zone *DeliveryZone) error {
	return r.db.Create(zone).Error
}

func (r *deliveryRepository) GetDeliveryZoneByID(id uint) (*DeliveryZone, error) {
	var zone DeliveryZone
	err := r.db.First(&zone, id).Error
	if err != nil {
		return nil, err
	}
	return &zone, nil
}

func (r *deliveryRepository) UpdateDeliveryZone(zone *DeliveryZone) error {
	return r.db.Save(zone).Error
}

func (r *deliveryRepository) DeleteDeliveryZone(id uint) error {
	return r.db.Delete(&DeliveryZone{}, id).Error
}

func (r *deliveryRepository) ListDeliveryZones(limit, offset int, isActive *bool) ([]DeliveryZone, int64, error) {
	var zones []DeliveryZone
	var total int64

	query := r.db.Model(&DeliveryZone{})
	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Limit(limit).Offset(offset).Order("created_at DESC").Find(&zones).Error
	return zones, total, err
}

func (r *deliveryRepository) GetZoneByCoordinates(lat, lng float64) (*DeliveryZone, error) {
	// Simplified implementation - in production, use PostGIS for proper polygon queries
	var zone DeliveryZone
	err := r.db.Where("is_active = ?", true).First(&zone).Error
	if err != nil {
		return nil, err
	}
	return &zone, nil
}

// Analytics methods implementation
func (r *deliveryRepository) GetDeliveryStats(startDate, endDate *time.Time) (*DeliveryStatsResponse, error) {
	stats := &DeliveryStatsResponse{}

	query := r.db.Model(&Delivery{})
	if startDate != nil && endDate != nil {
		query = query.Where("created_at BETWEEN ? AND ?", *startDate, *endDate)
	}

	// Total deliveries
	query.Count(&stats.TotalDeliveries)

	// Status counts
	query.Where("status = ?", DeliveryStatusPending).Count(&stats.PendingDeliveries)
	query.Where("status = ?", DeliveryStatusInTransit).Count(&stats.InTransitDeliveries)
	query.Where("status = ?", DeliveryStatusDelivered).Count(&stats.CompletedDeliveries)
	query.Where("status = ?", DeliveryStatusCancelled).Count(&stats.CancelledDeliveries)

	// Revenue calculation
	var totalRevenue struct {
		Sum int64
	}
	query.Select("SUM(delivery_fee) as sum").Where("status = ?", DeliveryStatusDelivered).Scan(&totalRevenue)
	stats.TotalRevenue = totalRevenue.Sum

	return stats, nil
}

func (r *deliveryRepository) GetDriverStats(driverID uint, startDate, endDate *time.Time) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	query := r.db.Model(&Delivery{}).Where("driver_id = ?", driverID)
	if startDate != nil && endDate != nil {
		query = query.Where("created_at BETWEEN ? AND ?", *startDate, *endDate)
	}

	var totalDeliveries int64
	var completedDeliveries int64
	var totalRevenue int64

	query.Count(&totalDeliveries)
	query.Where("status = ?", DeliveryStatusDelivered).Count(&completedDeliveries)
	query.Where("status = ?", DeliveryStatusDelivered).Select("SUM(delivery_fee)").Scan(&totalRevenue)

	stats["total_deliveries"] = totalDeliveries
	stats["completed_deliveries"] = completedDeliveries
	stats["total_revenue"] = totalRevenue

	return stats, nil
}
