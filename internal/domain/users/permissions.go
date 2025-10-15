package users

// Permission constants matching frontend expectations
const (
	PermProductsRead   = "products:read"
	PermProductsWrite  = "products:write"
	PermProductsDelete = "products:delete"
	PermOrdersRead     = "orders:read"
	PermOrdersWrite    = "orders:write"
	PermOrdersCancel   = "orders:cancel"
	PermChatRead       = "chat:read"
	PermChatWrite      = "chat:write"
	PermCouponsRead    = "coupons:read"
	PermCouponsCreate  = "coupons:create"
	PermReportsRead    = "reports:read"
)

// GetAvailablePermissions returns all available permissions
func GetAvailablePermissions() []map[string]string {
	return []map[string]string{
		{"id": PermProductsRead, "label": "View Products"},
		{"id": PermProductsWrite, "label": "Edit Products"},
		{"id": PermProductsDelete, "label": "Delete Products"},
		{"id": PermOrdersRead, "label": "View Orders"},
		{"id": PermOrdersWrite, "label": "Edit Orders"},
		{"id": PermOrdersCancel, "label": "Cancel Orders"},
		{"id": PermChatRead, "label": "View Chat"},
		{"id": PermChatWrite, "label": "Send Messages"},
		{"id": PermCouponsRead, "label": "View Coupons"},
		{"id": PermCouponsCreate, "label": "Create Coupons"},
		{"id": PermReportsRead, "label": "View Reports"},
	}
}

// GetDefaultPermissions returns default permissions based on role
func GetDefaultPermissions(role string) []string {
	switch role {
	case "superadmin":
		return []string{
			PermProductsRead, PermProductsWrite, PermProductsDelete,
			PermOrdersRead, PermOrdersWrite, PermOrdersCancel,
			PermChatRead, PermChatWrite,
			PermCouponsRead, PermCouponsCreate,
			PermReportsRead,
		}
	case "admin":
		return []string{
			PermProductsRead, PermProductsWrite,
			PermOrdersRead, PermOrdersWrite,
			PermChatRead, PermChatWrite,
			PermCouponsRead,
		}
	default:
		return []string{}
	}
}

// HasPermission checks if user has a specific permission
func (u *User) HasPermission(permission string) bool {
	// Superadmin has all permissions
	if u.Role == "superadmin" {
		return true
	}

	for _, perm := range u.Permissions {
		if perm == permission {
			return true
		}
	}
	return false
}
