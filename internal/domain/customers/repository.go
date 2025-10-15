package customers

import (
	"errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Create(customer *Customer) error
	GetByID(id uint) (*Customer, error)
	GetByUserID(userID uuid.UUID) (*Customer, error)
	Update(customer *Customer) error
	Delete(id uint) error
	List(limit, offset int) ([]Customer, error)

	// Address methods
	CreateAddress(address *Address) error
	GetAddressesByCustomerID(customerID uint) ([]Address, error)
	GetAddressByID(id uint) (*Address, error)
	UpdateAddress(address *Address) error
	DeleteAddress(id uint) error
	SetDefaultAddress(customerID, addressID uint) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(customer *Customer) error {
	return r.db.Create(customer).Error
}

func (r *repository) GetByID(id uint) (*Customer, error) {
	var customer Customer
	err := r.db.Preload("Addresses").First(&customer, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("customer not found")
		}
		return nil, err
	}
	return &customer, nil
}

func (r *repository) GetByUserID(userID uuid.UUID) (*Customer, error) {
	var customer Customer
	err := r.db.Preload("Addresses").Where("user_id = ?", userID).First(&customer).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("customer not found")
		}
		return nil, err
	}
	return &customer, nil
}

func (r *repository) Update(customer *Customer) error {
	return r.db.Save(customer).Error
}

func (r *repository) Delete(id uint) error {
	return r.db.Delete(&Customer{}, id).Error
}

func (r *repository) List(limit, offset int) ([]Customer, error) {
	var customers []Customer
	err := r.db.Preload("Addresses").Limit(limit).Offset(offset).Find(&customers).Error
	return customers, err
}

// Address methods
func (r *repository) CreateAddress(address *Address) error {
	return r.db.Create(address).Error
}

func (r *repository) GetAddressesByCustomerID(customerID uint) ([]Address, error) {
	var addresses []Address
	err := r.db.Where("customer_id = ?", customerID).Find(&addresses).Error
	return addresses, err
}

func (r *repository) GetAddressByID(id uint) (*Address, error) {
	var address Address
	err := r.db.First(&address, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("address not found")
		}
		return nil, err
	}
	return &address, nil
}

func (r *repository) UpdateAddress(address *Address) error {
	return r.db.Save(address).Error
}

func (r *repository) DeleteAddress(id uint) error {
	return r.db.Delete(&Address{}, id).Error
}

func (r *repository) SetDefaultAddress(customerID, addressID uint) error {
	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Remove default from all addresses
	if err := tx.Model(&Address{}).Where("customer_id = ?", customerID).Update("is_default", false).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Set new default
	if err := tx.Model(&Address{}).Where("id = ? AND customer_id = ?", addressID, customerID).Update("is_default", true).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
