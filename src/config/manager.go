// Package config provides centralized configuration management for EntityDB.
package config

import (
	"entitydb/models"
	"entitydb/logger"
	"encoding/json"
	"flag"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ConfigManager manages EntityDB's three-tier configuration hierarchy system.
//
// The configuration system provides flexible, runtime-configurable settings
// through a priority-based hierarchy:
//   1. Database configuration entities (highest priority)
//   2. Command-line flags (medium priority)  
//   3. Environment variables (lowest priority)
//
// Database Configuration:
//   Configuration stored as entities with type:config tags enables runtime
//   configuration changes without server restarts. Values are cached for
//   performance and refreshed every 5 minutes or on explicit refresh.
//
// Flag Processing:
//   Command-line flags use long names (--entitydb-*) to avoid conflicts.
//   Only explicitly set flags override environment values.
//
// Performance:
//   Database configuration is cached with configurable TTL to minimize
//   repository queries while maintaining reasonable update responsiveness.
//
// Thread Safety:
//   All operations are protected by read-write mutexes for safe concurrent
//   access from multiple goroutines.
type ConfigManager struct {
	// mu protects all ConfigManager fields from concurrent access
	mu sync.RWMutex
	
	// config holds the current active configuration after applying all hierarchy tiers
	config *Config
	
	// entityRepo provides access to configuration entities in the database
	entityRepo models.EntityRepository
	
	// flagValues stores parsed command-line flag values for priority checking
	flagValues map[string]interface{}
	
	// dbCache holds cached database configuration values to reduce repository queries
	dbCache map[string]string
	
	// cacheExpiry tracks when the database configuration cache should be refreshed
	cacheExpiry time.Time
	
	// cacheDuration defines how long database configuration values are cached
	// Default: 5 minutes (good balance between performance and responsiveness)
	cacheDuration time.Duration
}

// NewConfigManager creates a new configuration manager instance.
//
// The manager is initialized with empty caches and a default cache duration
// of 5 minutes for database configuration values. The entity repository
// can be nil initially and set later when the storage layer is available.
//
// Parameters:
//   entityRepo - Repository for accessing configuration entities (can be nil)
//
// Returns:
//   A new ConfigManager ready for initialization and flag registration
//
// Cache Duration:
//   The 5-minute cache duration provides a good balance between:
//   - Performance: Reduces database queries for configuration reads
//   - Responsiveness: Configuration changes take effect within reasonable time
//   - Consistency: Prevents stale configuration in long-running operations
func NewConfigManager(entityRepo models.EntityRepository) *ConfigManager {
	return &ConfigManager{
		entityRepo:    entityRepo,
		flagValues:    make(map[string]interface{}),
		dbCache:       make(map[string]string),
		cacheDuration: 5 * time.Minute,
	}
}

// Initialize builds the final configuration by applying the three-tier hierarchy.
//
// This method must be called after RegisterFlags() and flag.Parse() to properly
// apply command-line flag overrides. The initialization process follows this sequence:
//
//   1. Load base configuration from environment variables
//   2. Apply command-line flag overrides (only explicitly set flags)
//   3. Apply database configuration overrides (highest priority)
//
// The method is thread-safe and can be called multiple times, though subsequent
// calls will rebuild the entire configuration from scratch.
//
// Database Configuration Handling:
//   If the entity repository is not available or database configuration loading
//   fails, the method logs a warning and continues with environment and flag values.
//   This ensures the server can start even if the database is temporarily unavailable.
//
// Returns:
//   - *Config: The final configuration with all hierarchy tiers applied
//   - error: Only returned for critical errors (database errors are logged as warnings)
//
// Thread Safety:
//   This method acquires a write lock for the duration of configuration building.
func (cm *ConfigManager) Initialize() (*Config, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Start with environment defaults
	cm.config = Load()

	// Override with command-line flags if provided
	cm.applyFlags()

	// Override with database configuration (highest priority)
	if err := cm.applyDatabaseConfig(); err != nil {
		logger.Warn("Failed to load database configuration: %v", err)
		// Continue with env and flag values
	}

	return cm.config, nil
}

// RegisterFlags registers all command-line flags with long names
func (cm *ConfigManager) RegisterFlags() {
	// Initialize config if not already done
	if cm.config == nil {
		cm.config = Load()
	}
	
	// Server Configuration - all long flags
	flag.IntVar(&cm.config.Port, "entitydb-port", cm.config.Port, 
		"HTTP server port (default from ENTITYDB_PORT or 8085)")
	flag.IntVar(&cm.config.SSLPort, "entitydb-ssl-port", cm.config.SSLPort,
		"HTTPS server port (default from ENTITYDB_SSL_PORT or 8085)")
	flag.BoolVar(&cm.config.UseSSL, "entitydb-use-ssl", cm.config.UseSSL,
		"Enable SSL/TLS (default from ENTITYDB_USE_SSL or false)")
	flag.StringVar(&cm.config.SSLCert, "entitydb-ssl-cert", cm.config.SSLCert,
		"SSL certificate file path")
	flag.StringVar(&cm.config.SSLKey, "entitydb-ssl-key", cm.config.SSLKey,
		"SSL private key file path")

	// Paths - all long flags
	flag.StringVar(&cm.config.DataPath, "entitydb-data-path", cm.config.DataPath,
		"Data directory path")
	flag.StringVar(&cm.config.StaticDir, "entitydb-static-dir", cm.config.StaticDir,
		"Static files directory")

	// Security - all long flags
	flag.StringVar(&cm.config.TokenSecret, "entitydb-token-secret", cm.config.TokenSecret,
		"Secret key for JWT tokens")
	flag.IntVar(&cm.config.SessionTTLHours, "entitydb-session-ttl-hours", cm.config.SessionTTLHours,
		"Session timeout in hours")

	// Logging - all long flags
	flag.StringVar(&cm.config.LogLevel, "entitydb-log-level", cm.config.LogLevel,
		"Log level (trace, debug, info, warn, error)")

	// Performance - all long flags
	flag.BoolVar(&cm.config.HighPerformance, "entitydb-high-performance", cm.config.HighPerformance,
		"Enable high-performance mode")

	// Timeouts - all long flags
	flag.DurationVar(&cm.config.HTTPReadTimeout, "entitydb-http-read-timeout", cm.config.HTTPReadTimeout,
		"HTTP read timeout")
	flag.DurationVar(&cm.config.HTTPWriteTimeout, "entitydb-http-write-timeout", cm.config.HTTPWriteTimeout,
		"HTTP write timeout")
	flag.DurationVar(&cm.config.HTTPIdleTimeout, "entitydb-http-idle-timeout", cm.config.HTTPIdleTimeout,
		"HTTP idle timeout")
	flag.DurationVar(&cm.config.ShutdownTimeout, "entitydb-shutdown-timeout", cm.config.ShutdownTimeout,
		"Server shutdown timeout")

	// Metrics - all long flags
	flag.DurationVar(&cm.config.MetricsInterval, "entitydb-metrics-interval", cm.config.MetricsInterval,
		"Metrics collection interval")
	flag.DurationVar(&cm.config.AggregationInterval, "entitydb-metrics-aggregation-interval", cm.config.AggregationInterval,
		"Metrics aggregation interval")

	// API - all long flags
	flag.StringVar(&cm.config.SwaggerHost, "entitydb-swagger-host", cm.config.SwaggerHost,
		"Swagger API documentation host")
	
	// Rate Limiting - all long flags
	flag.BoolVar(&cm.config.EnableRateLimit, "entitydb-enable-rate-limit", cm.config.EnableRateLimit,
		"Enable rate limiting")
	flag.IntVar(&cm.config.RateLimitRequests, "entitydb-rate-limit-requests", cm.config.RateLimitRequests,
		"Number of requests allowed per window")
	flag.IntVar(&cm.config.RateLimitWindowMinutes, "entitydb-rate-limit-window-minutes", cm.config.RateLimitWindowMinutes,
		"Rate limit window in minutes")

	// File and Path Configuration - all long flags
	flag.StringVar(&cm.config.WALSuffix, "entitydb-wal-suffix", cm.config.WALSuffix,
		"Write-Ahead Log file suffix")
	flag.StringVar(&cm.config.IndexSuffix, "entitydb-index-suffix", cm.config.IndexSuffix,
		"Index file suffix")
	flag.StringVar(&cm.config.BackupPath, "entitydb-backup-path", cm.config.BackupPath,
		"Backup directory path")
	flag.StringVar(&cm.config.TempPath, "entitydb-temp-path", cm.config.TempPath,
		"Temporary files directory")
	flag.StringVar(&cm.config.PIDFile, "entitydb-pid-file", cm.config.PIDFile,
		"Process ID file path")
	flag.StringVar(&cm.config.LogFile, "entitydb-log-file", cm.config.LogFile,
		"Server log file path")

	// Development and Debugging - all long flags
	flag.BoolVar(&cm.config.DevMode, "entitydb-dev-mode", cm.config.DevMode,
		"Enable development mode")
	flag.IntVar(&cm.config.DebugPort, "entitydb-debug-port", cm.config.DebugPort,
		"Debug/profiling port")
	flag.BoolVar(&cm.config.ProfileEnabled, "entitydb-profile-enabled", cm.config.ProfileEnabled,
		"Enable CPU and memory profiling")

	// Trace Configuration - all long flags
	flag.StringVar(&cm.config.TraceSubsystems, "entitydb-trace-subsystems", cm.config.TraceSubsystems,
		"Comma-separated list of trace subsystems to enable")

	// Default Admin User Configuration - all long flags
	flag.StringVar(&cm.config.DefaultAdminUsername, "entitydb-default-admin-username", cm.config.DefaultAdminUsername,
		"Default admin username")
	flag.StringVar(&cm.config.DefaultAdminPassword, "entitydb-default-admin-password", cm.config.DefaultAdminPassword,
		"Default admin password")
	flag.StringVar(&cm.config.DefaultAdminEmail, "entitydb-default-admin-email", cm.config.DefaultAdminEmail,
		"Default admin email address")
	
	// System User Configuration - all long flags
	flag.StringVar(&cm.config.SystemUserID, "entitydb-system-user-id", cm.config.SystemUserID,
		"System user UUID (immutable)")
	flag.StringVar(&cm.config.SystemUsername, "entitydb-system-username", cm.config.SystemUsername,
		"System username")
	
	// Advanced Security Configuration - all long flags
	flag.IntVar(&cm.config.BcryptCost, "entitydb-bcrypt-cost", cm.config.BcryptCost,
		"Bcrypt cost for password hashing (4-31)")

	// Essential short flags only
	flag.Bool("v", false, "Show version information")
	flag.Bool("version", false, "Show version information")
	
	flag.Bool("h", false, "Show help")
	flag.Bool("help", false, "Show help")

	// Store flag values for priority handling
	flag.VisitAll(func(f *flag.Flag) {
		cm.flagValues[f.Name] = f.Value
	})
}

// applyFlags applies command-line flag values if they were explicitly set
func (cm *ConfigManager) applyFlags() {
	flag.Visit(func(f *flag.Flag) {
		// Flag was explicitly set, so it overrides environment
		switch f.Name {
		case "entitydb-port":
			if v, err := strconv.Atoi(f.Value.String()); err == nil {
				cm.config.Port = v
			}
		case "entitydb-ssl-port":
			if v, err := strconv.Atoi(f.Value.String()); err == nil {
				cm.config.SSLPort = v
			}
		case "entitydb-use-ssl":
			cm.config.UseSSL = f.Value.String() == "true"
		case "entitydb-ssl-cert":
			cm.config.SSLCert = f.Value.String()
		case "entitydb-ssl-key":
			cm.config.SSLKey = f.Value.String()
		case "entitydb-data-path":
			cm.config.DataPath = f.Value.String()
		case "entitydb-static-dir":
			cm.config.StaticDir = f.Value.String()
		case "entitydb-token-secret":
			cm.config.TokenSecret = f.Value.String()
		case "entitydb-session-ttl-hours":
			if v, err := strconv.Atoi(f.Value.String()); err == nil {
				cm.config.SessionTTLHours = v
			}
		case "entitydb-log-level":
			cm.config.LogLevel = f.Value.String()
		case "entitydb-high-performance":
			cm.config.HighPerformance = f.Value.String() == "true"
		case "entitydb-enable-rate-limit":
			cm.config.EnableRateLimit = f.Value.String() == "true"
		case "entitydb-rate-limit-requests":
			if v, err := strconv.Atoi(f.Value.String()); err == nil {
				cm.config.RateLimitRequests = v
			}
		case "entitydb-rate-limit-window-minutes":
			if v, err := strconv.Atoi(f.Value.String()); err == nil {
				cm.config.RateLimitWindowMinutes = v
			}
		case "entitydb-swagger-host":
			cm.config.SwaggerHost = f.Value.String()
		
		// File and Path Configuration
		case "entitydb-wal-suffix":
			cm.config.WALSuffix = f.Value.String()
		case "entitydb-index-suffix":
			cm.config.IndexSuffix = f.Value.String()
		case "entitydb-backup-path":
			cm.config.BackupPath = f.Value.String()
		case "entitydb-temp-path":
			cm.config.TempPath = f.Value.String()
		case "entitydb-pid-file":
			cm.config.PIDFile = f.Value.String()
		case "entitydb-log-file":
			cm.config.LogFile = f.Value.String()
		
		// Development and Debugging
		case "entitydb-dev-mode":
			cm.config.DevMode = f.Value.String() == "true"
		case "entitydb-debug-port":
			if v, err := strconv.Atoi(f.Value.String()); err == nil {
				cm.config.DebugPort = v
			}
		case "entitydb-profile-enabled":
			cm.config.ProfileEnabled = f.Value.String() == "true"
		
		// Trace Configuration
		case "entitydb-trace-subsystems":
			cm.config.TraceSubsystems = f.Value.String()
		
		// Default Admin User Configuration
		case "entitydb-default-admin-username":
			cm.config.DefaultAdminUsername = f.Value.String()
		case "entitydb-default-admin-password":
			cm.config.DefaultAdminPassword = f.Value.String()
		case "entitydb-default-admin-email":
			cm.config.DefaultAdminEmail = f.Value.String()
		
		// System User Configuration
		case "entitydb-system-user-id":
			cm.config.SystemUserID = f.Value.String()
		case "entitydb-system-username":
			cm.config.SystemUsername = f.Value.String()
		
		// Advanced Security Configuration
		case "entitydb-bcrypt-cost":
			if v, err := strconv.Atoi(f.Value.String()); err == nil {
				cm.config.BcryptCost = v
			}
		}
	})
}

// applyDatabaseConfig loads configuration from database entities (highest priority)
func (cm *ConfigManager) applyDatabaseConfig() error {
	// Check cache first
	if time.Now().Before(cm.cacheExpiry) && len(cm.dbCache) > 0 {
		return cm.applyCachedConfig()
	}

	// Skip if repository is not available yet
	if cm.entityRepo == nil {
		return nil
	}

	// Query configuration entities
	configEntities, err := cm.entityRepo.ListByTag("type:config")
	if err != nil {
		return fmt.Errorf("failed to query config entities: %v", err)
	}

	// Clear and rebuild cache
	cm.dbCache = make(map[string]string)

	for _, entity := range configEntities {
		// Parse configuration from entity tags
		for _, tag := range entity.GetTagsWithoutTimestamp() {
			if strings.HasPrefix(tag, "conf:") {
				parts := strings.SplitN(tag, ":", 3)
				if len(parts) == 3 {
					key := parts[1] + "." + parts[2]
					
					// Get value from entity content
					var configData map[string]interface{}
					if err := json.Unmarshal(entity.Content, &configData); err == nil {
						if value, ok := configData[parts[2]]; ok {
							cm.dbCache[key] = fmt.Sprintf("%v", value)
						}
					}
				}
			}
		}
	}

	// Update cache expiry
	cm.cacheExpiry = time.Now().Add(cm.cacheDuration)

	return cm.applyCachedConfig()
}

// applyCachedConfig applies cached database configuration values
func (cm *ConfigManager) applyCachedConfig() error {
	// Apply database values (highest priority)
	if v, ok := cm.dbCache["server.port"]; ok {
		if port, err := strconv.Atoi(v); err == nil {
			cm.config.Port = port
		}
	}
	if v, ok := cm.dbCache["server.ssl_port"]; ok {
		if port, err := strconv.Atoi(v); err == nil {
			cm.config.SSLPort = port
		}
	}
	if v, ok := cm.dbCache["server.use_ssl"]; ok {
		cm.config.UseSSL = v == "true"
	}
	if v, ok := cm.dbCache["server.ssl_cert"]; ok {
		cm.config.SSLCert = v
	}
	if v, ok := cm.dbCache["server.ssl_key"]; ok {
		cm.config.SSLKey = v
	}
	if v, ok := cm.dbCache["paths.data"]; ok {
		cm.config.DataPath = v
	}
	if v, ok := cm.dbCache["paths.static"]; ok {
		cm.config.StaticDir = v
	}
	if v, ok := cm.dbCache["security.token_secret"]; ok {
		cm.config.TokenSecret = v
	}
	if v, ok := cm.dbCache["security.session_ttl_hours"]; ok {
		if hours, err := strconv.Atoi(v); err == nil {
			cm.config.SessionTTLHours = hours
		}
	}
	if v, ok := cm.dbCache["logging.level"]; ok {
		cm.config.LogLevel = v
	}
	if v, ok := cm.dbCache["performance.high_performance"]; ok {
		cm.config.HighPerformance = v == "true"
	}
	if v, ok := cm.dbCache["rate_limit.enabled"]; ok {
		cm.config.EnableRateLimit = v == "true"
	}
	if v, ok := cm.dbCache["rate_limit.requests"]; ok {
		if requests, err := strconv.Atoi(v); err == nil {
			cm.config.RateLimitRequests = requests
		}
	}
	if v, ok := cm.dbCache["rate_limit.window_minutes"]; ok {
		if minutes, err := strconv.Atoi(v); err == nil {
			cm.config.RateLimitWindowMinutes = minutes
		}
	}
	if v, ok := cm.dbCache["api.swagger_host"]; ok {
		cm.config.SwaggerHost = v
	}

	return nil
}

// RefreshConfig refreshes configuration from database
func (cm *ConfigManager) RefreshConfig() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Force cache expiry
	cm.cacheExpiry = time.Time{}
	
	return cm.applyDatabaseConfig()
}

// GetConfig returns the current configuration
func (cm *ConfigManager) GetConfig() *Config {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.config
}

// SetDatabaseConfig sets a configuration value in the database
func (cm *ConfigManager) SetDatabaseConfig(namespace, key, value string) error {
	if cm.entityRepo == nil {
		return fmt.Errorf("entity repository not available")
	}

	// Create configuration entity
	configData := map[string]interface{}{
		key: value,
	}
	
	content, err := json.Marshal(configData)
	if err != nil {
		return fmt.Errorf("failed to marshal config data: %v", err)
	}

	entity := &models.Entity{
		ID: fmt.Sprintf("config_%s_%s", namespace, key),
		Tags: []string{
			"type:config",
			fmt.Sprintf("conf:%s:%s", namespace, key),
		},
		Content: content,
	}

	// Create or update the configuration entity
	if err := cm.entityRepo.Create(entity); err != nil {
		// Try update if create fails
		if err := cm.entityRepo.Update(entity); err != nil {
			return fmt.Errorf("failed to save config: %v", err)
		}
	}

	// Refresh configuration cache
	return cm.RefreshConfig()
}

// GetDatabaseConfig retrieves a configuration value from the database
func (cm *ConfigManager) GetDatabaseConfig(namespace, key string) (string, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	cacheKey := namespace + "." + key
	if value, ok := cm.dbCache[cacheKey]; ok {
		return value, nil
	}

	return "", fmt.Errorf("configuration not found: %s.%s", namespace, key)
}