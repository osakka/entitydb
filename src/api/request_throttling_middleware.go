// Package api provides HTTP request throttling middleware for protecting EntityDB
// against aggressive polling patterns and request abuse. This implements intelligent
// request pattern detection with adaptive response delays.
package api

import (
	"crypto/sha256"
	"entitydb/config"
	"entitydb/logger"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// RequestThrottlingMiddleware provides intelligent request throttling to protect
// against aggressive UI polling and request abuse patterns. It implements:
//
// - Request pattern detection per client IP
// - Client health scoring based on request frequency
// - Adaptive delays for abusive clients
// - Response caching for repeated requests
// - Zero impact on well-behaved clients
//
// The middleware uses a time-window based approach to track request patterns
// and applies graduated throttling responses based on client behavior.
type RequestThrottlingMiddleware struct {
	config     *config.Config
	clients    map[string]*ClientTracker // client IP -> tracker
	cache      map[string]*CachedResponse // request hash -> cached response
	mu         sync.RWMutex              // Protects clients and cache maps
	enabled    bool                      // Feature flag for throttling
}

// ClientTracker maintains request statistics and health scoring for a specific client
type ClientTracker struct {
	IP                string                 // Client IP address
	RequestTimes      []time.Time           // Sliding window of request timestamps
	EndpointCount     map[string]int        // endpoint -> request count in current window
	HealthScore       int                   // Current client health score (0=good, higher=worse)
	LastActivity      time.Time             // Last request timestamp
	TotalRequests     int                   // Total requests from this client
	ThrottledRequests int                   // Number of throttled requests
	mu                sync.RWMutex          // Protects client data
}

// CachedResponse stores a cached HTTP response for serving to repeated requests
type CachedResponse struct {
	StatusCode int                 // HTTP status code
	Headers    map[string]string   // Response headers
	Body       []byte             // Response body
	Timestamp  time.Time          // When this response was cached
	RequestHash string            // Hash of the original request
}

// NewRequestThrottlingMiddleware creates a new request throttling middleware instance
func NewRequestThrottlingMiddleware(cfg *config.Config) *RequestThrottlingMiddleware {
	return &RequestThrottlingMiddleware{
		config:  cfg,
		clients: make(map[string]*ClientTracker),
		cache:   make(map[string]*CachedResponse),
		enabled: cfg.ThrottleEnabled,
	}
}

// Handler returns the middleware function for use in HTTP routing
func (rtm *RequestThrottlingMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip throttling if disabled
		if !rtm.enabled {
			next.ServeHTTP(w, r)
			return
		}

		// Get client IP
		clientIP := rtm.getClientIP(r)
		
		// Track this request and get throttling decision
		shouldThrottle, delay, cachedResponse := rtm.analyzeRequest(clientIP, r)
		
		if cachedResponse != nil {
			// Serve cached response for repeated requests
			rtm.serveCachedResponse(w, cachedResponse)
			logger.Debug("Served cached response to %s for %s (health score: %d)", 
				clientIP, r.URL.Path, rtm.getClientHealthScore(clientIP))
			return
		}

		if shouldThrottle && delay > 0 {
			// Apply adaptive delay for throttled clients
			logger.Debug("Throttling request from %s for %s: delay=%v (health score: %d)", 
				clientIP, r.URL.Path, delay, rtm.getClientHealthScore(clientIP))
			time.Sleep(delay)
		}

		// Process request normally
		next.ServeHTTP(w, r)
	})
}

// analyzeRequest analyzes the incoming request and determines throttling strategy
func (rtm *RequestThrottlingMiddleware) analyzeRequest(clientIP string, r *http.Request) (bool, time.Duration, *CachedResponse) {
	rtm.mu.Lock()
	defer rtm.mu.Unlock()

	// Get or create client tracker
	client, exists := rtm.clients[clientIP]
	if !exists {
		client = &ClientTracker{
			IP:            clientIP,
			RequestTimes:  make([]time.Time, 0),
			EndpointCount: make(map[string]int),
			LastActivity:  time.Now(),
		}
		rtm.clients[clientIP] = client
	}

	// Update client activity
	client.mu.Lock()
	defer client.mu.Unlock()

	now := time.Now()
	client.LastActivity = now
	client.TotalRequests++

	// Clean old request timestamps (keep last minute)
	cutoff := now.Add(-time.Minute)
	newTimes := make([]time.Time, 0)
	for _, t := range client.RequestTimes {
		if t.After(cutoff) {
			newTimes = append(newTimes, t)
		}
	}
	client.RequestTimes = append(newTimes, now)

	// Update endpoint-specific counters
	endpoint := r.Method + " " + r.URL.Path
	client.EndpointCount[endpoint]++

	// Calculate health score based on request patterns
	healthScore := rtm.calculateHealthScore(client, endpoint)
	client.HealthScore = healthScore

	// Check for cached response first
	requestHash := rtm.generateRequestHash(r)
	if cached, exists := rtm.cache[requestHash]; exists {
		if time.Since(cached.Timestamp) < rtm.config.ThrottleCacheDuration {
			return true, 0, cached // Serve cached response
		} else {
			delete(rtm.cache, requestHash) // Expired cache entry
		}
	}

	// Determine throttling strategy based on health score
	shouldThrottle := healthScore > 2
	delay := rtm.calculateDelay(healthScore)

	if shouldThrottle {
		client.ThrottledRequests++
	}

	return shouldThrottle, delay, nil
}

// calculateHealthScore determines client health based on request patterns
func (rtm *RequestThrottlingMiddleware) calculateHealthScore(client *ClientTracker, endpoint string) int {
	score := 0

	// Factor 1: Overall request frequency (requests per minute)
	requestsPerMinute := len(client.RequestTimes)
	if requestsPerMinute > rtm.config.ThrottleRequestsPerMinute {
		score += (requestsPerMinute - rtm.config.ThrottleRequestsPerMinute) / 10
	}

	// Factor 2: Endpoint-specific polling (repeated requests to same endpoint)
	endpointRequests := client.EndpointCount[endpoint]
	if endpointRequests > rtm.config.ThrottlePollingThreshold {
		score += (endpointRequests - rtm.config.ThrottlePollingThreshold) / 3
	}

	// Factor 3: Very high frequency patterns (more than 2 requests/second sustained)
	if requestsPerMinute > 120 { // 2 requests/second for a full minute
		score += 5
	}

	// Cap the score at a reasonable maximum
	if score > 10 {
		score = 10
	}

	return score
}

// calculateDelay determines the delay to apply based on health score
func (rtm *RequestThrottlingMiddleware) calculateDelay(healthScore int) time.Duration {
	if healthScore <= 2 {
		return 0 // No delay for healthy clients
	}

	// Progressive delay calculation
	// Score 3-4: 50-200ms
	// Score 5-6: 200-500ms  
	// Score 7-8: 500-1000ms
	// Score 9-10: 1000-2000ms

	var baseDelay time.Duration
	switch {
	case healthScore <= 4:
		baseDelay = time.Duration(50+(healthScore-3)*75) * time.Millisecond
	case healthScore <= 6:
		baseDelay = time.Duration(200+(healthScore-5)*150) * time.Millisecond
	case healthScore <= 8:
		baseDelay = time.Duration(500+(healthScore-7)*250) * time.Millisecond
	default:
		baseDelay = time.Duration(1000+(healthScore-9)*500) * time.Millisecond
	}

	// Cap at configured maximum
	if baseDelay > rtm.config.ThrottleMaxDelayMs {
		baseDelay = rtm.config.ThrottleMaxDelayMs
	}

	return baseDelay
}

// getClientIP extracts the client IP address from the request
func (rtm *RequestThrottlingMiddleware) getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in the chain
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if colonIndex := strings.LastIndex(ip, ":"); colonIndex != -1 {
		ip = ip[:colonIndex] // Remove port
	}

	return ip
}

// generateRequestHash creates a hash for request caching
func (rtm *RequestThrottlingMiddleware) generateRequestHash(r *http.Request) string {
	// Create hash based on method, path, and key query parameters
	content := fmt.Sprintf("%s:%s:%s", r.Method, r.URL.Path, r.URL.RawQuery)
	hash := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", hash[:8]) // Use first 8 bytes for brevity
}

// serveCachedResponse writes a cached response to the client
func (rtm *RequestThrottlingMiddleware) serveCachedResponse(w http.ResponseWriter, cached *CachedResponse) {
	// Set headers
	for key, value := range cached.Headers {
		w.Header().Set(key, value)
	}
	
	// Add cache indication header
	w.Header().Set("X-EntityDB-Cached", "true")
	w.Header().Set("X-EntityDB-Cache-Age", fmt.Sprintf("%.0f", time.Since(cached.Timestamp).Seconds()))

	// Write status and body
	w.WriteHeader(cached.StatusCode)
	w.Write(cached.Body)
}

// getClientHealthScore returns the current health score for a client
func (rtm *RequestThrottlingMiddleware) getClientHealthScore(clientIP string) int {
	rtm.mu.RLock()
	defer rtm.mu.RUnlock()

	if client, exists := rtm.clients[clientIP]; exists {
		client.mu.RLock()
		defer client.mu.RUnlock()
		return client.HealthScore
	}
	return 0
}

// GetStats returns throttling statistics for monitoring
func (rtm *RequestThrottlingMiddleware) GetStats() map[string]interface{} {
	rtm.mu.RLock()
	defer rtm.mu.RUnlock()

	stats := map[string]interface{}{
		"enabled":      rtm.enabled,
		"total_clients": len(rtm.clients),
		"cached_responses": len(rtm.cache),
	}

	// Calculate aggregate statistics
	totalRequests := 0
	totalThrottled := 0
	activeClients := 0
	fiveMinutesAgo := time.Now().Add(-5 * time.Minute)

	for _, client := range rtm.clients {
		client.mu.RLock()
		totalRequests += client.TotalRequests
		totalThrottled += client.ThrottledRequests
		if client.LastActivity.After(fiveMinutesAgo) {
			activeClients++
		}
		client.mu.RUnlock()
	}

	stats["total_requests"] = totalRequests
	stats["total_throttled"] = totalThrottled
	stats["active_clients"] = activeClients
	if totalRequests > 0 {
		stats["throttle_rate"] = float64(totalThrottled) / float64(totalRequests)
	}

	return stats
}

// CleanupStaleClients removes inactive client trackers to prevent memory leaks
func (rtm *RequestThrottlingMiddleware) CleanupStaleClients() {
	rtm.mu.Lock()
	defer rtm.mu.Unlock()

	staleThreshold := 30 * time.Minute
	staleTime := time.Now().Add(-staleThreshold)

	for ip, client := range rtm.clients {
		client.mu.RLock()
		isStale := client.LastActivity.Before(staleTime)
		client.mu.RUnlock()

		if isStale {
			delete(rtm.clients, ip)
			logger.Debug("Cleaned up stale client tracker for IP: %s", ip)
		}
	}

	// Also cleanup stale cache entries
	cacheThreshold := rtm.config.ThrottleCacheDuration * 2
	cacheExpiry := time.Now().Add(-cacheThreshold)
	for hash, cached := range rtm.cache {
		if cached.Timestamp.Before(cacheExpiry) {
			delete(rtm.cache, hash)
		}
	}
}