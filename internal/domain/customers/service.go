package customers

import (
	"errors"
	"log"
	"github.com/google/uuid"
)

type Service interface {
	CreateCustomer(req interface{}) (interface{}, error)
	GetCustomerByID(id uint) (*CustomerResponse, error)
	GetCustomerByUserID(userID uuid.UUID) (*CustomerResponse, error)
	UpdateCustomer(id uint, req *UpdateCustomerRequest) (*CustomerResponse, error)
	DeleteCustomer(id uint) error
	ListCustomers(limit, offset int) ([]CustomerResponse, error)

	// Address methods
	CreateAddress(customerID uint, req *CreateAddressRequest) (*AddressResponse, error)
	GetCustomerAddresses(customerID uint) ([]AddressResponse, error)
	UpdateAddress(customerID, addressID uint, req *CreateAddressRequest) (*AddressResponse, error)
	DeleteAddress(customerID, addressID uint) error
	SetDefaultAddress(customerID, addressID uint) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) CreateCustomer(req interface{}) (interface{}, error) {
	// Handle both *CreateCustomerRequest and map[string]interface{} types
	var customerReq *CreateCustomerRequest
	
	switch v := req.(type) {
	case *CreateCustomerRequest:
		customerReq = v
	case map[string]interface{}:
		// Convert map to struct for auth service calls
		userID, ok := v["user_id"].(uuid.UUID)
		if !ok {
			return nil, errors.New("invalid user_id type")
		}
		customerReq = &CreateCustomerRequest{
			UserID:    userID,
			FirstName: v["first_name"].(string),
			LastName:  v["last_name"].(string),
			Phone:     v["phone"].(string),
		}
	default:
		return nil, errors.New("invalid request type")
	}

	return s.createCustomerInternal(customerReq)
}

// Internal method for actual customer creation
func (s *service) createCustomerInternal(req *CreateCustomerRequest) (*CustomerResponse, error) {
	// Check if customer already exists for this user
	existing, _ := s.repo.GetByUserID(req.UserID)
	if existing != nil {
		return nil, errors.New("customer profile already exists for this user")
	}

	customer := &Customer{
		UserID:      req.UserID,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		Phone:       req.Phone,
		DateOfBirth: req.DateOfBirth,
		Gender:      req.Gender,
		Status:      CustomerStatusActive,
	}

	if err := s.repo.Create(customer); err != nil {
		log.Printf("Error creating customer: %v", err)
		return nil, errors.New("failed to create customer")
	}

	return s.toCustomerResponse(customer), nil
}

func (s *service) GetCustomerByID(id uint) (*CustomerResponse, error) {
	customer, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return s.toCustomerResponse(customer), nil
}

func (s *service) GetCustomerByUserID(userID uuid.UUID) (*CustomerResponse, error) {
	customer, err := s.repo.GetByUserID(userID)
	if err != nil {
		// If customer not found, create a basic customer profile automatically
		if errors.Is(err, errors.New("customer not found")) || err.Error() == "customer not found" {
			log.Printf("Customer not found for user %d, creating basic profile", userID)
			
			// Create a basic customer profile
			newCustomer := &Customer{
				UserID:    userID,
				FirstName: "", // Will be updated when user provides info
				LastName:  "",
				Phone:     "",
				Status:    CustomerStatusActive,
			}
			
			if createErr := s.repo.Create(newCustomer); createErr != nil {
				log.Printf("Error creating customer profile for user %d: %v", userID, createErr)
				return nil, errors.New("failed to create customer profile")
			}
			
			log.Printf("Successfully created customer profile for user %d", userID)
			return s.toCustomerResponse(newCustomer), nil
		}
		return nil, err
	}
	return s.toCustomerResponse(customer), nil
}

func (s *service) UpdateCustomer(id uint, req *UpdateCustomerRequest) (*CustomerResponse, error) {
	customer, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.FirstName != nil {
		customer.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		customer.LastName = *req.LastName
	}
	if req.Phone != nil {
		customer.Phone = *req.Phone
	}
	if req.DateOfBirth != nil {
		customer.DateOfBirth = req.DateOfBirth
	}
	if req.Gender != nil {
		customer.Gender = *req.Gender
	}
	if req.Avatar != nil {
		customer.Avatar = *req.Avatar
	}

	if err := s.repo.Update(customer); err != nil {
		log.Printf("Error updating customer: %v", err)
		return nil, errors.New("failed to update customer")
	}

	return s.toCustomerResponse(customer), nil
}

func (s *service) DeleteCustomer(id uint) error {
	return s.repo.Delete(id)
}

func (s *service) ListCustomers(limit, offset int) ([]CustomerResponse, error) {
	customers, err := s.repo.List(limit, offset)
	if err != nil {
		return nil, err
	}

	responses := make([]CustomerResponse, len(customers))
	for i, customer := range customers {
		responses[i] = *s.toCustomerResponse(&customer)
	}

	return responses, nil
}

// Address methods
func (s *service) CreateAddress(customerID uint, req *CreateAddressRequest) (*AddressResponse, error) {
	// Verify customer exists and get customer info
	customer, err := s.repo.GetByID(customerID)
	if err != nil {
		return nil, err
	}

	address := &Address{
		CustomerID: customer.ID,
		UserID:     customer.UserID,
		Label:      req.Label,
		Type:       req.Type,
		Street:     req.Street,
		City:       req.City,
		State:      req.State,
		Country:    req.Country,
		PostalCode: req.PostalCode,
		ZipCode:    req.PostalCode,
		IsDefault:  req.IsDefault,
	}

	if err := s.repo.CreateAddress(address); err != nil {
		log.Printf("Error creating address: %v", err)
		return nil, errors.New("failed to create address")
	}

	// If this is set as default, update other addresses
	if req.IsDefault {
		if err := s.repo.SetDefaultAddress(customerID, address.ID); err != nil {
			log.Printf("Error setting default address: %v", err)
		}
	}

	return s.toAddressResponse(address), nil
}

func (s *service) GetCustomerAddresses(customerID uint) ([]AddressResponse, error) {
	addresses, err := s.repo.GetAddressesByCustomerID(customerID)
	if err != nil {
		return nil, err
	}

	responses := make([]AddressResponse, len(addresses))
	for i, address := range addresses {
		responses[i] = *s.toAddressResponse(&address)
	}

	return responses, nil
}

func (s *service) UpdateAddress(customerID, addressID uint, req *CreateAddressRequest) (*AddressResponse, error) {
	address, err := s.repo.GetAddressByID(addressID)
	if err != nil {
		return nil, err
	}

	// Verify address belongs to customer
	if address.CustomerID != customerID {
		return nil, errors.New("address not found")
	}

	address.Type = req.Type
	address.Street = req.Street
	address.City = req.City
	address.State = req.State
	address.Country = req.Country
	address.PostalCode = req.PostalCode
	address.IsDefault = req.IsDefault

	if err := s.repo.UpdateAddress(address); err != nil {
		log.Printf("Error updating address: %v", err)
		return nil, errors.New("failed to update address")
	}

	// If this is set as default, update other addresses
	if req.IsDefault {
		if err := s.repo.SetDefaultAddress(customerID, address.ID); err != nil {
			log.Printf("Error setting default address: %v", err)
		}
	}

	return s.toAddressResponse(address), nil
}

func (s *service) DeleteAddress(customerID, addressID uint) error {
	address, err := s.repo.GetAddressByID(addressID)
	if err != nil {
		return err
	}

	// Verify address belongs to customer
	if address.CustomerID != customerID {
		return errors.New("address not found")
	}

	return s.repo.DeleteAddress(addressID)
}

func (s *service) SetDefaultAddress(customerID, addressID uint) error {
	address, err := s.repo.GetAddressByID(addressID)
	if err != nil {
		return err
	}

	// Verify address belongs to customer
	if address.CustomerID != customerID {
		return errors.New("address not found")
	}

	return s.repo.SetDefaultAddress(customerID, addressID)
}

// Helper methods
func (s *service) toCustomerResponse(customer *Customer) *CustomerResponse {
	addresses := make([]AddressResponse, len(customer.Addresses))
	for i, addr := range customer.Addresses {
		addresses[i] = *s.toAddressResponse(&addr)
	}

	return &CustomerResponse{
		ID:          customer.ID,
		UserID:      customer.UserID,
		FirstName:   customer.FirstName,
		LastName:    customer.LastName,
		Phone:       customer.Phone,
		DateOfBirth: customer.DateOfBirth,
		Gender:      customer.Gender,
		Avatar:      customer.Avatar,
		Status:      customer.Status,
		Addresses:   addresses,
		CreatedAt:   customer.CreatedAt,
		UpdatedAt:   customer.UpdatedAt,
	}
}

func (s *service) toAddressResponse(address *Address) *AddressResponse {
	return &AddressResponse{
		ID:         address.ID,
		UserID:     address.UserID,
		Label:      address.Label,
		Type:       address.Type,
		Street:     address.Street,
		City:       address.City,
		State:      address.State,
		Country:    address.Country,
		PostalCode: address.PostalCode,
		IsDefault:  address.IsDefault,
		CreatedAt:  address.CreatedAt,
		UpdatedAt:  address.UpdatedAt,
	}
}
