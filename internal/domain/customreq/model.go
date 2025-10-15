package customreq

import "time"

type CustomRequestStatus string

const (
	CustomRequestStatusOpen      CustomRequestStatus = "open"
	CustomRequestStatusQuoted    CustomRequestStatus = "quoted"
	CustomRequestStatusAccepted  CustomRequestStatus = "accepted"
	CustomRequestStatusCompleted CustomRequestStatus = "completed"
	CustomRequestStatusCancelled CustomRequestStatus = "cancelled"
)

type CustomRequestQuote struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	CustomRequestID uint      `json:"customRequestId"`
	StoreID         uint      `json:"storeId"`
	Price           int64     `json:"price"`
	Description     string    `gorm:"type:text" json:"description"`
	EstimatedTime   string    `json:"estimatedTime"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type CustomRequest struct {
	ID         uint                 `gorm:"primaryKey" json:"id"`
	CustomerID uint                 `json:"customerId"`
	Title      string               `gorm:"size:200" json:"title"`
	Details    string               `gorm:"type:text" json:"details"`
	Status     CustomRequestStatus  `gorm:"type:varchar(20);default:'open'" json:"status"`
	Quotes     []CustomRequestQuote `json:"quotes"`
	CreatedAt  time.Time            `json:"createdAt"`
	UpdatedAt  time.Time            `json:"updatedAt"`
}
