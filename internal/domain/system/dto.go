package system

import "time"

// Request DTOs
type UpdateSystemConfigRequest struct {
	MaintenanceMode bool              `json:"maintenanceMode"`
	Settings        map[string]string `json:"settings"`
	Notifications   bool              `json:"notifications"`
}

type CreateAuditLogRequest struct {
	Action      string `json:"action" binding:"required"`
	Resource    string `json:"resource" binding:"required"`
	ResourceID  string `json:"resourceId"`
	Description string `json:"description"`
}

type SystemHealthRequest struct {
	Component string `json:"component"`
	Status    string `json:"status" binding:"required,oneof=healthy warning critical"`
	Message   string `json:"message"`
}

// Response DTOs
type SystemConfigResponse struct {
	ID              uint              `json:"id"`
	MaintenanceMode bool              `json:"maintenanceMode"`
	Settings        map[string]string `json:"settings"`
	Notifications   bool              `json:"notifications"`
	UpdatedAt       time.Time         `json:"updatedAt"`
}

type AuditLogResponse struct {
	ID          uint      `json:"id"`
	UserID      uint      `json:"userId"`
	Action      string    `json:"action"`
	Resource    string    `json:"resource"`
	ResourceID  string    `json:"resourceId"`
	Description string    `json:"description"`
	IPAddress   string    `json:"ipAddress"`
	UserAgent   string    `json:"userAgent"`
	CreatedAt   time.Time `json:"createdAt"`
}

type SystemHealthResponse struct {
	ID        uint      `json:"id"`
	Component string    `json:"component"`
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	CheckedAt time.Time `json:"checkedAt"`
}

type SystemStatsResponse struct {
	TotalUsers     int64 `json:"totalUsers"`
	ActiveUsers    int64 `json:"activeUsers"`
	TotalOrders    int64 `json:"totalOrders"`
	TotalProducts  int64 `json:"totalProducts"`
	SystemUptime   int64 `json:"systemUptime"`
	DatabaseHealth bool  `json:"databaseHealth"`
}

type AuditLogListResponse struct {
	Logs       []AuditLogResponse `json:"logs"`
	Total      int64              `json:"total"`
	Page       int                `json:"page"`
	Limit      int                `json:"limit"`
	TotalPages int                `json:"totalPages"`
}

// Helper functions
func ToSystemConfigResponse(config *SystemConfig) SystemConfigResponse {
	return SystemConfigResponse{
		ID:              config.ID,
		MaintenanceMode: config.MaintenanceMode,
		Settings:        config.Settings,
		Notifications:   config.Notifications,
		UpdatedAt:       config.UpdatedAt,
	}
}

func ToAuditLogResponse(log *AuditLog) AuditLogResponse {
	return AuditLogResponse{
		ID:          log.ID,
		UserID:      log.UserID,
		Action:      log.Action,
		Resource:    log.Resource,
		ResourceID:  log.ResourceID,
		Description: log.Description,
		IPAddress:   log.IPAddress,
		UserAgent:   log.UserAgent,
		CreatedAt:   log.CreatedAt,
	}
}

func ToAuditLogResponses(logs []AuditLog) []AuditLogResponse {
	responses := make([]AuditLogResponse, len(logs))
	for i, log := range logs {
		responses[i] = ToAuditLogResponse(&log)
	}
	return responses
}
