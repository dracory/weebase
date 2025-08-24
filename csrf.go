package weebase

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
)

const csrfCookieName = "wb_csrf"
const csrfFormKey = "csrf_token"
const csrfHeaderKey = "X-CSRF-Token"

// EnsureCSRFCookie ensures a CSRF base value cookie exists and returns a token derived from it.
func EnsureCSRFCookie(w http.ResponseWriter, r *http.Request, secret string) string {
	c, err := r.Cookie(csrfCookieName)
	if err != nil || c == nil || c.Value == "" {
		b := make([]byte, 32)
		_, _ = rand.Read(b)
		val := base64.RawURLEncoding.EncodeToString(b)
		http.SetCookie(w, &http.Cookie{
			Name:     csrfCookieName,
			Value:    val,
			Path:     "/",
			HttpOnly: true,
			Secure:   r.TLS != nil,
			SameSite: http.SameSiteLaxMode,
		})
		c = &http.Cookie{Name: csrfCookieName, Value: val}
	}
	return deriveToken(secret, c.Value)
}

// VerifyCSRF verifies the token from header or form using the double-submit cookie pattern.
func VerifyCSRF(r *http.Request, secret string) bool {
	c, err := r.Cookie(csrfCookieName)
	if err != nil || c == nil || c.Value == "" {
		return false
	}
	expected := deriveToken(secret, c.Value)
	token := r.Header.Get(csrfHeaderKey)
	if token == "" {
		_ = r.ParseForm()
		token = r.Form.Get(csrfFormKey)
	}
	if token == "" {
		return false
	}
	return hmac.Equal([]byte(token), []byte(expected))
}

func deriveToken(secret, base string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(base))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}
