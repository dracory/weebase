package api_connect_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/dracory/weebase/api/api_connect"
	"github.com/dracory/weebase/shared/types"
)

func TestApiConnect_ServeHTTP(t *testing.T) {
	tests := []struct {
		name           string
		setupRequest   func() *http.Request
		expectedStatus int
		expectedError  string
	}{
		{
			name: "valid connection request",
			setupRequest: func() *http.Request {
				form := make(url.Values)
				form.Add("driver", "sqlite")
				form.Add("dsn", ":memory:")
				req := httptest.NewRequest("POST", "/connect", strings.NewReader(form.Encode()))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				return req
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "missing driver",
			setupRequest: func() *http.Request {
				form := make(url.Values)
				form.Add("dsn", "test.db")
				req := httptest.NewRequest("POST", "/connect", strings.NewReader(form.Encode()))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				return req
			},
			expectedStatus: http.StatusOK,
			expectedError:  "unsupported driver",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test config
			cfg := types.Config{
				SessionSecret: "test-secret",
			}

			// Create handler
			handler := api_connect.New(cfg)

			// Create response recorder
			rr := httptest.NewRecorder()

			// Create request
			req := tt.setupRequest()

			// Call handler
			handler.ServeHTTP(rr, req)

			// Check status code
			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}

			// Check error message if expected
			if tt.expectedError != "" {
				var response map[string]interface{}
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Fatalf("failed to parse response: %v", err)
				}
				if message, ok := response["message"].(string); !ok || message != tt.expectedError {
					t.Errorf("unexpected error message: got %v want %v", message, tt.expectedError)
				}
			}
		})
	}
}

// TestConnectRequest tests the ConnectRequest struct JSON marshaling
func TestConnectRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *api_connect.ConnectRequest
		want    string
		wantErr bool
	}{
		{
			name: "basic request",
			req: &api_connect.ConnectRequest{
				Driver: "sqlite",
				DSN:    ":memory:",
			},
			want: `{"profile_id":"","driver":"sqlite","dsn":":memory:","server":"","port":"","username":"","password":"","database":""}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConnectRequest.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && string(got) != tt.want {
				t.Errorf("ConnectRequest.MarshalJSON() = %v, want %v", string(got), tt.want)
			}
		})
	}
}
