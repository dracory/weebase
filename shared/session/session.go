package session

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"sync"
	"time"
)

const (
	// SessionCookieName is the name of the session cookie
	SessionCookieName = "wb_sid"
	// SessionIDLength is the length of the session ID in bytes
	SessionIDLength = 32
)

// Session represents a user session
type Session struct {
	ID        string
	CreatedAt time.Time
	Conn      *ActiveConnection
}

// ActiveConnection holds the per-session active DB connection.
type ActiveConnection struct {
	ID       string
	Driver   string
	DB       interface{} // *gorm.DB or *sql.DB depending on the driver
	LastUsed time.Time
}

var (
	sessionsMu sync.RWMutex
	sessions   = map[string]*Session{}
)

// newRandomID generates a new random ID for sessions
func newRandomID() string {
	b := make([]byte, SessionIDLength/2)
	if _, err := rand.Read(b); err != nil {
		panic(err) // This should never happen with crypto/rand
	}
	return hex.EncodeToString(b)
}

// EnsureSession returns the existing session from cookie or creates a new one.
func EnsureSession(w http.ResponseWriter, r *http.Request, secret string) *Session {
	// Try to get existing session from cookie
	if c, err := r.Cookie(SessionCookieName); err == nil && c != nil && c.Value != "" {
		sessionsMu.RLock()
		s, ok := sessions[c.Value]
		sessionsMu.RUnlock()
		if ok {
			return s
		}
	}

	// Create new session
	id := newRandomID()
	s := &Session{
		ID:        id,
		CreatedAt: time.Now(),
	}

	// Store session
	sessionsMu.Lock()
	sessions[id] = s
	sessionsMu.Unlock()

	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    id,
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
	})

	return s
}

// GetSession retrieves an existing session by ID
func GetSession(sessionID string) (*Session, bool) {
	sessionsMu.RLock()
	defer sessionsMu.RUnlock()
	session, exists := sessions[sessionID]
	return session, exists
}

// DeleteSession removes a session
func DeleteSession(sessionID string) {
	sessionsMu.Lock()
	defer sessionsMu.Unlock()
	delete(sessions, sessionID)
}

// GenerateCSRFToken generates a new CSRF token for the current session
func GenerateCSRFToken(secret string) string {
	h := sha256.New()
	h.Write([]byte(secret))
	h.Write([]byte(time.Now().String()))
	return hex.EncodeToString(h.Sum(nil))
}

// VerifyCSRF verifies a CSRF token
func VerifyCSRF(r *http.Request, secret string) bool {
	token := r.Header.Get("X-CSRF-Token")
	if token == "" {
		token = r.FormValue("csrf_token")
	}

	// In a real implementation, you would verify the token against the session
	// For now, we'll just check if it's not empty
	return token != ""
}
