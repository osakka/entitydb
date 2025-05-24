package binary

import (
	"entitydb/models"
	"entitydb/logger"
	"os"
	"time"
)

// RepositoryFactory creates the appropriate repository based on configuration
type RepositoryFactory struct{}

// CreateRepository creates either a regular, high-performance, or temporal repository
func (f *RepositoryFactory) CreateRepository(dataPath string) (models.EntityRepository, error) {
	var baseRepo models.EntityRepository
	var err error
	
	// Check environment variables
	disableHighPerf := os.Getenv("ENTITYDB_DISABLE_HIGH_PERFORMANCE") == "true"
	enableHighPerf := os.Getenv("ENTITYDB_HIGH_PERFORMANCE") == "true"
	enableTemporal := os.Getenv("ENTITYDB_TEMPORAL") != "false" // Temporal by default
	enableWALOnly := os.Getenv("ENTITYDB_WAL_ONLY") == "true" // New WAL-only mode
	enableCache := os.Getenv("ENTITYDB_ENABLE_CACHE") != "false" // Cache by default
	
	// Determine cache settings
	cacheTTL := 5 * time.Minute
	if ttlStr := os.Getenv("ENTITYDB_CACHE_TTL"); ttlStr != "" {
		if ttl, err := time.ParseDuration(ttlStr); err == nil {
			cacheTTL = ttl
		}
	}
	
	// Create the base repository based on configuration
	switch {
	case enableTemporal && enableHighPerf:
		logger.Info("Creating TemporalRepository with high-performance optimizations")
		baseRepo, err = NewTemporalRepository(dataPath)
		if err != nil {
			logger.Error("Failed to create temporal repository: %v", err)
			return nil, err
		}
		
	case enableWALOnly:
		logger.Info("Creating WALOnlyRepository for O(1) write performance")
		baseRepo, err = NewWALOnlyRepository(dataPath)
		if err != nil {
			logger.Error("Failed to create WAL-only repository: %v", err)
			return nil, err
		}
		
	case disableHighPerf:
		logger.Info("Creating standard EntityRepository (high-performance disabled)")
		baseRepo, err = NewEntityRepository(dataPath)
		if err != nil {
			return nil, err
		}
		
	case enableHighPerf:
		logger.Info("Creating HighPerformanceRepository (explicitly enabled)")
		baseRepo, err = NewHighPerformanceRepository(dataPath)
		if err != nil {
			logger.Error("Failed to create high-performance repository: %v", err)
			return nil, err
		}
		
	case enableTemporal:
		logger.Info("Creating TemporalRepository for maximum temporal performance")
		baseRepo, err = NewTemporalRepository(dataPath)
		if err != nil {
			logger.Error("Failed to create temporal repository: %v", err)
			return nil, err
		}
		
	default:
		logger.Info("Creating HighPerformanceRepository for maximum performance")
		baseRepo, err = NewHighPerformanceRepository(dataPath)
		if err != nil {
			logger.Error("Failed to create high-performance repository: %v", err)
			return nil, err
		}
	}
	
	// Wrap with caching if enabled
	if enableCache {
		logger.Info("Wrapping repository with CachedRepository (TTL: %v)", cacheTTL)
		return NewCachedRepository(baseRepo, cacheTTL), nil
	}
	
	return baseRepo, nil
}