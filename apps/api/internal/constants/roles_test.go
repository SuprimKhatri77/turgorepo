package constants

import (
	"testing"

	"github.com/stretchr/testify/require"
)

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
		require.Equal(t, tt.want, IsValidRole(tt.role), "IsValidRole(%q)", tt.role)
	}
}
