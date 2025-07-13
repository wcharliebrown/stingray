package handlers

import (
	"net/http"
	"stingray/logging"
	"time"
)

type LoggingMiddleware struct {
	logger *logging.Logger
}

func NewLoggingMiddleware(logger *logging.Logger) *LoggingMiddleware {
	return &LoggingMiddleware{
		logger: logger,
	}
}

func (lm *LoggingMiddleware) Wrap(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Create a response writer wrapper to capture status code
		wrappedWriter := &responseWriter{ResponseWriter: w, statusCode: 200}
		
		// Call the next handler
		next(wrappedWriter, r)
		
		// Log the access
		duration := time.Since(start)
		remoteAddr := r.RemoteAddr
		if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
			remoteAddr = forwardedFor
		}
		
		lm.logger.LogAccess(
			r.Method,
			r.URL.Path,
			remoteAddr,
			r.UserAgent(),
			wrappedWriter.statusCode,
			duration,
		)
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
} 