package system

import (
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Setting operations
func (r *Repository) GetSetting(key string) (*Setting, error) {
	var setting Setting
	err := r.db.Where("key = ?", key).First(&setting).Error
	if err != nil {
		return nil, err
	}
	return &setting, nil
}

func (r *Repository) GetAllSettings() ([]Setting, error) {
	var settings []Setting
	err := r.db.Find(&settings).Error
	return settings, err
}

func (r *Repository) CreateOrUpdateSetting(setting *Setting) error {
	return r.db.Save(setting).Error
}

func (r *Repository) DeleteSetting(key string) error {
	return r.db.Where("key = ?", key).Delete(&Setting{}).Error
}

// SystemConfig operations
func (r *Repository) GetSystemConfig() (*SystemConfig, error) {
	var config SystemConfig
	err := r.db.First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *Repository) CreateOrUpdateSystemConfig(config *SystemConfig) error {
	return r.db.Save(config).Error
}

// AuditLog operations
func (r *Repository) CreateAuditLog(log *AuditLog) error {
	return r.db.Create(log).Error
}

func (r *Repository) GetAuditLogs(limit, offset int, userID *uint) ([]AuditLog, int64, error) {
	var logs []AuditLog
	var total int64

	query := r.db.Model(&AuditLog{})
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}

	// Get total count
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err = query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&logs).Error
	return logs, total, err
}

func (r *Repository) GetAuditLogsByResource(resource string, resourceID string) ([]AuditLog, error) {
	var logs []AuditLog
	query := r.db.Where("resource = ?", resource)
	if resourceID != "" {
		query = query.Where("resource_id = ?", resourceID)
	}
	err := query.Order("created_at DESC").Find(&logs).Error
	return logs, err
}

// SystemHealth operations
func (r *Repository) CreateOrUpdateSystemHealth(health *SystemHealth) error {
	return r.db.Save(health).Error
}

func (r *Repository) GetSystemHealthByComponent(component string) (*SystemHealth, error) {
	var health SystemHealth
	err := r.db.Where("component = ?", component).First(&health).Error
	if err != nil {
		return nil, err
	}
	return &health, nil
}

func (r *Repository) GetAllSystemHealth() ([]SystemHealth, error) {
	var healthChecks []SystemHealth
	err := r.db.Order("checked_at DESC").Find(&healthChecks).Error
	return healthChecks, err
}

func (r *Repository) DeleteOldAuditLogs(days int) error {
	return r.db.Where("created_at < NOW() - INTERVAL ? DAY", days).Delete(&AuditLog{}).Error
}
