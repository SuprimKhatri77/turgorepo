package middleware

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/suprimkhatri77/turgorepo/api/internal/config"
	"github.com/suprimkhatri77/turgorepo/api/internal/constants"
)

func TestBodyLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	serve := func(req *http.Request) (*httptest.ResponseRecorder, *bool, *error) {
		reached := false
		var readErr error
		r := gin.New()
		r.POST("/", BodyLimit(10), func(c *gin.Context) {
			reached = true
			_, readErr = io.ReadAll(c.Request.Body)
			c.Status(http.StatusOK)
		})
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w, &reached, &readErr
	}

	t.Run("allows a body within the limit", func(t *testing.T) {
		w, reached, readErr := serve(httptest.NewRequest(http.MethodPost, "/", strings.NewReader("short")))
		if w.Code != http.StatusOK || !*reached || *readErr != nil {
			t.Errorf("status = %d, reached = %v, err = %v", w.Code, *reached, *readErr)
		}
	})

	t.Run("refuses a declared oversized body without reading it", func(t *testing.T) {
		w, reached, _ := serve(httptest.NewRequest(http.MethodPost, "/", strings.NewReader(strings.Repeat("x", 11))))
		if w.Code != http.StatusRequestEntityTooLarge || *reached {
			t.Errorf("status = %d, reached = %v", w.Code, *reached)
		}
	})

	t.Run("stops reading a body that hides its size", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(strings.Repeat("x", 11)))
		req.ContentLength = -1
		_, reached, readErr := serve(req)
		if !*reached || *readErr == nil {
			t.Errorf("reached = %v, err = %v", *reached, *readErr)
		}
	})
}

func TestRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name     string
		incoming string
		reused   bool
	}{
		{"reuses a normal id", "abc-123_DEF", true},
		{"mints when missing", "", false},
		{"mints when too long", strings.Repeat("a", maxRequestIDLength+1), false},
		{"mints when it has log-unsafe characters", `abc" level=ERROR`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.GET("/", RequestID(), func(c *gin.Context) { c.Status(http.StatusOK) })
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set(constants.RequestIDHeader, tt.incoming)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			got := w.Header().Get(constants.RequestIDHeader)
			if got == "" {
				t.Fatal("no request id on the response")
			}
			if (got == tt.incoming) != tt.reused {
				t.Errorf("got %q, incoming %q, want reused = %v", got, tt.incoming, tt.reused)
			}
		})
	}
}

func TestAllowedOrigins(t *testing.T) {
	got := allowedOrigins(&config.Config{GinMode: "release", FrontendURL: "https://app.example.com"})
	if len(got) != 1 || got[0] != "https://app.example.com" {
		t.Errorf("configured origin: got %v", got)
	}
	got = allowedOrigins(&config.Config{})
	if len(got) != 1 || got[0] != "http://localhost:3000" {
		t.Errorf("default origin: got %v", got)
	}
}

func TestRecoveryReturns500(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/", Recovery(), func(*gin.Context) { panic(fmt.Errorf("boom")) })
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d", w.Code)
	}
	if strings.Contains(w.Body.String(), "boom") {
		t.Errorf("response leaks the panic: %s", w.Body)
	}
}
