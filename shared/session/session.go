package session

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"time"
)

const (
	// SessionCookieName is the name of the session cookie
	SessionCookieName = "wb_sess"
	// MaxCookieSize is the maximum size of a cookie in bytes (4KB - 1 for safety)
	MaxCookieSize = 4095
)

// Session represents a user session
type Session struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Conn      *ActiveConnection `json:"conn,omitempty"`
}

// ActiveConnection holds the per-session active DB connection.
type ActiveConnection struct {
	ID       string    `json:"id"`
	Driver   string    `json:"driver"`
	DSN      string    `json:"dsn"`
	LastUsed time.Time `json:"last_used"`
}

// NewRandomID generates a new random ID
func NewRandomID() string {
	b := make([]byte, 32) // 256 bits
	if _, err := rand.Read(b); err != nil {
		panic(err) // This should never happen with crypto/rand
	}
	return hex.EncodeToString(b)
}

// getSessionKey derives a secure key from the secret
func getSessionKey(secret string) []byte {
	h := sha256.New()
	h.Write([]byte(secret))
	return h.Sum(nil)
}

// DecodeSessionData decodes session data from a base64 string
// This is exposed for testing purposes
func DecodeSessionData(encodedData string, key []byte) (map[string]interface{}, error) {
	return decodeSessionData(encodedData, key)
}

// EnsureSession returns the existing session from cookie or creates a new one.
func EnsureSession(w http.ResponseWriter, r *http.Request, secret string) *Session {
	key := getSessionKey(secret)
	
	// Try to get existing session from cookie
	if c, err := r.Cookie(SessionCookieName); err == nil && c != nil && c.Value != "" {
		if data, err := decodeSessionData(c.Value, key); err == nil {
			session := &Session{
				ID:        data["id"].(string),
				CreatedAt: time.Unix(int64(data["created_at"].(float64)), 0),
			}
			
			if connData, ok := data["conn"].(map[string]interface{}); ok && connData != nil {
				session.Conn = &ActiveConnection{
					ID:       connData["id"].(string),
					Driver:   connData["driver"].(string),
					DSN:      connData["dsn"].(string),
					LastUsed: time.Unix(int64(connData["last_used"].(float64)), 0),
				}
			}
			return session
		}
	}

	// Create new session
	session := &Session{
		ID:        NewRandomID(),
		CreatedAt: time.Now(),
	}

	// Save session to cookie
	saveSessionToCookie(w, r, session, key)

	return session
}

// saveSessionToCookie saves the session data to an encrypted cookie
func saveSessionToCookie(w http.ResponseWriter, r *http.Request, session *Session, key []byte) {
	sessionData := map[string]interface{}{
		"id":         session.ID,
		"created_at": session.CreatedAt.Unix(),
	}

	if session.Conn != nil {
		sessionData["conn"] = map[string]interface{}{
			"id":        session.Conn.ID,
			"driver":    session.Conn.Driver,
			"dsn":       session.Conn.DSN,
			"last_used": session.Conn.LastUsed.Unix(),
		}
	}

	encoded, err := encodeSessionData(sessionData, key)
	if err != nil {
		// Log error but continue with empty session
		return
	}

	// Ensure the cookie isn't too large
	if len(encoded) > MaxCookieSize {
		// Handle error: session too large
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    encoded,
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   30 * 24 * 60 * 60, // 30 days
	})
}

// SaveSession saves the session to the response cookie
func SaveSession(w http.ResponseWriter, r *http.Request, session *Session, secret string) {
	saveSessionToCookie(w, r, session, getSessionKey(secret))
}

// DeleteSession removes the session cookie
func DeleteSession(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1, // Delete the cookie
	})
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
