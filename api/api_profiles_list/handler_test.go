package api_profiles_list_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/dracory/weebase/api/api_profiles_list"
	"github.com/dracory/weebase/shared/constants"
	"github.com/dracory/weebase/shared/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler_ServeHTTP(t *testing.T) {
	profile := api_profiles_list.Profile{
		ID:     "test-id",
		Name:   "test-profile",
		Driver: "postgres",
	}

	tests := []struct {
		name           string
		setupRequest   func() *http.Request
		expectedStatus int
		expectError    bool
		expectProfiles bool
	}{
		{
			name: "successful get with profiles",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("GET", "/api/profiles", nil)
				profiles := []api_profiles_list.Profile{profile}
				profilesJSON, _ := json.Marshal(profiles)
				req.AddCookie(&http.Cookie{
					Name:  constants.CookieProfiles,
					Value: url.QueryEscape(string(profilesJSON)),
				})
				return req
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
			expectProfiles: true,
		},
		{
			name: "invalid method",
			setupRequest: func() *http.Request {
				return httptest.NewRequest("POST", "/api/profiles", nil)
			},
			expectedStatus: http.StatusOK,
			expectError:    true,
			expectProfiles: false,
		},
		{
			name: "no profiles cookie",
			setupRequest: func() *http.Request {
				return httptest.NewRequest("GET", "/api/profiles", nil)
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
			expectProfiles: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			handler := api_profiles_list.New(types.Config{SessionSecret: "test-secret"})
			req := tt.setupRequest()
			rr := httptest.NewRecorder()

			// Test
			handler.ServeHTTP(rr, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, rr.Code, "handler returned wrong status code")

			// Parse response for all cases to check status
			var resp map[string]interface{}
			err := json.Unmarshal(rr.Body.Bytes(), &resp)
			require.NoError(t, err, "failed to parse response")

			// Check response status
			status, ok := resp["status"].(string)
			require.True(t, ok, "invalid status in response")

			if tt.expectError {
				assert.Equal(t, "error", status, "expected error status")
			} else {
				assert.Equal(t, "success", status, "expected success status")

				// Check if profiles exist in response
				if tt.expectProfiles {
					data, ok := resp["data"].(map[string]interface{})
					require.True(t, ok, "invalid data in response")
					profiles, ok := data["profiles"].([]interface{})
					require.True(t, ok, "invalid profiles in response")
					assert.Greater(t, len(profiles), 0, "expected at least one profile")
				}
			}
		})
	}
}

func TestGetProfilesFromCookie(t *testing.T) {
	tests := []struct {
		name          string
		setupCookie   func(*http.Request) *http.Request
		expectError   bool
		expectLength  int
		expectProfile *api_profiles_list.Profile
	}{
		{
			name: "valid cookie",
			setupCookie: func(req *http.Request) *http.Request {
				profiles := []api_profiles_list.Profile{{
					ID:     "test1",
					Name:   "Test DB",
					Driver: "postgres",
				}}
				profilesJSON, _ := json.Marshal(profiles)
				req.AddCookie(&http.Cookie{
					Name:  constants.CookieProfiles,
					Value: url.QueryEscape(string(profilesJSON)),
				})
				return req
			},
			expectError:  false,
			expectLength: 1,
			expectProfile: &api_profiles_list.Profile{
				ID:     "test1",
				Name:   "Test DB",
				Driver: "postgres",
			},
		},
		{
			name: "no cookie",
			setupCookie: func(req *http.Request) *http.Request {
				return req
			},
			expectError:  false,
			expectLength: 0,
		},
		{
			name: "invalid json",
			setupCookie: func(req *http.Request) *http.Request {
				req.AddCookie(&http.Cookie{
					Name:  constants.CookieProfiles,
					Value: "invalid-json",
				})
				return req
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			handler := api_profiles_list.New(types.Config{SessionSecret: "test-secret"})
			req := tt.setupCookie(httptest.NewRequest("GET", "/api/profiles", nil))

			// Test
			profiles, err := handler.GetProfilesFromCookie(req)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, profiles, tt.expectLength)

				if tt.expectProfile != nil && len(profiles) > 0 {
					assert.Equal(t, *tt.expectProfile, profiles[0])
				}
			}
		})
	}
}

func TestSetProfilesCookie(t *testing.T) {
	tests := []struct {
		name           string
		profiles       []api_profiles_list.Profile
		setupConfig    func(*types.Config)
		expectSecure   bool
		expectHttpOnly bool
	}{
		{
			name: "secure cookies enabled",
			profiles: []api_profiles_list.Profile{{
				ID:     "test1",
				Name:   "Test DB",
				Driver: "postgres",
			}},
			setupConfig: func(cfg *types.Config) {
				cfg.SecureCookies = true
			},
			expectSecure:   true,
			expectHttpOnly: true,
		},
		{
			name: "secure cookies disabled",
			profiles: []api_profiles_list.Profile{{
				ID:     "test1",
				Name:   "Test DB",
				Driver: "postgres",
			}},
			setupConfig: func(cfg *types.Config) {
				cfg.SecureCookies = false
			},
			expectSecure:   false,
			expectHttpOnly: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			cfg := types.Config{}
			if tt.setupConfig != nil {
				tt.setupConfig(&cfg)
			}

			handler := api_profiles_list.New(cfg)
			rr := httptest.NewRecorder()

			// Test
			err := handler.SetProfilesCookie(rr, tt.profiles)

			// Assert
			assert.NoError(t, err)

			// Check cookie was set
			cookies := rr.Result().Cookies()
			assert.Len(t, cookies, 1)

			cookie := cookies[0]
			assert.Equal(t, constants.CookieProfiles, cookie.Name)
			assert.Equal(t, "/", cookie.Path)
			assert.Equal(t, 30*24*60*60, cookie.MaxAge) // 30 days
			assert.Equal(t, tt.expectSecure, cookie.Secure)
			assert.Equal(t, tt.expectHttpOnly, cookie.HttpOnly)

			// Verify the cookie value can be unmarshaled back to profiles
			// First URL decode the cookie value
			decodedValue, err := url.QueryUnescape(cookie.Value)
			require.NoError(t, err, "failed to decode cookie value")

			var profiles []api_profiles_list.Profile
			err = json.Unmarshal([]byte(decodedValue), &profiles)
			assert.NoError(t, err, "failed to unmarshal profiles from cookie")
			assert.Len(t, profiles, len(tt.profiles))
		})
	}
}
