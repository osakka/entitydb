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
// Values are populated from environment variables with defaults.
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
	
	// Enhanced Metrics Configuration
	MetricsRetentionRaw    time.Duration // Raw data retention period
	MetricsRetention1Min   time.Duration // 1-minute aggregates retention
	MetricsRetention1Hour  time.Duration // 1-hour aggregates retention
	MetricsRetention1Day   time.Duration // Daily aggregates retention
	MetricsHistogramBuckets []float64     // Histogram bucket boundaries
	MetricsEnableRequestTracking bool     // Enable HTTP request metrics
	MetricsEnableStorageTracking bool     // Enable storage operation metrics
	
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