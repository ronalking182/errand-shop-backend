package system

import "time"

// Setting represents system configuration settings
type Setting struct {
	ID        uint                   `gorm:"primaryKey" json:"id"`
	Key       string                 `gorm:"uniqueIndex;size:100" json:"key"`
	Value     map[string]interface{} `gorm:"type:jsonb" json:"value"`
	UpdatedAt time.Time              `json:"updatedAt"`
}

// SystemConfig represents overall system configuration
type SystemConfig struct {
	ID              uint              `gorm:"primaryKey" json:"id"`
	MaintenanceMode bool              `gorm:"default:false" json:"maintenanceMode"`
	Settings        map[string]string `gorm:"type:jsonb" json:"settings"`
	Notifications   bool              `gorm:"default:true" json:"notifications"`
	CreatedAt       time.Time         `json:"createdAt"`
	UpdatedAt       time.Time         `json:"updatedAt"`
}

// AuditLog represents system audit logs
type AuditLog struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"index" json:"userId"`
	Action      string    `gorm:"size:100;not null" json:"action"`
	Resource    string    `gorm:"size:100;not null" json:"resource"`
	ResourceID  string    `gorm:"size:100" json:"resourceId"`
	Description string    `gorm:"type:text" json:"description"`
	IPAddress   string    `gorm:"size:45" json:"ipAddress"`
	UserAgent   string    `gorm:"type:text" json:"userAgent"`
	CreatedAt   time.Time `json:"createdAt"`
}

// SystemHealth represents system health checks
type SystemHealth struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Component string    `gorm:"size:100;not null" json:"component"`
	Status    string    `gorm:"size:20;not null" json:"status"` // healthy, warning, critical
	Message   string    `gorm:"type:text" json:"message"`
	CheckedAt time.Time `json:"checkedAt"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// TableName methods for custom table names
func (SystemConfig) TableName() string {
	return "system_configs"
}

func (AuditLog) TableName() string {
	return "audit_logs"
}

func (SystemHealth) TableName() string {
	return "system_health"
}
