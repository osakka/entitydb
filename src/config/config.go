package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all configuration values for EntityDB
type Config struct {
	// Server Configuration
	Port             int
	SSLPort          int
	UseSSL           bool
	SSLCert          string
	SSLKey           string
	
	// Paths
	DataPath         string
	StaticDir        string
	
	// Security
	TokenSecret      string
	SessionTTLHours  int
	
	// Timeouts
	HTTPReadTimeout  time.Duration
	HTTPWriteTimeout time.Duration
	HTTPIdleTimeout  time.Duration
	ShutdownTimeout  time.Duration
	
	// Metrics
	MetricsInterval  time.Duration
	AggregationInterval time.Duration
	
	// API
	SwaggerHost      string
	
	// Logging
	LogLevel         string
	
	// Performance
	HighPerformance  bool
	
	// Rate Limiting
	EnableRateLimit  bool
	RateLimitRequests int
	RateLimitWindowMinutes int
	
	// Application Info
	AppName          string
	AppVersion       string
}

// Load loads configuration from environment variables with defaults
func Load() *Config {
	return &Config{
		// Server Configuration
		Port:             getEnvInt("ENTITYDB_PORT", 8085),
		SSLPort:          getEnvInt("ENTITYDB_SSL_PORT", 8085),
		UseSSL:           getEnvBool("ENTITYDB_USE_SSL", false),
		SSLCert:          getEnv("ENTITYDB_SSL_CERT", "/etc/ssl/certs/server.pem"),
		SSLKey:           getEnv("ENTITYDB_SSL_KEY", "/etc/ssl/private/server.key"),
		
		// Paths
		DataPath:         getEnv("ENTITYDB_DATA_PATH", "/opt/entitydb/var"),
		StaticDir:        getEnv("ENTITYDB_STATIC_DIR", "/opt/entitydb/share/htdocs"),
		
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
		AppVersion:       getEnv("ENTITYDB_APP_VERSION", "2.24.0"),
	}
}

// DatabasePath returns the full path to the database file
func (c *Config) DatabasePath() string {
	return c.DataPath + "/data/entities.db"
}

// Helper functions for environment variable parsing

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true" || value == "1"
	}
	return defaultValue
}

func getEnvDuration(key string, defaultSeconds int) time.Duration {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return time.Duration(intValue) * time.Second
		}
	}
	return time.Duration(defaultSeconds) * time.Second
}