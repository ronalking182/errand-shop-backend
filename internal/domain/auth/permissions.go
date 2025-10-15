package auth

type Permission string

const (
	// User permissions
	PermissionReadProfile   Permission = "user:read:profile"
	PermissionUpdateProfile Permission = "user:update:profile"
	PermissionDeleteAccount Permission = "user:delete:account"

	// Order permissions
	PermissionCreateOrder Permission = "order:create"
	PermissionReadOrder   Permission = "order:read"
	PermissionUpdateOrder Permission = "order:update"
	PermissionCancelOrder Permission = "order:cancel"

	// Admin permissions
	PermissionManageUsers    Permission = "admin:manage:users"
	PermissionManageOrders   Permission = "admin:manage:orders"
	PermissionManageProducts Permission = "admin:manage:products"
	PermissionViewAnalytics  Permission = "admin:view:analytics"

	// Super admin
	PermissionAll Permission = "*"
)

var RolePermissions = map[string][]Permission{
	"customer": {
		PermissionReadProfile,
		PermissionUpdateProfile,
		PermissionCreateOrder,
		PermissionReadOrder,
		PermissionCancelOrder,
	},
	"admin": {
		PermissionReadProfile,
		PermissionUpdateProfile,
		PermissionManageUsers,
		PermissionManageOrders,
		PermissionManageProducts,
		PermissionViewAnalytics,
	},
	"superadmin": {
		PermissionAll,
	},
}

func GetUserPermissions(role string) []string {
	perms := RolePermissions[role]
	result := make([]string, len(perms))
	for i, perm := range perms {
		result[i] = string(perm)
	}
	return result
}
