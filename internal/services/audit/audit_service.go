package audit

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuditLog struct {
	gorm.Model
	UserID     *uuid.UUID             `json:"userID" gorm:"type:uuid;index"`
	Action     string                 `json:"action" gorm:"not null;index"`
	Resource   string                 `json:"resource" gorm:"not null;index"`
	ResourceID *string                `json:"resourceID" gorm:"index"`
	IPAddress  string                 `json:"ipAddress"`
	UserAgent  string                 `json:"userAgent"`
	Metadata   map[string]interface{} `json:"metadata" gorm:"type:jsonb"`
	Timestamp  time.Time              `json:"timestamp" gorm:"index"`
}

type AuditService struct {
	db *gorm.DB
}

func NewAuditService(db *gorm.DB) *AuditService {
	return &AuditService{db: db}
}

func (a *AuditService) Log(ctx context.Context, log *AuditLog) error {
	log.Timestamp = time.Now()
	return a.db.WithContext(ctx).Create(log).Error
}

func (a *AuditService) LogUserAction(ctx context.Context, userID uuid.UUID, action, resource string, metadata map[string]interface{}, ipAddress, userAgent string) error {
	log := &AuditLog{
		UserID:    &userID,
		Action:    action,
		Resource:  resource,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Metadata:  metadata,
	}
	return a.Log(ctx, log)
}

func (a *AuditService) LogSystemAction(ctx context.Context, action, resource string, metadata map[string]interface{}) error {
	log := &AuditLog{
		Action:   action,
		Resource: resource,
		Metadata: metadata,
	}
	return a.Log(ctx, log)
}
