package weebase

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"sync"
	"time"
)

const sessionCookieName = "wb_sid"

// Session holds per-user state. Extend as needed (user, role, connection, prefs).
type Session struct {
	ID        string
	CreatedAt time.Time
	// TODO: add user info, role, active connection id, etc.
}

var (
	sessionsMu sync.RWMutex
	sessions   = map[string]*Session{}
)

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
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
	return s
}

func newRandomID() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}
