package validator

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/suprimkhatri77/turgorepo/api/internal/constants"
	"github.com/suprimkhatri77/turgorepo/api/internal/types"
)

type bindRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Name     string `json:"name" binding:"required"`
	Password string `json:"password" trim:"-" binding:"required"`
	Count    int    `json:"count"`
}

// bind runs BindJSON on body and returns what it bound and the response.
func bind(t *testing.T, body string) (*bindRequest, *httptest.ResponseRecorder) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))
	req, ok := BindJSON[bindRequest](c)
	assert.Equal(t, ok, req != nil)
	return req, w
}

func TestBindJSON(t *testing.T) {
	t.Run("trims strings before validating them", func(t *testing.T) {
		req, w := bind(t, `{"email":"  ada@example.com ","name":" Ada ","password":" secret ","count":3}`)

		require.NotNil(t, req, w.Body.String())
		assert.Equal(t, "ada@example.com", req.Email)
		assert.Equal(t, "Ada", req.Name)
		assert.Equal(t, " secret ", req.Password, "trim:\"-\" fields keep their spaces")
		assert.Equal(t, 3, req.Count)
	})

	t.Run("a value of only spaces counts as missing", func(t *testing.T) {
		req, w := bind(t, `{"email":"ada@example.com","name":"   ","password":"secret"}`)

		assert.Nil(t, req)
		require.Equal(t, http.StatusBadRequest, w.Code)
		var res types.APIResponse
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &res))
		assert.Equal(t, constants.ValidationFailed, res.Code)
		if assert.Len(t, res.Errors, 1) {
			assert.Equal(t, "name", res.Errors[0].Field)
		}
	})

	t.Run("malformed and empty bodies are a 400", func(t *testing.T) {
		for _, body := range []string{`{"email":`, ``, `[]`} {
			req, w := bind(t, body)
			assert.Nil(t, req, "%q", body)
			assert.Equal(t, http.StatusBadRequest, w.Code, "%q", body)
		}
	})

	t.Run("a body over the limit is a 413", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"email":"`+strings.Repeat("a", 64)+`"}`))
		c.Request.Body = http.MaxBytesReader(w, c.Request.Body, 16)

		req, ok := BindJSON[bindRequest](c)
		assert.False(t, ok)
		assert.Nil(t, req)
		assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
	})
}
