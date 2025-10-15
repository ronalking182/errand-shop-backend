package system

import (
	"gorm.io/gorm"
	"time"
)

type Service struct {
	repo *Repository
	db   *gorm.DB
}

func NewService(repo *Repository, db *gorm.DB) *Service {
	return &Service{
		repo: repo,
		db:   db,
	}
}

// Setting operations
func (s *Service) GetSetting(key string) (*Setting, error) {
	return s.repo.GetSetting(key)
}

func (s *Service) GetAllSettings() ([]Setting, error) {
	return s.repo.GetAllSettings()
}

func (s *Service) UpdateSetting(key string, value map[string]interface{}) error {
	setting := &Setting{
		Key:       key,
		Value:     value,
		UpdatedAt: time.Now(),
	}
	return s.repo.CreateOrUpdateSetting(setting)
}

func (s *Service) DeleteSetting(key string) error {
	return s.repo.DeleteSetting(key)
}

// SystemConfig operations
func (s *Service) GetSystemConfig() (*SystemConfig, error) {
	config, err := s.repo.GetSystemConfig()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create default config if none exists
			defaultConfig := &SystemConfig{
				MaintenanceMode: false,
				Settings:        make(map[string]string),
				Notifications:   true,
			}
			err = s.repo.CreateOrUpdateSystemConfig(defaultConfig)
			if err != nil {
				return nil, err
			}
			return defaultConfig, nil
		}
		return nil, err
	}
	return config, nil
}

func (s *Service) UpdateSystemConfig(req UpdateSystemConfigRequest) error {
	config, err := s.GetSystemConfig()
	if err != nil {
		return err
	}

	config.MaintenanceMode = req.MaintenanceMode
	config.Settings = req.Settings
	config.Notifications = req.Notifications
	config.UpdatedAt = time.Now()

	return s.repo.CreateOrUpdateSystemConfig(config)
}

// AuditLog operations
func (s *Service) CreateAuditLog(userID uint, action, resource, resourceID, description, ipAddress, userAgent string) error {
	log := &AuditLog{
		UserID:      userID,
		Action:      action,
		Resource:    resource,
		ResourceID:  resourceID,
		Description: description,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		CreatedAt:   time.Now(),
	}
	return s.repo.CreateAuditLog(log)
}

func (s *Service) GetAuditLogs(page, limit int, userID *uint) ([]AuditLog, int64, error) {
	offset := (page - 1) * limit
	return s.repo.GetAuditLogs(limit, offset, userID)
}

func (s *Service) GetAuditLogsByResource(resource, resourceID string) ([]AuditLog, error) {
	return s.repo.GetAuditLogsByResource(resource, resourceID)
}

// SystemHealth operations
func (s *Service) UpdateSystemHealth(component, status, message string) error {
	health := &SystemHealth{
		Component: component,
		Status:    status,
		Message:   message,
		CheckedAt: time.Now(),
	}
	return s.repo.CreateOrUpdateSystemHealth(health)
}

func (s *Service) GetSystemHealth() ([]SystemHealth, error) {
	return s.repo.GetAllSystemHealth()
}

func (s *Service) GetSystemStats() (*SystemStatsResponse, error) {
	stats := &SystemStatsResponse{}

	// Get user counts
	s.db.Table("users").Count(&stats.TotalUsers)
	s.db.Table("users").Where("is_active = ?", true).Count(&stats.ActiveUsers)

	// Get order count
	s.db.Table("orders").Count(&stats.TotalOrders)

	// Get product count
	s.db.Table("products").Count(&stats.TotalProducts)

	// Check database health
	sqlDB, err := s.db.DB()
	if err != nil {
		stats.DatabaseHealth = false
	} else {
		err = sqlDB.Ping()
		stats.DatabaseHealth = err == nil
	}

	// System uptime (placeholder - would need to track actual start time)
	stats.SystemUptime = time.Now().Unix()

	return stats, nil
}

func (s *Service) CleanupOldAuditLogs(days int) error {
	return s.repo.DeleteOldAuditLogs(days)
}

// Helper method to log system actions
func (s *Service) LogSystemAction(userID uint, action, description, ipAddress, userAgent string) {
	s.CreateAuditLog(userID, action, "system", "", description, ipAddress, userAgent)
}
