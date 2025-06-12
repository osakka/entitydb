// Package config provides centralized configuration management for EntityDB.
//
// This package implements a three-tier configuration hierarchy:
//   1. Database configuration entities (highest priority)
//   2. CLI flags
//   3. Environment variables (lowest priority)
//
// All configuration values are loaded from environment variables with
// sensible defaults. Tools and utilities should use this package for
// consistent configuration across the entire system.
package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all configuration values for EntityDB.
//
// Configuration follows a three-tier hierarchy:
//  1. Database configuration entities (highest priority)
//  2. Command-line flags
//  3. Environment variables (lowest priority)
//
// All values have sensible defaults and can be overridden through
// environment variables or command-line flags.
type Config struct {
	// Server Configuration
	// ===================
	
	// Port is the HTTP server listening port.
	// Environment: ENTITYDB_PORT
	// Default: 8085
	// Valid range: 1-65535
	// Recommendation: Use 8085 for development, 80/443 for production
	Port int
	
	// SSLPort is the HTTPS server listening port.
	// Environment: ENTITYDB_SSL_PORT
	// Default: 8085 (same as Port for development)
	// Valid range: 1-65535
	// Recommendation: Use 8443 for development, 443 for production
	SSLPort int
	
	// UseSSL enables SSL/TLS encryption for all connections.
	// Environment: ENTITYDB_USE_SSL
	// Default: false
	// Recommendation: Always enable for production environments
	// Note: Required for proper CORS functionality with web applications
	UseSSL bool
	
	// SSLCert is the path to the SSL certificate file.
	// Environment: ENTITYDB_SSL_CERT
	// Default: "./certs/server.pem"
	// Format: PEM-encoded X.509 certificate
	SSLCert string
	
	// SSLKey is the path to the SSL private key file.
	// Environment: ENTITYDB_SSL_KEY
	// Default: "./certs/server.key"
	// Format: PEM-encoded private key
	// Security: Ensure proper file permissions (600)
	SSLKey string
	
	// File System Paths
	// =================
	
	// DataPath is the root directory for all EntityDB data files.
	// Environment: ENTITYDB_DATA_PATH
	// Default: "./var"
	// Contains: entities database, WAL files, indexes, metrics
	// Recommendation: Use absolute paths in production
	DataPath string
	
	// StaticDir is the directory containing web UI and static files.
	// Environment: ENTITYDB_STATIC_DIR
	// Default: "./share/htdocs"
	// Contains: dashboard HTML, JavaScript, CSS, Swagger documentation
	StaticDir string
	
	// Security Configuration
	// ======================
	
	// TokenSecret is the secret key used for JWT token signing and validation.
	// Environment: ENTITYDB_TOKEN_SECRET
	// Default: "entitydb-secret-key" (CHANGE IN PRODUCTION)
	// Minimum length: 32 characters
	// Recommendation: Use cryptographically secure random string
	// Security: Never commit production secrets to version control
	TokenSecret string
	
	// SessionTTLHours defines session timeout in hours.
	// Environment: ENTITYDB_SESSION_TTL_HOURS
	// Default: 2 hours
	// Valid range: 1-168 (1 week maximum)
	// Recommendation: 2-8 hours for web applications, 1-2 hours for APIs
	SessionTTLHours int
	
	// HTTP Server Timeouts
	// ====================
	
	// HTTPReadTimeout is the maximum duration for reading the entire request.
	// Environment: ENTITYDB_HTTP_READ_TIMEOUT (seconds)
	// Default: 15 seconds
	// Recommendation: 15-30 seconds for most applications
	HTTPReadTimeout time.Duration
	
	// HTTPWriteTimeout is the maximum duration before timing out writes.
	// Environment: ENTITYDB_HTTP_WRITE_TIMEOUT (seconds)
	// Default: 15 seconds
	// Recommendation: 15-60 seconds depending on expected response sizes
	HTTPWriteTimeout time.Duration
	
	// HTTPIdleTimeout is the maximum time to wait for the next request.
	// Environment: ENTITYDB_HTTP_IDLE_TIMEOUT (seconds)
	// Default: 60 seconds
	// Recommendation: 60-120 seconds for connection reuse optimization
	HTTPIdleTimeout time.Duration
	
	// ShutdownTimeout is the maximum time to wait for graceful shutdown.
	// Environment: ENTITYDB_SHUTDOWN_TIMEOUT (seconds)
	// Default: 30 seconds
	// Recommendation: 30-60 seconds to allow active requests to complete
	ShutdownTimeout time.Duration
	
	// Metrics Collection Configuration
	// ================================
	
	// MetricsInterval defines how often to collect system metrics.
	// Environment: ENTITYDB_METRICS_INTERVAL (seconds)
	// Default: 30 seconds
	// Recommendation: 10-60 seconds depending on monitoring requirements
	MetricsInterval time.Duration
	
	// AggregationInterval defines how often to aggregate collected metrics.
	// Environment: ENTITYDB_METRICS_AGGREGATION_INTERVAL (seconds)
	// Default: 30 seconds
	// Should match or be multiple of MetricsInterval
	AggregationInterval time.Duration
	
	// Advanced Metrics Configuration
	// ==============================
	
	// MetricsRetentionRaw defines retention period for raw metric data.
	// Environment: ENTITYDB_METRICS_RETENTION_RAW (minutes)
	// Default: 1440 minutes (24 hours)
	// Recommendation: 1-7 days depending on storage capacity
	MetricsRetentionRaw time.Duration
	
	// MetricsRetention1Min defines retention period for 1-minute aggregates.
	// Environment: ENTITYDB_METRICS_RETENTION_1MIN (minutes)
	// Default: 10080 minutes (7 days)
	// Recommendation: 7-30 days for operational monitoring
	MetricsRetention1Min time.Duration
	
	// MetricsRetention1Hour defines retention period for 1-hour aggregates.
	// Environment: ENTITYDB_METRICS_RETENTION_1HOUR (minutes)
	// Default: 43200 minutes (30 days)
	// Recommendation: 30-90 days for trend analysis
	MetricsRetention1Hour time.Duration
	
	// MetricsRetention1Day defines retention period for daily aggregates.
	// Environment: ENTITYDB_METRICS_RETENTION_1DAY (minutes)
	// Default: 525600 minutes (365 days)
	// Recommendation: 365-1095 days for long-term analysis
	MetricsRetention1Day time.Duration
	
	// MetricsHistogramBuckets defines latency histogram bucket boundaries in seconds.
	// Environment: ENTITYDB_METRICS_HISTOGRAM_BUCKETS (comma-separated floats)
	// Default: [0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 5, 10]
	// Covers: 1ms to 10s latency ranges
	// Recommendation: Adjust based on expected latency characteristics
	MetricsHistogramBuckets []float64
	
	// MetricsEnableRequestTracking enables HTTP request metrics collection.
	// Environment: ENTITYDB_METRICS_ENABLE_REQUEST_TRACKING
	// Default: true
	// Impact: Minimal performance overhead (<1%)
	MetricsEnableRequestTracking bool
	
	// MetricsEnableStorageTracking enables storage operation metrics collection.
	// Environment: ENTITYDB_METRICS_ENABLE_STORAGE_TRACKING
	// Default: true
	// Impact: Minimal performance overhead (<2%)
	MetricsEnableStorageTracking bool
	
	// API Documentation Configuration
	// ===============================
	
	// SwaggerHost defines the host:port for Swagger API documentation.
	// Environment: ENTITYDB_SWAGGER_HOST
	// Default: "localhost:8085"
	// Format: "hostname:port" (no protocol)
	// Used in generated OpenAPI specifications
	SwaggerHost string
	
	// Logging Configuration
	// =====================
	
	// LogLevel sets the minimum log level for message output.
	// Environment: ENTITYDB_LOG_LEVEL
	// Default: "info"
	// Valid values: "trace", "debug", "info", "warn", "error"
	// Recommendation: "info" for production, "debug" for development
	LogLevel string
	
	// Performance Configuration
	// =========================
	
	// HighPerformance enables optimizations for high-throughput scenarios.
	// Environment: ENTITYDB_HIGH_PERFORMANCE
	// Default: false
	// Features: Memory-mapped files, optimized indexing, reduced safety checks
	// Trade-off: Higher memory usage, faster query processing
	// Recommendation: Enable for read-heavy workloads with sufficient RAM
	HighPerformance bool
	
	// Rate Limiting Configuration
	// ===========================
	
	// EnableRateLimit activates request rate limiting per client IP.
	// Environment: ENTITYDB_ENABLE_RATE_LIMIT
	// Default: false
	// Recommendation: Enable for public-facing deployments
	EnableRateLimit bool
	
	// RateLimitRequests defines maximum requests allowed per window.
	// Environment: ENTITYDB_RATE_LIMIT_REQUESTS
	// Default: 100
	// Applies per client IP address
	// Recommendation: 100-1000 depending on expected usage patterns
	RateLimitRequests int
	
	// RateLimitWindowMinutes defines the rate limiting window duration.
	// Environment: ENTITYDB_RATE_LIMIT_WINDOW_MINUTES
	// Default: 1 minute
	// Recommendation: 1-5 minutes for API protection
	RateLimitWindowMinutes int
	
	// Application Metadata
	// ====================
	
	// AppName is the application name used in logs and responses.
	// Environment: ENTITYDB_APP_NAME
	// Default: "EntityDB Server"
	// Used in HTTP headers and API responses
	AppName string
	
	// AppVersion is the application version for API documentation.
	// Environment: ENTITYDB_APP_VERSION
	// Default: "2.28.0"
	// Should match build version for consistency
	AppVersion string
}

// Load creates a new Config instance with values loaded from environment variables.
//
// This function applies the lowest priority tier of the configuration hierarchy,
// reading from environment variables with sensible defaults. Values returned
// by this function can be overridden by command-line flags or database configuration.
//
// Environment Variable Format:
//   All environment variables use the ENTITYDB_ prefix followed by uppercase
//   parameter names with underscores. For example:
//     - ENTITYDB_PORT=8085
//     - ENTITYDB_USE_SSL=true
//     - ENTITYDB_LOG_LEVEL=debug
//
// Duration Values:
//   Timeout and interval values are specified in seconds as integers.
//   They are automatically converted to time.Duration internally.
//
// Boolean Values:
//   Accept "true", "1" for true; anything else is considered false.
//
// Array Values:
//   Comma-separated values for histogram buckets and similar arrays.
//
// Returns:
//   A new Config instance with all values populated from environment
//   variables or their documented defaults.
func Load() *Config {
	return &Config{
		// Server Configuration
		Port:             getEnvInt("ENTITYDB_PORT", 8085),
		SSLPort:          getEnvInt("ENTITYDB_SSL_PORT", 8085),
		UseSSL:           getEnvBool("ENTITYDB_USE_SSL", false),
		SSLCert:          getEnv("ENTITYDB_SSL_CERT", "./certs/server.pem"),
		SSLKey:           getEnv("ENTITYDB_SSL_KEY", "./certs/server.key"),
		
		// Paths - use relative paths as defaults
		DataPath:         getEnv("ENTITYDB_DATA_PATH", "./var"),
		StaticDir:        getEnv("ENTITYDB_STATIC_DIR", "./share/htdocs"),
		
		// Security
		TokenSecret:      getEnv("ENTITYDB_TOKEN_SECRET", "entitydb-secret-key"),
		SessionTTLHours:  getEnvInt("ENTITYDB_SESSION_TTL_HOURS", 2),
		
		// Timeouts
		HTTPReadTimeout:  getEnvDuration("ENTITYDB_HTTP_READ_TIMEOUT", 15),
		HTTPWriteTimeout: getEnvDuration("ENTITYDB_HTTP_WRITE_TIMEOUT", 15),
		HTTPIdleTimeout:  getEnvDuration("ENTITYDB_HTTP_IDLE_TIMEOUT", 60),
		ShutdownTimeout:  getEnvDuration("ENTITYDB_SHUTDOWN_TIMEOUT", 30),
		
		// Metrics
		MetricsInterval:  getEnvDuration("ENTITYDB_METRICS_INTERVAL", 30),
		AggregationInterval: getEnvDuration("ENTITYDB_METRICS_AGGREGATION_INTERVAL", 30),
		
		// Enhanced Metrics Configuration
		MetricsRetentionRaw:    getEnvDuration("ENTITYDB_METRICS_RETENTION_RAW", 24*60), // 24 hours in minutes
		MetricsRetention1Min:   getEnvDuration("ENTITYDB_METRICS_RETENTION_1MIN", 7*24*60), // 7 days in minutes
		MetricsRetention1Hour:  getEnvDuration("ENTITYDB_METRICS_RETENTION_1HOUR", 30*24*60), // 30 days in minutes
		MetricsRetention1Day:   getEnvDuration("ENTITYDB_METRICS_RETENTION_1DAY", 365*24*60), // 365 days in minutes
		MetricsHistogramBuckets: getEnvFloatSlice("ENTITYDB_METRICS_HISTOGRAM_BUCKETS", []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 5, 10}),
		MetricsEnableRequestTracking: getEnvBool("ENTITYDB_METRICS_ENABLE_REQUEST_TRACKING", true),
		MetricsEnableStorageTracking: getEnvBool("ENTITYDB_METRICS_ENABLE_STORAGE_TRACKING", true),
		
		// API
		SwaggerHost:      getEnv("ENTITYDB_SWAGGER_HOST", "localhost:8085"),
		
		// Logging
		LogLevel:         getEnv("ENTITYDB_LOG_LEVEL", "info"),
		
		// Performance
		HighPerformance:  getEnvBool("ENTITYDB_HIGH_PERFORMANCE", false),
		
		// Rate Limiting
		EnableRateLimit:  getEnvBool("ENTITYDB_ENABLE_RATE_LIMIT", false),
		RateLimitRequests: getEnvInt("ENTITYDB_RATE_LIMIT_REQUESTS", 100),
		RateLimitWindowMinutes: getEnvInt("ENTITYDB_RATE_LIMIT_WINDOW_MINUTES", 1),
		
		// Application Info
		AppName:          getEnv("ENTITYDB_APP_NAME", "EntityDB Server"),
		AppVersion:       getEnv("ENTITYDB_APP_VERSION", "2.28.0"),
	}
}

// DatabasePath returns the full path to the main EntityDB database file.
//
// The database file uses the custom EBF (EntityDB Binary Format) and contains
// all entity data, relationships, and metadata. This path is constructed by
// combining the configured DataPath with the standard database subdirectory
// and filename.
//
// Path Structure:
//   {DataPath}/data/entities.db
//
// For example, with default DataPath of "./var":
//   ./var/data/entities.db
//
// The database file is accompanied by related files in the same directory:
//   - entities.db      - Main database file (EBF format)
//   - entities.db.wal  - Write-Ahead Log for durability
//   - *.idx            - Various index files for performance
//
// Returns:
//   Complete filesystem path to the EntityDB database file.
//
// Note:
//   The parent directory (DataPath/data) must exist and be writable
//   by the EntityDB process. The server will create the database file
//   if it doesn't exist but won't create parent directories.
func (c *Config) DatabasePath() string {
	return c.DataPath + "/data/entities.db"
}

// =============================================================================
// Environment Variable Parsing Utilities
// =============================================================================
//
// These helper functions provide type-safe parsing of environment variables
// with fallback to default values when variables are unset or invalid.
// All functions follow the pattern of returning the default value if the
// environment variable is missing or cannot be parsed.

// getEnv retrieves a string environment variable with a default fallback.
//
// Parameters:
//   key - Environment variable name
//   defaultValue - Value to return if variable is unset or empty
//
// Returns:
//   Environment variable value or defaultValue if unset/empty
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt retrieves an integer environment variable with a default fallback.
//
// The function attempts to parse the environment variable as an integer using
// strconv.Atoi. If parsing fails, the default value is returned.
//
// Parameters:
//   key - Environment variable name
//   defaultValue - Value to return if variable is unset or invalid
//
// Returns:
//   Parsed integer value or defaultValue if unset/invalid
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvBool retrieves a boolean environment variable with a default fallback.
//
// The function considers "true" and "1" as true values; all other values
// (including empty string) are considered false.
//
// Parameters:
//   key - Environment variable name
//   defaultValue - Value to return if variable is unset
//
// Returns:
//   Boolean value based on environment variable or defaultValue if unset
//
// Examples:
//   ENTITYDB_USE_SSL=true  -> true
//   ENTITYDB_USE_SSL=1     -> true
//   ENTITYDB_USE_SSL=false -> false
//   ENTITYDB_USE_SSL=0     -> false
//   ENTITYDB_USE_SSL=      -> defaultValue
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true" || value == "1"
	}
	return defaultValue
}

// getEnvDuration retrieves a duration environment variable with a default fallback.
//
// The function expects the environment variable to contain an integer representing
// seconds, which is then converted to a time.Duration. This simplifies configuration
// while providing the flexibility of duration types internally.
//
// Parameters:
//   key - Environment variable name
//   defaultSeconds - Default duration in seconds if variable is unset/invalid
//
// Returns:
//   time.Duration based on environment variable or default
//
// Examples:
//   ENTITYDB_HTTP_READ_TIMEOUT=30 -> 30 * time.Second
//   ENTITYDB_HTTP_READ_TIMEOUT=   -> defaultSeconds * time.Second
func getEnvDuration(key string, defaultSeconds int) time.Duration {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return time.Duration(intValue) * time.Second
		}
	}
	return time.Duration(defaultSeconds) * time.Second
}

// getEnvFloatSlice retrieves a comma-separated float slice with a default fallback.
//
// The function parses a comma-separated list of floating-point numbers from
// the environment variable. Invalid numbers are skipped, and if no valid
// numbers are found, the default value is returned.
//
// Parameters:
//   key - Environment variable name
//   defaultValue - Default slice to return if variable is unset/invalid
//
// Returns:
//   Slice of float64 values parsed from environment variable or defaultValue
//
// Examples:
//   ENTITYDB_METRICS_HISTOGRAM_BUCKETS="0.001,0.01,0.1,1.0" -> [0.001, 0.01, 0.1, 1.0]
//   ENTITYDB_METRICS_HISTOGRAM_BUCKETS="0.001,invalid,0.1"  -> [0.001, 0.1]
//   ENTITYDB_METRICS_HISTOGRAM_BUCKETS=""                   -> defaultValue
func getEnvFloatSlice(key string, defaultValue []float64) []float64 {
	if value := os.Getenv(key); value != "" {
		parts := strings.Split(value, ",")
		result := make([]float64, 0, len(parts))
		for _, part := range parts {
			if f, err := strconv.ParseFloat(strings.TrimSpace(part), 64); err == nil {
				result = append(result, f)
			}
		}
		if len(result) > 0 {
			return result
		}
	}
	return defaultValue
}