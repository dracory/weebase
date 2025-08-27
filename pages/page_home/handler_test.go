package page_home

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/dracory/weebase/shared/session"
	"github.com/dracory/weebase/shared/types"
	"github.com/dracory/weebase/shared/urls"
	"github.com/stretchr/testify/assert"
)

// testContext holds test dependencies
type testContext struct {
	t        *testing.T
	handler  *pageHomeController
	sessions map[string]*session.Session
}

// newTestContext creates a new test context
func newTestContext(t *testing.T) *testContext {
	return &testContext{
		t: t,
		handler: New(types.Config{
			BasePath:        "/test",
			SessionSecret:   "test-secret",
			EnabledDrivers:  []string{"sqlite"},
			SafeModeDefault: true,
		}),
		sessions: make(map[string]*session.Session),
	}
}

// createTestSession creates a test session with the given ID and connection
func (tc *testContext) createTestSession(id string, conn *session.ActiveConnection) *session.Session {
	sess := &session.Session{
		ID:        id,
		CreatedAt: time.Now(),
		Conn:      conn,
	}
	tc.sessions[id] = sess
	return sess
}

// testSessionStore mimics the session package's internal storage for testing
var testSessions = struct {
	sync.RWMutex
	sessions map[string]*session.Session
}{
	sessions: make(map[string]*session.Session),
}

// testGetSession is a test helper to get a session from the test store
func testGetSession(sessionID string) (*session.Session, bool) {
	testSessions.RLock()
	defer testSessions.RUnlock()
	sess, exists := testSessions.sessions[sessionID]
	return sess, exists
}

// testSetSession is a test helper to set a session in the test store
func testSetSession(sess *session.Session) {
	testSessions.Lock()
	defer testSessions.Unlock()
	testSessions.sessions[sess.ID] = sess
}

// testClearSessions clears all test sessions
func testClearSessions() {
	testSessions.Lock()
	defer testSessions.Unlock()
	testSessions.sessions = make(map[string]*session.Session)
}

func TestServeHTTP_WithActiveConnection(t *testing.T) {
	// Setup test context
	tc := newTestContext(t)

	// Create test request
	req := httptest.NewRequest("GET", "/test?action=page_home", nil)
	w := httptest.NewRecorder()

	// Create and store a session with active connection
	sess := tc.createTestSession("test-session-with-conn", &session.ActiveConnection{
		Driver:   "sqlite",
		LastUsed: time.Now(),
	})
	testSetSession(sess)

	// Set session cookie
	req.AddCookie(&http.Cookie{
		Name:  session.SessionCookieName,
		Value: sess.ID,
		Path:  "/",
	})

	// Execute handler
	tc.handler.ServeHTTP(w, req)

	// Verify response
	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Content-Type"), "text/html")

	// Clean up
	testClearSessions()
}

func TestServeHTTP_WithoutActiveConnection(t *testing.T) {
	// Setup test context
	tc := newTestContext(t)

	// Create test request
	req := httptest.NewRequest("GET", "/test?action=page_home", nil)
	w := httptest.NewRecorder()

	// Create and store a session without active connection
	sess := tc.createTestSession("test-session-no-conn", nil)
	testSetSession(sess)

	// Set session cookie
	req.AddCookie(&http.Cookie{
		Name:  session.SessionCookieName,
		Value: sess.ID,
		Path:  "/",
	})

	// Execute handler
	tc.handler.ServeHTTP(w, req)

	// Verify response
	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	loginURL := urls.Login("/test")
	assert.Equal(t, loginURL, resp.Header.Get("Location"))

	// Clean up
	testClearSessions()
}

func TestHandle(t *testing.T) {
	// Setup test context
	tc := newTestContext(t)

	// Execute handler
	html, err := tc.handler.Handle()

	// Verify response
	assert.NoError(t, err)
	// Check for case-insensitive HTML5 doctype
	assert.Regexp(t, `(?i)<!doctype\s+html>`, string(html))
}
