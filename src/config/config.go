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
	
	// Specific File Path Configuration
	// ================================
	// These paths are used literally by the binary - no path joining
	
	// DatabaseFilename is the full path to the main database file.
	// Environment: ENTITYDB_DATABASE_FILE
	// Default: "./var/entities.edb"
	// Used for the main entity storage file (unified format)
	DatabaseFilename string
	
	// WALFilename is the full path to the Write-Ahead Log file.
	// Environment: ENTITYDB_WAL_FILE
	// Default: "./var/entitydb.wal"
	// Used for transactional safety and recovery
	WALFilename string
	
	// IndexFilename is the full path to the tag index file.
	// Environment: ENTITYDB_INDEX_FILE
	// Default: "./var/entities.edb" (embedded in unified file)
	// Used for fast tag-based queries
	IndexFilename string
	
	// MetricsFilename is the full path to the metrics storage file.
	// Environment: ENTITYDB_METRICS_FILE
	// Default: "./var/metrics.json"
	// Used for metrics persistence
	MetricsFilename string
	
	
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

	// MetricsGentlePauseMs defines pause duration between metric collection blocks.
	// Environment: ENTITYDB_METRICS_GENTLE_PAUSE_MS (milliseconds)
	// Default: 100 milliseconds
	// Purpose: Smooth CPU usage spikes during metrics collection
	// Recommendation: 50-200ms to balance smoothness vs collection speed
	MetricsGentlePauseMs time.Duration

	// Request Throttling Configuration
	// =================================
	
	// ThrottleEnabled enables intelligent request throttling and abuse protection.
	// Environment: ENTITYDB_THROTTLE_ENABLED
	// Default: true
	// Purpose: Protect against aggressive UI polling and request abuse patterns
	ThrottleEnabled bool

	// ThrottleRequestsPerMinute defines the baseline requests/minute before throttling activates.
	// Environment: ENTITYDB_THROTTLE_REQUESTS_PER_MINUTE
	// Default: 60 requests/minute (1 request/second average)
	// Purpose: Detect aggressive polling patterns
	ThrottleRequestsPerMinute int

	// ThrottlePollingThreshold defines repeated requests to same endpoint that trigger throttling.
	// Environment: ENTITYDB_THROTTLE_POLLING_THRESHOLD
	// Default: 10 requests to same endpoint within time window
	// Purpose: Detect UI polling patterns specifically
	ThrottlePollingThreshold int

	// ThrottleMaxDelayMs defines maximum delay applied to throttled requests.
	// Environment: ENTITYDB_THROTTLE_MAX_DELAY_MS (milliseconds)
	// Default: 2000ms (2 seconds)
	// Purpose: Prevent complete blocking while discouraging abuse
	ThrottleMaxDelayMs time.Duration

	// ThrottleCacheDuration defines how long to cache responses for repeated requests.
	// Environment: ENTITYDB_THROTTLE_CACHE_DURATION (seconds)
	// Default: 30 seconds
	// Purpose: Serve cached responses to rapid repeated requests
	ThrottleCacheDuration time.Duration
	
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
	// Default: false (disabled to prevent feedback loops)
	// Impact: When enabled, minimal performance overhead (<1%)
	// Security: Can create entity storage loops if misconfigured
	MetricsEnableRequestTracking bool
	
	// MetricsEnableStorageTracking enables storage operation metrics collection.
	// Environment: ENTITYDB_METRICS_ENABLE_STORAGE_TRACKING
	// Default: false (disabled to prevent feedback loops)
	// Impact: When enabled, minimal performance overhead (<2%)
	// Security: Can create metric recursion if misconfigured
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
	
	// StringCacheSize defines maximum number of interned strings.
	// Environment: ENTITYDB_STRING_CACHE_SIZE
	// Default: 100000 (100k strings)
	// Purpose: Prevent unbounded memory growth from string interning
	// Recommendation: Adjust based on workload - higher for tag-heavy usage
	StringCacheSize int
	
	// StringCacheMemoryLimit defines memory limit for interned strings in bytes.
	// Environment: ENTITYDB_STRING_CACHE_MEMORY_LIMIT
	// Default: 104857600 (100MB)
	// Purpose: Secondary limit to prevent memory exhaustion
	// Recommendation: Set to 5-10% of available memory
	StringCacheMemoryLimit int64
	
	// EntityCacheSize defines maximum number of cached entities.
	// Environment: ENTITYDB_ENTITY_CACHE_SIZE
	// Default: 10000 (10k entities)
	// Purpose: Prevent unbounded memory growth from entity caching
	// Recommendation: Higher values for read-heavy workloads
	EntityCacheSize int
	
	// EntityCacheMemoryLimit defines memory limit for cached entities in bytes.
	// Environment: ENTITYDB_ENTITY_CACHE_MEMORY_LIMIT
	// Default: 1073741824 (1GB)
	// Purpose: Prevent memory exhaustion from large entities
	// Recommendation: Set to 10-20% of available memory
	EntityCacheMemoryLimit int64
	
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
	// Default: "2.32.5"
	// Should match build version for consistency
	AppVersion string
	
	// Default Admin User Configuration
	// ================================
	
	// DefaultAdminUsername is the username for the default admin user.
	// Environment: ENTITYDB_DEFAULT_ADMIN_USERNAME
	// Default: "admin"
	// Security: Change in production environments
	DefaultAdminUsername string
	
	// DefaultAdminPassword is the password for the default admin user.
	// Environment: ENTITYDB_DEFAULT_ADMIN_PASSWORD
	// Default: "admin"
	// Security: MUST be changed in production environments
	DefaultAdminPassword string
	
	// DefaultAdminEmail is the email address for the default admin user.
	// Environment: ENTITYDB_DEFAULT_ADMIN_EMAIL
	// Default: "admin@entitydb.local"
	// Used for admin notifications and account identification
	DefaultAdminEmail string
	
	// System User Configuration
	// =========================
	
	// SystemUserID is the immutable UUID for the system user.
	// Environment: ENTITYDB_SYSTEM_USER_ID
	// Default: "00000000000000000000000000000001"
	// Warning: Changing this in existing databases requires migration
	SystemUserID string
	
	// SystemUsername is the username for the system user.
	// Environment: ENTITYDB_SYSTEM_USERNAME
	// Default: "system"
	// Used in identity tags and system operations
	SystemUsername string
	
	// Advanced Security Configuration
	// ===============================
	
	// BcryptCost is the computational cost for bcrypt password hashing.
	// Environment: ENTITYDB_BCRYPT_COST
	// Default: 10 (bcrypt.DefaultCost)
	// Valid range: 4-31 (higher = more secure but slower)
	// Recommendation: 10-12 for most applications
	BcryptCost int
	
	// File and Path Configuration
	// ===========================
	
	
	// WALSuffix is the suffix appended to database path for WAL files.
	// Environment: ENTITYDB_WAL_SUFFIX
	// Default: "" (embedded in unified file)
	// Example: entities.edb (unified format)
	WALSuffix string
	
	// IndexSuffix is the suffix for index files.
	// Environment: ENTITYDB_INDEX_SUFFIX
	// Default: "" (embedded in unified file)
	// Example: entities.edb (unified format)
	IndexSuffix string
	
	// BackupPath is the directory for database backups.
	// Environment: ENTITYDB_BACKUP_PATH
	// Default: "./backup"
	// Relative to DataPath or absolute path
	BackupPath string
	
	// BackupInterval defines how often routine backups are created.
	// Environment: ENTITYDB_BACKUP_INTERVAL
	// Default: 1h (1 hour)
	// Examples: 5m, 30m, 1h, 2h, 24h
	BackupInterval time.Duration
	
	// BackupRetentionHours defines how many hourly backups to keep.
	// Environment: ENTITYDB_BACKUP_RETENTION_HOURS
	// Default: 24 (keep last 24 hourly backups)
	BackupRetentionHours int
	
	// BackupRetentionDays defines how many daily backups to keep.
	// Environment: ENTITYDB_BACKUP_RETENTION_DAYS
	// Default: 7 (keep last 7 daily backups)
	BackupRetentionDays int
	
	// BackupRetentionWeeks defines how many weekly backups to keep.
	// Environment: ENTITYDB_BACKUP_RETENTION_WEEKS
	// Default: 4 (keep last 4 weekly backups)
	BackupRetentionWeeks int
	
	// BackupMaxSize defines maximum total size of all backups in MB.
	// Environment: ENTITYDB_BACKUP_MAX_SIZE_MB
	// Default: 1000 (1GB)
	// When exceeded, oldest backups are removed
	BackupMaxSizeMB int64
	
	// TempPath is the directory for temporary files.
	// Environment: ENTITYDB_TEMP_PATH
	// Default: "./tmp"
	// Relative to DataPath or absolute path
	TempPath string
	
	// PIDFile is the path to the server process ID file.
	// Environment: ENTITYDB_PID_FILE
	// Default: "./var/entitydb.pid"
	// Used by daemon scripts for process management
	PIDFile string
	
	// LogFile is the path to the server log file.
	// Environment: ENTITYDB_LOG_FILE
	// Default: "./var/entitydb.log"
	// Used when running as daemon
	LogFile string
	
	// Development and Debugging Configuration
	// ======================================
	
	// DevMode enables development mode features.
	// Environment: ENTITYDB_DEV_MODE
	// Default: false
	// Enables additional logging, debug endpoints, relaxed security
	DevMode bool
	
	// DebugPort is the port for debug/pprof endpoints.
	// Environment: ENTITYDB_DEBUG_PORT
	// Default: 6060
	// Only active when DevMode is enabled
	DebugPort int
	
	// ProfileEnabled enables CPU and memory profiling.
	// Environment: ENTITYDB_PROFILE_ENABLED
	// Default: false
	// Useful for performance analysis
	ProfileEnabled bool
	
	// Trace Subsystems Configuration
	// ==============================
	
	// TraceSubsystems is a comma-separated list of trace subsystems to enable.
	// Environment: ENTITYDB_TRACE_SUBSYSTEMS
	// Default: "" (none enabled)
	// Available: auth, storage, wal, chunking, metrics, locks, query, dataset, relationship, temporal
	TraceSubsystems string
	
	// Deletion Collector Configuration
	// ================================
	
	// DeletionCollectorEnabled controls whether the deletion collector runs.
	// Environment: ENTITYDB_DELETION_COLLECTOR_ENABLED
	// Default: true
	// Purpose: Enables automatic entity lifecycle management based on retention policies
	DeletionCollectorEnabled bool
	
	// DeletionCollectorInterval defines how often the collector runs.
	// Environment: ENTITYDB_DELETION_COLLECTOR_INTERVAL (seconds)
	// Default: 3600 seconds (1 hour)
	// Purpose: Controls frequency of retention policy evaluation
	DeletionCollectorInterval time.Duration
	
	// DeletionCollectorBatchSize limits entities processed per cycle.
	// Environment: ENTITYDB_DELETION_COLLECTOR_BATCH_SIZE
	// Default: 100
	// Purpose: Controls memory usage and processing chunks
	DeletionCollectorBatchSize int
	
	// DeletionCollectorMaxRuntime limits single collection cycle duration.
	// Environment: ENTITYDB_DELETION_COLLECTOR_MAX_RUNTIME (seconds)
	// Default: 1800 seconds (30 minutes)
	// Purpose: Prevents collection cycles from running too long
	DeletionCollectorMaxRuntime time.Duration
	
	// DeletionCollectorDryRun enables dry run mode (logs without changes).
	// Environment: ENTITYDB_DELETION_COLLECTOR_DRY_RUN
	// Default: false
	// Purpose: Test retention policies without actual modifications
	DeletionCollectorDryRun bool
	
	// DeletionCollectorConcurrency controls parallel entity processing.
	// Environment: ENTITYDB_DELETION_COLLECTOR_CONCURRENCY
	// Default: 4
	// Purpose: Balance performance vs resource usage
	DeletionCollectorConcurrency int
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
		
		// Specific file paths - binary uses these literally
		DatabaseFilename: getEnv("ENTITYDB_DATABASE_FILE", "./var/entities.edb"),
		WALFilename:      getEnv("ENTITYDB_WAL_FILE", "./var/entitydb.wal"),
		IndexFilename:    getEnv("ENTITYDB_INDEX_FILE", "./var/entities.edb.idx"),
		MetricsFilename:  getEnv("ENTITYDB_METRICS_FILE", "./var/metrics.json"),
		
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
		MetricsGentlePauseMs: getEnvDurationMs("ENTITYDB_METRICS_GENTLE_PAUSE_MS", 100),
		AggregationInterval: getEnvDuration("ENTITYDB_METRICS_AGGREGATION_INTERVAL", 30),

		// Request Throttling
		ThrottleEnabled:           getEnvBool("ENTITYDB_THROTTLE_ENABLED", true),
		ThrottleRequestsPerMinute: getEnvInt("ENTITYDB_THROTTLE_REQUESTS_PER_MINUTE", 60),
		ThrottlePollingThreshold:  getEnvInt("ENTITYDB_THROTTLE_POLLING_THRESHOLD", 10),
		ThrottleMaxDelayMs:        getEnvDurationMs("ENTITYDB_THROTTLE_MAX_DELAY_MS", 2000),
		ThrottleCacheDuration:     getEnvDuration("ENTITYDB_THROTTLE_CACHE_DURATION", 30),
		
		// Enhanced Metrics Configuration
		MetricsRetentionRaw:    getEnvDuration("ENTITYDB_METRICS_RETENTION_RAW", 24*60), // 24 hours in minutes
		MetricsRetention1Min:   getEnvDuration("ENTITYDB_METRICS_RETENTION_1MIN", 2*60), // 2 hours in minutes (was 7 days - performance fix)
		MetricsRetention1Hour:  getEnvDuration("ENTITYDB_METRICS_RETENTION_1HOUR", 24*60), // 24 hours in minutes (was 30 days - performance fix)
		MetricsRetention1Day:   getEnvDuration("ENTITYDB_METRICS_RETENTION_1DAY", 7*24*60), // 7 days in minutes (was 365 days - performance fix)
		MetricsHistogramBuckets: getEnvFloatSlice("ENTITYDB_METRICS_HISTOGRAM_BUCKETS", []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 5, 10}),
		MetricsEnableRequestTracking: getEnvBool("ENTITYDB_METRICS_ENABLE_REQUEST_TRACKING", false),
		MetricsEnableStorageTracking: getEnvBool("ENTITYDB_METRICS_ENABLE_STORAGE_TRACKING", false),
		
		// API
		SwaggerHost:      getEnv("ENTITYDB_SWAGGER_HOST", "localhost:8085"),
		
		// Logging
		LogLevel:         getEnv("ENTITYDB_LOG_LEVEL", "info"),
		
		// Performance
		HighPerformance:  getEnvBool("ENTITYDB_HIGH_PERFORMANCE", false),
		StringCacheSize: getEnvInt("ENTITYDB_STRING_CACHE_SIZE", 100000),
		StringCacheMemoryLimit: getEnvInt64("ENTITYDB_STRING_CACHE_MEMORY_LIMIT", 100*1024*1024),
		EntityCacheSize: getEnvInt("ENTITYDB_ENTITY_CACHE_SIZE", 10000),
		EntityCacheMemoryLimit: getEnvInt64("ENTITYDB_ENTITY_CACHE_MEMORY_LIMIT", 1024*1024*1024),
		
		// Rate Limiting
		EnableRateLimit:  getEnvBool("ENTITYDB_ENABLE_RATE_LIMIT", false),
		RateLimitRequests: getEnvInt("ENTITYDB_RATE_LIMIT_REQUESTS", 100),
		RateLimitWindowMinutes: getEnvInt("ENTITYDB_RATE_LIMIT_WINDOW_MINUTES", 1),
		
		// Application Info
		AppName:          getEnv("ENTITYDB_APP_NAME", "EntityDB Server"),
		AppVersion:       getEnv("ENTITYDB_APP_VERSION", "2.34.0"),
		
		// Default Admin User Configuration
		DefaultAdminUsername: getEnv("ENTITYDB_DEFAULT_ADMIN_USERNAME", "admin"),
		DefaultAdminPassword: getEnv("ENTITYDB_DEFAULT_ADMIN_PASSWORD", "admin"),
		DefaultAdminEmail:    getEnv("ENTITYDB_DEFAULT_ADMIN_EMAIL", "admin@entitydb.local"),
		
		// System User Configuration
		SystemUserID:    getEnv("ENTITYDB_SYSTEM_USER_ID", "00000000000000000000000000000001"),
		SystemUsername:  getEnv("ENTITYDB_SYSTEM_USERNAME", "system"),
		
		// Advanced Security Configuration
		BcryptCost:      getEnvInt("ENTITYDB_BCRYPT_COST", 10),
		
		// File and Path Configuration
		WALSuffix:        getEnv("ENTITYDB_WAL_SUFFIX", ".wal"),
		IndexSuffix:      getEnv("ENTITYDB_INDEX_SUFFIX", ".idx"),
		BackupPath:       getEnv("ENTITYDB_BACKUP_PATH", "./backup"),
		BackupInterval:   getEnvDuration("ENTITYDB_BACKUP_INTERVAL", 3600), // 1 hour
		BackupRetentionHours: getEnvInt("ENTITYDB_BACKUP_RETENTION_HOURS", 24),
		BackupRetentionDays:  getEnvInt("ENTITYDB_BACKUP_RETENTION_DAYS", 7),
		BackupRetentionWeeks: getEnvInt("ENTITYDB_BACKUP_RETENTION_WEEKS", 4),
		BackupMaxSizeMB:      getEnvInt64("ENTITYDB_BACKUP_MAX_SIZE_MB", 1000),
		TempPath:         getEnv("ENTITYDB_TEMP_PATH", "./tmp"),
		PIDFile:          getEnv("ENTITYDB_PID_FILE", "./var/entitydb.pid"),
		LogFile:          getEnv("ENTITYDB_LOG_FILE", "./var/entitydb.log"),
		
		// Development and Debugging
		DevMode:          getEnvBool("ENTITYDB_DEV_MODE", false),
		DebugPort:        getEnvInt("ENTITYDB_DEBUG_PORT", 6060),
		ProfileEnabled:   getEnvBool("ENTITYDB_PROFILE_ENABLED", false),
		
		// Trace Subsystems
		TraceSubsystems:  getEnv("ENTITYDB_TRACE_SUBSYSTEMS", ""),
		
		// Deletion Collector
		DeletionCollectorEnabled:     getEnvBool("ENTITYDB_DELETION_COLLECTOR_ENABLED", true),
		DeletionCollectorInterval:    getEnvDuration("ENTITYDB_DELETION_COLLECTOR_INTERVAL", 3600),
		DeletionCollectorBatchSize:   getEnvInt("ENTITYDB_DELETION_COLLECTOR_BATCH_SIZE", 100),
		DeletionCollectorMaxRuntime:  getEnvDuration("ENTITYDB_DELETION_COLLECTOR_MAX_RUNTIME", 1800),
		DeletionCollectorDryRun:      getEnvBool("ENTITYDB_DELETION_COLLECTOR_DRY_RUN", false),
		DeletionCollectorConcurrency: getEnvInt("ENTITYDB_DELETION_COLLECTOR_CONCURRENCY", 4),
	}
}



// BackupFullPath returns the full path to the backup directory.
//
// If BackupPath is relative, it's resolved relative to DataPath.
// If BackupPath is absolute, it's used as-is.
//
// Returns:
//   Complete filesystem path to the backup directory
//
// Example:
//   ./var/backup (relative)
//   /opt/entitydb/backup (absolute)
func (c *Config) BackupFullPath() string {
	if strings.HasPrefix(c.BackupPath, "/") {
		return c.BackupPath
	}
	return c.DataPath + "/" + strings.TrimPrefix(c.BackupPath, "./")
}

// TempFullPath returns the full path to the temporary files directory.
//
// If TempPath is relative, it's resolved relative to DataPath.
// If TempPath is absolute, it's used as-is.
//
// Returns:
//   Complete filesystem path to the temporary directory
//
// Example:
//   ./var/tmp (relative)
//   /tmp/entitydb (absolute)
func (c *Config) TempFullPath() string {
	if strings.HasPrefix(c.TempPath, "/") {
		return c.TempPath
	}
	return c.DataPath + "/" + strings.TrimPrefix(c.TempPath, "./")
}

// PIDFullPath returns the full path to the PID file.
//
// If PIDFile is relative, it's resolved relative to DataPath.
// If PIDFile is absolute, it's used as-is.
//
// Returns:
//   Complete filesystem path to the PID file
func (c *Config) PIDFullPath() string {
	if strings.HasPrefix(c.PIDFile, "/") {
		return c.PIDFile
	}
	return c.DataPath + "/" + strings.TrimPrefix(c.PIDFile, "./")
}

// LogFullPath returns the full path to the log file.
//
// If LogFile is relative, it's resolved relative to DataPath.
// If LogFile is absolute, it's used as-is.
//
// Returns:
//   Complete filesystem path to the log file
func (c *Config) LogFullPath() string {
	if strings.HasPrefix(c.LogFile, "/") {
		return c.LogFile
	}
	return c.DataPath + "/" + strings.TrimPrefix(c.LogFile, "./")
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

// getEnvInt64 retrieves an int64 environment variable with a default fallback.
// Supports parsing of large numbers for memory limits and similar values.
//
// Parameters:
//   key - Environment variable name
//   defaultValue - Value to return if variable is unset or invalid
//
// Returns:
//   Parsed int64 value or defaultValue if unset/invalid
func getEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if int64Value, err := strconv.ParseInt(value, 10, 64); err == nil {
			return int64Value
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

// getEnvDurationMs retrieves a duration environment variable in milliseconds with a default fallback.
//
// The function expects the environment variable to contain an integer representing
// milliseconds, which is then converted to a time.Duration. This provides fine-grained
// control for timing-sensitive operations like gentle pacing.
//
// Parameters:
//   key - Environment variable name
//   defaultMs - Default duration in milliseconds if variable is unset/invalid
//
// Returns:
//   time.Duration based on environment variable or default
//
// Examples:
//   ENTITYDB_METRICS_GENTLE_PAUSE_MS=100 -> 100 * time.Millisecond
//   ENTITYDB_METRICS_GENTLE_PAUSE_MS=    -> defaultMs * time.Millisecond
func getEnvDurationMs(key string, defaultMs int) time.Duration {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return time.Duration(intValue) * time.Millisecond
		}
	}
	return time.Duration(defaultMs) * time.Millisecond
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