package types

// DeliveryZone represents a delivery zone with pricing and locations
type DeliveryZone struct {
	ZoneID    int      `json:"zoneId"`
	Price     int      `json:"price"`
	Locations []string `json:"locations"`
}

// MatchResult represents the result of address matching
type MatchResult struct {
	ZoneID         int     `json:"zoneId"`
	ZoneName       string  `json:"zoneName"`
	MatchedKeyword string  `json:"matchedKeyword"`
	MatchedBy      string  `json:"matchedBy"` // "exact" or "fuzzy"
	Confidence     float64 `json:"confidence"`
	Price          int     `json:"price"`
}

// NoMatchResult represents when no match is found
type NoMatchResult struct {
	MatchedBy   string            `json:"matchedBy"` // "none"
	Message     string            `json:"message"`
	Suggestions []MatchSuggestion `json:"suggestions"`
}

// MatchSuggestion represents a suggested match
type MatchSuggestion struct {
	ZoneID     int     `json:"zoneId"`
	Keyword    string  `json:"keyword"`
	Price      int     `json:"price"`
	Confidence float64 `json:"confidence"`
}

// DeliveryEstimateRequest represents the request for delivery estimation
type DeliveryEstimateRequest struct {
	AddressID string `json:"addressId" validate:"required"`
	UserID    string `json:"userId,omitempty"` // Optional, used when not authenticated
}

// OrderConfirmRequest represents the request for order confirmation
type OrderConfirmRequest struct {
	AddressID   string `json:"addressId" validate:"required"`
	ClientPrice int    `json:"clientPrice" validate:"required,min=1"`
}

// Address represents a user's stored address
type Address struct {
	ID     string `json:"id"`
	UserID string `json:"userId"`
	Text   string `json:"text"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}