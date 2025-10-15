package delivery

import (
	"time"

	"gorm.io/gorm"
)

// DeliveryStatus represents the status of a delivery
type DeliveryStatus string

const (
	DeliveryStatusPending     DeliveryStatus = "pending"
	DeliveryStatusSentOut     DeliveryStatus = "sent_out"
	DeliveryStatusInProgress  DeliveryStatus = "in_progress"
	DeliveryStatusPickedUp    DeliveryStatus = "picked_up"
	DeliveryStatusInTransit   DeliveryStatus = "in_transit"
	DeliveryStatusAssigned    DeliveryStatus = "assigned"
	DeliveryStatusDelivered   DeliveryStatus = "delivered"
	DeliveryStatusCancelled   DeliveryStatus = "cancelled"
	DeliveryStatusReturned    DeliveryStatus = "returned"
)

// DeliveryType represents the type of delivery service
type DeliveryType string

const (
	DeliveryTypeStandard  DeliveryType = "standard"
	DeliveryTypeExpress   DeliveryType = "express"
	DeliveryTypeSameDay   DeliveryType = "same_day"
	DeliveryTypeScheduled DeliveryType = "scheduled"
)

// LogisticsProvider represents third-party logistics providers
type LogisticsProvider string

const (
	ProviderDHL     LogisticsProvider = "dhl"
	ProviderFedEx   LogisticsProvider = "fedex"
	ProviderUPS     LogisticsProvider = "ups"
	ProviderGIG     LogisticsProvider = "gig"
	ProviderKwik    LogisticsProvider = "kwik"
	ProviderSendbox LogisticsProvider = "sendbox"
)

// VehicleType represents the type of delivery vehicle
type VehicleType string

const (
	VehicleTypeBike       VehicleType = "bike"
	VehicleTypeMotorcycle VehicleType = "motorcycle"
	VehicleTypeCar        VehicleType = "car"
	VehicleTypeVan        VehicleType = "van"
	VehicleTypeTruck      VehicleType = "truck"
)

// Delivery represents a delivery order
type Delivery struct {
	ID                 uint              `json:"id" gorm:"primaryKey"`
	OrderID            string            `json:"order_id" gorm:"type:uuid;not null;index"`
	TrackingNumber     string            `json:"tracking_number" gorm:"uniqueIndex;size:50"`
	DeliveryType       DeliveryType      `json:"delivery_type" gorm:"type:varchar(20);not null"`
	Status             DeliveryStatus    `json:"status" gorm:"type:varchar(20);not null;default:'pending'"`
	LogisticsProvider  LogisticsProvider `json:"logistics_provider" gorm:"type:varchar(20)"`
	ProviderTrackingID string            `json:"provider_tracking_id" gorm:"size:100"`

	// Pickup Information
	PickupAddress   string     `json:"pickup_address" gorm:"type:text;not null"`
	PickupLatitude  *float64   `json:"pickup_latitude"`
	PickupLongitude *float64   `json:"pickup_longitude"`
	PickupTime      *time.Time `json:"pickup_time"`
	PickupNotes     string     `json:"pickup_notes" gorm:"type:text"`

	// Delivery Information
	DeliveryAddress   string     `json:"delivery_address" gorm:"type:text;not null"`
	DeliveryLatitude  *float64   `json:"delivery_latitude"`
	DeliveryLongitude *float64   `json:"delivery_longitude"`
	DeliveryTime      *time.Time `json:"delivery_time"`
	DeliveryNotes     string     `json:"delivery_notes" gorm:"type:text"`
	RecipientName     string     `json:"recipient_name" gorm:"size:100;not null"`
	RecipientPhone    string     `json:"recipient_phone" gorm:"size:20;not null"`

	// Scheduling
	ScheduledDate *time.Time `json:"scheduled_date"`
	EstimatedTime *time.Time `json:"estimated_time"`
	ActualTime    *time.Time `json:"actual_time"`

	// Pricing
	DeliveryFee int64    `json:"delivery_fee" gorm:"not null"` // in kobo
	Distance    *float64 `json:"distance"`                     // in kilometers
	Duration    *int     `json:"duration"`                     // in minutes

	// Driver assignment
	DriverID *uint           `json:"driver_id" gorm:"index"`
	Driver   *DeliveryDriver `json:"driver,omitempty" gorm:"foreignKey:DriverID"`

	// Tracking
	TrackingUpdates []TrackingUpdate `json:"tracking_updates,omitempty" gorm:"foreignKey:DeliveryID"` 

	// Admin fields
	AssignedBy    *uint  `json:"assigned_by" gorm:"index"`        // Admin who assigned to logistics provider
	LastUpdatedBy *uint  `json:"last_updated_by" gorm:"index"`    // Admin who last updated status
	InternalNotes string `json:"internal_notes" gorm:"type:text"` // Internal admin notes

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// TrackingUpdate represents a delivery tracking update
type TrackingUpdate struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	DeliveryID  uint           `json:"delivery_id" gorm:"not null;index"`
	Status      DeliveryStatus `json:"status" gorm:"type:varchar(20);not null"`
	Message     string         `json:"message" gorm:"type:text"`
	UpdatedBy   *uint          `json:"updated_by" gorm:"index"`           // Admin who made the update
	IsAutomatic bool           `json:"is_automatic" gorm:"default:false"` // If update came from provider API
	Timestamp   time.Time      `json:"timestamp" gorm:"not null"`
	CreatedAt   time.Time      `json:"created_at"`
}

// DeliveryZone represents delivery service areas
type DeliveryZone struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"size:100;not null"`
	Description string         `json:"description" gorm:"type:text"`
	IsActive    bool           `json:"is_active" gorm:"default:true"`
	BaseFee     int64          `json:"base_fee" gorm:"not null"`     // in kobo
	PerKmFee    int64          `json:"per_km_fee" gorm:"not null"`   // in kobo
	MaxDistance float64        `json:"max_distance" gorm:"not null"` // in kilometers
	Boundaries  string         `json:"boundaries" gorm:"type:text"`  // JSON polygon coordinates
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// DeliveryDriver represents a delivery driver
type DeliveryDriver struct {
	ID                 uint           `json:"id" gorm:"primaryKey"`
	UserID             uint           `json:"user_id" gorm:"not null;uniqueIndex"`
	LicenseNumber      string         `json:"license_number" gorm:"size:50;not null;uniqueIndex"`
	VehicleType        VehicleType    `json:"vehicle_type" gorm:"type:varchar(20);not null"`
	VehicleNumber      string         `json:"vehicle_number" gorm:"size:20;not null"`
	VehiclePlate       string         `json:"vehicle_plate" gorm:"size:20;not null"`
	VehicleModel       string         `json:"vehicle_model" gorm:"size:100;not null"`
	VehicleColor       string         `json:"vehicle_color" gorm:"size:50;not null"`
	EmergencyContact   string         `json:"emergency_contact" gorm:"size:20;not null"`
	IsActive           bool           `json:"is_active" gorm:"default:true"`
	IsAvailable        bool           `json:"is_available" gorm:"default:false"`
	CurrentLatitude    *float64       `json:"current_latitude"`
	CurrentLongitude   *float64       `json:"current_longitude"`
	LastLocationUpdate *time.Time     `json:"last_location_update"`
	Rating             *float64       `json:"rating" gorm:"type:decimal(3,2)"`
	TotalDeliveries    int64          `json:"total_deliveries" gorm:"default:0"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName returns the table name for DeliveryDriver
func (DeliveryDriver) TableName() string {
	return "delivery_drivers"
}
