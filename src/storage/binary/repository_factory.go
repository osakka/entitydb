package binary

import (
	"entitydb/models"
	"entitydb/config"
	"entitydb/logger"
	"os"
	"time"
)

// RepositoryFactory creates the appropriate repository based on configuration
type RepositoryFactory struct{}

// CreateRepository creates either a regular, high-performance, or temporal repository
func (f *RepositoryFactory) CreateRepository(cfg *config.Config) (models.EntityRepository, error) {
	var baseRepo models.EntityRepository
	var err error
	
	// Check environment variables
	disableHighPerf := os.Getenv("ENTITYDB_DISABLE_HIGH_PERFORMANCE") == "true"
	enableHighPerf := os.Getenv("ENTITYDB_HIGH_PERFORMANCE") == "true"
	enableTemporal := os.Getenv("ENTITYDB_TEMPORAL") != "false" // Temporal by default
	enableWALOnly := os.Getenv("ENTITYDB_WAL_ONLY") == "true" // New WAL-only mode
	enableCache := os.Getenv("ENTITYDB_ENABLE_CACHE") != "false" // Cache by default
	enableDataset := os.Getenv("ENTITYDB_ENABLE_DATASET") == "true" // Dataset isolation
	enableUnified := os.Getenv("ENTITYDB_ENABLE_UNIFIED") == "true" // Unified format
	
	// Determine cache settings
	cacheTTL := 5 * time.Minute
	if ttlStr := os.Getenv("ENTITYDB_CACHE_TTL"); ttlStr != "" {
		if ttl, err := time.ParseDuration(ttlStr); err == nil {
			cacheTTL = ttl
		}
	}
	
	// Create the EntityRepository with appropriate feature configuration
	// All variants now merged into unified EntityRepository
	switch {
	case enableUnified:
		logger.Info("Creating EntityRepository with unified file format")
		baseRepo, err = NewUnifiedRepositoryWithConfig(cfg)
		
	case enableDataset:
		logger.Info("Creating EntityRepository with dataset isolation features")
		baseRepo, err = NewDatasetRepositoryWithConfig(cfg)
		
	case enableTemporal && enableHighPerf:
		logger.Info("Creating EntityRepository with temporal and high-performance features")
		baseRepo, err = NewTemporalRepositoryWithConfig(cfg)
		
	case enableWALOnly:
		logger.Info("Creating EntityRepository with WAL-only optimization")
		baseRepo, err = NewWALOnlyRepositoryWithConfig(cfg)
		
	case disableHighPerf:
		logger.Info("Creating EntityRepository with standard features")
		baseRepo, err = NewEntityRepositoryWithConfig(cfg)
		
	case enableHighPerf:
		logger.Info("Creating EntityRepository with high-performance features")
		baseRepo, err = NewHighPerformanceRepositoryWithConfig(cfg)
		
	case enableTemporal:
		logger.Info("Creating EntityRepository with temporal features")
		baseRepo, err = NewTemporalRepositoryWithConfig(cfg)
		
	default:
		logger.Info("Creating EntityRepository with high-performance features (default)")
		baseRepo, err = NewHighPerformanceRepositoryWithConfig(cfg)
	}
	
	if err != nil {
		logger.Error("Failed to create repository: %v", err)
		return nil, err
	}
	
	// Wrap with caching if enabled
	if enableCache {
		logger.Info("Wrapping repository with CachedRepository (TTL: %v)", cacheTTL)
		return NewCachedRepository(baseRepo, cacheTTL), nil
	}
	
	return baseRepo, nil
}