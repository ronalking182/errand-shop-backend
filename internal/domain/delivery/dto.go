package delivery

import "time"

// CreateDeliveryRequest represents request to create a delivery
type CreateDeliveryRequest struct {
	OrderID           string       `json:"order_id" validate:"required"`
	DeliveryType      DeliveryType `json:"delivery_type" validate:"required,oneof=standard express same_day scheduled"`
	PickupAddress     string       `json:"pickup_address" validate:"required,min=10,max=500"`
	PickupLatitude    *float64     `json:"pickup_latitude"`
	PickupLongitude   *float64     `json:"pickup_longitude"`
	PickupNotes       string       `json:"pickup_notes" validate:"max=500"`
	DeliveryAddress   string       `json:"delivery_address" validate:"required,min=10,max=500"`
	DeliveryLatitude  *float64     `json:"delivery_latitude"`
	DeliveryLongitude *float64     `json:"delivery_longitude"`
	DeliveryNotes     string       `json:"delivery_notes" validate:"max=500"`
	RecipientName     string       `json:"recipient_name" validate:"required,min=2,max=100"`
	RecipientPhone    string       `json:"recipient_phone" validate:"required,min=10,max=20"`
	ScheduledDate     *time.Time   `json:"scheduled_date"`
}

// UpdateDeliveryStatusRequest represents request to update delivery status (Admin only)
type UpdateDeliveryStatusRequest struct {
	Status             DeliveryStatus    `json:"status" validate:"required,oneof=pending sent_out in_progress delivered cancelled returned"`
	Message            string            `json:"message" validate:"max=500"`
	LogisticsProvider  LogisticsProvider `json:"logistics_provider" validate:"omitempty,oneof=dhl fedex ups gig kwik sendbox"`
	ProviderTrackingID string            `json:"provider_tracking_id" validate:"max=100"`
	InternalNotes      string            `json:"internal_notes" validate:"max=1000"`
	Latitude           *float64          `json:"latitude"`
	Longitude          *float64          `json:"longitude"`
}

// AssignLogisticsProviderRequest represents request to assign logistics provider
type AssignLogisticsProviderRequest struct {
	LogisticsProvider  LogisticsProvider `json:"logistics_provider" validate:"required,oneof=dhl fedex ups gig kwik sendbox"`
	ProviderTrackingID string            `json:"provider_tracking_id" validate:"required,min=3,max=100"`
	InternalNotes      string            `json:"internal_notes" validate:"max=1000"`
}

// DeliveryResponse represents delivery response
type DeliveryResponse struct {
	ID                 uint                     `json:"id"`
	OrderID            string                   `json:"order_id"`
	TrackingNumber     string                   `json:"tracking_number"`
	DeliveryType       DeliveryType             `json:"delivery_type"`
	Status             DeliveryStatus           `json:"status"`
	LogisticsProvider  LogisticsProvider        `json:"logistics_provider"`
	ProviderTrackingID string                   `json:"provider_tracking_id"`
	PickupAddress      string                   `json:"pickup_address"`
	PickupLatitude     *float64                 `json:"pickup_latitude"`
	PickupLongitude    *float64                 `json:"pickup_longitude"`
	PickupTime         *time.Time               `json:"pickup_time"`
	PickupNotes        string                   `json:"pickup_notes"`
	DeliveryAddress    string                   `json:"delivery_address"`
	DeliveryLatitude   *float64                 `json:"delivery_latitude"`
	DeliveryLongitude  *float64                 `json:"delivery_longitude"`
	DeliveryTime       *time.Time               `json:"delivery_time"`
	DeliveryNotes      string                   `json:"delivery_notes"`
	RecipientName      string                   `json:"recipient_name"`
	RecipientPhone     string                   `json:"recipient_phone"`
	ScheduledDate      *time.Time               `json:"scheduled_date"`
	EstimatedTime      *time.Time               `json:"estimated_time"`
	ActualTime         *time.Time               `json:"actual_time"`
	DeliveryFee        int64                    `json:"delivery_fee"`
	Distance           *float64                 `json:"distance"`
	Duration           *int                     `json:"duration"`
	InternalNotes      string                   `json:"internal_notes,omitempty"` // Only for admins
	TrackingUpdates    []TrackingUpdateResponse `json:"tracking_updates,omitempty"`
	CreatedAt          time.Time                `json:"created_at"`
	UpdatedAt          time.Time                `json:"updated_at"`
}

// TrackingUpdateResponse represents tracking update response
type TrackingUpdateResponse struct {
	ID          uint           `json:"id"`
	Status      DeliveryStatus `json:"status"`
	Message     string         `json:"message"`
	IsAutomatic bool           `json:"is_automatic"`
	Timestamp   time.Time      `json:"timestamp"`
}

// DeliveryStatsResponse represents delivery statistics
type DeliveryStatsResponse struct {
	TotalDeliveries      int64                       `json:"total_deliveries"`
	PendingDeliveries    int64                       `json:"pending_deliveries"`
	SentOutDeliveries    int64                       `json:"sent_out_deliveries"`
	InProgressDeliveries int64                       `json:"in_progress_deliveries"`
	InTransitDeliveries  int64                       `json:"in_transit_deliveries"`
	CompletedDeliveries  int64                       `json:"completed_deliveries"`
	CancelledDeliveries  int64                       `json:"cancelled_deliveries"`
	AverageDeliveryTime  float64                     `json:"average_delivery_time"`
	TotalRevenue         int64                       `json:"total_revenue"`
	ProviderBreakdown    map[LogisticsProvider]int64 `json:"provider_breakdown"`
}

// DeliveryQuoteRequest represents request for delivery quote
type DeliveryQuoteRequest struct {
	PickupLatitude    float64      `json:"pickup_latitude" validate:"required,min=-90,max=90"`
	PickupLongitude   float64      `json:"pickup_longitude" validate:"required,min=-180,max=180"`
	DeliveryLatitude  float64      `json:"delivery_latitude" validate:"required,min=-90,max=90"`
	DeliveryLongitude float64      `json:"delivery_longitude" validate:"required,min=-180,max=180"`
	DeliveryType      DeliveryType `json:"delivery_type" validate:"required,oneof=standard express same_day scheduled"`
	ScheduledDate     *time.Time   `json:"scheduled_date"`
}

// DeliveryQuoteResponse represents delivery quote response
type DeliveryQuoteResponse struct {
	Distance          float64    `json:"distance"`
	EstimatedTime     int        `json:"estimated_time"`
	DeliveryFee       int64      `json:"delivery_fee"`
	EstimatedPickup   *time.Time `json:"estimated_pickup"`
	EstimatedDelivery *time.Time `json:"estimated_delivery"`
}

// CreateDriverRequest represents request to create a delivery driver
type CreateDriverRequest struct {
	UserID           uint        `json:"user_id" validate:"required"`
	LicenseNumber    string      `json:"license_number" validate:"required,min=5,max=50"`
	VehicleType      VehicleType `json:"vehicle_type" validate:"required,oneof=bike motorcycle car van truck"`
	VehicleNumber    string      `json:"vehicle_number" validate:"required,min=3,max=50"`
	VehiclePlate     string      `json:"vehicle_plate" validate:"required,min=3,max=20"`
	VehicleModel     string      `json:"vehicle_model" validate:"required,min=2,max=100"`
	VehicleColor     string      `json:"vehicle_color" validate:"required,min=2,max=50"`
	EmergencyContact string      `json:"emergency_contact" validate:"required,min=10,max=20"`
	IsActive         bool        `json:"is_active"`
}

// UpdateDriverLocationRequest represents request to update driver location
type UpdateDriverLocationRequest struct {
	Latitude  float64  `json:"latitude" validate:"required,min=-90,max=90"`
	Longitude float64  `json:"longitude" validate:"required,min=-180,max=180"`
	Heading   *float64 `json:"heading" validate:"omitempty,min=0,max=360"`
	Speed     *float64 `json:"speed" validate:"omitempty,min=0"`
}

// DeliveryDriverResponse represents delivery driver response
type DeliveryDriverResponse struct {
	ID                 uint        `json:"id"`
	UserID             uint        `json:"user_id"`
	LicenseNumber      string      `json:"license_number"`
	VehicleType        VehicleType `json:"vehicle_type"`
	VehicleNumber      string      `json:"vehicle_number"`
	VehiclePlate       string      `json:"vehicle_plate"`
	VehicleModel       string      `json:"vehicle_model"`
	VehicleColor       string      `json:"vehicle_color"`
	EmergencyContact   string      `json:"emergency_contact"`
	IsActive           bool        `json:"is_active"`
	IsAvailable        bool        `json:"is_available"`
	CurrentLatitude    *float64    `json:"current_latitude"`
	CurrentLongitude   *float64    `json:"current_longitude"`
	LastLocationUpdate *time.Time  `json:"last_location_update"`
	Rating             *float64    `json:"rating"`
	TotalDeliveries    int64       `json:"total_deliveries"`
	CreatedAt          time.Time   `json:"created_at"`
	UpdatedAt          time.Time   `json:"updated_at"`
}

// AssignDriverRequest represents request to assign driver to delivery
type AssignDriverRequest struct {
	DriverID      uint   `json:"driver_id" validate:"required"`
	InternalNotes string `json:"internal_notes" validate:"max=1000"`
}
