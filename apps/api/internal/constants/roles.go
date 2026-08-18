package constants

const (
	RoleMember     = "member"
	RoleAdmin      = "admin"
	RoleSuperAdmin = "superadmin"
)

// ValidRoles are roles allowed in access-token claims.
var ValidRoles = []string{RoleMember, RoleAdmin, RoleSuperAdmin}

func IsValidRole(role string) bool {
	for _, r := range ValidRoles {
		if role == r {
			return true
		}
	}
	return false
}
