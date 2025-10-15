package users

import (
	"errors"
	"gorm.io/gorm"
)

type repository struct {
	db *gorm.DB
}

// NewRepository creates a new user repository
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(user *User) error {
	return r.db.Create(user).Error
}

func (r *repository) GetByID(id uint) (*User, error) {
	var user User
	err := r.db.First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

func (r *repository) GetByEmail(email string) (*User, error) {
	var user User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

func (r *repository) Update(user *User) error {
	return r.db.Model(user).Select("name", "email", "phone", "gender", "avatar", "updated_at").Updates(user).Error
}

func (r *repository) Delete(id uint) error {
	return r.db.Delete(&User{}, id).Error
}

func (r *repository) List(offset, limit int) ([]User, int64, error) {
	var users []User
	var total int64
	
	// Count total records
	if err := r.db.Model(&User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// Get paginated results
	err := r.db.Offset(offset).Limit(limit).Find(&users).Error
	return users, total, err
}

func (r *repository) UpdatePermissions(userID uint, permissions []string) error {
	return r.db.Model(&User{}).Where("id = ?", userID).Update("permissions", permissions).Error
}