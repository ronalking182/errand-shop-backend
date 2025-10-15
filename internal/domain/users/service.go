package users

import (
	"golang.org/x/crypto/bcrypt"
)

// Service represents the user service
type Service struct {
	repo Repository
}

// NewService creates a new user service
func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

// hashedPassword hashes a password using bcrypt
func hashedPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashed), err
}

// CreateUser creates a new user
func (s *Service) CreateUser(req CreateUserRequest) (*User, error) {
	// Hash the password
	hashed, err := hashedPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// Set default permissions if none provided
	permissions := req.Permissions
	if len(permissions) == 0 {
		permissions = GetDefaultPermissions(req.Role)
	}

	phone := req.Phone
	user := &User{
		Name:        req.Name,
		Email:       req.Email,
		Phone:       &phone,
		Password:    hashed,
		Role:        req.Role,
		Permissions: permissions,
		Status:      "active",
		IsVerified:  false,
	}

	if err := s.repo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

// UpdateUserPermissions updates user permissions
func (s *Service) UpdateUserPermissions(userID uint, permissions []string) error {
	return s.repo.UpdatePermissions(userID, permissions)
}

// ToggleUserPermission toggles a specific permission for a user
func (s *Service) ToggleUserPermission(userID uint, permission string, grant bool) error {
	user, err := s.repo.GetByID(userID)
	if err != nil {
		return err
	}

	// Remove permission if it exists
	newPermissions := []string{}
	for _, perm := range user.Permissions {
		if perm != permission {
			newPermissions = append(newPermissions, perm)
		}
	}

	// Add permission if granting
	if grant {
		newPermissions = append(newPermissions, permission)
	}

	return s.repo.UpdatePermissions(userID, newPermissions)
}
