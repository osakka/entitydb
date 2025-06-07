package api

import (
	"net/http"
	"entitydb/logger"
)

// TEHeaderMiddleware strips the problematic TE header that causes hangs
// The TE: trailers header can cause Go's HTTP server to hang in certain configurations
type TEHeaderMiddleware struct{}

// NewTEHeaderMiddleware creates a new TE header middleware
func NewTEHeaderMiddleware() *TEHeaderMiddleware {
	return &TEHeaderMiddleware{}
}

// Middleware returns the HTTP middleware function
func (m *TEHeaderMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if TE header is present
		if teHeader := r.Header.Get("TE"); teHeader != "" {
			logger.Debug("Removing TE header to prevent hang: %s", teHeader)
			// Remove the TE header to prevent server hangs
			r.Header.Del("TE")
		}
		
		// Continue processing
		next.ServeHTTP(w, r)
	})
}