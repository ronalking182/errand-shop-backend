package repos

import (
	"errors"
	"sync"

	"errandShop/internal/core/types"
)

// AddressRepo defines the interface for address repository
type AddressRepo interface {
	GetByID(userID, addressID string) (*types.Address, error)
	GetByUserID(userID string) ([]*types.Address, error)
}

// InMemoryAddressRepo is an in-memory implementation of AddressRepo
type InMemoryAddressRepo struct {
	addresses map[string]*types.Address
	mu        sync.RWMutex
}

// NewInMemoryAddressRepo creates a new in-memory address repository
func NewInMemoryAddressRepo() *InMemoryAddressRepo {
	repo := &InMemoryAddressRepo{
		addresses: make(map[string]*types.Address),
	}
	
	// Seed with test addresses for user U-TEST
	testAddresses := []*types.Address{
		{
			ID:     "ADDR-1",
			UserID: "U-TEST",
			Text:   "No. 4, Wuse Zone II, Abuja",
		},
		{
			ID:     "ADDR-2",
			UserID: "U-TEST",
			Text:   "Block 8, Gwarinpa 6th Ave, Abuja",
		},
		{
			ID:     "ADDR-3",
			UserID: "U-TEST",
			Text:   "River Park Lugbe Estate, FCT",
		},
		{
			ID:     "ADDR-4",
			UserID: "U-TEST",
			Text:   "House 2, Maitama Abuja",
		},
	}
	
	for _, addr := range testAddresses {
		repo.addresses[addr.ID] = addr
	}
	
	return repo
}

// GetByID retrieves an address by ID for a specific user
func (r *InMemoryAddressRepo) GetByID(userID, addressID string) (*types.Address, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	addr, exists := r.addresses[addressID]
	if !exists {
		return nil, errors.New("address not found")
	}
	
	// Ensure the address belongs to the user
	if addr.UserID != userID {
		return nil, errors.New("address not found")
	}
	
	return addr, nil
}

// GetByUserID retrieves all addresses for a specific user
func (r *InMemoryAddressRepo) GetByUserID(userID string) ([]*types.Address, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var userAddresses []*types.Address
	for _, addr := range r.addresses {
		if addr.UserID == userID {
			userAddresses = append(userAddresses, addr)
		}
	}
	
	return userAddresses, nil
}