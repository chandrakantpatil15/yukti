package models

// Role represents user roles in the system
type Role string

const (
	RoleOwner  Role = "owner"
	RoleAdmin  Role = "admin"
	RoleEditor Role = "editor"
	RoleViewer Role = "viewer"
)

// Permission represents specific permissions
type Permission string

const (
	// AWS Management
	PermViewAWS       Permission = "view_aws"
	PermManageAWS     Permission = "manage_aws"
	PermScanResources Permission = "scan_resources"

	// Findings
	PermViewFindings   Permission = "view_findings"
	PermManageFindings Permission = "manage_findings"

	// Whitelists
	PermViewWhitelists   Permission = "view_whitelists"
	PermManageWhitelists Permission = "manage_whitelists"

	// Budgets
	PermViewBudgets   Permission = "view_budgets"
	PermManageBudgets Permission = "manage_budgets"

	// IaC
	PermGenerateIaC Permission = "generate_iac"

	// Team
	PermViewTeam   Permission = "view_team"
	PermManageTeam Permission = "manage_team"

	// Billing
	PermViewBilling   Permission = "view_billing"
	PermManageBilling Permission = "manage_billing"
)

// RolePermissions maps roles to their permissions
var RolePermissions = map[Role][]Permission{
	RoleOwner: {
		PermViewAWS, PermManageAWS, PermScanResources,
		PermViewFindings, PermManageFindings,
		PermViewWhitelists, PermManageWhitelists,
		PermViewBudgets, PermManageBudgets,
		PermGenerateIaC,
		PermViewTeam, PermManageTeam,
		PermViewBilling, PermManageBilling,
	},
	RoleAdmin: {
		PermViewAWS, PermManageAWS, PermScanResources,
		PermViewFindings, PermManageFindings,
		PermViewWhitelists, PermManageWhitelists,
		PermViewBudgets, PermManageBudgets,
		PermGenerateIaC,
		PermViewTeam, PermManageTeam,
		PermViewBilling, // Can view but not manage billing
	},
	RoleEditor: {
		PermViewAWS, PermScanResources,
		PermViewFindings, PermManageFindings,
		PermViewWhitelists, PermManageWhitelists,
		PermViewBudgets,
		PermGenerateIaC,
		PermViewTeam,
	},
	RoleViewer: {
		PermViewAWS,
		PermViewFindings,
		PermViewWhitelists,
		PermViewBudgets,
		PermViewTeam,
	},
}

// HasPermission checks if a role has a specific permission
func HasPermission(role Role, permission Permission) bool {
	permissions, exists := RolePermissions[role]
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

// CanManageUser checks if a role can manage another role
func CanManageUser(managerRole, targetRole Role) bool {
	// Owner can manage everyone except other owners
	if managerRole == RoleOwner && targetRole != RoleOwner {
		return true
	}
	// Admin can manage editors and viewers
	if managerRole == RoleAdmin && (targetRole == RoleEditor || targetRole == RoleViewer) {
		return true
	}
	return false
}
