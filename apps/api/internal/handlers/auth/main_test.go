package auth

import (
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/suprimkhatri77/turgorepo/api/internal/validator"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	validator.Init()
	os.Exit(m.Run())
}
