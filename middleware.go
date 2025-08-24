package weebase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"time"
)

type ctxKey string

const ctxKeyRequestID ctxKey = "request_id"

// GetRequestID returns the request id from context if present.
func GetRequestID(ctx context.Context) string {
	if v, ok := ctx.Value(ctxKeyRequestID).(string); ok {
		return v
	}
	return ""
}

// RequestLogger adds a request id to the context and logs basic request info.
func RequestLogger(next http.Handler) http.Handler {
	logger := slog.Default()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := newReqID()
		ctx := context.WithValue(r.Context(), ctxKeyRequestID, reqID)

		ww := &statusRecorder{ResponseWriter: w, status: 200}
		start := time.Now()

		ww.Header().Set("X-Request-Id", reqID)
		next.ServeHTTP(ww, r.WithContext(ctx))

		dur := time.Since(start)
		logger.Info("http_request",
			slog.String("id", reqID),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", ww.status),
			slog.String("remote", r.RemoteAddr),
			slog.String("ua", r.UserAgent()),
			slog.Duration("duration", dur),
		)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

func newReqID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return time.Now().Format("20060102150405.000000")
	}
	return hex.EncodeToString(b)
}
