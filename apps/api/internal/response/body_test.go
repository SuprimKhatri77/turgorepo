package response

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func run(write func(c *gin.Context)) (*httptest.ResponseRecorder, *gin.Context) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	write(c)
	return w, c
}

func TestHelpers(t *testing.T) {
	t.Run("OK leaves out code and empty data", func(t *testing.T) {
		w, _ := run(func(c *gin.Context) { OK(c, http.StatusCreated, "Created", nil) })
		assert.Equal(t, http.StatusCreated, w.Code)
		assert.JSONEq(t, `{"success":true,"message":"Created"}`, w.Body.String())
	})

	t.Run("OK includes data", func(t *testing.T) {
		w, _ := run(func(c *gin.Context) { OK(c, http.StatusOK, "Found", map[string]string{"id": "1"}) })
		assert.JSONEq(t, `{"success":true,"message":"Found","data":{"id":"1"}}`, w.Body.String())
	})

	t.Run("Error sets the code", func(t *testing.T) {
		w, c := run(func(c *gin.Context) { Error(c, http.StatusConflict, "USER_ALREADY_EXISTS", "User already exists") })
		assert.Equal(t, http.StatusConflict, w.Code)
		assert.JSONEq(t, `{"success":false,"message":"User already exists","code":"USER_ALREADY_EXISTS"}`, w.Body.String())
		assert.False(t, c.IsAborted())
	})

	t.Run("Abort also stops the chain", func(t *testing.T) {
		w, c := run(func(c *gin.Context) { Abort(c, http.StatusUnauthorized, "TOKEN_INVALID", "Invalid access token") })
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.True(t, c.IsAborted())
	})

	t.Run("Internal hides the cause", func(t *testing.T) {
		w, _ := run(Internal)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.JSONEq(t, `{"success":false,"message":"Something went wrong","code":"INTERNAL_SERVER_ERROR"}`, w.Body.String())
	})
}
