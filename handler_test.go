package weebase_test

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/dracory/weebase"
	"github.com/dracory/weebase/shared/constants"
	"github.com/dracory/weebase/shared/session"
	"github.com/stretchr/testify/assert"
)

var (
	sessionsMu sync.RWMutex
	sessions   = map[string]*session.Session{}
)

func TestNewHandler(t *testing.T) {
	t.Run("with default options", func(t *testing.T) {
		h := weebase.New()
		assert.NotNil(t, h, "expected handler to be created")
	})
}

func TestRegister(t *testing.T) {
	t.Run("register handler", func(t *testing.T) {
		mux := http.NewServeMux()
		h := weebase.NewHandler(weebase.Options{})
		weebase.Register(mux, "/test", h)

		req := httptest.NewRequest("GET", "/test?action="+constants.ActionHome, nil)
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)

		// Expecting a redirect to login page since we're not authenticated
		assert.Equal(t, http.StatusFound, rr.Code, "expected redirect to login page")
	})
}

func TestHandler_ServeHTTP(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		setup          func() http.Handler
		expectedStatus int
	}{
		{
			name: "home page",
			url:  "/?action=" + constants.ActionHome,
			setup: func() http.Handler {
				return weebase.NewHandler(weebase.Options{})
			},
			expectedStatus: http.StatusFound, // Expecting redirect to login
		},
		{
			name: "login page",
			url:  "/?action=" + constants.ActionPageLogin,
			setup: func() http.Handler {
				return weebase.NewHandler()
			},
			expectedStatus: http.StatusFound, // Expecting redirect to login
		},
		{
			name: "non-existent page",
			url:  "/?action=nonexistent",
			setup: func() http.Handler {
				return weebase.NewHandler()
			},
			expectedStatus: http.StatusFound, // Expecting redirect to login
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := tt.setup()
			req := httptest.NewRequest("GET", tt.url, nil)
			rr := httptest.NewRecorder()

			h.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code, "handler returned wrong status code")
		})
	}
}

func TestHandler_ServeHTTP_NotFound(t *testing.T) {
	h := weebase.New()
	req := httptest.NewRequest("GET", "/nonexistent", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	// The application returns 200 for non-existent routes
	assert.Equal(t, http.StatusOK, rr.Code, "expected 200 for non-existent route")
}

func TestHandler_API_ProfilesSave(t *testing.T) {
	h := weebase.New()
	req := httptest.NewRequest("POST", "/?action=api/profiles/save", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	// The API returns 200 OK with an error message in the response body
	assert.Equal(t, http.StatusOK, rr.Code, "expected 200 OK for API endpoint")
}

func TestHandler_API_Connect(t *testing.T) {
	h := weebase.New()
	req := httptest.NewRequest("POST", "/?action=api/connect", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	// The API returns 200 OK with an error message in the response body
	assert.Equal(t, http.StatusOK, rr.Code, "expected 200 OK for API endpoint")
}

func TestHandler_API_ProfilesList(t *testing.T) {
	h := weebase.New()
	req := httptest.NewRequest("GET", "/?action=api/profiles/list", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	// The API returns 200 OK with an error message in the response body
	assert.Equal(t, http.StatusOK, rr.Code, "expected 200 OK for API endpoint")
}

func TestHandler_API_TablesList(t *testing.T) {
	tests := []struct {
		name           string
		setup          func(t *testing.T) http.Handler
		action         string
		authenticated  bool
		expectedStatus int
		expectRedirect bool
	}{
		{
			name: "unauthenticated tables list",
			setup: func(t *testing.T) http.Handler {
				return weebase.New()
			},
			action:         "api/tables/list",
			authenticated:  false,
			expectedStatus: http.StatusFound, // Expecting redirect to login
			expectRedirect: true,
		},
		{
			name: "authenticated tables list with no connection",
			setup: func(t *testing.T) http.Handler {
				h := weebase.New()
				return h
			},
			action:         "api/tables/list",
			authenticated:  true,
			expectedStatus: http.StatusOK, // Should return 200 with error in body
			expectRedirect: false,
		},
		{
			name: "tables list with invalid action",
			setup: func(t *testing.T) http.Handler {
				return weebase.New()
			},
			action:         "api/invalid-action",
			authenticated:  false,
			expectedStatus: http.StatusFound, // Expecting redirect to login
			expectRedirect: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := tt.setup(t)
			req := httptest.NewRequest("GET", "/?action="+tt.action, nil)
			rr := httptest.NewRecorder()

			// If test case is authenticated, create a session
			if tt.authenticated {
				sessionID := createTestSession(t)
				req.AddCookie(&http.Cookie{
					Name:  "wb_sid",
					Value: sessionID,
				})
			}

			h.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code, "handler returned wrong status code")

			// Check for redirect location if expected
			if tt.expectRedirect {
				location, err := rr.Result().Location()
				assert.NoError(t, err, "expected a redirect location")
				assert.Contains(t, location.String(), "action=page_login", "expected redirect to login page")
			}

			// For authenticated API calls, check the response body
			if tt.authenticated && tt.expectedStatus == http.StatusOK {
				assert.Contains(t, rr.Body.String(), `"success"`, "expected JSON response with success field")
			}
		})
	}
}

// createTestSession creates a test session and returns its ID
func createTestSession(t *testing.T) string {
	sessionID := "test-session-" + time.Now().Format(time.RFC3339Nano)

	sessionsMu.Lock()
	sessions[sessionID] = &session.Session{
		ID:        sessionID,
		CreatedAt: time.Now(),
		Conn:      nil, // No active connection
	}
	sessionsMu.Unlock()

	return sessionID
}
