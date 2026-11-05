package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/suprimkhatri77/turgorepo/api/internal/constants"
)

func TestRequireRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		role string
		want int
	}{
		{constants.RoleAdmin, http.StatusOK},
		{constants.RoleSuperAdmin, http.StatusOK},
		{constants.RoleMember, http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.role, func(t *testing.T) {
			r := gin.New()
			r.GET("/",
				func(c *gin.Context) { c.Set(constants.RoleKey, tt.role) },
				RequireRole(constants.RoleAdmin, constants.RoleSuperAdmin),
				func(c *gin.Context) { c.Status(http.StatusOK) },
			)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
			if w.Code != tt.want {
				t.Errorf("status = %d, want %d", w.Code, tt.want)
			}
		})
	}
}
