package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestHealthHandler(t *testing.T) {
	tests := []struct {
		name           string
		expectedStatus int
		expectedBody   map[string]string
	}{
		{
			name:           "health check returns ok",
			expectedStatus: http.StatusOK,
			expectedBody:   map[string]string{"status": "ok"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Execute
			err := HealthHandler(c)
			if err != nil {
				t.Errorf("HealthHandler() error = %v", err)
				return
			}

			// Assert status code
			if rec.Code != tt.expectedStatus {
				t.Errorf("HealthHandler() status = %v, want %v", rec.Code, tt.expectedStatus)
			}

			// Assert response body
			var response map[string]string
			if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
				t.Errorf("Failed to unmarshal response body: %v", err)
				return
			}

			if response["status"] != tt.expectedBody["status"] {
				t.Errorf("HealthHandler() response = %v, want %v", response, tt.expectedBody)
			}
		})
	}
}
