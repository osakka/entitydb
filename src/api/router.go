package api

import (
	"context"
	"entitydb/logger"
	"net/http"
	"strings"
	"time"
)

// Router handles API routing
type Router struct {
	mux             *http.ServeMux
	middleware      []MiddlewareFunc
	staticPaths     map[string]string
	handlers        map[string]http.HandlerFunc // Used to track registered handlers
	registeredPaths map[string]bool            // Tracks registered paths to avoid duplicates
}

// MiddlewareFunc defines a middleware function
type MiddlewareFunc func(http.HandlerFunc) http.HandlerFunc

// RouterMiddleware alias for backward compatibility
type RouterMiddleware = MiddlewareFunc

// NewRouter creates a new router
func NewRouter() *Router {
	return &Router{
		mux:             http.NewServeMux(),
		middleware:      []MiddlewareFunc{},
		staticPaths:     make(map[string]string),
		handlers:        make(map[string]http.HandlerFunc),
		registeredPaths: make(map[string]bool),
	}
}

// Use adds middleware to the router
func (r *Router) Use(middleware MiddlewareFunc) {
	r.middleware = append(r.middleware, middleware)
}

// applyMiddleware applies all middleware to a handler
func (r *Router) applyMiddleware(handler http.HandlerFunc) http.HandlerFunc {
	// Apply middleware in reverse order (so they execute in registration order)
	for i := len(r.middleware) - 1; i >= 0; i-- {
		handler = r.middleware[i](handler)
	}
	return handler
}

// Handle registers a handler for a path and method
func (r *Router) Handle(method, path string, handler http.HandlerFunc) {
	// Check if the path is already registered (for the same HTTP method)
	pathKey := method + ":" + path
	if _, exists := r.registeredPaths[pathKey]; exists {
		logger.Warn("Path %s %s is already registered, skipping duplicate registration", method, path)
		return
	}
	
	// Log the route being registered
	logger.Debug("Registering route: %s %s", method, path)
	
	// Create an exact path pattern for ServeMux
	exactPath := path
	
	// Create a unique key for the handler based on method and path
	handlerKey := method + ":" + path
	
	// Store the handler
	r.handlers[handlerKey] = handler
	
	// Mark this path as registered
	r.registeredPaths[pathKey] = true
	
	// Handle function to check method and path
	handleRequest := func(w http.ResponseWriter, req *http.Request) {
		// Log the request
		logger.Debug("Handler called for: %s %s (registered as %s %s)", 
			req.Method, req.URL.Path, method, path)
		
		// Check if method matches
		if req.Method != method {
			logger.Debug("Method mismatch: received %s, expected %s", req.Method, method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		
		// Apply middleware and execute handler
		logger.Debug("Executing handler for %s %s", method, path)
		r.applyMiddleware(handler)(w, req)
	}
	
	// Register the path with the ServeMux
	r.mux.HandleFunc(exactPath, func(w http.ResponseWriter, req *http.Request) {
		// Check if the path matches exactly
		if req.URL.Path != path {
			logger.Debug("Path mismatch: received %s, expected %s", req.URL.Path, path)
			http.NotFound(w, req)
			return
		}
		
		handleRequest(w, req)
	})
}

// GET registers a GET handler
func (r *Router) GET(path string, handler http.HandlerFunc) {
	r.Handle(http.MethodGet, path, handler)
}

// POST registers a POST handler
func (r *Router) POST(path string, handler http.HandlerFunc) {
	r.Handle(http.MethodPost, path, handler)
}

// PUT registers a PUT handler
func (r *Router) PUT(path string, handler http.HandlerFunc) {
	r.Handle(http.MethodPut, path, handler)
}

// DELETE registers a DELETE handler
func (r *Router) DELETE(path string, handler http.HandlerFunc) {
	r.Handle(http.MethodDelete, path, handler)
}

// ServeStatic registers a path for serving static files
func (r *Router) ServeStatic(urlPath, fsPath string) {
	logger.Debug("Registering static file server for %s -> %s", urlPath, fsPath)
	r.staticPaths[urlPath] = fsPath
	
	fileServer := http.FileServer(http.Dir(fsPath))
	
	// Create a handler that only serves static files for non-API requests
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Log request path
		logger.Debug("Static handler received request: %s", req.URL.Path)
		
		// Skip API requests, as they're handled by the API routes
		if strings.HasPrefix(req.URL.Path, "/api/") {
			logger.Debug("Skipping static file handling for API request: %s", req.URL.Path)
			http.NotFound(w, req)
			return
		}
		
		// For requests to the root, redirect to dashboard.html by default
		if req.URL.Path == "/" {
			logger.Debug("Redirecting root request to dashboard.html")
			http.Redirect(w, req, "/dashboard.html", http.StatusFound)
			return
		}
		
		// For all other requests, serve the static file
		logger.Debug("Serving static file for path: %s", req.URL.Path)
		fileServer.ServeHTTP(w, req)
	})
	
	// Register the handler without stripping the prefix for the root path
	if urlPath == "/" {
		logger.Debug("Registering root static handler at: %s", urlPath)
		r.mux.Handle(urlPath, handler)
	} else {
		logger.Debug("Registering static handler with prefix strip at: %s", urlPath)
		r.mux.Handle(urlPath, http.StripPrefix(urlPath, handler))
	}
}

// ServeHTTP implements the http.Handler interface
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	start := time.Now()
	
	// Log received request
	logger.Debug("Router received request: %s %s", req.Method, req.URL.Path)
	
	// Set enhanced CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Type, Authorization")
	w.Header().Set("Access-Control-Max-Age", "86400")
	
	// Handle preflight requests
	if req.Method == "OPTIONS" {
		logger.Debug("Responding to OPTIONS request for %s", req.URL.Path)
		w.WriteHeader(http.StatusOK)
		return
	}
	
	// First, try the API status endpoint
	if req.URL.Path == "/api/v1/status" && req.Method == "GET" {
		logger.Debug("Handling API status request directly")
		StatusHandler()(w, req)
		return
	}
	
	// Let the ServeMux handle the request
	logger.Debug("Delegating request to mux: %s %s", req.Method, req.URL.Path)
	r.mux.ServeHTTP(w, req)
	
	logger.Debug("%s %s %s - %v", req.RemoteAddr, req.Method, req.URL.Path, time.Since(start))
}

// Note: RespondJSON and RespondError moved to auth.go
// Use those functions instead of duplicating them here

// StatusHandler returns a simple status handler for API health checks
func StatusHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := map[string]interface{}{
			"status":     "ok",
			"timestamp":  time.Now().Format(time.RFC3339),
			"api_status": "connected",
		}
		RespondJSON(w, http.StatusOK, status)
	}
}

// LoggingMiddleware logs request details
func LoggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Call the next handler
		next(w, r)
		
		// Log the request
		duration := time.Since(start)
		logger.Debug("%s %s %s - %v", r.RemoteAddr, r.Method, r.URL.Path, duration)
	}
}

// AuthMiddleware handles authentication
// DEPRECATED: Use Auth.AuthMiddleware() instead for proper token validation
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Log deprecation warning
		logger.Debug("Warning: Using deprecated global AuthMiddleware. Use Auth.AuthMiddleware() instead")
		
		// Get the authentication token from the request
		authHeader := r.Header.Get("Authorization")
		
		// Check if Authorization header is present
		if authHeader == "" {
			RespondError(w, http.StatusUnauthorized, "Authentication required")
			return
		}
		
		// Extract token from Bearer format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			RespondError(w, http.StatusUnauthorized, "Invalid token format")
			return
		}
		
		// Add a warning context value
		ctx := context.WithValue(r.Context(), "auth_warning", 
			"Using deprecated middleware without proper token validation")
		
		// Call the next handler with the updated context
		next(w, r.WithContext(ctx))
	}
}