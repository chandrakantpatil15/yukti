package models

// AdminRole represents platform admin roles
type AdminRole string

const (
	AdminRoleSuperAdmin AdminRole = "super_admin"
	AdminRoleSupport    AdminRole = "support"
	AdminRoleAnalyst    AdminRole = "analyst"
)

// AdminPermission represents admin-level permissions
type AdminPermission string

const (
	// Tenant Management
	PermAdminViewTenants    AdminPermission = "admin_view_tenants"
	PermAdminManageTenants  AdminPermission = "admin_manage_tenants"
	PermAdminSuspendTenants AdminPermission = "admin_suspend_tenants"
	PermAdminDeleteTenants  AdminPermission = "admin_delete_tenants"

	// User Management
	PermAdminViewUsers   AdminPermission = "admin_view_users"
	PermAdminManageUsers AdminPermission = "admin_manage_users"

	// Impersonation
	PermAdminImpersonate AdminPermission = "admin_impersonate"

	// Analytics
	PermAdminViewAnalytics AdminPermission = "admin_view_analytics"

	// System
	PermAdminViewAuditLogs AdminPermission = "admin_view_audit_logs"
	PermAdminSystemConfig  AdminPermission = "admin_system_config"
)

// AdminRolePermissions maps admin roles to permissions
var AdminRolePermissions = map[AdminRole][]AdminPermission{
	AdminRoleSuperAdmin: {
		PermAdminViewTenants, PermAdminManageTenants, PermAdminSuspendTenants, PermAdminDeleteTenants,
		PermAdminViewUsers, PermAdminManageUsers,
		PermAdminImpersonate,
		PermAdminViewAnalytics,
		PermAdminViewAuditLogs, PermAdminSystemConfig,
	},
	AdminRoleSupport: {
		PermAdminViewTenants, PermAdminManageTenants,
		PermAdminViewUsers, PermAdminManageUsers,
		PermAdminImpersonate,
		PermAdminViewAuditLogs,
	},
	AdminRoleAnalyst: {
		PermAdminViewTenants,
		PermAdminViewUsers,
		PermAdminViewAnalytics,
		PermAdminViewAuditLogs,
	},
}

// HasAdminPermission checks if admin role has permission
func HasAdminPermission(role AdminRole, permission AdminPermission) bool {
	permissions, exists := AdminRolePermissions[role]
	if !exists {
		return false
	}
	for _, p := range permissions {
		if p == permission {
			return true
		}
	}
	return false
}
