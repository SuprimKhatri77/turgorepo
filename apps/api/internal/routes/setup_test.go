package routes

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type fakePinger struct{ err error }

func (p fakePinger) Ping(context.Context) error { return p.err }

func TestReady(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		dependencies []Pinger
		want         int
	}{
		{"all reachable", []Pinger{fakePinger{}, fakePinger{}}, http.StatusOK},
		{"database down", []Pinger{fakePinger{err: errors.New("down")}, fakePinger{}}, http.StatusServiceUnavailable},
		{"redis down", []Pinger{fakePinger{}, fakePinger{err: errors.New("down")}}, http.StatusServiceUnavailable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.GET("/", ready(tt.dependencies...))
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
			if w.Code != tt.want {
				t.Errorf("status = %d, want %d", w.Code, tt.want)
			}
		})
	}
}
