package constants

import "testing"

func TestIsValidRole(t *testing.T) {
	tests := []struct {
		role string
		want bool
	}{
		{RoleMember, true},
		{RoleAdmin, true},
		{RoleSuperAdmin, true},
		{"", false},
		{"guest", false},
		{"Admin", false},
	}

	for _, tt := range tests {
		if got := IsValidRole(tt.role); got != tt.want {
			t.Fatalf("IsValidRole(%q) = %v, want %v", tt.role, got, tt.want)
		}
	}
}
