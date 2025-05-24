package binary

import (
	"entitydb/models"
	"entitydb/logger"
	"os"
)

// RepositoryFactory creates the appropriate repository based on configuration
type RepositoryFactory struct{}

// CreateRepository creates either a regular, high-performance, or temporal repository
func (f *RepositoryFactory) CreateRepository(dataPath string) (models.EntityRepository, error) {
	// Check environment variables
	disableHighPerf := os.Getenv("ENTITYDB_DISABLE_HIGH_PERFORMANCE") == "true"
	enableHighPerf := os.Getenv("ENTITYDB_HIGH_PERFORMANCE") == "true"
	enableTemporal := os.Getenv("ENTITYDB_TEMPORAL") != "false" // Temporal by default
	enableWALOnly := os.Getenv("ENTITYDB_WAL_ONLY") == "true" // New WAL-only mode
	
	// WAL-only mode for maximum write performance
	if enableWALOnly {
		logger.Info("Creating WALOnlyRepository for O(1) write performance")
		walRepo, err := NewWALOnlyRepository(dataPath)
		if err != nil {
			logger.Error("Failed to create WAL-only repository: %v, falling back to temporal", err)
			return NewTemporalRepository(dataPath)
		}
		return walRepo, nil
	}
	
	// Explicit disable overrides enable
	if disableHighPerf {
		logger.Info("Creating standard EntityRepository (high-performance disabled)")
		return NewEntityRepository(dataPath)
	}
	
	// If high performance mode is explicitly requested, use it regardless of temporal setting
	if enableHighPerf {
		logger.Info("Creating HighPerformanceRepository (explicitly enabled)")
		highPerfRepo, err := NewHighPerformanceRepository(dataPath)
		if err != nil {
			logger.Error("Failed to create high-performance repository: %v, falling back to standard", err)
			return NewEntityRepository(dataPath)
		}
		return highPerfRepo, nil
	}
	
	if enableTemporal {
		// Use temporal mode by default for maximum performance
		logger.Info("Creating TemporalRepository for maximum temporal performance")
		temporalRepo, err := NewTemporalRepository(dataPath)
		if err != nil {
			logger.Error("Failed to create temporal repository: %v, falling back to high-performance", err)
			return NewHighPerformanceRepository(dataPath)
		}
		return temporalRepo, nil
	}
	
	// Use regular high-performance mode
	logger.Info("Creating HighPerformanceRepository for maximum performance")
	highPerfRepo, err := NewHighPerformanceRepository(dataPath)
	if err != nil {
		logger.Error("Failed to create high-performance repository: %v, falling back to standard", err)
		return NewEntityRepository(dataPath)
	}
	
	return highPerfRepo, nil
}