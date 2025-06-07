package api

import (
	"net/http"
)

// ConnectionCloseMiddleware ensures connections are closed after each request
// This prevents hanging connections, especially with simple curl commands
type ConnectionCloseMiddleware struct{}

// NewConnectionCloseMiddleware creates a new connection close middleware
func NewConnectionCloseMiddleware() *ConnectionCloseMiddleware {
	return &ConnectionCloseMiddleware{}
}

// closeWriter wraps http.ResponseWriter to set Connection: close before first write
type closeWriter struct {
	http.ResponseWriter
	headerSet bool
}

func (cw *closeWriter) Write(data []byte) (int, error) {
	if !cw.headerSet {
		cw.Header().Set("Connection", "close")
		cw.headerSet = true
	}
	return cw.ResponseWriter.Write(data)
}

func (cw *closeWriter) WriteHeader(statusCode int) {
	if !cw.headerSet {
		cw.Header().Set("Connection", "close")
		cw.headerSet = true
	}
	cw.ResponseWriter.WriteHeader(statusCode)
}

// Middleware returns the HTTP middleware function
func (m *ConnectionCloseMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Wrap the response writer to set Connection: close just before writing
		wrapped := &closeWriter{ResponseWriter: w}
		
		// Continue processing
		next.ServeHTTP(wrapped, r)
	})
}