package shared

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
	"time"
)

const (
	sessionCookieName = "session"
	sessionIDLength  = 32
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
	b := make([]byte, sessionIDLength/2)
	if _, err := rand.Read(b); err != nil {
		panic(err) // This should never happen with crypto/rand
	}
	return hex.EncodeToString(b)
}

// EnsureSession returns the existing session from cookie or creates a new one.
func EnsureSession(w http.ResponseWriter, r *http.Request, secret string) *Session {
	// For now we don't sign the cookie because the server stores all state; the ID is opaque.
	// We still set secure attributes on cookie.
	if c, err := r.Cookie(sessionCookieName); err == nil && c != nil && c.Value != "" {
		sessionsMu.RLock()
		s, ok := sessions[c.Value]
		sessionsMu.RUnlock()
		if ok {
			return s
		}
	}

	id := newRandomID()
	s := &Session{ID: id, CreatedAt: time.Now()}
	sessionsMu.Lock()
	sessions[id] = s
	sessionsMu.Unlock()

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
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

// GetActiveConnection returns the active connection for the session
func (s *Session) GetActiveConnection() *ActiveConnection {
	if s == nil {
		return nil
	}
	return s.Conn
}

// SetActiveConnection sets the active connection for the session
func (s *Session) SetActiveConnection(conn *ActiveConnection) {
	if s != nil {
		s.Conn = conn
	}
}
