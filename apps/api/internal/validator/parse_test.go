package validator

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/suprimkhatri77/turgorepo/api/internal/response"
)

type messageRequest struct {
	Tags  []string `json:"tags" binding:"omitempty,min=2,max=3"`
	Day   string   `json:"day" label:"Start date" binding:"omitempty,date_format"`
	Role  string   `json:"role" binding:"omitempty,oneof=member admin"`
	Age   int      `json:"age"`
	Admin bool     `json:"admin"`
}

func TestParseMessages(t *testing.T) {
	gin.SetMode(gin.TestMode)
	Init()

	tests := []struct {
		name string
		body string
		want response.FieldError
	}{
		{"too few items", `{"tags":["a"]}`,
			response.FieldError{Code: "TOO_SHORT", Field: "tags", Message: "Tags needs at least 2 items"}},
		{"too many items", `{"tags":["a","b","c","d"]}`,
			response.FieldError{Code: "TOO_LONG", Field: "tags", Message: "Tags can have at most 3 items"}},
		{"bad date", `{"day":"21/11/2026"}`,
			response.FieldError{Code: "INVALID_DATE", Field: "day", Message: "Start date must be a valid date"}},
		{"value not in the list", `{"role":"owner"}`,
			response.FieldError{Code: "INVALID_VALUE", Field: "role", Message: "Choose a valid role"}},
		{"number for text, with a label", `{"day":5}`,
			response.FieldError{Code: "INVALID_TYPE", Field: "day", Message: "Start date must be text"}},
		{"text for a number", `{"age":"ten"}`,
			response.FieldError{Code: "INVALID_TYPE", Field: "age", Message: "Age must be a number"}},
		{"text for true or false", `{"admin":"yes"}`,
			response.FieldError{Code: "INVALID_TYPE", Field: "admin", Message: "Admin must be true or false"}},
		{"text for a list", `{"tags":"a"}`,
			response.FieldError{Code: "INVALID_TYPE", Field: "tags", Message: "Tags must be a list"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body))

			_, ok := BindJSON[messageRequest](c)
			require.False(t, ok)
			require.Equal(t, http.StatusBadRequest, w.Code)
			var res response.Body
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &res))
			assert.Equal(t, []response.FieldError{tt.want}, res.Errors)
		})
	}
}

func TestItems(t *testing.T) {
	assert.Equal(t, "1 item", items("1"))
	assert.Equal(t, "3 items", items("3"))
}
