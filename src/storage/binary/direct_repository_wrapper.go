package binary

import (
	"entitydb/logger"
	"entitydb/models"
	"strings"
)

// DirectRepositoryWrapper provides non-instrumented access to repository operations
// This prevents circular dependencies when metrics collection needs to store data
type DirectRepositoryWrapper struct {
	underlying     models.EntityRepository
	metricsBackend *MetricsBackend
}

// NewDirectRepositoryWrapper creates a new direct repository wrapper
func NewDirectRepositoryWrapper(repo models.EntityRepository) DirectRepository {
	// Get data path from repository if possible
	dataPath := "/opt/entitydb/var" // fallback default
	if entityRepo, ok := repo.(*EntityRepository); ok {
		dataPath = entityRepo.dataPath
	}
	
	return &DirectRepositoryWrapper{
		underlying:     repo,
		metricsBackend: NewMetricsBackend(dataPath),
	}
}

// CreateDirect creates an entity without triggering any metrics instrumentation
func (drw *DirectRepositoryWrapper) CreateDirect(entity *models.Entity) error {
	logger.Trace("DirectRepositoryWrapper operating without instrumentation: CreateDirect")
	
	// DECOMMISSIONED: All entities now use single source of truth (main database)
	// No more isolated metrics backend - everything goes to main repository
	return drw.underlying.Create(entity)
}

// GetByIDDirect gets an entity by ID without triggering any metrics instrumentation
func (drw *DirectRepositoryWrapper) GetByIDDirect(id string) (*models.Entity, error) {
	logger.Trace("DirectRepositoryWrapper operating without instrumentation: GetByIDDirect")
	
	// DECOMMISSIONED: All entities now use single source of truth (main database)
	// No more isolated metrics backend - everything goes to main repository
	return drw.underlying.GetByID(id)
}

// ListByTagDirect lists entities by tag without triggering any metrics instrumentation
func (drw *DirectRepositoryWrapper) ListByTagDirect(tag string) ([]*models.Entity, error) {
	logger.Trace("DirectRepositoryWrapper operating without instrumentation: ListByTagDirect")
	
	// DECOMMISSIONED: All entities now use single source of truth (main database)
	// No more isolated metrics backend - everything goes to main repository  
	return drw.underlying.ListByTag(tag)
}

// AddTagDirect adds a tag to an entity without triggering any metrics instrumentation
func (drw *DirectRepositoryWrapper) AddTagDirect(entityID, tag string) error {
	logger.Trace("DirectRepositoryWrapper operating without instrumentation: AddTagDirect")
	
	// DECOMMISSIONED: All entities now use single source of truth (main database)
	// No more isolated metrics backend - everything goes to main repository
	return drw.underlying.AddTag(entityID, tag)
}

// isMetricsEntity checks if an entity is metrics-related
func (drw *DirectRepositoryWrapper) isMetricsEntity(entity *models.Entity) bool {
	for _, tag := range entity.Tags {
		cleanTag := tag
		if pipePos := strings.Index(tag, "|"); pipePos != -1 {
			cleanTag = tag[pipePos+1:]
		}
		
		if strings.HasPrefix(cleanTag, "type:metric") {
			return true
		}
	}
	return false
}

// isMetricsEntityID checks if an entity ID is metrics-related
func (drw *DirectRepositoryWrapper) isMetricsEntityID(entityID string) bool {
	return strings.HasPrefix(entityID, "metric_") || strings.Contains(entityID, "metric")
}

// isMetricsTag checks if a tag is metrics-related
func (drw *DirectRepositoryWrapper) isMetricsTag(tag string) bool {
	cleanTag := tag
	if pipePos := strings.Index(tag, "|"); pipePos != -1 {
		cleanTag = tag[pipePos+1:]
	}
	
	return strings.HasPrefix(cleanTag, "name:") && 
		   (strings.Contains(cleanTag, "metric") || 
		    strings.Contains(cleanTag, "storage") ||
		    strings.Contains(cleanTag, "request") ||
		    strings.Contains(cleanTag, "performance"))
}