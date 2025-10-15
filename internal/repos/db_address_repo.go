package repos

import (
	"fmt"
	"strconv"

	"errandShop/internal/core/types"
	"errandShop/internal/domain/customers"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DBAddressRepo is a database-backed implementation of AddressRepo
type DBAddressRepo struct {
	db *gorm.DB
}

// NewDBAddressRepo creates a new database address repository
func NewDBAddressRepo(db *gorm.DB) *DBAddressRepo {
	return &DBAddressRepo{db: db}
}

// GetByID retrieves an address by ID for a specific user
func (r *DBAddressRepo) GetByID(userID, addressID string) (*types.Address, error) {
	// Parse address ID to uint
	id, err := strconv.ParseUint(addressID, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid address ID: %v", err)
	}

	// Parse user ID to UUID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %v", err)
	}

	// Query the database
	var dbAddress customers.Address
	err = r.db.Where("id = ? AND user_id = ?", uint(id), userUUID).First(&dbAddress).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("address not found")
		}
		return nil, fmt.Errorf("database error: %v", err)
	}

	// Convert to types.Address format
	typesAddress := &types.Address{
		ID:     addressID,
		UserID: userID,
		Text:   formatAddressText(&dbAddress),
	}

	return typesAddress, nil
}

// GetByUserID retrieves all addresses for a specific user
func (r *DBAddressRepo) GetByUserID(userID string) ([]*types.Address, error) {
	// Parse user ID to UUID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %v", err)
	}

	// Query the database
	var dbAddresses []customers.Address
	err = r.db.Where("user_id = ?", userUUID).Find(&dbAddresses).Error
	if err != nil {
		return nil, fmt.Errorf("database error: %v", err)
	}

	// Convert to types.Address format
	var typesAddresses []*types.Address
	for _, dbAddr := range dbAddresses {
		typesAddr := &types.Address{
			ID:     strconv.FormatUint(uint64(dbAddr.ID), 10),
			UserID: userID,
			Text:   formatAddressText(&dbAddr),
		}
		typesAddresses = append(typesAddresses, typesAddr)
	}

	return typesAddresses, nil
}

// formatAddressText converts a database address to a text format suitable for delivery zone matching
func formatAddressText(addr *customers.Address) string {
	// Create a comprehensive address text that includes all relevant parts
	// This format should match what the delivery zone matcher expects
	text := addr.Street
	if addr.City != "" {
		if text != "" {
			text += ", "
		}
		text += addr.City
	}
	if addr.State != "" {
		if text != "" {
			text += ", "
		}
		text += addr.State
	}
	if addr.Country != "" {
		if text != "" {
			text += ", "
		}
		text += addr.Country
	}
	return text
}