package notifications

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type NotificationRepository interface {
	Create(notification *Notification) error
	GetByID(id uint) (*Notification, error)
	GetByRecipient(recipientID uuid.UUID, recipientType NotificationRecipient, page, limit int) ([]Notification, int64, error)
	GetUnreadCount(recipientID uuid.UUID, recipientType NotificationRecipient) (int64, error)
	MarkAsRead(id uint) error
	MarkAllAsRead(recipientID uuid.UUID, recipientType NotificationRecipient) error
	Delete(id uint) error
	GetPendingNotifications(limit int) ([]Notification, error)
	UpdateStatus(id uint, status NotificationStatus) error
}

type TemplateRepository interface {
	Create(template *NotificationTemplate) error
	GetByID(id uint) (*NotificationTemplate, error)
	GetByType(notificationType NotificationType) (*NotificationTemplate, error)
	GetAll() ([]NotificationTemplate, error)
	Update(id uint, template *NotificationTemplate) error
	Delete(id uint) error
}

type PushTokenRepository interface {
	Create(token *PushToken) error
	GetByUserID(userID uuid.UUID, userType string) ([]PushToken, error)
	GetActiveTokens(userID uuid.UUID, userType string) ([]PushToken, error)
	UpdateLastUsed(id uint) error
	Deactivate(id uint) error
	DeleteByToken(token string) error
}

// FCM-specific repositories
type FCMTokenRepository interface {
	Create(token *FCMToken) error
	GetByUserID(userID uuid.UUID, userType string) ([]FCMToken, error)
	GetActiveTokens(userID uuid.UUID, userType string) ([]FCMToken, error)
	GetAllActiveTokens() ([]FCMToken, error)
	DeleteByToken(token string) error
	Deactivate(id uint) error
	GetByToken(token string) (*FCMToken, error)
}

type FCMMessageRepository interface {
	Create(message *FCMMessage) error
	GetByID(id uint) (*FCMMessage, error)
	GetMessages(page, limit int) ([]FCMMessage, int64, error)
	GetMessagesByUser(userID uuid.UUID, page, limit int) ([]FCMMessage, int64, error)
	Update(id uint, message *FCMMessage) error
	Delete(id uint) error
}

type FCMMessageRecipientRepository interface {
	Create(recipient *FCMMessageRecipient) error
	CreateBatch(recipients []FCMMessageRecipient) error
	GetByMessageID(messageID uint) ([]FCMMessageRecipient, error)
	UpdateStatus(id uint, status string, error string) error
	MarkAsDelivered(id uint) error
	GetStats() (map[string]int64, error)
}

type notificationRepository struct {
	db *gorm.DB
}

type templateRepository struct {
	db *gorm.DB
}

type pushTokenRepository struct {
	db *gorm.DB
}

type fcmTokenRepository struct {
	db *gorm.DB
}

type fcmMessageRepository struct {
	db *gorm.DB
}

type fcmMessageRecipientRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

func NewTemplateRepository(db *gorm.DB) TemplateRepository {
	return &templateRepository{db: db}
}

func NewFCMTokenRepository(db *gorm.DB) FCMTokenRepository {
	return &fcmTokenRepository{db: db}
}

func NewFCMMessageRepository(db *gorm.DB) FCMMessageRepository {
	return &fcmMessageRepository{db: db}
}

func NewFCMMessageRecipientRepository(db *gorm.DB) FCMMessageRecipientRepository {
	return &fcmMessageRecipientRepository{db: db}
}

func NewPushTokenRepository(db *gorm.DB) PushTokenRepository {
	return &pushTokenRepository{db: db}
}

// Notification Repository Implementation
func (r *notificationRepository) Create(notification *Notification) error {
	return r.db.Create(notification).Error
}

func (r *notificationRepository) GetByID(id uint) (*Notification, error) {
	var notification Notification
	err := r.db.First(&notification, id).Error
	return &notification, err
}

func (r *notificationRepository) GetByRecipient(recipientID uuid.UUID, recipientType NotificationRecipient, page, limit int) ([]Notification, int64, error) {
	var notifications []Notification
	var total int64

	query := r.db.Model(&Notification{}).Where("recipient_id = ? AND recipient_type = ?", recipientID, recipientType)
	query.Count(&total)

	offset := (page - 1) * limit
	err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&notifications).Error

	return notifications, total, err
}

func (r *notificationRepository) GetUnreadCount(recipientID uuid.UUID, recipientType NotificationRecipient) (int64, error) {
	var count int64
	err := r.db.Model(&Notification{}).Where("recipient_id = ? AND recipient_type = ? AND status = ?", recipientID, recipientType, StatusPending).Count(&count).Error
	return count, err
}

func (r *notificationRepository) MarkAsRead(id uint) error {
	return r.db.Model(&Notification{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":  StatusRead,
		"read_at": "NOW()",
	}).Error
}

func (r *notificationRepository) MarkAllAsRead(recipientID uuid.UUID, recipientType NotificationRecipient) error {
	return r.db.Model(&Notification{}).Where("recipient_id = ? AND recipient_type = ? AND status = ?", recipientID, recipientType, StatusPending).Updates(map[string]interface{}{
		"status":  StatusRead,
		"read_at": "NOW()",
	}).Error
}

func (r *notificationRepository) Delete(id uint) error {
	return r.db.Delete(&Notification{}, id).Error
}

func (r *notificationRepository) GetPendingNotifications(limit int) ([]Notification, error) {
	var notifications []Notification
	err := r.db.Where("status = ?", StatusPending).Limit(limit).Find(&notifications).Error
	return notifications, err
}

func (r *notificationRepository) UpdateStatus(id uint, status NotificationStatus) error {
	updates := map[string]interface{}{"status": status}
	if status == StatusSent {
		updates["sent_at"] = "NOW()"
	}
	return r.db.Model(&Notification{}).Where("id = ?", id).Updates(updates).Error
}

// Template Repository Implementation
func (r *templateRepository) Create(template *NotificationTemplate) error {
	return r.db.Create(template).Error
}

func (r *templateRepository) GetByID(id uint) (*NotificationTemplate, error) {
	var template NotificationTemplate
	err := r.db.First(&template, id).Error
	return &template, err
}

func (r *templateRepository) GetByType(notificationType NotificationType) (*NotificationTemplate, error) {
	var template NotificationTemplate
	err := r.db.Where("type = ? AND is_active = ?", notificationType, true).First(&template).Error
	return &template, err
}

func (r *templateRepository) GetAll() ([]NotificationTemplate, error) {
	var templates []NotificationTemplate
	err := r.db.Find(&templates).Error
	return templates, err
}

func (r *templateRepository) Update(id uint, template *NotificationTemplate) error {
	return r.db.Model(&NotificationTemplate{}).Where("id = ?", id).Updates(template).Error
}

func (r *templateRepository) Delete(id uint) error {
	return r.db.Delete(&NotificationTemplate{}, id).Error
}

// Push Token Repository Implementation
func (r *pushTokenRepository) Create(token *PushToken) error {
	return r.db.Create(token).Error
}

func (r *pushTokenRepository) GetByUserID(userID uuid.UUID, userType string) ([]PushToken, error) {
	var tokens []PushToken
	err := r.db.Where("user_id = ? AND user_type = ?", userID, userType).Find(&tokens).Error
	return tokens, err
}

func (r *pushTokenRepository) GetActiveTokens(userID uuid.UUID, userType string) ([]PushToken, error) {
	var tokens []PushToken
	err := r.db.Where("user_id = ? AND user_type = ? AND is_active = ?", userID, userType, true).Find(&tokens).Error
	return tokens, err
}

func (r *pushTokenRepository) UpdateLastUsed(id uint) error {
	return r.db.Model(&PushToken{}).Where("id = ?", id).Update("last_used_at", "NOW()").Error
}

func (r *pushTokenRepository) Deactivate(id uint) error {
	return r.db.Model(&PushToken{}).Where("id = ?", id).Update("is_active", false).Error
}

func (r *pushTokenRepository) DeleteByToken(token string) error {
	return r.db.Where("token = ?", token).Delete(&PushToken{}).Error
}

// FCM Token Repository Implementation
func (r *fcmTokenRepository) Create(token *FCMToken) error {
	return r.db.Create(token).Error
}

func (r *fcmTokenRepository) GetByUserID(userID uuid.UUID, userType string) ([]FCMToken, error) {
	var tokens []FCMToken
	err := r.db.Where("user_id = ? AND user_type = ?", userID, userType).Find(&tokens).Error
	return tokens, err
}

func (r *fcmTokenRepository) GetActiveTokens(userID uuid.UUID, userType string) ([]FCMToken, error) {
	var tokens []FCMToken
	err := r.db.Where("user_id = ? AND user_type = ? AND is_active = ?", userID, userType, true).Find(&tokens).Error
	return tokens, err
}

func (r *fcmTokenRepository) GetAllActiveTokens() ([]FCMToken, error) {
	var tokens []FCMToken
	err := r.db.Where("is_active = ?", true).Find(&tokens).Error
	return tokens, err
}

func (r *fcmTokenRepository) DeleteByToken(token string) error {
	return r.db.Where("token = ?", token).Delete(&FCMToken{}).Error
}

func (r *fcmTokenRepository) Deactivate(id uint) error {
	return r.db.Model(&FCMToken{}).Where("id = ?", id).Update("is_active", false).Error
}

func (r *fcmTokenRepository) GetByToken(token string) (*FCMToken, error) {
	var fcmToken FCMToken
	err := r.db.Where("token = ?", token).First(&fcmToken).Error
	return &fcmToken, err
}

// FCM Message Repository Implementation
func (r *fcmMessageRepository) Create(message *FCMMessage) error {
	return r.db.Create(message).Error
}

func (r *fcmMessageRepository) GetByID(id uint) (*FCMMessage, error) {
	var message FCMMessage
	err := r.db.Preload("Recipients").First(&message, id).Error
	return &message, err
}

func (r *fcmMessageRepository) GetMessages(page, limit int) ([]FCMMessage, int64, error) {
	var messages []FCMMessage
	var total int64
	
	offset := (page - 1) * limit
	
	err := r.db.Model(&FCMMessage{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	
	err = r.db.Preload("Recipients").Order("created_at DESC").Offset(offset).Limit(limit).Find(&messages).Error
	return messages, total, err
}

func (r *fcmMessageRepository) GetMessagesByUser(userID uuid.UUID, page, limit int) ([]FCMMessage, int64, error) {
	var messages []FCMMessage
	var total int64
	
	offset := (page - 1) * limit
	
	subQuery := r.db.Model(&FCMMessageRecipient{}).Select("message_id").Where("user_id = ?", userID)
	
	err := r.db.Model(&FCMMessage{}).Where("id IN (?)", subQuery).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	
	err = r.db.Preload("Recipients").Where("id IN (?)", subQuery).Order("created_at DESC").Offset(offset).Limit(limit).Find(&messages).Error
	return messages, total, err
}

func (r *fcmMessageRepository) Update(id uint, message *FCMMessage) error {
	return r.db.Model(&FCMMessage{}).Where("id = ?", id).Updates(message).Error
}

func (r *fcmMessageRepository) Delete(id uint) error {
	return r.db.Delete(&FCMMessage{}, id).Error
}

// FCM Message Recipient Repository Implementation
func (r *fcmMessageRecipientRepository) Create(recipient *FCMMessageRecipient) error {
	return r.db.Create(recipient).Error
}

func (r *fcmMessageRecipientRepository) CreateBatch(recipients []FCMMessageRecipient) error {
	return r.db.CreateInBatches(recipients, 100).Error
}

func (r *fcmMessageRecipientRepository) GetByMessageID(messageID uint) ([]FCMMessageRecipient, error) {
	var recipients []FCMMessageRecipient
	err := r.db.Preload("Token").Where("message_id = ?", messageID).Find(&recipients).Error
	return recipients, err
}

func (r *fcmMessageRecipientRepository) UpdateStatus(id uint, status string, errorMsg string) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if errorMsg != "" {
		updates["error"] = errorMsg
	}
	return r.db.Model(&FCMMessageRecipient{}).Where("id = ?", id).Updates(updates).Error
}

func (r *fcmMessageRecipientRepository) MarkAsDelivered(id uint) error {
	return r.db.Model(&FCMMessageRecipient{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":       "delivered",
		"delivered_at": "NOW()",
	}).Error
}

func (r *fcmMessageRecipientRepository) GetStats() (map[string]int64, error) {
	stats := make(map[string]int64)
	
	// Count by status
	var results []struct {
		Status string
		Count  int64
	}
	
	err := r.db.Model(&FCMMessageRecipient{}).Select("status, COUNT(*) as count").Group("status").Scan(&results).Error
	if err != nil {
		return nil, err
	}
	
	for _, result := range results {
		stats[result.Status] = result.Count
	}
	
	return stats, nil
}
