package chat

import (
	"time"

	"gorm.io/gorm"
)

// ChatRoomRepository interface
type ChatRoomRepository interface {
	Create(room *ChatRoom) error
	GetByID(id uint) (*ChatRoom, error)
	GetByCustomerID(customerID uint, page, limit int) ([]ChatRoom, int64, error)
	GetAll(page, limit int) ([]ChatRoom, int64, error)
	GetByStatus(status ChatStatus, page, limit int) ([]ChatRoom, int64, error)
	Update(id uint, updates map[string]interface{}) error
	Delete(id uint) error
	GetStats() (*ChatStatsResponse, error)
	AssignAdmin(roomID, adminID uint) error
	UnassignAdmin(roomID uint) error
}

// ChatMessageRepository interface
type ChatMessageRepository interface {
	Create(message *ChatMessage) error
	GetByID(id uint) (*ChatMessage, error)
	GetByRoomID(roomID uint, page, limit int) ([]ChatMessage, int64, error)
	MarkAsRead(messageID uint) error
	MarkRoomMessagesAsRead(roomID, userID uint, userType SenderType) error
	GetUnreadCount(roomID, userID uint, userType SenderType) (int64, error)
	GetLastMessage(roomID uint) (*ChatMessage, error)
	Delete(id uint) error
	GetTodayMessageCount() (int64, error)
}

// Repository implementations
type chatRoomRepository struct {
	db *gorm.DB
}

type chatMessageRepository struct {
	db *gorm.DB
}

// Constructor functions
func NewChatRoomRepository(db *gorm.DB) ChatRoomRepository {
	return &chatRoomRepository{db: db}
}

func NewChatMessageRepository(db *gorm.DB) ChatMessageRepository {
	return &chatMessageRepository{db: db}
}

// ChatRoomRepository implementation
func (r *chatRoomRepository) Create(room *ChatRoom) error {
	return r.db.Create(room).Error
}

func (r *chatRoomRepository) GetByID(id uint) (*ChatRoom, error) {
	var room ChatRoom
	err := r.db.Preload("Messages", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at DESC").Limit(1)
	}).First(&room, id).Error
	return &room, err
}

func (r *chatRoomRepository) GetByCustomerID(customerID uint, page, limit int) ([]ChatRoom, int64, error) {
	var rooms []ChatRoom
	var total int64

	query := r.db.Model(&ChatRoom{}).Where("customer_id = ?", customerID)
	query.Count(&total)

	offset := (page - 1) * limit
	err := query.Preload("Messages", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at DESC").Limit(1)
	}).Order("updated_at DESC").Offset(offset).Limit(limit).Find(&rooms).Error

	return rooms, total, err
}

func (r *chatRoomRepository) GetAll(page, limit int) ([]ChatRoom, int64, error) {
	var rooms []ChatRoom
	var total int64

	query := r.db.Model(&ChatRoom{})
	query.Count(&total)

	offset := (page - 1) * limit
	err := query.Preload("Messages", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at DESC").Limit(1)
	}).Order("updated_at DESC").Offset(offset).Limit(limit).Find(&rooms).Error

	return rooms, total, err
}

func (r *chatRoomRepository) GetByStatus(status ChatStatus, page, limit int) ([]ChatRoom, int64, error) {
	var rooms []ChatRoom
	var total int64

	query := r.db.Model(&ChatRoom{}).Where("status = ?", status)
	query.Count(&total)

	offset := (page - 1) * limit
	err := query.Preload("Messages", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at DESC").Limit(1)
	}).Order("updated_at DESC").Offset(offset).Limit(limit).Find(&rooms).Error

	return rooms, total, err
}

func (r *chatRoomRepository) Update(id uint, updates map[string]interface{}) error {
	return r.db.Model(&ChatRoom{}).Where("id = ?", id).Updates(updates).Error
}

func (r *chatRoomRepository) Delete(id uint) error {
	return r.db.Delete(&ChatRoom{}, id).Error
}

func (r *chatRoomRepository) GetStats() (*ChatStatsResponse, error) {
	stats := &ChatStatsResponse{}

	// Total rooms
	r.db.Model(&ChatRoom{}).Count(&stats.TotalRooms)

	// Active rooms
	r.db.Model(&ChatRoom{}).Where("status = ?", ChatStatusActive).Count(&stats.ActiveRooms)

	// Closed rooms
	r.db.Model(&ChatRoom{}).Where("status = ?", ChatStatusClosed).Count(&stats.ClosedRooms)

	// Unread messages (messages not read by admins)
	r.db.Model(&ChatMessage{}).Where("sender_type = ? AND is_read = ?", SenderTypeCustomer, false).Count(&stats.UnreadMessages)

	// Today's messages
	today := time.Now().Truncate(24 * time.Hour)
	r.db.Model(&ChatMessage{}).Where("created_at >= ?", today).Count(&stats.TodayMessages)

	return stats, nil
}

func (r *chatRoomRepository) AssignAdmin(roomID, adminID uint) error {
	return r.db.Model(&ChatRoom{}).Where("id = ?", roomID).Update("admin_id", adminID).Error
}

func (r *chatRoomRepository) UnassignAdmin(roomID uint) error {
	return r.db.Model(&ChatRoom{}).Where("id = ?", roomID).Update("admin_id", nil).Error
}

// ChatMessageRepository implementation
func (r *chatMessageRepository) Create(message *ChatMessage) error {
	return r.db.Create(message).Error
}

func (r *chatMessageRepository) GetByID(id uint) (*ChatMessage, error) {
	var message ChatMessage
	err := r.db.Preload("Room").First(&message, id).Error
	return &message, err
}

func (r *chatMessageRepository) GetByRoomID(roomID uint, page, limit int) ([]ChatMessage, int64, error) {
	var messages []ChatMessage
	var total int64

	query := r.db.Model(&ChatMessage{}).Where("room_id = ?", roomID)
	query.Count(&total)

	offset := (page - 1) * limit
	err := query.Order("created_at ASC").Offset(offset).Limit(limit).Find(&messages).Error

	return messages, total, err
}

func (r *chatMessageRepository) MarkAsRead(messageID uint) error {
	now := time.Now()
	return r.db.Model(&ChatMessage{}).Where("id = ?", messageID).Updates(map[string]interface{}{
		"is_read": true,
		"read_at": &now,
	}).Error
}

func (r *chatMessageRepository) MarkRoomMessagesAsRead(roomID, userID uint, userType SenderType) error {
	now := time.Now()
	// Mark messages as read that were NOT sent by the current user
	return r.db.Model(&ChatMessage{}).Where(
		"room_id = ? AND (sender_id != ? OR sender_type != ?) AND is_read = ?",
		roomID, userID, userType, false,
	).Updates(map[string]interface{}{
		"is_read": true,
		"read_at": &now,
	}).Error
}

func (r *chatMessageRepository) GetUnreadCount(roomID, userID uint, userType SenderType) (int64, error) {
	var count int64
	// Count messages that were NOT sent by the current user and are unread
	err := r.db.Model(&ChatMessage{}).Where(
		"room_id = ? AND (sender_id != ? OR sender_type != ?) AND is_read = ?",
		roomID, userID, userType, false,
	).Count(&count).Error
	return count, err
}

func (r *chatMessageRepository) GetLastMessage(roomID uint) (*ChatMessage, error) {
	var message ChatMessage
	err := r.db.Where("room_id = ?", roomID).Order("created_at DESC").First(&message).Error
	if err != nil {
		return nil, err
	}
	return &message, nil
}

func (r *chatMessageRepository) Delete(id uint) error {
	return r.db.Delete(&ChatMessage{}, id).Error
}

func (r *chatMessageRepository) GetTodayMessageCount() (int64, error) {
	var count int64
	today := time.Now().Truncate(24 * time.Hour)
	err := r.db.Model(&ChatMessage{}).Where("created_at >= ?", today).Count(&count).Error
	return count, err
}