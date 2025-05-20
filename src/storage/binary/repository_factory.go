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
	enableTemporal := os.Getenv("ENTITYDB_TEMPORAL") != "false" // Temporal by default
	
	if disableHighPerf {
		logger.Info("Creating standard EntityRepository (high-performance disabled)")
		return NewEntityRepository(dataPath)
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